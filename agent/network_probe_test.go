//go:build testing

package agent

import (
	"context"
	"errors"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/fxamacker/cbor/v2"
	"github.com/henrygd/beszel/internal/common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRunNetworkProbeTCPingSuccess(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	defer func() { _ = ln.Close() }()
	go func() {
		conn, err := ln.Accept()
		if err == nil {
			_ = conn.Close()
		}
	}()

	result := runNetworkProbe(context.Background(), common.NetworkProbeRequest{
		ProbeID:        "probe-1",
		Type:           common.NetworkProbeTCPing,
		Target:         ln.Addr().String(),
		TimeoutSeconds: 1,
	})

	assert.True(t, result.Success)
	assert.Equal(t, "probe-1", result.ProbeID)
	assert.Equal(t, common.NetworkProbeTCPing, result.Type)
	assert.GreaterOrEqual(t, result.LatencyMs, float64(0))
	assert.Empty(t, result.Error)
	assert.Empty(t, result.FailureCategory)
	assert.NotEmpty(t, result.CheckedAt)
}

func TestRunNetworkProbeTCPingInvalidTargetUsesSafeFailureCategory(t *testing.T) {
	result := runNetworkProbe(context.Background(), common.NetworkProbeRequest{
		ProbeID:        "probe-invalid",
		Type:           common.NetworkProbeTCPing,
		Target:         "example.com",
		TimeoutSeconds: 1,
	})

	assert.False(t, result.Success)
	assert.Equal(t, string(common.NetworkProbeFailureInvalidTarget), result.FailureCategory)
	assert.Contains(t, result.Error, "host:port")
}

func TestRunNetworkProbeHTTPGetRecordsStatus(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	result := runNetworkProbe(context.Background(), common.NetworkProbeRequest{
		ProbeID:        "probe-http",
		Type:           common.NetworkProbeHTTPGet,
		Target:         server.URL,
		TimeoutSeconds: 1,
	})

	require.True(t, result.Success)
	require.NotNil(t, result.HTTPStatus)
	assert.Equal(t, http.StatusNoContent, *result.HTTPStatus)
	assert.GreaterOrEqual(t, result.LatencyMs, float64(0))
	assert.Empty(t, result.FailureCategory)
}

func TestRunNetworkProbeICMPFailureDoesNotPanic(t *testing.T) {
	result := runNetworkProbe(context.Background(), common.NetworkProbeRequest{
		ProbeID:        "probe-icmp",
		Type:           common.NetworkProbeICMPPing,
		Target:         "192.0.2.1",
		TimeoutSeconds: 1,
	})

	assert.Equal(t, "probe-icmp", result.ProbeID)
	assert.Equal(t, common.NetworkProbeICMPPing, result.Type)
	assert.NotEmpty(t, result.CheckedAt)
	if !result.Success {
		assert.NotEmpty(t, result.Error)
		assert.NotEmpty(t, result.FailureCategory)
	}
}

func TestRunNetworkProbeUnsupportedTypeFails(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	result := runNetworkProbe(ctx, common.NetworkProbeRequest{
		ProbeID:        "probe-bad",
		Type:           common.NetworkProbeType("bad"),
		Target:         "example.com",
		TimeoutSeconds: 1,
	})

	assert.False(t, result.Success)
	assert.Contains(t, result.Error, "unsupported")
	assert.Equal(t, string(common.NetworkProbeFailureUnsupported), result.FailureCategory)
}

func TestNetworkProbeErrorMapsCommonFailures(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		category string
	}{
		{name: "timeout", err: context.DeadlineExceeded, category: string(common.NetworkProbeFailureTimeout)},
		{name: "dns", err: errors.New("lookup example.com: no such host"), category: string(common.NetworkProbeFailureDNSFailure)},
		{name: "refused", err: errors.New("connect: connection refused"), category: string(common.NetworkProbeFailureConnectionRefused)},
		{name: "invalid target", err: errors.New("invalid target: tcping target must use host:port format"), category: string(common.NetworkProbeFailureInvalidTarget)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, category := networkProbeError(tt.err)
			assert.Equal(t, tt.category, category)
		})
	}
}

func TestRunNetworkProbeHandlerSendsResult(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	defer func() { _ = ln.Close() }()
	go func() {
		conn, err := ln.Accept()
		if err == nil {
			_ = conn.Close()
		}
	}()

	payload, err := cbor.Marshal(common.NetworkProbeRequest{
		ProbeID:        "probe-handler",
		Type:           common.NetworkProbeTCPing,
		Target:         ln.Addr().String(),
		TimeoutSeconds: 1,
	})
	require.NoError(t, err)

	requestID := uint32(42)
	var sent any
	err = (&RunNetworkProbeHandler{}).Handle(&HandlerContext{
		Request: &common.HubRequest[cbor.RawMessage]{
			Action: common.RunNetworkProbe,
			Data:   payload,
		},
		RequestID: &requestID,
		SendResponse: func(data any, gotRequestID *uint32) error {
			require.Equal(t, &requestID, gotRequestID)
			sent = data
			return nil
		},
	})
	require.NoError(t, err)

	result, ok := sent.(common.NetworkProbeResult)
	require.True(t, ok)
	assert.True(t, result.Success)
	assert.Equal(t, "probe-handler", result.ProbeID)
	assert.Equal(t, common.NetworkProbeTCPing, result.Type)
}
