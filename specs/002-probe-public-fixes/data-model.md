# Data Model: Probe Chart and Public Dashboard Fixes

## Latency Chart Series

Represents one line in a combined latency chart.

**Fields**

- `series_id`: stable identifier derived from probe and executing node/path.
- `label`: user-facing line label, usually configured probe name plus target or node label when needed for disambiguation.
- `probe_id`: source probe identifier.
- `system_id`: executing node identifier.
- `type`: probe type; latency charts primarily include `tcping` and latency-capable `icmp_ping`.
- `target_label`: public-safe target label when allowed.
- `points`: ordered list of timestamped result points.

**Relationships**

- Derived from `Network Probe` and `Network Probe Result`.
- Belongs to a node detail context or public dashboard system context.

**Validation Rules**

- Multiple series may share a chart.
- Empty series must be allowed and rendered as pending/no data.
- Series labels must be safe for UI display and must avoid exposing private data on public pages.

## Probe Result Point

Represents one chartable result sample.

**Fields**

- `created`: result timestamp.
- `success`: whether the probe succeeded.
- `latency_ms`: latency value when available.
- `failure_category`: optional normalized category for failed checks.
- `error_label`: optional user-safe error message.
- `http_status`: optional HTTP response status.
- `packet_loss_percent`: optional packet loss for ping-like checks.

**Relationships**

- Belongs to one `Latency Chart Series`.

**Validation Rules**

- Failed results remain chartable.
- Runtime failures must not break other points or series.
- Error labels must not expose local paths, credentials, or private transport details.

## Probe Failure Category

User-facing classification for a failed probe.

**Values**

- `invalid_target`
- `dns_failure`
- `timeout`
- `connection_refused`
- `target_unreachable`
- `execution_node_unavailable`
- `unsupported`
- `unknown_failure`

**Validation Rules**

- Save-time validation failures use `invalid_target`.
- Offline or unsupported execution nodes use execution-node categories.
- Unknown agent errors must fall back to `unknown_failure` with a safe label.

## Public Metric Summary

Public-safe node health summary for the anonymous dashboard.

**Fields**

- `system_id`: public-safe system identifier.
- `name`: public display name.
- `status`: availability state.
- `freshness`: latest public-safe report timestamp or stale marker.
- `cpu_percent`: optional latest CPU percentage.
- `memory_percent`: optional latest memory percentage.
- `disk_percent`: optional latest disk percentage.
- `unavailable_fields`: list of metric fields that are genuinely unavailable.

**Relationships**

- Composed from public visibility settings and the latest available system report.

**Validation Rules**

- Only systems enabled for public display are included.
- Private host, port, address, token, user, and admin fields are excluded.
- Missing metrics are represented explicitly rather than blank UI values.

## Public Dashboard Refresh State

Client-side state for anonymous dashboard refresh.

**Fields**

- `last_loaded_at`: timestamp of the last successful public-status fetch.
- `generated_at`: server timestamp from the public-status response.
- `refresh_interval`: configured client polling interval.
- `error_state`: optional safe error state if refresh fails.

**Validation Rules**

- A failed refresh must not clear currently displayed public data.
- Refresh must not require login or authenticated realtime subscriptions.
