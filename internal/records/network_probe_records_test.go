//go:build testing

package records_test

import (
	"testing"
	"time"

	"github.com/henrygd/beszel/internal/records"
	"github.com/henrygd/beszel/internal/tests"
	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/tools/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateLongerNetworkProbeRecords(t *testing.T) {
	hub, err := tests.NewTestHub(t.TempDir())
	require.NoError(t, err)
	defer hub.Cleanup()

	probe, systemRecord := createNetworkProbeFixture(t, hub)
	now := waitForStableProbeRollupWindow()
	windowStart := completedProbeRollupWindowSampleStart(now, 10*time.Minute)

	for i := range 10 {
		createNetworkProbeResultRecord(t, hub, probe.Id, systemRecord.Id, windowStart.Add(time.Duration(i)*100*time.Millisecond), "1m", float64(10+i))
	}

	rm := records.NewRecordManager(hub)
	rm.CreateLongerNetworkProbeRecords()

	results, err := hub.FindRecordsByFilter(
		"network_probe_results",
		"probe = {:probe} && system = {:system} && bucket = {:bucket}",
		"created",
		-1,
		0,
		dbx.Params{
			"probe":  probe.Id,
			"system": systemRecord.Id,
			"bucket": "10m",
		},
	)
	require.NoError(t, err)
	require.Len(t, results, 1, "expected a rolled-up 10m probe result from recent 1m samples")
	assert.Equal(t, probe.Id, results[0].GetString("probe"))
	assert.Equal(t, systemRecord.Id, results[0].GetString("system"))
	assert.Equal(t, "10m", results[0].GetString("bucket"))
	assert.True(t, results[0].GetBool("success"))
	require.NotZero(t, results[0].GetFloat("latency_ms"))
	assert.InDelta(t, 14.5, results[0].GetFloat("latency_ms"), 0.01)
}

func TestCreateLongerNetworkProbeRecordsMixedWindowUsesSuccessfulLatencyAverage(t *testing.T) {
	hub, err := tests.NewTestHub(t.TempDir())
	require.NoError(t, err)
	defer hub.Cleanup()

	probe, systemRecord := createNetworkProbeFixture(t, hub)
	now := waitForStableProbeRollupWindowFor(10 * time.Minute)
	windowStart := completedProbeRollupWindowSampleStart(now, 10*time.Minute)

	inputs := []networkProbeResultInput{
		{
			probeID:   probe.Id,
			systemID:  systemRecord.Id,
			created:   windowStart,
			bucket:    "1m",
			success:   true,
			latencyMs: 10,
		},
		{
			probeID:           probe.Id,
			systemID:          systemRecord.Id,
			created:           windowStart.Add(100 * time.Millisecond),
			bucket:            "1m",
			success:           false,
			errorMessage:      "timeout",
			failureCategory:   "timeout",
			packetLossPercent: floatPtr(25),
		},
		{
			probeID:    probe.Id,
			systemID:   systemRecord.Id,
			created:    windowStart.Add(200 * time.Millisecond),
			bucket:     "1m",
			success:    true,
			latencyMs:  20,
			httpStatus: intPtr(204),
		},
	}
	for i := 0; i < 7; i++ {
		inputs = append(inputs, networkProbeResultInput{
			probeID:           probe.Id,
			systemID:          systemRecord.Id,
			created:           windowStart.Add(time.Duration(i+3) * 100 * time.Millisecond),
			bucket:            "1m",
			success:           false,
			errorMessage:      "timeout",
			failureCategory:   "timeout",
			packetLossPercent: floatPtr(25),
		})
	}
	for _, input := range inputs {
		createNetworkProbeResultWithOptions(t, hub, input)
	}

	rm := records.NewRecordManager(hub)
	rm.CreateLongerNetworkProbeRecords()

	results, err := hub.FindRecordsByFilter(
		"network_probe_results",
		"probe = {:probe} && system = {:system} && bucket = {:bucket}",
		"created",
		-1,
		0,
		dbx.Params{
			"probe":  probe.Id,
			"system": systemRecord.Id,
			"bucket": "10m",
		},
	)
	require.NoError(t, err)
	require.Len(t, results, 1)

	assert.True(t, results[0].GetBool("success"))
	assert.InDelta(t, 15.0, results[0].GetFloat("latency_ms"), 0.01)
	assert.Empty(t, results[0].GetString("error"))
	assert.Empty(t, results[0].GetString("failure_category"))
}

func TestCreateLongerNetworkProbeRecordsAllFailureUsesLatestFailureMetadata(t *testing.T) {
	hub, err := tests.NewTestHub(t.TempDir())
	require.NoError(t, err)
	defer hub.Cleanup()

	probe, systemRecord := createNetworkProbeFixture(t, hub)
	now := waitForStableProbeRollupWindowFor(10 * time.Minute)
	windowStart := completedProbeRollupWindowSampleStart(now, 10*time.Minute)

	createNetworkProbeResultWithOptions(t, hub, networkProbeResultInput{
		probeID:           probe.Id,
		systemID:          systemRecord.Id,
		created:           windowStart,
		bucket:            "1m",
		success:           false,
		errorMessage:      "connection refused",
		failureCategory:   "connection_refused",
		httpStatus:        intPtr(503),
		packetLossPercent: floatPtr(100),
	})
	for i := 0; i < 8; i++ {
		createNetworkProbeResultWithOptions(t, hub, networkProbeResultInput{
			probeID:           probe.Id,
			systemID:          systemRecord.Id,
			created:           windowStart.Add(time.Duration(i+1) * 100 * time.Millisecond),
			bucket:            "1m",
			success:           false,
			errorMessage:      "connection refused",
			failureCategory:   "connection_refused",
			httpStatus:        intPtr(503),
			packetLossPercent: floatPtr(100),
		})
	}
	createNetworkProbeResultWithOptions(t, hub, networkProbeResultInput{
		probeID:           probe.Id,
		systemID:          systemRecord.Id,
		created:           windowStart.Add(900 * time.Millisecond),
		bucket:            "1m",
		success:           false,
		errorMessage:      "timeout",
		failureCategory:   "timeout",
		httpStatus:        intPtr(504),
		packetLossPercent: floatPtr(75),
	})

	rm := records.NewRecordManager(hub)
	rm.CreateLongerNetworkProbeRecords()

	results, err := hub.FindRecordsByFilter(
		"network_probe_results",
		"probe = {:probe} && system = {:system} && bucket = {:bucket}",
		"created",
		-1,
		0,
		dbx.Params{
			"probe":  probe.Id,
			"system": systemRecord.Id,
			"bucket": "10m",
		},
	)
	require.NoError(t, err)
	require.Len(t, results, 1)

	assert.False(t, results[0].GetBool("success"))
	assert.Zero(t, results[0].GetFloat("latency_ms"))
	assert.Equal(t, "timeout", results[0].GetString("error"))
	assert.Equal(t, "timeout", results[0].GetString("failure_category"))
	assert.Equal(t, 504, results[0].GetInt("http_status"))
	assert.InDelta(t, 75.0, results[0].GetFloat("packet_loss_percent"), 0.01)
}

func TestCreateLongerNetworkProbeRecordsSkipsDuplicateWindow(t *testing.T) {
	hub, err := tests.NewTestHub(t.TempDir())
	require.NoError(t, err)
	defer hub.Cleanup()

	probe, systemRecord := createNetworkProbeFixture(t, hub)
	now := waitForStableProbeRollupWindowFor(10 * time.Minute)
	windowStart := completedProbeRollupWindowSampleStart(now, 10*time.Minute)

	for i := range 10 {
		createNetworkProbeResultRecord(t, hub, probe.Id, systemRecord.Id, windowStart.Add(time.Duration(i)*100*time.Millisecond), "1m", float64(5+i))
	}

	rm := records.NewRecordManager(hub)
	rm.CreateLongerNetworkProbeRecords()
	rm.CreateLongerNetworkProbeRecords()

	results, err := hub.FindRecordsByFilter(
		"network_probe_results",
		"probe = {:probe} && system = {:system} && bucket = {:bucket}",
		"created",
		-1,
		0,
		dbx.Params{
			"probe":  probe.Id,
			"system": systemRecord.Id,
			"bucket": "10m",
		},
	)
	require.NoError(t, err)
	assert.Len(t, results, 1)
}

func TestCreateLongerNetworkProbeRecordsIncludesLegacyUnbucketedSamples(t *testing.T) {
	hub, err := tests.NewTestHub(t.TempDir())
	require.NoError(t, err)
	defer hub.Cleanup()

	probe, systemRecord := createNetworkProbeFixture(t, hub)
	now := waitForStableProbeRollupWindowFor(10 * time.Minute)
	windowStart := completedProbeRollupWindowSampleStart(now, 10*time.Minute)

	for i := range 10 {
		createNetworkProbeResultRecord(t, hub, probe.Id, systemRecord.Id, windowStart.Add(time.Duration(i)*100*time.Millisecond), "", float64(20+i))
	}

	rm := records.NewRecordManager(hub)
	rm.CreateLongerNetworkProbeRecords()

	results, err := hub.FindRecordsByFilter(
		"network_probe_results",
		"probe = {:probe} && system = {:system} && bucket = {:bucket}",
		"created",
		-1,
		0,
		dbx.Params{
			"probe":  probe.Id,
			"system": systemRecord.Id,
			"bucket": "10m",
		},
	)
	require.NoError(t, err)
	require.Len(t, results, 1)
	assert.InDelta(t, 24.5, results[0].GetFloat("latency_ms"), 0.01)
}

func TestCreateLongerNetworkProbeRecordsSkipsActiveWindow(t *testing.T) {
	hub, err := tests.NewTestHub(t.TempDir())
	require.NoError(t, err)
	defer hub.Cleanup()

	probe, systemRecord := createNetworkProbeFixture(t, hub)
	now := waitForStableProbeRollupWindow()

	for i := range 9 {
		createNetworkProbeResultRecord(
			t,
			hub,
			probe.Id,
			systemRecord.Id,
			now.Add(-time.Duration(8-i)*100*time.Millisecond),
			"1m",
			float64(10+i),
		)
	}

	rm := records.NewRecordManager(hub)
	rm.CreateLongerNetworkProbeRecords()

	results, err := hub.FindRecordsByFilter(
		"network_probe_results",
		"probe = {:probe} && system = {:system} && bucket = {:bucket}",
		"created",
		-1,
		0,
		dbx.Params{
			"probe":  probe.Id,
			"system": systemRecord.Id,
			"bucket": "10m",
		},
	)
	require.NoError(t, err)
	assert.Len(t, results, 0, "expected no 10m rollup for the active in-progress window")
}

func TestDeleteOldNetworkProbeResults(t *testing.T) {
	hub, err := tests.NewTestHub(t.TempDir())
	require.NoError(t, err)
	defer hub.Cleanup()

	probe, systemRecord := createNetworkProbeFixture(t, hub)
	now := time.Now().UTC()

	testCases := []struct {
		bucket       string
		ageFromNow   time.Duration
		shouldBeKept bool
	}{
		{bucket: "", ageFromNow: 30 * time.Minute, shouldBeKept: true},
		{bucket: "", ageFromNow: 2 * time.Hour, shouldBeKept: false},
		{bucket: "1m", ageFromNow: 30 * time.Minute, shouldBeKept: true},
		{bucket: "1m", ageFromNow: 2 * time.Hour, shouldBeKept: false},
		{bucket: "10m", ageFromNow: 6 * time.Hour, shouldBeKept: true},
		{bucket: "10m", ageFromNow: 24 * time.Hour, shouldBeKept: false},
		{bucket: "20m", ageFromNow: 12 * time.Hour, shouldBeKept: true},
		{bucket: "20m", ageFromNow: 48 * time.Hour, shouldBeKept: false},
		{bucket: "120m", ageFromNow: 3 * 24 * time.Hour, shouldBeKept: true},
		{bucket: "120m", ageFromNow: 10 * 24 * time.Hour, shouldBeKept: false},
		{bucket: "480m", ageFromNow: 15 * 24 * time.Hour, shouldBeKept: true},
		{bucket: "480m", ageFromNow: 45 * 24 * time.Hour, shouldBeKept: false},
	}

	recordIDs := make([]string, 0, len(testCases))
	for i, tc := range testCases {
		record := createNetworkProbeResultRecord(t, hub, probe.Id, systemRecord.Id, now.Add(-tc.ageFromNow), tc.bucket, float64(i+1))
		recordIDs = append(recordIDs, record.Id)
	}

	rm := records.NewRecordManager(hub)
	rm.DeleteOldRecords()

	for i, tc := range testCases {
		_, err := hub.FindRecordById("network_probe_results", recordIDs[i])
		if tc.shouldBeKept {
			assert.NoError(t, err, "expected recent %s network probe result to remain", tc.bucket)
		} else {
			require.Error(t, err, "expected old %s network probe result to be deleted", tc.bucket)
		}
	}
}

func waitForStableProbeRollupWindow() time.Time {
	return waitForStableProbeRollupWindowFor(10 * time.Minute)
}

func waitForStableProbeRollupWindowFor(window time.Duration) time.Time {
	now := time.Now().UTC()
	windowStart := now.Truncate(window)
	elapsed := now.Sub(windowStart)
	if elapsed >= 2*time.Second {
		return now
	}
	time.Sleep(2*time.Second - elapsed)
	return time.Now().UTC()
}

func completedProbeRollupWindowSampleStart(now time.Time, window time.Duration) time.Time {
	return now.Truncate(window).Add(-1500 * time.Millisecond)
}

func createNetworkProbeFixture(t testing.TB, app core.App) (*core.Record, *core.Record) {
	t.Helper()

	user, err := tests.CreateUser(app, "probe-records@example.com", "password123")
	require.NoError(t, err)

	systemRecord, err := tests.CreateRecord(app, "systems", map[string]any{
		"name":   "node",
		"host":   "127.0.0.1",
		"status": "up",
		"users":  []string{user.Id},
	})
	require.NoError(t, err)

	probe, err := tests.CreateRecord(app, "network_probes", map[string]any{
		"name":             "line",
		"type":             "tcping",
		"scope":            "fixed",
		"target":           "example.com:443",
		"interval_seconds": 60,
		"timeout_seconds":  5,
		"enabled":          true,
		"public_visible":   true,
	})
	require.NoError(t, err)

	return probe, systemRecord
}

func createNetworkProbeResultRecord(t testing.TB, app core.App, probeID string, systemID string, created time.Time, bucket string, latency float64) *core.Record {
	t.Helper()
	return createNetworkProbeResultWithOptions(t, app, networkProbeResultInput{
		probeID:   probeID,
		systemID:  systemID,
		created:   created,
		bucket:    bucket,
		success:   true,
		latencyMs: latency,
	})
}

type networkProbeResultInput struct {
	probeID           string
	systemID          string
	created           time.Time
	bucket            string
	success           bool
	latencyMs         float64
	errorMessage      string
	failureCategory   string
	packetLossPercent *float64
	httpStatus        *int
}

func createNetworkProbeResultWithOptions(t testing.TB, app core.App, input networkProbeResultInput) *core.Record {
	t.Helper()

	record, err := tests.CreateRecord(app, "network_probe_results", map[string]any{
		"probe":            input.probeID,
		"system":           input.systemID,
		"type":             "tcping",
		"target":           "example.com:443",
		"success":          input.success,
		"latency_ms":       input.latencyMs,
		"error":            input.errorMessage,
		"failure_category": input.failureCategory,
		"bucket":           input.bucket,
	})
	require.NoError(t, err)

	if input.packetLossPercent != nil {
		record.Set("packet_loss_percent", *input.packetLossPercent)
	}
	if input.httpStatus != nil {
		record.Set("http_status", *input.httpStatus)
	}
	record.SetRaw("created", input.created.Format(types.DefaultDateLayout))
	require.NoError(t, app.SaveNoValidate(record))
	return record
}

func floatPtr(v float64) *float64 {
	return &v
}

func intPtr(v int) *int {
	return &v
}
