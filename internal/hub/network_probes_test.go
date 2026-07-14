//go:build testing

package hub

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/henrygd/beszel/internal/common"
	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/tests"
	"github.com/pocketbase/pocketbase/tools/router"
	"github.com/pocketbase/pocketbase/tools/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidateNetworkProbeConfig(t *testing.T) {
	tests := []struct {
		name    string
		cfg     networkProbeConfig
		wantErr bool
		wantMsg string
	}{
		{
			name: "valid tcping",
			cfg: networkProbeConfig{
				Name:            "tcp",
				Type:            common.NetworkProbeTCPing,
				Target:          "example.com:443",
				IntervalSeconds: 60,
				TimeoutSeconds:  5,
			},
		},
		{
			name: "valid icmp",
			cfg: networkProbeConfig{
				Name:            "icmp",
				Type:            common.NetworkProbeICMPPing,
				Target:          "1.1.1.1",
				IntervalSeconds: 60,
				TimeoutSeconds:  5,
			},
		},
		{
			name: "valid http",
			cfg: networkProbeConfig{
				Name:            "http",
				Type:            common.NetworkProbeHTTPGet,
				Target:          "https://example.com",
				IntervalSeconds: 60,
				TimeoutSeconds:  5,
			},
		},
		{
			name: "reject fixed invalid tcp target",
			cfg: networkProbeConfig{
				Name:            "bad",
				Type:            common.NetworkProbeTCPing,
				Target:          "example.com",
				IntervalSeconds: 60,
				TimeoutSeconds:  5,
			},
			wantErr: true,
			wantMsg: "tcping target must use host:port format",
		},
		{
			name: "reject tcp target with invalid port",
			cfg: networkProbeConfig{
				Name:            "bad",
				Type:            common.NetworkProbeTCPing,
				Target:          "example.com:0",
				IntervalSeconds: 60,
				TimeoutSeconds:  5,
			},
			wantErr: true,
			wantMsg: "tcping target must include a valid port from 1 to 65535",
		},
		{
			name: "reject timeout above interval",
			cfg: networkProbeConfig{
				Name:            "bad",
				Type:            common.NetworkProbeHTTPGet,
				Target:          "https://example.com",
				IntervalSeconds: 5,
				TimeoutSeconds:  5,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateNetworkProbeConfig(tt.cfg)
			if tt.wantErr {
				require.Error(t, err)
				if tt.wantMsg != "" {
					assert.EqualError(t, err, tt.wantMsg)
				}
				return
			}
			require.NoError(t, err)
		})
	}
}

func TestNormalizeProbeFailureCategory(t *testing.T) {
	assert.Equal(t, ProbeFailureInvalidTarget, normalizeProbeFailureCategory("", "invalid target: expected host:port"))
	assert.Equal(t, ProbeFailureDNSFailure, normalizeProbeFailureCategory("", "no such host"))
	assert.Equal(t, ProbeFailureTimeout, normalizeProbeFailureCategory("", context.DeadlineExceeded.Error()))
	assert.Equal(t, ProbeFailureConnectionRefused, normalizeProbeFailureCategory("", "connection refused"))
	assert.Equal(t, ProbeFailureTargetUnreachable, normalizeProbeFailureCategory("", "network is unreachable"))
	assert.Equal(t, ProbeFailureExecutionNodeUnavailable, normalizeProbeFailureCategory("", "agent offline or unsupported"))
	assert.Equal(t, ProbeFailureUnsupported, normalizeProbeFailureCategory("", "unsupported probe type"))
	assert.Equal(t, ProbeFailureUnknown, normalizeProbeFailureCategory("", "some other error"))
}

func TestNetworkProbeResultPointIncludesFailureCategory(t *testing.T) {
	record := core.NewRecord(&core.Collection{})
	record.Id = "result123"
	record.Set("system", "system1")
	record.Set("success", false)
	record.Set("error", "connection refused")
	record.Set("failure_category", "")

	point := networkProbeResultPoint(record)
	assert.Equal(t, ProbeFailureConnectionRefused, point.FailureCategory)
	assert.Equal(t, "connection refused", point.Error)
}

func TestFailedNetworkProbeResultCarriesCategory(t *testing.T) {
	result := failedNetworkProbeResult(common.NetworkProbeRequest{
		ProbeID: "probe1",
		Type:    common.NetworkProbeTCPing,
		Target:  "example.com:443",
	}, "agent offline or unsupported")

	assert.Equal(t, ProbeFailureExecutionNodeUnavailable, result.FailureCategory)
	assert.False(t, result.Success)
	assert.NotEmpty(t, result.Error)
}

func TestNormalizePersistedProbeFailureCategory(t *testing.T) {
	assert.Equal(t, ProbeFailureInvalidTarget, normalizePersistedProbeFailureCategory("INVALID_TARGET", ""))
	assert.Equal(t, ProbeFailureTimeout, normalizePersistedProbeFailureCategory("", "deadline exceeded"))
	assert.Equal(t, ProbeFailureUnknown, normalizePersistedProbeFailureCategory("", ""))
}

func TestSafeProbeResultError(t *testing.T) {
	assert.Equal(t, "connection refused", safeProbeResultError("connection refused", ProbeFailureConnectionRefused))
	assert.Equal(t, "timeout", safeProbeResultError("a/very/long/error/message", ProbeFailureTimeout))
}

func TestNetworkProbeResponseDoesNotPanicWithEmptyAssignments(t *testing.T) {
	record := core.NewRecord(&core.Collection{})
	record.Id = "probe1"
	record.Set("name", "probe")
	record.Set("type", string(common.NetworkProbeTCPing))
	record.Set("target", "example.com:443")
	record.Set("interval_seconds", 60)
	record.Set("timeout_seconds", 5)
	record.Set("enabled", true)
	record.Set("public_visible", true)
	record.Set("scope", NetworkProbeScopeGlobal)

	resp := networkProbeResponse(record, nil)
	require.Empty(t, resp.Systems)
	encoded, err := json.Marshal(resp)
	require.NoError(t, err)
	assert.Contains(t, string(encoded), `"systems":[]`)
	assert.Equal(t, "probe1", resp.ID)
	assert.Equal(t, NetworkProbeScopeGlobal, resp.Scope)
}

func TestNormalizeNetworkProbeScopeDefaultsFromSystems(t *testing.T) {
	assert.Equal(t, NetworkProbeScopeGlobal, normalizeNetworkProbeScope("", nil))
	assert.Equal(t, NetworkProbeScopeGlobal, normalizeNetworkProbeScope("", []string{}))
	assert.Equal(t, NetworkProbeScopeFixed, normalizeNetworkProbeScope("", []string{"system1"}))
	assert.Equal(t, NetworkProbeScopeGlobal, normalizeNetworkProbeScope("GLOBAL", []string{"system1"}))
	assert.Equal(t, NetworkProbeScopeFixed, normalizeNetworkProbeScope("fixed", nil))
}

func TestNetworkProbeAssignmentDueHonorsProbeInterval(t *testing.T) {
	probe := core.NewRecord(&core.Collection{})
	probe.Set("enabled", true)
	probe.Set("interval_seconds", 20)
	now := time.Date(2026, 7, 1, 12, 0, 0, 0, time.UTC)

	require.True(t, networkProbeAssignmentDue(probe, nil, now))

	recent := core.NewRecord(&core.Collection{})
	recent.SetRaw("created", now.Add(-5*time.Second).Format(types.DefaultDateLayout))
	require.False(t, networkProbeAssignmentDue(probe, recent, now))

	stale := core.NewRecord(&core.Collection{})
	stale.SetRaw("created", now.Add(-25*time.Second).Format(types.DefaultDateLayout))
	require.True(t, networkProbeAssignmentDue(probe, stale, now))
}

func TestLiveNetworkProbeRequestBoundsTimeoutWithoutMutation(t *testing.T) {
	probe := core.NewRecord(&core.Collection{})
	probe.Id = "probe-live"
	probe.Set("type", string(common.NetworkProbeTCPing))
	probe.Set("target", "example.com:443")
	probe.Set("timeout_seconds", 5)

	req := liveNetworkProbeRequest(probe)
	assert.Equal(t, "probe-live", req.ProbeID)
	assert.Equal(t, uint16(liveProbeTimeoutSeconds), req.TimeoutSeconds)
	assert.Equal(t, 5, probe.GetInt("timeout_seconds"))
}

func TestIsLiveLatencyProbeType(t *testing.T) {
	assert.True(t, isLiveLatencyProbeType(common.NetworkProbeTCPing))
	assert.True(t, isLiveLatencyProbeType(common.NetworkProbeICMPPing))
	assert.False(t, isLiveLatencyProbeType(common.NetworkProbeHTTPGet))
}

func TestLiveNetworkProbeAssignmentsFilterEnabledLatencyLines(t *testing.T) {
	hub := newNetworkProbeTestHub(t)
	user := createNetworkProbeTestRecord(t, hub, "users", map[string]any{
		"email":    "latency-user@example.com",
		"password": "password123",
	})
	systemRecord := createNetworkProbeTestRecord(t, hub, "systems", map[string]any{
		"name":   "node",
		"host":   "127.0.0.1",
		"status": "up",
		"users":  []string{user.Id},
	})
	tcpProbe := createNetworkProbeTestRecord(t, hub, CollectionNetworkProbes, map[string]any{
		"name":             "tcp",
		"type":             NetworkProbeTypeTCPing,
		"scope":            NetworkProbeScopeFixed,
		"target":           "example.com:443",
		"interval_seconds": 20,
		"timeout_seconds":  5,
		"enabled":          true,
		"public_visible":   true,
	})
	httpProbe := createNetworkProbeTestRecord(t, hub, CollectionNetworkProbes, map[string]any{
		"name":             "http",
		"type":             NetworkProbeTypeHTTPGet,
		"scope":            NetworkProbeScopeFixed,
		"target":           "https://example.com",
		"interval_seconds": 20,
		"timeout_seconds":  5,
		"enabled":          true,
		"public_visible":   true,
	})
	disabledProbe := createNetworkProbeTestRecord(t, hub, CollectionNetworkProbes, map[string]any{
		"name":             "icmp",
		"type":             NetworkProbeTypeICMPPing,
		"scope":            NetworkProbeScopeFixed,
		"target":           "1.1.1.1",
		"interval_seconds": 20,
		"timeout_seconds":  5,
		"enabled":          false,
		"public_visible":   true,
	})
	createNetworkProbeTestRecord(t, hub, CollectionNetworkProbeAssignments, map[string]any{
		"probe":   tcpProbe.Id,
		"system":  systemRecord.Id,
		"enabled": true,
	})
	createNetworkProbeTestRecord(t, hub, CollectionNetworkProbeAssignments, map[string]any{
		"probe":   httpProbe.Id,
		"system":  systemRecord.Id,
		"enabled": true,
	})
	createNetworkProbeTestRecord(t, hub, CollectionNetworkProbeAssignments, map[string]any{
		"probe":   disabledProbe.Id,
		"system":  systemRecord.Id,
		"enabled": true,
	})

	assignments, err := hub.liveNetworkProbeAssignments(systemRecord.Id)
	require.NoError(t, err)
	require.Len(t, assignments, 1)
	assert.Equal(t, tcpProbe.Id, assignments[0].ProbeID)
}

func TestLiveNetworkProbeAssignmentsIncludeGlobalProbeWithoutAssignment(t *testing.T) {
	hub := newNetworkProbeTestHub(t)
	user := createNetworkProbeTestRecord(t, hub, "users", map[string]any{
		"email":    "global-live@example.com",
		"password": "password123",
	})
	systemRecord := createNetworkProbeTestRecord(t, hub, "systems", map[string]any{
		"name":   "node",
		"host":   "127.0.0.1",
		"status": "up",
		"users":  []string{user.Id},
	})
	globalProbe := createNetworkProbeTestRecord(t, hub, CollectionNetworkProbes, map[string]any{
		"name":             "global tcp",
		"type":             NetworkProbeTypeTCPing,
		"scope":            NetworkProbeScopeGlobal,
		"target":           "example.com:443",
		"interval_seconds": 20,
		"timeout_seconds":  5,
		"enabled":          true,
		"public_visible":   true,
	})

	assignments, err := hub.liveNetworkProbeAssignments(systemRecord.Id)
	require.NoError(t, err)
	require.Len(t, assignments, 1)
	assert.Equal(t, globalProbe.Id, assignments[0].ProbeID)
	assert.Equal(t, systemRecord.Id, assignments[0].SystemID)
}

func TestEffectiveNetworkProbeAssignmentsKeepFixedProbeScoped(t *testing.T) {
	hub := newNetworkProbeTestHub(t)
	user := createNetworkProbeTestRecord(t, hub, "users", map[string]any{
		"email":    "fixed-scope@example.com",
		"password": "password123",
	})
	systemOne := createNetworkProbeTestRecord(t, hub, "systems", map[string]any{
		"name":   "node-1",
		"host":   "127.0.0.1",
		"status": "up",
		"users":  []string{user.Id},
	})
	systemTwo := createNetworkProbeTestRecord(t, hub, "systems", map[string]any{
		"name":   "node-2",
		"host":   "127.0.0.2",
		"status": "up",
		"users":  []string{user.Id},
	})
	fixedProbe := createNetworkProbeTestRecord(t, hub, CollectionNetworkProbes, map[string]any{
		"name":             "fixed",
		"type":             NetworkProbeTypeTCPing,
		"scope":            NetworkProbeScopeFixed,
		"target":           "example.com:443",
		"interval_seconds": 20,
		"timeout_seconds":  5,
		"enabled":          true,
		"public_visible":   true,
	})
	globalProbe := createNetworkProbeTestRecord(t, hub, CollectionNetworkProbes, map[string]any{
		"name":             "global",
		"type":             NetworkProbeTypeTCPing,
		"scope":            NetworkProbeScopeGlobal,
		"target":           "example.org:443",
		"interval_seconds": 20,
		"timeout_seconds":  5,
		"enabled":          true,
		"public_visible":   true,
	})
	createNetworkProbeTestRecord(t, hub, CollectionNetworkProbeAssignments, map[string]any{
		"probe":   fixedProbe.Id,
		"system":  systemOne.Id,
		"enabled": true,
	})

	assignments, err := hub.effectiveNetworkProbeAssignments(systemTwo.Id)
	require.NoError(t, err)
	require.Len(t, assignments, 1)
	assert.Equal(t, globalProbe.Id, assignments[0].ProbeID)
	assert.Equal(t, systemTwo.Id, assignments[0].SystemID)
}

func TestReplaceProbeAssignmentsHonorsScopeTransitions(t *testing.T) {
	hub := newNetworkProbeTestHub(t)
	user := createNetworkProbeTestRecord(t, hub, "users", map[string]any{
		"email":    "scope-transition@example.com",
		"password": "password123",
	})
	systemOne := createNetworkProbeTestRecord(t, hub, "systems", map[string]any{
		"name":   "node-1",
		"host":   "127.0.0.1",
		"status": "up",
		"users":  []string{user.Id},
	})
	systemTwo := createNetworkProbeTestRecord(t, hub, "systems", map[string]any{
		"name":   "node-2",
		"host":   "127.0.0.2",
		"status": "up",
		"users":  []string{user.Id},
	})
	probe := createNetworkProbeTestRecord(t, hub, CollectionNetworkProbes, map[string]any{
		"name":             "line",
		"type":             NetworkProbeTypeTCPing,
		"scope":            NetworkProbeScopeFixed,
		"target":           "example.com:443",
		"interval_seconds": 20,
		"timeout_seconds":  5,
		"enabled":          true,
		"public_visible":   true,
	})

	assignments, err := hub.replaceProbeAssignments(probe.Id, NetworkProbeScopeFixed, []string{systemOne.Id, systemTwo.Id})
	require.NoError(t, err)
	require.Len(t, assignments, 2)

	assignments, err = hub.replaceProbeAssignments(probe.Id, NetworkProbeScopeGlobal, nil)
	require.NoError(t, err)
	assert.Empty(t, assignments)

	remaining, err := hub.FindRecordsByFilter(CollectionNetworkProbeAssignments, "probe = {:probe}", "", -1, 0, dbx.Params{"probe": probe.Id})
	require.NoError(t, err)
	assert.Empty(t, remaining)
}

func TestRunLiveNetworkProbeAssignmentPersistsOfflineFailure(t *testing.T) {
	hub := newNetworkProbeTestHub(t)
	user := createNetworkProbeTestRecord(t, hub, "users", map[string]any{
		"email":    "offline-live@example.com",
		"password": "password123",
	})
	systemRecord := createNetworkProbeTestRecord(t, hub, "systems", map[string]any{
		"name":   "node",
		"host":   "127.0.0.1",
		"status": "up",
		"users":  []string{user.Id},
	})
	probe := createNetworkProbeTestRecord(t, hub, CollectionNetworkProbes, map[string]any{
		"name":             "tcp",
		"type":             NetworkProbeTypeTCPing,
		"scope":            NetworkProbeScopeFixed,
		"target":           "example.com:443",
		"interval_seconds": 20,
		"timeout_seconds":  5,
		"enabled":          true,
		"public_visible":   true,
	})
	assignment := createNetworkProbeTestRecord(t, hub, CollectionNetworkProbeAssignments, map[string]any{
		"probe":   probe.Id,
		"system":  systemRecord.Id,
		"enabled": true,
	})

	require.NoError(t, hub.runLiveNetworkProbeAssignment(context.Background(), networkProbeAssignment{
		ID:       assignment.Id,
		ProbeID:  probe.Id,
		SystemID: systemRecord.Id,
		Enabled:  true,
	}))

	result, err := hub.latestNetworkProbeResult(probe.Id, systemRecord.Id)
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.False(t, result.GetBool("success"))
	assert.Equal(t, ProbeFailureExecutionNodeUnavailable, result.GetString("failure_category"))
}

func TestGetNetworkProbeResultsUsesRangeAndReturnsNewestAscending(t *testing.T) {
	hub := newNetworkProbeTestHub(t)
	user := createNetworkProbeTestRecord(t, hub, "users", map[string]any{
		"email":    "network-probe@example.com",
		"password": "password123",
	})
	probe := createNetworkProbeTestRecord(t, hub, CollectionNetworkProbes, map[string]any{
		"name":             "line",
		"type":             NetworkProbeTypeTCPing,
		"scope":            NetworkProbeScopeFixed,
		"target":           "example.com:443",
		"interval_seconds": 20,
		"timeout_seconds":  5,
		"enabled":          true,
		"public_visible":   true,
	})
	systemRecord := createNetworkProbeTestRecord(t, hub, "systems", map[string]any{
		"name":   "node",
		"host":   "127.0.0.1",
		"status": "up",
		"users":  []string{user.Id},
	})

	now := time.Now().UTC()
	createNetworkProbeResult(t, hub, probe.Id, systemRecord.Id, now.Add(-45*time.Minute), 45)
	createNetworkProbeResult(t, hub, probe.Id, systemRecord.Id, now.Add(-20*time.Minute), 20)
	createNetworkProbeResult(t, hub, probe.Id, systemRecord.Id, now.Add(-5*time.Minute), 5)

	response := requestNetworkProbeResults(t, hub, probe.Id, systemRecord.Id, "30m")
	require.Len(t, response.Series, 2)
	require.NotNil(t, response.Series[0].LatencyMs)
	require.NotNil(t, response.Series[1].LatencyMs)
	assert.Equal(t, 20.0, *response.Series[0].LatencyMs)
	assert.Equal(t, 5.0, *response.Series[1].LatencyMs)
	assert.Less(t, response.Series[0].Created, response.Series[1].Created)
}

func TestGetNetworkProbeResultsUseCompatibleBucketsForLongRanges(t *testing.T) {
	hub := newNetworkProbeTestHub(t)
	user := createNetworkProbeTestRecord(t, hub, "users", map[string]any{
		"email":    "network-probe-bucket@example.com",
		"password": "password123",
	})
	probe := createNetworkProbeTestRecord(t, hub, CollectionNetworkProbes, map[string]any{
		"name":             "line",
		"type":             NetworkProbeTypeTCPing,
		"scope":            NetworkProbeScopeFixed,
		"target":           "example.com:443",
		"interval_seconds": 20,
		"timeout_seconds":  5,
		"enabled":          true,
		"public_visible":   true,
	})
	systemRecord := createNetworkProbeTestRecord(t, hub, "systems", map[string]any{
		"name":   "node",
		"host":   "127.0.0.1",
		"status": "up",
		"users":  []string{user.Id},
	})

	now := time.Now().UTC().Truncate(120 * time.Minute)
	createNetworkProbeResultWithBucket(t, hub, probe.Id, systemRecord.Id, now.Add(-6*time.Hour+30*time.Minute), "120m", 120)
	createNetworkProbeResultWithBucket(t, hub, probe.Id, systemRecord.Id, now.Add(-6*time.Hour+50*time.Minute), "20m", 25)
	createNetworkProbeResultWithBucket(t, hub, probe.Id, systemRecord.Id, now.Add(-4*time.Hour+20*time.Minute), "20m", 40)
	createNetworkProbeLegacyResult(t, hub, probe.Id, systemRecord.Id, now.Add(-2*time.Hour+20*time.Minute), 60)

	response := requestNetworkProbeResults(t, hub, probe.Id, systemRecord.Id, "1w")
	require.Len(t, response.Series, 3)
	require.NotNil(t, response.Series[0].LatencyMs)
	require.NotNil(t, response.Series[1].LatencyMs)
	require.NotNil(t, response.Series[2].LatencyMs)
	assert.Equal(t, 120.0, *response.Series[0].LatencyMs)
	assert.Equal(t, 40.0, *response.Series[1].LatencyMs)
	assert.Equal(t, 60.0, *response.Series[2].LatencyMs)
	assert.Equal(t, "120m", response.Series[0].RetentionBucket)
	assert.Equal(t, "20m", response.Series[1].RetentionBucket)
	assert.Empty(t, response.Series[2].RetentionBucket)
}

func TestGetNetworkProbeResultsWithoutSystemKeepsCompatibleRecordsPerSystem(t *testing.T) {
	hub := newNetworkProbeTestHub(t)
	user := createNetworkProbeTestRecord(t, hub, "users", map[string]any{
		"email":    "network-probe-multi-system@example.com",
		"password": "password123",
	})
	probe := createNetworkProbeTestRecord(t, hub, CollectionNetworkProbes, map[string]any{
		"name":             "line",
		"type":             NetworkProbeTypeTCPing,
		"scope":            NetworkProbeScopeGlobal,
		"target":           "example.com:443",
		"interval_seconds": 20,
		"timeout_seconds":  5,
		"enabled":          true,
		"public_visible":   true,
	})
	systemA := createNetworkProbeTestRecord(t, hub, "systems", map[string]any{
		"name":   "node-a",
		"host":   "127.0.0.1",
		"status": "up",
		"users":  []string{user.Id},
	})
	systemB := createNetworkProbeTestRecord(t, hub, "systems", map[string]any{
		"name":   "node-b",
		"host":   "127.0.0.2",
		"status": "up",
		"users":  []string{user.Id},
	})

	now := time.Now().UTC().Truncate(120 * time.Minute)
	createNetworkProbeResultWithBucket(t, hub, probe.Id, systemA.Id, now.Add(-4*time.Hour+30*time.Minute), "120m", 120)
	createNetworkProbeResultWithBucket(t, hub, probe.Id, systemB.Id, now.Add(-4*time.Hour+45*time.Minute), "20m", 45)

	request := httptest.NewRequest(http.MethodGet, "/api/beszel/network-probes/"+probe.Id+"/results?range=1w", nil)
	request.SetPathValue("probeId", probe.Id)
	recorder := httptest.NewRecorder()
	event := &core.RequestEvent{
		App:  hub,
		Auth: user,
		Event: router.Event{
			Request:  request,
			Response: recorder,
		},
	}

	require.NoError(t, hub.getNetworkProbeResults(event))

	var response NetworkProbeResultsResponse
	require.NoError(t, json.Unmarshal(recorder.Body.Bytes(), &response))
	require.Len(t, response.Series, 2)
	assert.ElementsMatch(t, []string{systemA.Id, systemB.Id}, []string{response.Series[0].SystemID, response.Series[1].SystemID})
}

func TestGetNetworkProbeResultsCollapseCompatibleFallbackDensityPerWindow(t *testing.T) {
	hub := newNetworkProbeTestHub(t)
	user := createNetworkProbeTestRecord(t, hub, "users", map[string]any{
		"email":    "network-probe-density@example.com",
		"password": "password123",
	})
	probe := createNetworkProbeTestRecord(t, hub, CollectionNetworkProbes, map[string]any{
		"name":             "line",
		"type":             NetworkProbeTypeTCPing,
		"scope":            NetworkProbeScopeFixed,
		"target":           "example.com:443",
		"interval_seconds": 20,
		"timeout_seconds":  5,
		"enabled":          true,
		"public_visible":   true,
	})
	systemRecord := createNetworkProbeTestRecord(t, hub, "systems", map[string]any{
		"name":   "node",
		"host":   "127.0.0.1",
		"status": "up",
		"users":  []string{user.Id},
	})

	now := time.Now().UTC().Truncate(120 * time.Minute)
	createNetworkProbeLegacyResult(t, hub, probe.Id, systemRecord.Id, now.Add(-2*time.Hour+10*time.Minute), 20)
	createNetworkProbeLegacyResult(t, hub, probe.Id, systemRecord.Id, now.Add(-2*time.Hour+30*time.Minute), 30)
	createNetworkProbeLegacyResult(t, hub, probe.Id, systemRecord.Id, now.Add(-2*time.Hour+50*time.Minute), 50)

	response := requestNetworkProbeResults(t, hub, probe.Id, systemRecord.Id, "1w")
	require.Len(t, response.Series, 1)
	require.NotNil(t, response.Series[0].LatencyMs)
	assert.Equal(t, 50.0, *response.Series[0].LatencyMs)
}

func TestGetNetworkProbeResultsRejectsInvalidRange(t *testing.T) {
	hub := newNetworkProbeTestHub(t)
	probe := createNetworkProbeTestRecord(t, hub, CollectionNetworkProbes, map[string]any{
		"name":             "line",
		"type":             NetworkProbeTypeTCPing,
		"scope":            NetworkProbeScopeFixed,
		"target":           "example.com:443",
		"interval_seconds": 20,
		"timeout_seconds":  5,
		"enabled":          true,
		"public_visible":   true,
	})
	req := httptest.NewRequest(http.MethodGet, "/api/beszel/network-probes/"+probe.Id+"/results?range=2h", nil)
	req.SetPathValue("probeId", probe.Id)
	recorder := httptest.NewRecorder()

	err := hub.getNetworkProbeResults(&core.RequestEvent{
		App: hub,
		Event: router.Event{
			Request:  req,
			Response: recorder,
		},
	})
	require.Error(t, err)

	apiErr := router.ToApiError(err)
	assert.Equal(t, http.StatusBadRequest, apiErr.Status)
	assert.Equal(t, "Invalid network probe chart range.", apiErr.Message)
}

func newNetworkProbeTestHub(t testing.TB) *Hub {
	t.Helper()
	app, err := tests.NewTestApp(t.TempDir())
	require.NoError(t, err)
	t.Cleanup(app.Cleanup)
	return NewHub(app)
}

func createNetworkProbeTestRecord(t testing.TB, app core.App, collectionName string, fields map[string]any) *core.Record {
	t.Helper()
	collection, err := app.FindCachedCollectionByNameOrId(collectionName)
	require.NoError(t, err)
	record := core.NewRecord(collection)
	record.Load(fields)
	require.NoError(t, app.Save(record))
	return record
}

func createNetworkProbeResult(t testing.TB, app core.App, probeID string, systemID string, created time.Time, latency float64) {
	t.Helper()
	createNetworkProbeResultWithBucket(t, app, probeID, systemID, created, "1m", latency)
}

func createNetworkProbeResultWithBucket(t testing.TB, app core.App, probeID string, systemID string, created time.Time, bucket string, latency float64) {
	t.Helper()
	record := createNetworkProbeTestRecord(t, app, CollectionNetworkProbeResults, map[string]any{
		"probe":            probeID,
		"system":           systemID,
		"type":             NetworkProbeTypeTCPing,
		"success":          true,
		"latency_ms":       latency,
		"failure_category": "",
		"bucket":           bucket,
	})
	record.SetRaw("created", created.Format(types.DefaultDateLayout))
	require.NoError(t, app.SaveNoValidate(record))
}

func createNetworkProbeLegacyResult(t testing.TB, app core.App, probeID string, systemID string, created time.Time, latency float64) {
	t.Helper()
	record := createNetworkProbeTestRecord(t, app, CollectionNetworkProbeResults, map[string]any{
		"probe":            probeID,
		"system":           systemID,
		"type":             NetworkProbeTypeTCPing,
		"success":          true,
		"latency_ms":       latency,
		"failure_category": "",
	})
	record.SetRaw("created", created.Format(types.DefaultDateLayout))
	require.NoError(t, app.SaveNoValidate(record))
}

func requestNetworkProbeResults(t testing.TB, hub *Hub, probeID string, systemID string, rangeValue string) NetworkProbeResultsResponse {
	t.Helper()
	req := httptest.NewRequest(http.MethodGet, "/api/beszel/network-probes/"+probeID+"/results?system="+systemID+"&range="+rangeValue, nil)
	req.SetPathValue("probeId", probeID)
	recorder := httptest.NewRecorder()
	require.NoError(t, hub.getNetworkProbeResults(&core.RequestEvent{
		App: hub,
		Event: router.Event{
			Request:  req,
			Response: recorder,
		},
	}))
	require.Equal(t, http.StatusOK, recorder.Code)
	var response NetworkProbeResultsResponse
	require.NoError(t, json.Unmarshal(recorder.Body.Bytes(), &response))
	return response
}
