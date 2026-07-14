# Network Probe Retention And Chart Stability Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Fix the long-running network probe chart failure by introducing range-aware probe retention/aggregation on the backend and resilient partial rendering on the frontend.

**Architecture:** Keep probe execution unchanged in the hub and agent path, but stop serving unbounded raw `1m` probe rows for historical views. Add probe-result rollups and retention in `internal/records`, teach both public and authenticated probe APIs to query the correct bucket for the selected range, and harden the React data loader so a single failed probe request does not blank the whole chart area.

**Tech Stack:** Go 1.26.3, PocketBase/SQLite, React 19, TypeScript 5.9, Vite 7, Recharts, Biome, Go tests, Node unit tests.

---

## Problem Summary

Confirmed diagnosis from local reproduction:

- `network_probe_results` currently stores only raw `1m` samples and never aggregates them into `10m/20m/120m/480m` buckets.
- `network_probe_results` also has no retention cleanup, so data grows without bound.
- Public status currently returns full historical probe series for the requested range via `internal/hub/public_status.go`, using `FindRecordsByFilter(..., -1, 0, ...)`.
- With local data cloned and amplified to 10 probe lines and ~305k probe result rows, `/api/beszel/public/status?range=30d` returned about `22MB` and took about `2.44s`.
- Authenticated detail probe results already cap to `500` rows, but the frontend uses `Promise.all`, so a single failing probe request flips the entire probe section into error.

## File Structure

- Modify: `internal/records/records.go`
  Add actual probe rollup creation instead of the current no-op hook.
- Modify: `internal/records/records_deletion.go`
  Add retention cleanup for `network_probe_results` by bucket.
- Modify: `internal/hub/hub.go`
  Ensure probe rollup creation runs with the existing long-record cron path.
- Modify: `internal/hub/network_probes.go`
  Make authenticated probe history endpoint choose bucket by requested range.
- Modify: `internal/hub/public_status.go`
  Make public probe summaries query the matching bucket for the selected range instead of always scanning raw `1m` rows.
- Modify: `internal/hub/network_probes_test.go`
  Cover range-to-bucket mapping and ordered result output.
- Modify: `internal/hub/public_status_test.go`
  Cover public range queries using aggregated buckets.
- Create: `internal/records/network_probe_records_test.go`
  Focused backend tests for probe rollup and retention cleanup.
- Modify: `internal/site/src/components/routes/system/use-network-probe-data.ts`
  Change probe history fetch to tolerate per-line failures and preserve successful lines.
- Create: `internal/site/src/components/routes/system/use-network-probe-data.partial-failure.test.ts`
  Frontend regression test for partial probe request failure.
- Optionally modify: `internal/site/src/components/routes/system.tsx`
  If needed, surface a softer per-line warning instead of a full probe section failure.

## Range/Bucket Contract

Use the same range cadence already defined for public metric charts:

- `1m` -> live session / raw realtime only
- `30m` -> `1m`
- `1h` -> `1m`
- `12h` -> `10m`
- `24h` -> `20m`
- `1w` -> `120m`
- `30d` -> `480m`

The bugfix should make probe queries follow this same mapping consistently in:

- authenticated detail charts
- public status charts
- future retention cleanup logic

### Task 1: Lock Down The Regression With Backend Tests

**Files:**
- Create: `internal/records/network_probe_records_test.go`
- Modify: `internal/hub/network_probes_test.go`
- Modify: `internal/hub/public_status_test.go`

- [ ] **Step 1: Write failing probe rollup/retention tests**

```go
func TestCreateLongerNetworkProbeRecordsCreatesTenMinuteBuckets(t *testing.T) {
	hub, _ := tests.GetHubWithUser(t)
	now := time.Now().UTC()
	system := createNetworkProbeTestRecord(t, hub, "systems", map[string]any{
		"name": "probe-node",
		"host": "127.0.0.1",
		"port": "45876",
	})
	probe := createNetworkProbeTestRecord(t, hub, CollectionNetworkProbes, map[string]any{
		"name": "probe-a",
		"type": "tcping",
		"target": "example.com:80",
		"enabled": true,
		"interval_seconds": 10,
		"timeout_seconds": 5,
	})
	for i := 0; i < 10; i++ {
		createNetworkProbeTestRecord(t, hub, CollectionNetworkProbeResults, map[string]any{
			"probe": probe.Id,
			"system": system.Id,
			"type": "tcping",
			"target": "example.com:80",
			"success": true,
			"latency_ms": 10 + i,
			"bucket": "1m",
			"created": now.Add(-time.Duration(9-i) * time.Minute),
		})
	}

	hub.GetRecordManager().CreateLongerNetworkProbeRecords()

	count, err := hub.CountRecords(CollectionNetworkProbeResults, dbx.NewExp("probe = {:probe} AND system = {:system} AND bucket = '10m'", dbx.Params{
		"probe": probe.Id,
		"system": system.Id,
	}))
	require.NoError(t, err)
	assert.EqualValues(t, 1, count)
}
```

- [ ] **Step 2: Write failing public/auth range tests**

```go
func TestGetNetworkProbeResultsUsesBucketForLongRanges(t *testing.T) {
	hub, user := tests.GetHubWithUser(t)
	system := createNetworkProbeTestRecord(t, hub, "systems", map[string]any{
		"name": "probe-node",
		"host": "127.0.0.1",
		"port": "45876",
		"users": []string{user.Id},
	})
	probe := createNetworkProbeTestRecord(t, hub, CollectionNetworkProbes, map[string]any{
		"name": "probe-a",
		"type": "tcping",
		"target": "example.com:80",
		"enabled": true,
	})
	createNetworkProbeTestRecord(t, hub, CollectionNetworkProbeResults, map[string]any{
		"probe": probe.Id,
		"system": system.Id,
		"type": "tcping",
		"target": "example.com:80",
		"success": true,
		"latency_ms": 18,
		"bucket": "120m",
	})

	response := requestNetworkProbeResults(t, hub, probe.Id, system.Id, "1w")
	require.Len(t, response.Series, 1)
	assert.Equal(t, 18.0, response.Series[0].LatencyMs)
}
```

- [ ] **Step 3: Run targeted tests and confirm they fail before implementation**

Run:

```bash
/usr/local/go/bin/go test -tags=testing ./internal/records ./internal/hub -run 'Test(CreateLongerNetworkProbeRecords|GetNetworkProbeResultsUsesBucketForLongRanges|PublicStatus.*Probe)'
```

Expected:

- FAIL because probe rollup is still a no-op
- FAIL because long-range probe queries still look only at raw `1m` rows

- [ ] **Step 4: Commit test-only red state**

```bash
git add internal/records/network_probe_records_test.go internal/hub/network_probes_test.go internal/hub/public_status_test.go
git commit -m "test: cover network probe history retention regressions"
```

### Task 2: Implement Probe Result Rollups And Retention

**Files:**
- Modify: `internal/records/records.go`
- Modify: `internal/records/records_deletion.go`
- Modify: `internal/hub/hub.go`

- [ ] **Step 1: Implement probe rollup creation**

```go
func (rm *RecordManager) CreateLongerNetworkProbeRecords() {
	longerRecordData := []LongerRecordData{
		{shorterType: "1m", minShorterRecords: 9, longerType: "10m", longerTimeDuration: -10 * time.Minute},
		{shorterType: "10m", minShorterRecords: 2, longerType: "20m", longerTimeDuration: -20 * time.Minute},
		{shorterType: "20m", minShorterRecords: 6, longerType: "120m", longerTimeDuration: -120 * time.Minute},
		{shorterType: "120m", minShorterRecords: 4, longerType: "480m", longerTimeDuration: -480 * time.Minute},
	}
	_ = rm.app.RunInTransaction(func(txApp core.App) error {
		return createLongerProbeRecords(txApp, longerRecordData)
	})
}
```

- [ ] **Step 2: Add probe aggregation helper with stable averaging semantics**

```go
func createLongerProbeRecords(app core.App, definitions []LongerRecordData) error {
	// group by probe + system, read shorter bucket rows in window,
	// create one longer bucket row with averaged latency for success rows,
	// failed aggregate row when no successful sample exists in the window.
	return nil
}
```

Implementation rules:

- average `latency_ms` across successful rows only
- if all rows failed, write one failed aggregate row with `latency_ms = 0/null`
- preserve last meaningful `failure_category`, `error`, `http_status`, `packet_loss_percent`
- skip if longer bucket row for the same probe/system/window already exists

- [ ] **Step 3: Add probe retention cleanup**

```go
func deleteOldNetworkProbeResults(app core.App) error {
	recordData := []RecordDeletionData{
		{recordType: "1m", retention: time.Hour},
		{recordType: "10m", retention: 12 * time.Hour},
		{recordType: "20m", retention: 24 * time.Hour},
		{recordType: "120m", retention: 7 * 24 * time.Hour},
		{recordType: "480m", retention: 30 * 24 * time.Hour},
	}
	// DELETE FROM network_probe_results WHERE (bucket = ... AND created < ...)
	return nil
}
```

- [ ] **Step 4: Wire probe rollup/retention into existing cron path**

```go
func (rm *RecordManager) CreateLongerRecords() {
	// existing system/container rollup
	rm.CreateLongerNetworkProbeRecords()
}

func (rm *RecordManager) DeleteOldRecords() {
	// existing cleanup
	_ = deleteOldNetworkProbeResults(txApp)
}
```

- [ ] **Step 5: Run targeted backend tests**

Run:

```bash
/usr/local/go/bin/go test -tags=testing ./internal/records ./internal/hub -run 'Test(CreateLongerNetworkProbeRecords|DeleteOldNetworkProbeResults|GetNetworkProbeResultsUsesBucketForLongRanges|PublicStatus.*Probe)'
```

Expected:

- PASS for the new probe rollup and retention tests

- [ ] **Step 6: Commit backend storage fix**

```bash
git add internal/records/records.go internal/records/records_deletion.go internal/hub/hub.go internal/records/network_probe_records_test.go internal/hub/network_probes_test.go internal/hub/public_status_test.go
git commit -m "feat: add probe history rollups and retention"
```

### Task 3: Make Public And Authenticated Probe Queries Range-Aware

**Files:**
- Modify: `internal/hub/network_probes.go`
- Modify: `internal/hub/public_status.go`
- Modify: `internal/hub/network_probes_test.go`
- Modify: `internal/hub/public_status_test.go`

- [ ] **Step 1: Add one shared range-to-bucket helper**

```go
func probeBucketForRange(rangeSpec publicChartRange) string {
	switch rangeSpec.Name {
	case "12h":
		return "10m"
	case "24h":
		return "20m"
	case "1w":
		return "120m"
	case "30d":
		return "480m"
	default:
		return "1m"
	}
}
```

- [ ] **Step 2: Use the helper in authenticated detail results**

```go
bucket := probeBucketForRange(rangeSpec)
filter := "probe = {:probe} && bucket = {:bucket} && created >= {:created}"
params := dbx.Params{
	"probe": probeID,
	"bucket": bucket,
	"created": time.Now().UTC().Add(-rangeSpec.Duration),
}
```

Keep the `500` row limit for authenticated detail charts.

- [ ] **Step 3: Use the helper in public status probe summaries**

```go
results, err := h.FindRecordsByFilter(
	CollectionNetworkProbeResults,
	"probe = {:probe} && system = {:system} && bucket = {:bucket} && created >= {:created}",
	"created",
	-1,
	0,
	dbx.Params{
		"probe": probe.Id,
		"system": systemID,
		"bucket": probeBucketForRange(rangeSpec),
		"created": time.Now().UTC().Add(-rangeSpec.Duration),
	},
)
```

- [ ] **Step 4: Re-run the red-capable HTTP loop**

Run:

```bash
for r in 30m 1h 12h 24h 1w 30d; do
  curl -sS -o /tmp/probe-$r.json -w "$r status=%{http_code} time=%{time_total} size=%{size_download}\n" \
    "http://127.0.0.1:8090/api/beszel/public/status?range=$r"
done
```

Expected after fix:

- `1w` and `30d` payload size should drop materially compared with the diagnosed `~16.7MB/~22MB` 10-line reproduction
- response time should also drop materially

- [ ] **Step 5: Commit range-aware query change**

```bash
git add internal/hub/network_probes.go internal/hub/public_status.go internal/hub/network_probes_test.go internal/hub/public_status_test.go
git commit -m "feat: query probe history by retention bucket"
```

### Task 4: Harden The Detail Page Against Partial Probe Failures

**Files:**
- Modify: `internal/site/src/components/routes/system/use-network-probe-data.ts`
- Create: `internal/site/src/components/routes/system/use-network-probe-data.partial-failure.test.ts`
- Optionally modify: `internal/site/src/components/routes/system.tsx`

- [ ] **Step 1: Write a failing frontend regression test**

```ts
it("keeps successful probe lines when one probe request fails", async () => {
	const assigned = [
		{ id: "probe-a", name: "A" },
		{ id: "probe-b", name: "B" },
	]
	// mock one success and one rejection
	// expect probe A to render while error state stays non-fatal
})
```

- [ ] **Step 2: Replace `Promise.all` with `Promise.allSettled`**

```ts
const settled = await Promise.allSettled(
	assigned.map(async (probe) => {
		const data = await getNetworkProbeResults(probe.id, { system: systemId, range })
		return [probe.id, data] as const
	})
)

const successfulEntries = settled
	.filter((entry): entry is PromiseFulfilledResult<readonly [string, { series: NetworkProbeResultPoint[] }]> => entry.status === "fulfilled")
	.map((entry) => entry.value)

const hadFailures = settled.some((entry) => entry.status === "rejected")
setResults((current) => mergeProbeResults(current, Object.fromEntries(successfulEntries), range))
setError(hadFailures && successfulEntries.length === 0)
```

- [ ] **Step 3: Surface a soft warning when some lines fail**

```tsx
const partialProbeFailure = hadFailures && successfulEntries.length > 0
```

If needed, render a small muted warning above the chart instead of hiding the chart.

- [ ] **Step 4: Run frontend unit tests**

Run:

```bash
npm --prefix ./internal/site run test:unit
```

Expected:

- PASS including the new partial-failure regression test

- [ ] **Step 5: Commit frontend resilience fix**

```bash
git add internal/site/src/components/routes/system/use-network-probe-data.ts internal/site/src/components/routes/system/use-network-probe-data.partial-failure.test.ts internal/site/src/components/routes/system.tsx
git commit -m "fix: keep probe charts visible on partial request failure"
```

### Task 5: Full Verification

**Files:**
- No source changes expected

- [ ] **Step 1: Run backend verification**

```bash
/usr/local/go/bin/go test -tags=testing ./...
```

Expected:

- PASS

- [ ] **Step 2: Run lint**

```bash
GOTOOLCHAIN=go1.26.3 /usr/local/go/bin/go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@latest run --build-tags testing
```

Expected:

- PASS with `0 issues`

- [ ] **Step 3: Run frontend verification**

```bash
npm --prefix ./internal/site run test:unit
npm --prefix ./internal/site run check
npm --prefix ./internal/site run build
```

Expected:

- unit tests pass
- Biome check passes
- production build passes

- [ ] **Step 4: Re-run manual probe scale validation**

Use the existing local compose or a cloned data set and verify:

- public `12h/24h/1w/30d` probe charts load without blanking
- detail page still shows partial data if one line fails
- probe history payload size is materially lower for `1w` and `30d`

- [ ] **Step 5: Commit final verification if code changed during validation**

```bash
git add -A
git commit -m "chore: verify probe history stability fix"
```

## Self-Review

Spec coverage for this repair plan:

- Historical probe growth: covered by Task 2.
- Public chart instability after several days: covered by Tasks 2 and 3.
- Detail page fragility when one probe request fails: covered by Task 4.
- Existing stack constraints: preserved; backend remains Go and frontend remains React/TypeScript/Vite.
- Unit-test requirement: covered in Tasks 1 and 4.

Placeholder scan:

- No `TODO`/`TBD` placeholders left.
- All tasks include exact files and exact verification commands.

Type consistency:

- Range mapping uses existing `publicChartRange`.
- Probe bucket values match existing `network_probe_results.bucket` values.
- Frontend partial-failure logic stays inside existing `useNetworkProbeData` seam.

## Execution Handoff

Plan complete and saved to `docs/superpowers/plans/2026-07-09-network-probe-retention-and-chart-stability.md`. Two execution options:

**1. Subagent-Driven (recommended)** - I dispatch a fresh subagent per task, review between tasks, fast iteration

**2. Inline Execution** - Execute tasks in this session using executing-plans, batch execution with checkpoints

Which approach?
