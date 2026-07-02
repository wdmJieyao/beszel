# Data Model: Latency Sample Cadence

## Live Latency Session

Represents one browser's active request for high-cadence latency measurements on a node detail `1 分钟` chart.

**Fields**

- `session_id`: server-generated opaque identifier.
- `system_id`: node being observed.
- `user_id`: authenticated user that created the session.
- `range`: always `1m` for this feature.
- `cadence_seconds`: expected live cadence, normally `1`.
- `started_at`: when the live observation was created.
- `last_seen_at`: last create or heartbeat time from the browser.
- `expires_at`: server-side TTL cutoff used to clean up abandoned sessions.

**Relationships**

- Belongs to one authenticated user and one visible system.
- Activates high-cadence execution for enabled latency assignments on the same system.
- Multiple sessions for the same system are coalesced into one execution stream.

**Validation Rules**

- User must be authenticated.
- User must be allowed to view the system.
- Only `range = 1m` is valid.
- Expired sessions do not trigger live probe execution.
- Deleting or expiring one session must not stop execution if another active session remains for the same system.

**State Transitions**

1. `Created`: frontend enters `1 分钟` and creates the session.
2. `Active`: frontend renews before `expires_at`; hub live worker includes the system.
3. `Ended`: frontend leaves `1 分钟` or the session expires.
4. `Recreated`: frontend re-enters `1 分钟`; a new session id and live observation are used.

## Live Cadence System

Represents a system that currently has at least one active live latency session.

**Fields**

- `system_id`: node being observed.
- `active_session_count`: number of unexpired sessions.
- `next_due_at`: next time the live cadence worker should run latency assignments.
- `last_run_at`: last high-cadence execution attempt.
- `in_flight_assignments`: assignment identifiers currently running, used to prevent overlap.

**Relationships**

- Derived from active Live Latency Sessions.
- Runs enabled Live Latency Assignments.
- Persists High-Cadence Latency Samples.

**Validation Rules**

- One system should be executed once per cadence tick even when multiple users are watching.
- A still-running assignment should not be started again until the previous attempt completes or times out.
- Normal background probe scheduling remains independent.

## Live Latency Assignment

Represents an enabled latency-capable probe assignment eligible for high-cadence execution.

**Fields**

- `assignment_id`: existing probe assignment identifier.
- `probe_id`: existing probe identifier.
- `system_id`: node assigned to execute the probe.
- `type`: latency-capable probe type.
- `target`: configured target.
- `enabled`: whether the assignment and probe are enabled.
- `live_timeout_seconds`: bounded timeout used during high-cadence execution.

**Relationships**

- References existing Network Probe and Network Probe Assignment records.
- Produces High-Cadence Latency Samples through the existing agent request path.

**Validation Rules**

- Probe and assignment must both be enabled.
- Assignment must belong to the live system.
- Only latency chart line types are included.
- Invalid, offline, or unsupported execution produces a failed sample rather than blocking other assignments.

## High-Cadence Latency Sample

Represents one fresh result produced while a Live Latency Session is active.

**Fields**

- `probe_id`
- `system_id`
- `created`
- `success`
- `latency_ms`
- `packet_loss_percent`
- `failure_category`
- `error`
- `source`: live `1m` cadence source marker when available.

**Relationships**

- Stored in the existing probe result collection.
- Delivered to browser charts through existing realtime result subscriptions.
- Also remains available for historical queries unless a later implementation explicitly downsamples or filters it.

**Validation Rules**

- Successful samples can draw connected waveform segments.
- Failed samples preserve failure state and must not become fake successful latency values.
- Duplicate stale samples must not be emitted as new measurements.
- Sample timestamps should reflect actual check completion time.

## Normal Probe Schedule

Represents the configured background probe behavior outside active live observation.

**Fields**

- `interval_seconds`: administrator-configured normal interval.
- `timeout_seconds`: administrator-configured normal timeout.
- `enabled`: configured probe state.

**Relationships**

- Continues to use existing due checks and persisted results.
- Is not replaced by Live Latency Session behavior.

**Validation Rules**

- Existing normal interval validation remains in force.
- Longer ranges and public dashboard do not require high-cadence density.
- Live cadence must not permanently mutate normal probe configuration.

## Latency Chart Range

Represents the selected time range for the node detail `线路检测` chart.

**Fields**

- `range_key`: one of `1m`, `30m`, `1h`, `12h`, `24h`, `1w`, or `30d`.
- `display_label`: Chinese UI label, such as `1 分钟` or `30 分钟`.
- `rendering_mode`: `live-realtime` for `1m`; `historical-range` for `30m` and longer.
- `window_seconds`: duration covered by the selected range.
- `target_point_count`: bounded number of display points used for historical chart readability.

**Relationships**

- Controls whether a Live Latency Session is created.
- Controls how persisted High-Cadence Latency Samples and normal background samples are queried and drawn.
- Shares the node detail page-wide range selection; the latency chart does not introduce an independent selector for this feature.

**Validation Rules**

- `30m` must remain present and selectable.
- Only `1m` may create or renew a Live Latency Session.
- Switching from `1m` to any historical range must end or stop renewing the live session.
- Historical ranges must load persisted results for the selected window instead of continuing to draw only the current one-minute observation.
- Historical ranges must use range-appropriate density/downsampling so high-cadence samples do not overcrowd the chart.

## Latency Rendering Mode

Represents the chart data policy derived from the selected Latency Chart Range.

**Fields**

- `mode`: `live-realtime` or `historical-range`.
- `source`: realtime PocketBase append stream for `live-realtime`; historical result query for `historical-range`.
- `starts_empty`: true for `live-realtime`, false for `historical-range` when historical data exists.
- `downsampling_bucket_seconds`: selected according to the historical range and chart width.

**Validation Rules**

- `live-realtime` starts from an empty current observation window and appends only fresh samples observed after entering `1 分钟`.
- `historical-range` may include previously persisted high-cadence samples but must aggregate or thin them for readability.
- Failure samples remain failures in both modes.
