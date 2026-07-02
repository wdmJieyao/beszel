# Data Model: Latency Realtime Window

## Latency Live Session

Represents a temporary, in-memory observation window that begins when an operator selects `1 分钟` on a node detail page.

**Fields**:

- `systemId`: The node being observed.
- `range`: Always `1m` for this session.
- `startedAt`: The moment the current browser page entered the `1 分钟` session, used for reset and chart-window semantics.
- `probeResultsByProbeId`: Fresh latency result points received from the active browser realtime subscription after `startedAt`, grouped by configured probe.

**Relationships**:

- Belongs to one node detail page instance.
- Contains zero or more fresh results for each configured latency line.
- Resets when the user leaves `1 分钟`, changes node, or re-enters `1 分钟`.

**Validation Rules**:

- Historical fetch results are excluded from plotted data in `1 分钟`.
- Realtime events received before the current session, including points from a previous `1 分钟` session, are excluded from plotted data.
- Points for other systems are ignored.
- Deleting or replacing a probe removes or updates that probe's session data without affecting unrelated probes.

**State Transitions**:

1. `Not Active`: User is viewing a non-`1 分钟` range.
2. `Started Empty`: User switches to `1 分钟`; configured lines are known, plotted results are empty.
3. `Collecting`: Realtime results received by the active browser session arrive and are appended.
4. `Reset`: User leaves and later re-enters `1 分钟`; a new `startedAt` is created and plotted results clear.

## Configured Latency Line

Represents an enabled latency probe assigned to the current node.

**Fields**:

- `probeId`: Stable probe identifier.
- `label`: User-visible probe name.
- `type`: Latency-capable probe type.
- `systemId`: Assigned node.
- `enabled`: Whether the line should be represented.

**Relationships**:

- Appears in the线路检测 legend even when it has no live-session points.
- May have zero, one, or many fresh results in a Latency Live Session.

**Validation Rules**:

- Only latency-capable configured probes are included in the combined线路检测 group.
- Lines not assigned to the current node are excluded.

## Fresh Latency Result

Represents one result received through the active browser realtime subscription during the active Latency Live Session.

**Fields**:

- `probeId`
- `systemId`
- `created`
- `success`
- `latencyMs`
- `failureCategory`
- `error`
- optional protocol-specific values such as packet loss or HTTP status

**Relationships**:

- Belongs to one Configured Latency Line.
- Is included in the active Latency Live Session only if it matches the current node and was received by the browser after the current session became active.

**Validation Rules**:

- Successful latency points can contribute to a plotted line.
- Failed points preserve failure status but must not cause stale successful history to be drawn.
- Duplicate realtime events with the same creation time replace older copies for the same probe.

## Historical Latency Result

Represents stored latency data used by non-`1 分钟` ranges.

**Fields**: Same shape as Fresh Latency Result.

**Relationships**:

- Available to longer historical ranges.
- Excluded from a new `1 分钟` Latency Live Session because it comes from historical fetches rather than the active browser realtime stream.

**Validation Rules**:

- Historical range behavior remains unchanged for longer ranges.
