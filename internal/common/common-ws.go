package common

import (
	"github.com/fxamacker/cbor/v2"
	"github.com/henrygd/beszel/internal/entities/smart"
	"github.com/henrygd/beszel/internal/entities/system"
	"github.com/henrygd/beszel/internal/entities/systemd"
)

type WebSocketAction = uint8

const (
	// Request system data from agent
	GetData WebSocketAction = iota
	// Check the fingerprint of the agent
	CheckFingerprint
	// Request container logs from agent
	GetContainerLogs
	// Request container info from agent
	GetContainerInfo
	// Request SMART data from agent
	GetSmartData
	// Request detailed systemd service info from agent
	GetSystemdInfo
	// Request network probe execution from agent
	RunNetworkProbe
	// Add new actions here...
)

// HubRequest defines the structure for requests sent from hub to agent.
type HubRequest[T any] struct {
	Action WebSocketAction `cbor:"0,keyasint"`
	Data   T               `cbor:"1,keyasint,omitempty,omitzero"`
	Id     *uint32         `cbor:"2,keyasint,omitempty"`
}

// AgentResponse defines the structure for responses sent from agent to hub.
type AgentResponse struct {
	Id          *uint32                    `cbor:"0,keyasint,omitempty"`
	SystemData  *system.CombinedData       `cbor:"1,keyasint,omitempty,omitzero"` // Legacy (<= 0.17)
	Fingerprint *FingerprintResponse       `cbor:"2,keyasint,omitempty,omitzero"` // Legacy (<= 0.17)
	Error       string                     `cbor:"3,keyasint,omitempty,omitzero"`
	String      *string                    `cbor:"4,keyasint,omitempty,omitzero"` // Legacy (<= 0.17)
	SmartData   map[string]smart.SmartData `cbor:"5,keyasint,omitempty,omitzero"` // Legacy (<= 0.17)
	ServiceInfo systemd.ServiceDetails     `cbor:"6,keyasint,omitempty,omitzero"` // Legacy (<= 0.17)
	// Data is the generic response payload for new endpoints (0.18+)
	Data cbor.RawMessage `cbor:"7,keyasint,omitempty,omitzero"`
}

type FingerprintRequest struct {
	Signature   []byte `cbor:"0,keyasint"`
	NeedSysInfo bool   `cbor:"1,keyasint"` // For universal token system creation
}

type FingerprintResponse struct {
	Fingerprint string `cbor:"0,keyasint"`
	// Optional system info for universal token system creation
	Hostname string `cbor:"1,keyasint,omitzero"`
	Port     string `cbor:"2,keyasint,omitzero"`
	Name     string `cbor:"3,keyasint,omitzero"`
}

type DataRequestOptions struct {
	CacheTimeMs    uint16 `cbor:"0,keyasint"`
	IncludeDetails bool   `cbor:"1,keyasint"`
}

type ContainerLogsRequest struct {
	ContainerID string `cbor:"0,keyasint"`
}

type ContainerInfoRequest struct {
	ContainerID string `cbor:"0,keyasint"`
}

type SystemdInfoRequest struct {
	ServiceName string `cbor:"0,keyasint"`
}

type NetworkProbeType string

const (
	NetworkProbeTCPing   NetworkProbeType = "tcping"
	NetworkProbeICMPPing NetworkProbeType = "icmp_ping"
	NetworkProbeHTTPGet  NetworkProbeType = "http_get"
)

type NetworkProbeFailureCategory string

const (
	NetworkProbeFailureInvalidTarget            NetworkProbeFailureCategory = "invalid_target"
	NetworkProbeFailureDNSFailure               NetworkProbeFailureCategory = "dns_failure"
	NetworkProbeFailureTimeout                  NetworkProbeFailureCategory = "timeout"
	NetworkProbeFailureConnectionRefused        NetworkProbeFailureCategory = "connection_refused"
	NetworkProbeFailureTargetUnreachable        NetworkProbeFailureCategory = "target_unreachable"
	NetworkProbeFailureExecutionNodeUnavailable NetworkProbeFailureCategory = "execution_node_unavailable"
	NetworkProbeFailureUnsupported              NetworkProbeFailureCategory = "unsupported"
	NetworkProbeFailureUnknown                  NetworkProbeFailureCategory = "unknown_failure"
)

type NetworkProbeRequest struct {
	ProbeID        string           `json:"probeId" cbor:"0,keyasint"`
	Type           NetworkProbeType `json:"type" cbor:"1,keyasint"`
	Target         string           `json:"target" cbor:"2,keyasint"`
	TimeoutSeconds uint16           `json:"timeoutSeconds" cbor:"3,keyasint"`
}

type NetworkProbeResult struct {
	ProbeID           string           `json:"probeId" cbor:"0,keyasint"`
	Type              NetworkProbeType `json:"type" cbor:"1,keyasint"`
	Target            string           `json:"target" cbor:"2,keyasint"`
	Success           bool             `json:"success" cbor:"3,keyasint"`
	LatencyMs         float64          `json:"latencyMs,omitempty" cbor:"4,keyasint,omitempty,omitzero"`
	PacketLossPercent *float64         `json:"packetLossPercent,omitempty" cbor:"5,keyasint,omitempty,omitzero"`
	HTTPStatus        *int             `json:"httpStatus,omitempty" cbor:"6,keyasint,omitempty,omitzero"`
	Error             string           `json:"error,omitempty" cbor:"7,keyasint,omitempty,omitzero"`
	FailureCategory   string           `json:"failureCategory,omitempty" cbor:"8,keyasint,omitempty,omitzero"`
	CheckedAt         string           `json:"checkedAt" cbor:"9,keyasint"`
}
