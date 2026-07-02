package ws

import (
	"context"
	"errors"

	"github.com/fxamacker/cbor/v2"
	"github.com/henrygd/beszel/internal/common"
	"github.com/lxzan/gws"
	"golang.org/x/crypto/ssh"
)

// ResponseHandler defines interface for handling agent responses.
// This is used by handleAgentRequest for legacy response handling.
type ResponseHandler interface {
	Handle(agentResponse common.AgentResponse) error
	HandleLegacy(rawData []byte) error
}

// BaseHandler provides a default implementation that can be embedded to make HandleLegacy optional
type BaseHandler struct{}

func (h *BaseHandler) HandleLegacy(rawData []byte) error {
	return errors.New("legacy format not supported")
}

////////////////////////////////////////////////////////////////////////////
// Fingerprint handling (used for WebSocket authentication)
////////////////////////////////////////////////////////////////////////////

// fingerprintHandler implements ResponseHandler for fingerprint requests
type fingerprintHandler struct {
	result *common.FingerprintResponse
}

func (h *fingerprintHandler) HandleLegacy(rawData []byte) error {
	return cbor.Unmarshal(rawData, h.result)
}

func (h *fingerprintHandler) Handle(agentResponse common.AgentResponse) error {
	if agentResponse.Fingerprint != nil {
		*h.result = *agentResponse.Fingerprint
		return nil
	}
	return errors.New("no fingerprint data in response")
}

type networkProbeHandler struct {
	BaseHandler
	result *common.NetworkProbeResult
}

func (h *networkProbeHandler) Handle(agentResponse common.AgentResponse) error {
	if len(agentResponse.Data) == 0 {
		return errors.New("no network probe data in response")
	}
	return cbor.Unmarshal(agentResponse.Data, h.result)
}

// GetFingerprint authenticates with the agent using SSH signature and returns the agent's fingerprint.
func (ws *WsConn) GetFingerprint(ctx context.Context, token string, signer ssh.Signer, needSysInfo bool) (common.FingerprintResponse, error) {
	if !ws.IsConnected() {
		return common.FingerprintResponse{}, gws.ErrConnClosed
	}

	challenge := []byte(token)
	signature, err := signer.Sign(nil, challenge)
	if err != nil {
		return common.FingerprintResponse{}, err
	}

	req, err := ws.requestManager.SendRequest(ctx, common.CheckFingerprint, common.FingerprintRequest{
		Signature:   signature.Blob,
		NeedSysInfo: needSysInfo,
	})
	if err != nil {
		return common.FingerprintResponse{}, err
	}

	var result common.FingerprintResponse
	handler := &fingerprintHandler{result: &result}
	err = ws.handleAgentRequest(req, handler)
	return result, err
}

// RunNetworkProbe requests one network probe execution from a connected agent.
func (ws *WsConn) RunNetworkProbe(ctx context.Context, probe common.NetworkProbeRequest) (common.NetworkProbeResult, error) {
	if !ws.IsConnected() {
		return common.NetworkProbeResult{}, gws.ErrConnClosed
	}
	req, err := ws.requestManager.SendRequest(ctx, common.RunNetworkProbe, probe)
	if err != nil {
		return common.NetworkProbeResult{}, err
	}
	var result common.NetworkProbeResult
	handler := &networkProbeHandler{result: &result}
	err = ws.handleAgentRequest(req, handler)
	return result, err
}
