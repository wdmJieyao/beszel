# Quickstart: Validate Latency Sample Cadence

## Prerequisites

- Local Docker deployment is available via `docker compose`.
- At least one node is connected and visible in the hub.
- The node has three enabled latency-capable probes assigned, such as 广东电信、广东联通、广东移动.
- The frontend can be accessed at `http://127.0.0.1:8090`.

## Quality Gates

Run these before browser validation:

```bash
/usr/local/go/bin/go test -tags=testing ./...
npm --prefix ./internal/site run test:unit
npm --prefix ./internal/site run check
npm --prefix ./internal/site run build
```

If `golangci-lint` is available in the environment, also run:

```bash
golangci-lint run
```

Current workspace note: `go` is not on `PATH`, so use `/usr/local/go/bin/go`. `golangci-lint` may be unavailable in the current environment; if so, record that as an environment blocker instead of treating it as a feature failure.

## Local Deployment

```bash
docker compose up -d --build beszel beszel-agent
docker compose ps
```

Expected outcome: hub and agent containers are running, and the hub is reachable at `http://127.0.0.1:8090`.

## Scenario 1: `1 分钟` Starts Live Cadence

1. Sign in to the hub.
2. Open a node detail page with three enabled latency lines.
3. Switch the page-wide range selector to `1 分钟`.
4. Observe `线路检测` for 20 seconds.

Expected outcome:

- The chart still starts from the current live observation window.
- All three configured line labels remain visible.
- Fresh samples begin arriving at roughly 1-second spacing.
- Reachable lines render connected segments instead of isolated points.

## Scenario 2: 60-Second Density Check

1. Stay on the same node detail page in `1 分钟`.
2. Observe for 60 seconds.
3. Use browser automation or realtime/result inspection to count fresh successful samples per line.

Expected outcome:

- Each reachable line receives at least 40 successful samples in 60 seconds.
- At least 90% of adjacent successful samples for each reachable line are no more than 4 seconds apart.
- The chart displays a continuous-looking waveform for each reachable line.

## Scenario 3: Leaving `1 分钟` Stops Live Cadence

1. While `1 分钟` live cadence is active, switch the page-wide range to `30 分钟`.
2. Wait at least 10 seconds.
3. Inspect the chart and recent probe result cadence.

Expected outcome:

- The `30 分钟` option is visible and selectable.
- The `30 分钟` chart shows historical data for the selected range when records exist.
- The page no longer maintains a live latency session.
- Probe execution returns to normal configured interval behavior unless another browser is still actively watching `1 分钟` for the same node.
- The waveform no longer uses the crowded one-minute realtime plotting density.

## Scenario 4: Longer Historical Ranges Stay Readable

1. After collecting at least 60 seconds of high-cadence `1 分钟` samples, switch the page-wide range to `1 小时`.
2. Then switch to `12 小时`, `24 小时`, `1 周`, and `30 天` if data exists.
3. Inspect the horizontal axis and line density after each switch.

Expected outcome:

- Each historical range loads persisted results for the selected window.
- The chart dynamically adjusts horizontal-axis ticks and point density for the selected range.
- The chart does not keep the one-minute realtime trace density across wider ranges.
- Existing high-cadence samples may contribute to historical data, but they are bucketed or thinned enough that the chart remains readable.

## Scenario 5: Failure Accuracy

1. Configure one reachable latency line and one intentionally failing latency line.
2. Open the node detail page and switch to `1 分钟`.
3. Observe the chart for 30 seconds.

Expected outcome:

- The reachable line becomes smooth.
- The failing line remains represented in the legend and shows failure/gap behavior.
- Failed measurements are not drawn as fake successful latency segments.

## Scenario 6: Public Dashboard Does Not Start Live Cadence

1. Open the anonymous public dashboard.
2. Select a public chart range if available.
3. Observe probe result creation cadence.

Expected outcome:

- Public dashboard latency charts keep historical behavior.
- Viewing the public dashboard does not create live latency sessions.
- No 1-second cadence starts unless an authenticated node detail page is actively in `1 分钟`.

## Optional Playwright Checks

Use Playwright against `http://127.0.0.1:8090` to assert:

- The frontend calls the live-session create endpoint after switching node detail range to `1 分钟`.
- The frontend renews the session while it remains in `1 分钟`.
- The frontend ends or lets the session expire after leaving `1 分钟`.
- The range selector includes `30 分钟`.
- Switching to `30 分钟` and `1 小时` does not call the live-session create/renew endpoints.
- Historical ranges render with bounded visible points and readable x-axis ticks instead of one-minute density.
- After 20 seconds, `.recharts-line-curve` exists for reachable lines.
- After 60 seconds, each reachable configured line has at least 40 fresh samples.
- Result inspection can additionally confirm approximately 20 samples per line in the first 20 seconds and that counts drop back to normal after leaving `1 分钟`.
