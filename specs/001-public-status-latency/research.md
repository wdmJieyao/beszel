# Research: Public Status Page and Network Probe Trends

## Decision: Public data served by custom sanitized REST handlers

**Rationale**: PocketBase collection rules are broad record-level gates, while
the public page must expose a narrow field-level subset. Custom handlers can
compose system status, visibility settings, and probe summaries without
opening private fields such as host, port, user relationships, tokens, or admin
controls.

**Alternatives considered**:

- Public PocketBase collection rules on `systems`: rejected because field-level
  privacy is too easy to get wrong.
- Static exported JSON: rejected because visibility and probe state must update
  after admin changes.

## Decision: Add per-system public visibility setting, private by default

**Rationale**: Existing systems should not become public during upgrade. An
explicit opt-in setting gives admins control and supports acceptance tests that
prove private systems never appear anonymously.

**Alternatives considered**:

- Global public dashboard switch: rejected because it cannot expose only part
  of the fleet.
- Public by default: rejected as a security regression.

## Decision: Persist network probes and probe results as PocketBase collections

**Rationale**: The hub already stores time-series records in PocketBase and the
frontend already reads chart data from collections/custom handlers. Persisting
probe definitions and results enables trend windows, public summaries, and
admin management without adding a separate database.

**Alternatives considered**:

- In-memory probe results: rejected because history and restart persistence are
  required.
- External time-series database: rejected because it adds an unsupported stack
  for this feature.

## Decision: Execution nodes run assigned probes while the default UI hides agent binding

**Rationale**: The feature is meant to show each node's own network path to a
target. Hub-side probing would only measure hub connectivity and would not
represent VPS-specific latency. The user-facing setup should describe this as
线路, 观测点, and 执行节点; explicit agent binding belongs in advanced settings
because the default flow should not read like an internal transport tool.

**Alternatives considered**:

- Hub-only probing: rejected because it gives the wrong measurement.
- Both hub and agent probing in v1: rejected to keep scope focused; hub probing
  can be a later extension if needed.
- Requiring agent selection in the primary setup flow: rejected because it
  exposes implementation details too early and does not match the desired
  product experience.

## Decision: Use existing hub-to-agent websocket/CBOR request pattern

**Rationale**: The project already sends typed requests from hub to agent for
data, SMART, containers, and systemd. Adding a `RunNetworkProbe` action keeps
authentication, request IDs, timeout handling, and connection lifecycle aligned
with existing code.

**Alternatives considered**:

- Agent pushes probe results on its own schedule: rejected for v1 because it
  requires a new scheduler and backpressure model on every agent.
- Separate HTTP listener on agents: rejected because agents already maintain
  the hub websocket connection.

## Decision: Support TCPing, ICMP Ping, and HTTP GET probe types

**Rationale**: These types cover port reachability, network latency, packet
loss/timeout, and service availability. This matches the clarified requirement
and common status-panel practice.

**Alternatives considered**:

- TCPing only: rejected because it cannot represent ICMP packet loss or simple
  HTTP service status.
- TCPing plus ICMP only: rejected because HTTP GET is useful for service
  reachability checks and was explicitly selected.

## Decision: Probe public visibility defaults to visible for public execution nodes, with per-probe opt-out

**Rationale**: The anonymous home dashboard should show the new network charts by default
while still allowing admins to hide internal or sensitive targets.

**Alternatives considered**:

- All probes hidden by default: rejected because it weakens the requested
  public page behavior.
- Global public probe switch only: rejected because individual targets may be
  sensitive.

## Decision: Reuse existing frontend chart primitives

**Rationale**: `internal/site` already has Recharts wrappers, chart time
selection, and cached data-fetching patterns. Reusing these keeps the UI
consistent and satisfies the constitution's stack requirement.

**Alternatives considered**:

- New charting library: rejected as unnecessary stack expansion.
- Static tables only: rejected because latency trend visualization is a core
  requirement.

## Decision: Add or decide focused frontend test harness during task planning

**Rationale**: The repo has Go tests and Biome checks, but no obvious frontend
unit-test script. The constitution requires unit coverage for behavior changes,
so tasks must either add a minimal frontend test harness for pure functions and
visibility filtering or explicitly justify a temporary frontend test gap.

**Alternatives considered**:

- Rely only on manual UI validation: rejected for filtering/security behavior.
- Add broad browser E2E immediately: deferred to tasks because it may be larger
  than the minimal frontend unit coverage needed for the first slice.
