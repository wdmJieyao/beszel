# Quickstart: Validate Latency Realtime Window

## Prerequisites

- Local Docker deployment is available via `docker compose`.
- At least one node is connected and visible in the hub.
- The node has at least three enabled latency probes assigned.
- The frontend can be accessed at `http://127.0.0.1:8090`.

## Quality Gates

Run these before manual validation:

```bash
npm --prefix ./internal/site run test:unit
npm --prefix ./internal/site run check
npm --prefix ./internal/site run build
```

If Go files are changed during implementation, also run the focused Go tests for touched packages and the repository Go quality gate available in the environment.

Implementation note: this 004 live-window slice does not require new Go backend, agent, storage, or HTTP API changes. The current repository may contain pre-existing backend/agent work from earlier latency-probe features; validate this slice by reviewing the frontend files listed in [tasks.md](./tasks.md) and by confirming no additional backend/API changes were introduced for the realtime-window display behavior.

## Local Deployment

```bash
docker compose up -d --build beszel beszel-agent
docker compose ps
```

Expected outcome: both hub and local agent containers are running.

## Scenario 1: `1 分钟` Starts Empty

1. Sign in to the hub.
2. Open a node detail page with existing latency history.
3. Confirm the chart range is a longer range such as `1 小时`.
4. Switch the page-wide chart range to `1 分钟`.

Expected outcome:

- The线路检测 card title and legend remain visible.
- All configured latency line labels are visible in the legend.
- The plot area does not draw a line from historical latency points recorded before the switch.

## Scenario 2: Fresh Results Append

1. Stay on the same node detail page in `1 分钟`.
2. Wait for fresh probe results to arrive.
3. Observe the chart over multiple probe intervals.

Expected outcome:

- Lines begin only after fresh realtime samples are received by the browser after switching.
- A line with no fresh samples remains in the legend but does not draw stale history.
- Once all configured lines have enough fresh samples, all lines are plotted.

## Scenario 3: Re-entering `1 分钟` Resets

1. While in `1 分钟`, wait until at least one latency line is plotted.
2. Switch to `1 小时`.
3. Switch back to `1 分钟`.

Expected outcome:

- The线路检测 chart starts a new live observation session.
- Previous `1 分钟` session points are not reused as plotted data.
- New lines appear only as fresh realtime results are received after re-entering `1 分钟`.

Validation note: the `1 分钟` view is intentionally driven by browser realtime events received after entering the range. Do not validate it by expecting records from a historical query window such as `created >= switch time`.

## Scenario 4: Longer Ranges Still Show History

1. Switch from `1 分钟` to `30 分钟` or `1 小时`.
2. Inspect the线路检测 chart.

Expected outcome:

- The selected longer range displays historical latency data for that range.
- Historical behavior remains available outside `1 分钟`.

## Optional Browser Automation Checks

Use Playwright or an equivalent browser runner to assert:

- After switching to `1 分钟`, legend labels include all configured probes.
- Immediately after switching, the线路检测 SVG contains no plotted line path from previously fetched or previously received data.
- After fresh samples arrive, plotted line paths correspond only to realtime events received after the browser entered the current `1 分钟` session.
