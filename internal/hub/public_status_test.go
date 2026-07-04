//go:build testing

package hub

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/henrygd/beszel/internal/entities/system"
	"github.com/pocketbase/pocketbase/core"
	pbTests "github.com/pocketbase/pocketbase/tests"
	"github.com/pocketbase/pocketbase/tools/router"
	"github.com/pocketbase/pocketbase/tools/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSanitizePublicSystemOmitsPrivateFields(t *testing.T) {
	record := core.NewRecord(&core.Collection{})
	record.Id = "system123"
	record.Set("name", "private-name")
	record.Set("host", "10.0.0.1")
	record.Set("port", "45876")
	record.Set("users", []string{"user123"})
	record.Set("status", "up")
	record.Set("info", system.Info{
		Cpu:     12.5,
		MemPct:  50.1,
		DiskPct: 70.2,
	})

	visibility := PublicSystemVisibility{
		PublicEnabled: true,
		PublicName:    "Public VPS",
		ShowCPU:       true,
		ShowMemory:    false,
		ShowDisk:      true,
	}

	summary := sanitizePublicSystem(record, visibility, nil, nil)
	payload, err := json.Marshal(summary)
	require.NoError(t, err)
	body := string(payload)

	for _, private := range []string{"10.0.0.1", "45876", "user123", "private-name", "host", "port", "users", "token"} {
		if strings.Contains(body, private) {
			t.Fatalf("public summary leaked %q in %s", private, body)
		}
	}
	assert.Equal(t, "system123", summary.ID)
	assert.Equal(t, "Public VPS", summary.Name)
	assert.Equal(t, "up", summary.Status)
	require.NotNil(t, summary.Metrics.CPUPercent)
	assert.Equal(t, 12.5, *summary.Metrics.CPUPercent)
	assert.Nil(t, summary.Metrics.MemoryPercent)
	require.NotNil(t, summary.Metrics.DiskPercent)
	assert.Equal(t, 70.2, *summary.Metrics.DiskPercent)
}

func TestValidatePublicVisibilityDefaultsAndName(t *testing.T) {
	visibility, err := normalizePublicVisibilityInput(PublicVisibilityInput{PublicEnabled: true})
	require.NoError(t, err)
	assert.True(t, visibility.ShowCPU)
	assert.True(t, visibility.ShowMemory)
	assert.True(t, visibility.ShowDisk)

	_, err = normalizePublicVisibilityInput(PublicVisibilityInput{
		PublicEnabled: true,
		PublicName:    string(make([]byte, publicNameMaxLength+1)),
	})
	require.Error(t, err)
}

func TestSanitizePublicSystemIncludesLatestMetrics(t *testing.T) {
	record := core.NewRecord(&core.Collection{})
	record.Id = "system123"
	record.Set("name", "node")
	record.Set("status", "up")
	record.Set("updated", time.Now().UTC())
	record.Set("info", system.Info{
		Cpu:     11.1,
		MemPct:  22.2,
		DiskPct: 33.3,
	})

	summary := sanitizePublicSystem(record, PublicSystemVisibility{ShowCPU: true, ShowMemory: true, ShowDisk: true}, nil, nil)
	require.NotNil(t, summary.Metrics.CPUPercent)
	require.NotNil(t, summary.Metrics.MemoryPercent)
	require.NotNil(t, summary.Metrics.DiskPercent)
	assert.NotEmpty(t, summary.Updated)
	assert.Equal(t, summary.Freshness, summary.Updated)
	assert.Empty(t, summary.Metrics.Unavailable)
}

func TestSanitizePublicSystemReportsUnavailableMetrics(t *testing.T) {
	record := core.NewRecord(&core.Collection{})
	record.Id = "system123"
	record.Set("name", "node")
	record.Set("status", "up")
	record.Set("updated", time.Now().UTC())

	summary := sanitizePublicSystem(record, PublicSystemVisibility{ShowCPU: true, ShowMemory: true, ShowDisk: true}, nil, nil)
	require.Nil(t, summary.Metrics.CPUPercent)
	require.Nil(t, summary.Metrics.MemoryPercent)
	require.Nil(t, summary.Metrics.DiskPercent)
	assert.Equal(t, []string{"cpu", "memory", "disk"}, summary.Metrics.Unavailable)
}

func TestSanitizePublicSystemAllowsZeroMetricValues(t *testing.T) {
	record := core.NewRecord(&core.Collection{})
	record.Id = "system123"
	record.Set("name", "node")
	record.Set("status", "up")
	record.Set("updated", time.Now().UTC())
	record.Set("info", system.Info{
		Cpu:     0,
		MemPct:  0,
		DiskPct: 0,
	})

	summary := sanitizePublicSystem(record, PublicSystemVisibility{ShowCPU: true, ShowMemory: true, ShowDisk: true}, nil, nil)
	require.NotNil(t, summary.Metrics.CPUPercent)
	require.NotNil(t, summary.Metrics.MemoryPercent)
	require.NotNil(t, summary.Metrics.DiskPercent)
	assert.Equal(t, 0.0, *summary.Metrics.CPUPercent)
	assert.Equal(t, 0.0, *summary.Metrics.MemoryPercent)
	assert.Equal(t, 0.0, *summary.Metrics.DiskPercent)
	assert.Empty(t, summary.Metrics.Unavailable)
}

func TestSanitizePublicSystemReadsJSONInfoString(t *testing.T) {
	record := core.NewRecord(&core.Collection{})
	record.Id = "system123"
	record.Set("name", "node")
	record.Set("status", "up")
	record.Set("updated", time.Now().UTC())
	record.Set("info", `{"cpu":8,"mp":45.37,"dp":34.5}`)

	summary := sanitizePublicSystem(record, PublicSystemVisibility{ShowCPU: true, ShowMemory: true, ShowDisk: true}, nil, nil)
	require.NotNil(t, summary.Metrics.CPUPercent)
	require.NotNil(t, summary.Metrics.MemoryPercent)
	require.NotNil(t, summary.Metrics.DiskPercent)
	assert.Equal(t, 8.0, *summary.Metrics.CPUPercent)
	assert.Equal(t, 45.37, *summary.Metrics.MemoryPercent)
	assert.Equal(t, 34.5, *summary.Metrics.DiskPercent)
	assert.Empty(t, summary.Metrics.Unavailable)
}

func TestPublicProbeFailurePointIncludesSafeReason(t *testing.T) {
	record := core.NewRecord(&core.Collection{})
	record.Set("success", false)
	record.Set("error", "context deadline exceeded")
	record.Set("failure_category", "")

	errorMessage := record.GetString("error")
	failureCategory := normalizeProbeFailureCategory(record.GetString("failure_category"), errorMessage)
	point := PublicProbeSeriesPoint{
		Success:         record.GetBool("success"),
		Error:           safeProbeResultError(errorMessage, failureCategory),
		FailureCategory: failureCategory,
	}

	assert.False(t, point.Success)
	assert.Equal(t, ProbeFailureTimeout, point.FailureCategory)
	assert.Equal(t, "context deadline exceeded", point.Error)
}

func TestPublicFreshnessReadsStoredStringTimestamp(t *testing.T) {
	record := core.NewRecord(&core.Collection{})
	record.Set("updated", "2026-06-29 13:52:42.217Z")

	assert.Equal(t, "2026-06-29T13:52:42Z", publicFreshnessFromRecord(record))
}

func TestPublicProbeSuccessPointDoesNotExposeFailureFields(t *testing.T) {
	record := core.NewRecord(&core.Collection{})
	record.Set("success", true)
	record.Set("latency_ms", 14.27)
	record.Set("error", "")
	record.Set("failure_category", ProbeFailureUnknown)

	errorMessage := record.GetString("error")
	failureCategory := ""
	if !record.GetBool("success") {
		failureCategory = normalizeProbeFailureCategory(record.GetString("failure_category"), errorMessage)
	}
	point := PublicProbeSeriesPoint{
		Success:         record.GetBool("success"),
		LatencyMs:       optionalFloat(record, "latency_ms"),
		FailureCategory: failureCategory,
	}
	if !point.Success {
		point.Error = safeProbeResultError(errorMessage, failureCategory)
	}

	assert.True(t, point.Success)
	require.NotNil(t, point.LatencyMs)
	assert.Equal(t, 14.27, *point.LatencyMs)
	assert.Empty(t, point.Error)
	assert.Empty(t, point.FailureCategory)
}

func TestPublicProbeSummaryDoesNotExposeTarget(t *testing.T) {
	summary := PublicProbeSummary{
		ID:   "probe123",
		Name: "广东电信",
		Type: NetworkProbeTypeTCPing,
		Series: []PublicProbeSeriesPoint{{
			Created: "2026-06-29T12:00:00Z",
			Success: true,
		}},
	}
	payload, err := json.Marshal(summary)
	require.NoError(t, err)
	body := string(payload)

	assert.NotContains(t, body, "target")
	assert.NotContains(t, body, "gd-ct-v4.ip.zstaticcdn.com")
}

func TestParsePublicChartRangeDefaultsAndValidates(t *testing.T) {
	t.Parallel()

	rangeSpec, err := parsePublicChartRange("")
	require.NoError(t, err)
	assert.Equal(t, "30m", rangeSpec.Name)
	assert.Equal(t, 30*time.Minute, rangeSpec.Duration)
	assert.Equal(t, "1m", rangeSpec.StatsType)

	for _, value := range []string{"30m", "1m", "1h", "12h", "24h", "1w", "30d"} {
		t.Run(value, func(t *testing.T) {
			rangeSpec, err := parsePublicChartRange(value)
			require.NoError(t, err)
			assert.Equal(t, value, rangeSpec.Name)
			assert.Positive(t, rangeSpec.Duration)
			assert.NotEmpty(t, rangeSpec.StatsType)
		})
	}

	_, err = parsePublicChartRange("2h")
	require.Error(t, err)
}

func TestGetPublicStatusRejectsInvalidExplicitRange(t *testing.T) {
	hub := newPublicStatusTestHub(t)
	req := httptest.NewRequest(http.MethodGet, "/api/beszel/public/status?range=2h", nil)
	recorder := httptest.NewRecorder()

	err := hub.getPublicStatus(&core.RequestEvent{
		App: hub,
		Event: router.Event{
			Request:  req,
			Response: recorder,
		},
	})
	require.Error(t, err)

	apiErr := router.ToApiError(err)
	assert.Equal(t, http.StatusBadRequest, apiErr.Status)
	assert.Equal(t, "Invalid public chart range.", apiErr.Message)
	assert.Equal(t, map[string]string{"range": "2h"}, apiErr.RawData())
}

func TestGetPublicStatusAcceptsExplicitRangesAndUsesRangeBucket(t *testing.T) {
	for rangeValue, statsType := range map[string]string{
		"1m":  "1m",
		"30m": "1m",
		"1h":  "1m",
		"12h": "10m",
		"24h": "20m",
		"1w":  "120m",
		"30d": "480m",
	} {
		t.Run(rangeValue, func(t *testing.T) {
			hub := newPublicStatusTestHub(t)
			user := createPublicStatusUser(t, hub)
			systemRecord := createPublicStatusRecord(t, hub, "systems", map[string]any{
				"name":   "private-node",
				"host":   "10.0.0.2",
				"status": "up",
				"users":  []string{user.Id},
			})
			createPublicStatusRecord(t, hub, CollectionPublicSystemVisibility, map[string]any{
				"system":         systemRecord.Id,
				"public_enabled": true,
				"public_name":    "public-node",
				"show_cpu":       true,
				"show_memory":    true,
				"show_disk":      true,
			})
			now := time.Now().UTC()
			createPublicSystemStats(t, hub, systemRecord.Id, now.Add(-30*time.Second), statsType, 42)
			otherStatsType := "10m"
			if statsType == otherStatsType {
				otherStatsType = "1m"
			}
			createPublicSystemStats(t, hub, systemRecord.Id, now.Add(-30*time.Second), otherStatsType, 99)

			response := requestPublicStatus(t, hub, rangeValue)
			require.Len(t, response.Systems, 1)
			require.Len(t, response.Systems[0].History, 1)
			require.NotNil(t, response.Systems[0].History[0].CPUPercent)
			assert.Equal(t, 42.0, *response.Systems[0].History[0].CPUPercent)
		})
	}
}

func TestPublicMetricHistoryUsesSelectedRangeAndAscendingPoints(t *testing.T) {
	hub := newPublicStatusTestHub(t)
	user := createPublicStatusUser(t, hub)

	systemRecord := createPublicStatusRecord(t, hub, "systems", map[string]any{
		"name":   "node",
		"host":   "127.0.0.1",
		"status": "up",
		"users":  []string{user.Id},
	})

	now := time.Now().UTC()
	createPublicSystemStats(t, hub, systemRecord.Id, now.Add(-45*time.Minute), "1m", 11)
	createPublicSystemStats(t, hub, systemRecord.Id, now.Add(-20*time.Minute), "1m", 22)
	createPublicSystemStats(t, hub, systemRecord.Id, now.Add(-5*time.Minute), "1m", 33)

	history, err := hub.publicMetricHistory(systemRecord.Id, PublicSystemVisibility{
		ShowCPU:    true,
		ShowMemory: true,
		ShowDisk:   true,
	}, publicChartRange{Name: "30m", Duration: 30 * time.Minute, StatsType: "1m"})
	require.NoError(t, err)
	require.Len(t, history, 2)

	assert.Equal(t, 22.0, *history[0].CPUPercent)
	assert.Equal(t, 33.0, *history[1].CPUPercent)
	assert.Less(t, history[0].Created, history[1].Created)
}

func TestPublicProbeSummariesFilterSeriesByRangeKeepLatestAndHideTarget(t *testing.T) {
	hub := newPublicStatusTestHub(t)
	user := createPublicStatusUser(t, hub)

	systemRecord := createPublicStatusRecord(t, hub, "systems", map[string]any{
		"name":   "node",
		"host":   "127.0.0.1",
		"status": "up",
		"users":  []string{user.Id},
	})
	probe := createPublicStatusRecord(t, hub, CollectionNetworkProbes, map[string]any{
		"name":             "Public line",
		"type":             NetworkProbeTypeTCPing,
		"scope":            NetworkProbeScopeFixed,
		"target":           "secret.example.com:443",
		"interval_seconds": 20,
		"timeout_seconds":  5,
		"enabled":          true,
		"public_visible":   true,
	})
	createPublicStatusRecord(t, hub, CollectionNetworkProbeAssignments, map[string]any{
		"probe":   probe.Id,
		"system":  systemRecord.Id,
		"enabled": true,
	})

	now := time.Now().UTC()
	createPublicProbeResult(t, hub, probe.Id, systemRecord.Id, now.Add(-45*time.Minute), true, 45, "")
	createPublicProbeResult(t, hub, probe.Id, systemRecord.Id, now.Add(-20*time.Minute), true, 20, "")
	createPublicProbeResult(t, hub, probe.Id, systemRecord.Id, now.Add(-5*time.Minute), false, 0, "dial tcp secret.example.com:443: i/o timeout")

	summaries, err := hub.publicProbeSummaries(systemRecord.Id, publicChartRange{Name: "30m", Duration: 30 * time.Minute, StatsType: "1m"})
	require.NoError(t, err)
	require.Len(t, summaries, 1)
	summary := summaries[0]

	require.NotNil(t, summary.Latest)
	assert.False(t, summary.Latest.Success)
	assert.Equal(t, "timeout", summary.Latest.FailureCategory)
	assert.Equal(t, "timeout", summary.Latest.Error)
	require.Len(t, summary.Series, 2)
	assert.True(t, summary.Series[0].Success)
	assert.Equal(t, 20.0, *summary.Series[0].LatencyMs)
	assert.False(t, summary.Series[1].Success)
	assert.Less(t, summary.Series[0].Created, summary.Series[1].Created)

	payload, err := json.Marshal(summary)
	require.NoError(t, err)
	assert.NotContains(t, string(payload), "secret.example.com")
	assert.NotContains(t, string(payload), "443")
}

func TestPublicProbeSummariesIncludeGlobalPublicProbeWithoutAssignment(t *testing.T) {
	hub := newPublicStatusTestHub(t)
	user := createPublicStatusUser(t, hub)

	systemRecord := createPublicStatusRecord(t, hub, "systems", map[string]any{
		"name":   "node",
		"host":   "127.0.0.1",
		"status": "up",
		"users":  []string{user.Id},
	})
	probe := createPublicStatusRecord(t, hub, CollectionNetworkProbes, map[string]any{
		"name":             "Global line",
		"type":             NetworkProbeTypeTCPing,
		"scope":            NetworkProbeScopeGlobal,
		"target":           "secret.example.com:443",
		"interval_seconds": 20,
		"timeout_seconds":  5,
		"enabled":          true,
		"public_visible":   true,
	})
	createPublicProbeResult(t, hub, probe.Id, systemRecord.Id, time.Now().UTC().Add(-30*time.Second), true, 8, "")

	summaries, err := hub.publicProbeSummaries(systemRecord.Id, publicChartRange{Name: "30m", Duration: 30 * time.Minute, StatsType: "1m"})
	require.NoError(t, err)
	require.Len(t, summaries, 1)
	assert.Equal(t, "Global line", summaries[0].Name)
	require.NotNil(t, summaries[0].Latest)
	require.NotNil(t, summaries[0].Latest.LatencyMs)
	assert.Equal(t, 8.0, *summaries[0].Latest.LatencyMs)
}

func TestPublicProbeSummariesExcludeHiddenGlobalProbe(t *testing.T) {
	hub := newPublicStatusTestHub(t)
	user := createPublicStatusUser(t, hub)

	systemRecord := createPublicStatusRecord(t, hub, "systems", map[string]any{
		"name":   "node",
		"host":   "127.0.0.1",
		"status": "up",
		"users":  []string{user.Id},
	})
	probe := createPublicStatusRecord(t, hub, CollectionNetworkProbes, map[string]any{
		"name":             "Hidden line",
		"type":             NetworkProbeTypeTCPing,
		"scope":            NetworkProbeScopeGlobal,
		"target":           "secret.example.com:443",
		"interval_seconds": 20,
		"timeout_seconds":  5,
		"enabled":          true,
		"public_visible":   false,
	})
	createPublicProbeResult(t, hub, probe.Id, systemRecord.Id, time.Now().UTC().Add(-30*time.Second), true, 8, "")

	summaries, err := hub.publicProbeSummaries(systemRecord.Id, publicChartRange{Name: "30m", Duration: 30 * time.Minute, StatsType: "1m"})
	require.NoError(t, err)
	assert.Empty(t, summaries)
}

func TestGetPublicStatusResponseHidesProbeTargetMetadata(t *testing.T) {
	hub := newPublicStatusTestHub(t)
	user := createPublicStatusUser(t, hub)
	systemRecord := createPublicStatusRecord(t, hub, "systems", map[string]any{
		"name":   "private-node",
		"host":   "10.0.0.2",
		"status": "up",
		"users":  []string{user.Id},
	})
	createPublicStatusRecord(t, hub, CollectionPublicSystemVisibility, map[string]any{
		"system":         systemRecord.Id,
		"public_enabled": true,
		"public_name":    "public-node",
		"show_cpu":       true,
		"show_memory":    true,
		"show_disk":      true,
	})
	probe := createPublicStatusRecord(t, hub, CollectionNetworkProbes, map[string]any{
		"name":             "Public line",
		"type":             NetworkProbeTypeTCPing,
		"scope":            NetworkProbeScopeFixed,
		"target":           "secret.example.com:443",
		"interval_seconds": 20,
		"timeout_seconds":  5,
		"enabled":          true,
		"public_visible":   true,
	})
	createPublicStatusRecord(t, hub, CollectionNetworkProbeAssignments, map[string]any{
		"probe":   probe.Id,
		"system":  systemRecord.Id,
		"enabled": true,
	})
	createPublicProbeResult(t, hub, probe.Id, systemRecord.Id, time.Now().UTC().Add(-30*time.Second), false, 0, "dial tcp secret.example.com:443: i/o timeout")

	req := httptest.NewRequest(http.MethodGet, "/api/beszel/public/status?range=30m", nil)
	recorder := httptest.NewRecorder()
	require.NoError(t, hub.getPublicStatus(&core.RequestEvent{
		App: hub,
		Event: router.Event{
			Request:  req,
			Response: recorder,
		},
	}))

	body := recorder.Body.String()
	assert.NotContains(t, body, "secret.example.com")
	assert.NotContains(t, body, "10.0.0.2")
	assert.NotContains(t, body, "443")
}

func newPublicStatusTestHub(t testing.TB) *Hub {
	t.Helper()
	app, err := pbTests.NewTestApp(t.TempDir())
	require.NoError(t, err)
	t.Cleanup(app.Cleanup)
	return NewHub(app)
}

func createPublicStatusUser(t testing.TB, app core.App) *core.Record {
	t.Helper()
	return createPublicStatusRecord(t, app, "users", map[string]any{
		"email":    "public-status@example.com",
		"password": "password123",
	})
}

func createPublicStatusRecord(t testing.TB, app core.App, collectionName string, fields map[string]any) *core.Record {
	t.Helper()
	collection, err := app.FindCachedCollectionByNameOrId(collectionName)
	require.NoError(t, err)
	record := core.NewRecord(collection)
	record.Load(fields)
	require.NoError(t, app.Save(record))
	return record
}

func requestPublicStatus(t testing.TB, hub *Hub, rangeValue string) PublicStatusResponse {
	t.Helper()
	req := httptest.NewRequest(http.MethodGet, "/api/beszel/public/status?range="+rangeValue, nil)
	recorder := httptest.NewRecorder()
	require.NoError(t, hub.getPublicStatus(&core.RequestEvent{
		App: hub,
		Event: router.Event{
			Request:  req,
			Response: recorder,
		},
	}))
	require.Equal(t, http.StatusOK, recorder.Code)
	var response PublicStatusResponse
	require.NoError(t, json.Unmarshal(recorder.Body.Bytes(), &response))
	return response
}

func createPublicSystemStats(t testing.TB, app core.App, systemID string, created time.Time, statsType string, cpu float64) {
	t.Helper()
	record := createPublicStatusRecord(t, app, "system_stats", map[string]any{
		"system": systemID,
		"type":   statsType,
		"stats": system.Stats{
			Cpu:     cpu,
			MemPct:  cpu + 1,
			DiskPct: cpu + 2,
		},
	})
	record.SetRaw("created", created.Format(types.DefaultDateLayout))
	require.NoError(t, app.SaveNoValidate(record))
}

func createPublicProbeResult(t testing.TB, app core.App, probeID string, systemID string, created time.Time, success bool, latency float64, errorMessage string) {
	t.Helper()
	record := createPublicStatusRecord(t, app, CollectionNetworkProbeResults, map[string]any{
		"probe":            probeID,
		"system":           systemID,
		"type":             NetworkProbeTypeTCPing,
		"success":          success,
		"latency_ms":       latency,
		"error":            errorMessage,
		"failure_category": "",
	})
	record.SetRaw("created", created.Format(types.DefaultDateLayout))
	require.NoError(t, app.SaveNoValidate(record))
}
