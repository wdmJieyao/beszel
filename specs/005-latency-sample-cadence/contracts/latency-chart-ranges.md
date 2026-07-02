# Contract: Latency Chart Range Rendering

## Scope

The authenticated node detail `线路检测` chart follows the page-wide time range. It must not introduce a separate latency-only range selector for this feature.

## Supported Ranges

The range list must include:

| Range key | Chinese label | Rendering mode |
| --- | --- | --- |
| `1m` | `1 分钟` | `live-realtime` |
| `30m` | `30 分钟` | `historical-range` |
| `1h` | `1 小时` | `historical-range` |
| `12h` | `12 小时` | `historical-range` |
| `24h` | `24 小时` | `historical-range` |
| `1w` | `1 周` | `historical-range` |
| `30d` | `30 天` | `historical-range` |

## `1 分钟` Live Realtime Mode

- Starts from an empty current observation window when the page-wide range changes to `1m`.
- Creates or renews a live latency session for the current system.
- Appends only fresh realtime probe results observed after entering the range.
- Keeps all configured enabled latency line legends visible, even before every line has a fresh point.
- Does not backfill older historical samples into the current one-minute view.

## Historical Range Mode

- Applies to `30m`, `1h`, `12h`, `24h`, `1w`, and `30d`.
- Does not create or renew live latency sessions.
- Loads persisted probe results for the selected time window.
- Uses range-appropriate bucketing, thinning, or chart point limits so high-cadence samples collected during earlier live sessions do not overcrowd the chart.
- Dynamically adjusts the horizontal axis to the selected range. It must not keep the one-minute plotting density after the operator expands the time range.
- Preserves failure semantics: failed samples are not converted into successful latency values.

## Public Dashboard

Public dashboard latency charts keep historical behavior and do not use the authenticated live-session API.

## Validation Expectations

- Switching from `1 分钟` to `30 分钟` shows historical data for the last 30 minutes when records exist.
- `30 分钟` remains visible and selectable after any range switch.
- Switching from `1 分钟` to `1 小时` or longer stops live-session renewal and redraws using historical density.
- The chart remains readable after high-cadence samples exist in storage; it must not render a visually overcrowded one-minute trace across wider ranges.
