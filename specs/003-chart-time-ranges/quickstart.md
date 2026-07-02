# Quickstart: Public Chart Time Ranges

## Prerequisites

- Local Beszel development workspace.
- Existing Docker Compose setup with a hub and agent.
- At least one public system with CPU, memory, disk, and public latency probe
  history.
- Playwright available through the existing validation setup or temporary
  runner.

## Validation Commands

Run focused backend tests:

```bash
go test -tags=testing ./internal/hub ./internal/common ./internal/hub/ws
```

Run frontend quality and build checks:

```bash
npm --prefix ./internal/site run check
npm --prefix ./internal/site run build
```

If the full Go lint tool is available in the environment, run:

```bash
golangci-lint run
```

Start the local validation stack:

```bash
docker compose up -d --build
```

## API Validation

Default range should behave as 30 minutes:

```bash
curl -fsS http://127.0.0.1:8090/api/beszel/public/status | jq '.systems[0] | {historyCount:(.history|length), first:.history[0].created, last:.history[-1].created, probeSeries:(.probes|map({name, count:(.series|length), first:.series[0].created, last:.series[-1].created}))}'
```

Explicit 30-minute range should match the default semantics:

```bash
curl -fsS 'http://127.0.0.1:8090/api/beszel/public/status?range=30m' | jq '.systems[0] | {historyCount:(.history|length), probeCount:(.probes|length)}'
```

Longer ranges should return range-appropriate data without exposing targets:

```bash
curl -fsS 'http://127.0.0.1:8090/api/beszel/public/status?range=1h' | jq '.systems[0].probes[] | {name, type, targetLabel, seriesCount:(.series|length)}'
```

Expected:

- No `targetLabel`, hostname, IP, or port is present in public probe summaries.
- Invalid explicit range values return a clear client error or documented
  fallback behavior.

## Browser Validation

Open:

```text
http://127.0.0.1:8090/
```

Validate default public latency chart:

- The chart defaults to the latest 30-minute window.
- The x-axis is visible and uses hour-minute-second labels.
- Labels do not wrap or overlap at desktop and mobile widths.
- The chart refreshes within the 20-second cadence while preserving the selected range.

Validate latency range selector:

- Open the range selector in the public latency chart area.
- Select `1小时`, `12小时`, `24小时`, `1周`, and `30天` where available.
- Confirm plotted points and x-axis labels update without a page reload.
- Confirm selecting a range on one node does not change another node unexpectedly.

Validate resource trend dialog:

- Click a public node's CPU, memory, or disk summary.
- Confirm the dialog opens with CPU, memory, and disk charts.
- Confirm each chart defaults to 30 minutes.
- Confirm x-axis labels are visible in hour-minute-second format.
- Change the range and confirm all three resource charts update consistently.
- Confirm the latest chart value remains consistent with the card's current
  CPU, memory, and disk summary values.

## Playwright Smoke Checks

Recommended automated checks:

- Public page shows one range selector for the public latency chart.
- Resource trend dialog shows a range selector and three charts.
- Default chart data excludes points older than 30 minutes when more history exists.
- Visible x-axis labels contain time text and do not create multiple-line tick labels.
- Public page text and serialized response do not include hidden probe targets.
- Selecting a non-default range changes the public latency chart x-axis labels
  and preserves the selection across the next 20-second refresh.
- Opening the resource trend dialog after selecting a range shows CPU, memory,
  and disk charts using that same range.

## Expected Completion

The feature is considered validated when:

- Backend tests pass.
- Frontend check/build pass, or pre-existing unrelated diagnostics are documented.
- Docker stack runs with the updated public dashboard.
- API contract behavior matches `contracts/public-status-range.md`.
- Browser validation confirms default 30-minute charts, 20-second refresh,
  readable x-axis labels, range selection, and target sanitization.
