package agent

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/henrygd/beszel/internal/common"
	"golang.org/x/net/icmp"
	"golang.org/x/net/ipv4"
)

const defaultNetworkProbeTimeout = 5 * time.Second

func runNetworkProbe(ctx context.Context, req common.NetworkProbeRequest) common.NetworkProbeResult {
	result := common.NetworkProbeResult{
		ProbeID:   req.ProbeID,
		Type:      req.Type,
		Target:    req.Target,
		CheckedAt: time.Now().UTC().Format(time.RFC3339),
	}

	timeout := networkProbeTimeout(req.TimeoutSeconds)
	probeCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	var latency time.Duration
	var err error

	switch req.Type {
	case common.NetworkProbeTCPing:
		latency, err = runTCPing(probeCtx, req.Target)
	case common.NetworkProbeHTTPGet:
		var status int
		latency, status, err = runHTTPGet(probeCtx, req.Target, timeout)
		if status != 0 {
			result.HTTPStatus = &status
		}
	case common.NetworkProbeICMPPing:
		var loss float64
		latency, loss, err = runICMPPing(probeCtx, req.Target, timeout)
		result.PacketLossPercent = &loss
	default:
		err = fmt.Errorf("unsupported probe type: %s", req.Type)
	}

	if err != nil {
		result.Error, result.FailureCategory = networkProbeError(err)
		return result
	}

	result.Success = true
	result.LatencyMs = float64(latency) / float64(time.Millisecond)
	return result
}

func networkProbeTimeout(timeoutSeconds uint16) time.Duration {
	if timeoutSeconds == 0 {
		return defaultNetworkProbeTimeout
	}
	return time.Duration(timeoutSeconds) * time.Second
}

func runTCPing(ctx context.Context, target string) (time.Duration, error) {
	host, port, err := splitTCPingTarget(target)
	if err != nil {
		return 0, errors.New("invalid target: tcping target must use host:port format")
	}
	if strings.TrimSpace(host) == "" || strings.TrimSpace(port) == "" {
		return 0, errors.New("invalid target: tcping target must use host:port format")
	}

	dialer := net.Dialer{}
	start := time.Now()
	conn, err := dialer.DialContext(ctx, "tcp", net.JoinHostPort(host, port))
	if err != nil {
		return 0, err
	}
	latency := time.Since(start)
	_ = conn.Close()
	return latency, nil
}

func splitTCPingTarget(target string) (string, string, error) {
	target = strings.TrimSpace(target)
	host, port, err := net.SplitHostPort(target)
	if err == nil {
		return host, port, nil
	}
	if strings.Count(target, ":") == 1 && !strings.Contains(target, "://") {
		parts := strings.SplitN(target, ":", 2)
		return parts[0], parts[1], nil
	}
	return "", "", err
}

func runHTTPGet(ctx context.Context, target string, timeout time.Duration) (time.Duration, int, error) {
	if !strings.HasPrefix(target, "http://") && !strings.HasPrefix(target, "https://") {
		return 0, 0, errors.New("invalid target: expected http or https URL")
	}

	client := http.Client{Timeout: timeout}
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, target, nil)
	if err != nil {
		return 0, 0, fmt.Errorf("invalid target: %w", err)
	}

	start := time.Now()
	response, err := client.Do(request)
	if err != nil {
		return 0, 0, err
	}
	defer func() {
		_ = response.Body.Close()
	}()
	_, _ = io.Copy(io.Discard, response.Body)

	return time.Since(start), response.StatusCode, nil
}

func runICMPPing(ctx context.Context, target string, timeout time.Duration) (time.Duration, float64, error) {
	ip, err := resolveIPv4(ctx, target)
	if err != nil {
		return 0, 100, err
	}

	conn, err := icmp.ListenPacket("ip4:icmp", "0.0.0.0")
	if err != nil {
		return 0, 100, err
	}
	defer func() {
		_ = conn.Close()
	}()

	deadline := time.Now().Add(timeout)
	if ctxDeadline, ok := ctx.Deadline(); ok && ctxDeadline.Before(deadline) {
		deadline = ctxDeadline
	}
	_ = conn.SetDeadline(deadline)

	message := icmp.Message{
		Type: ipv4.ICMPTypeEcho,
		Code: 0,
		Body: &icmp.Echo{
			ID:   os.Getpid() & 0xffff,
			Seq:  1,
			Data: []byte("beszel-network-probe"),
		},
	}
	bytes, err := message.Marshal(nil)
	if err != nil {
		return 0, 100, err
	}

	start := time.Now()
	if _, err = conn.WriteTo(bytes, &net.IPAddr{IP: ip}); err != nil {
		return 0, 100, err
	}

	buffer := make([]byte, 1500)
	for {
		n, _, err := conn.ReadFrom(buffer)
		if err != nil {
			return 0, 100, err
		}
		reply, err := icmp.ParseMessage(ipv4.ICMPTypeEchoReply.Protocol(), buffer[:n])
		if err != nil {
			continue
		}
		echo, ok := reply.Body.(*icmp.Echo)
		if reply.Type == ipv4.ICMPTypeEchoReply && ok && echo.ID == os.Getpid()&0xffff && echo.Seq == 1 {
			return time.Since(start), 0, nil
		}
	}
}

func resolveIPv4(ctx context.Context, target string) (net.IP, error) {
	ips, err := net.DefaultResolver.LookupIPAddr(ctx, target)
	if err != nil {
		return nil, err
	}
	for _, ipAddr := range ips {
		if ip := ipAddr.IP.To4(); ip != nil {
			return ip, nil
		}
	}
	return nil, errors.New("icmp requires an IPv4 target")
}

func networkProbeError(err error) (string, string) {
	if err == nil {
		return "", ""
	}
	if errors.Is(err, context.DeadlineExceeded) || errors.Is(err, os.ErrDeadlineExceeded) {
		return "timeout", string(common.NetworkProbeFailureTimeout)
	}
	var netErr net.Error
	if errors.As(err, &netErr) && netErr.Timeout() {
		return "timeout", string(common.NetworkProbeFailureTimeout)
	}
	if errors.Is(err, os.ErrPermission) {
		return "permission denied", string(common.NetworkProbeFailureUnknown)
	}

	message := err.Error()
	switch {
	case strings.Contains(message, "unsupported probe type"):
		return message, string(common.NetworkProbeFailureUnsupported)
	case strings.Contains(message, "permission denied"), strings.Contains(message, "operation not permitted"):
		return "permission denied", string(common.NetworkProbeFailureUnknown)
	case strings.Contains(message, "no such host"):
		return "dns failure", string(common.NetworkProbeFailureDNSFailure)
	case strings.Contains(message, "connection refused"):
		return "connection refused", string(common.NetworkProbeFailureConnectionRefused)
	case strings.Contains(message, "network is unreachable"), strings.Contains(message, "no route to host"), strings.Contains(message, "host is unreachable"):
		return "target unreachable", string(common.NetworkProbeFailureTargetUnreachable)
	case strings.Contains(message, "invalid target"), strings.Contains(message, "missing port"), strings.Contains(message, "host:port"):
		return message, string(common.NetworkProbeFailureInvalidTarget)
	default:
		return "probe failed", string(common.NetworkProbeFailureUnknown)
	}
}
