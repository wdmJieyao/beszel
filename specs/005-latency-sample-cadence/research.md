# Research: Latency Sample Cadence

## Decision: Use active live sessions rather than lowering configured probe intervals

**Decision**: Add a short-lived live observation session that applies an approximately 1-second cadence only while a signed-in operator is viewing a node detail `1 分钟` latency chart.

**Rationale**: Current configured probes have a minimum interval of 10 seconds and are scheduled by a 5-second hub ticker. Lowering all configured intervals to 1 second would permanently increase load, affect public/historical data collection, and change administrator intent. The user complaint is tied to the active `1 分钟` live view, so a temporary session provides dense data only when it is useful.

**Alternatives considered**:

- Lower `minProbeIntervalSeconds` and the PocketBase field minimum from 10 to 2. Rejected because it changes global background behavior and can overload nodes with many configured probes.
- Draw smoother lines by interpolation in the browser. Rejected because it fabricates continuity without real measurements and conflicts with failure accuracy.
- Keep current scheduler and only increase chart update frequency. Rejected because no new probe results exist to draw.

## Decision: Hub owns cadence; agent remains request/response executor

**Decision**: Keep the existing `RunNetworkProbe` agent action unchanged. The hub decides when to request probes and sends one request per configured latency assignment during live cadence ticks.

**Rationale**: The current architecture already has hub-to-agent request/response semantics, timeout handling, and result persistence. Moving timers into the agent would require new configuration distribution, cancellation, and compatibility handling. Hub-side cadence also lets the browser's active live session directly control when higher frequency is needed.

**Alternatives considered**:

- Add a new agent-side streaming probe mode. Rejected as a larger protocol change with harder cancellation and version compatibility.
- Let the frontend call the agent directly. Rejected because agents are not exposed to browsers and hub auth/visibility must remain authoritative.

## Decision: Coalesce active sessions per system

**Decision**: Track live sessions by session id but execute probes once per system per cadence tick when one or more unexpired sessions exist for that system.

**Rationale**: Multiple browser tabs or users may watch the same node. Running duplicate probes per viewer would multiply load and store duplicate samples. Coalescing keeps waveform density stable while still allowing each client to maintain its own frontend live observation window.

**Alternatives considered**:

- Execute probes per browser session. Rejected because it scales poorly and creates duplicate result records.
- Allow only one live viewer per system. Rejected because it would be surprising and unnecessary.

## Decision: Use RESTful live-session endpoints with TTL heartbeat

**Decision**: Add authenticated REST endpoints to create, renew, and end live latency sessions. Sessions have a short TTL; the frontend renews while `1 分钟` remains active, and the hub expires abandoned sessions automatically.

**Rationale**: The hub cannot infer the selected chart range from PocketBase result subscriptions. An explicit resource gives the hub the missing intent while remaining compatible with normal REST API rules. TTL protects against lost browser unload events.

**Alternatives considered**:

- Infer active live mode from `network_probe_results` subscriptions. Rejected because subscriptions do not encode page range and could include other consumers.
- Use only a long-running websocket command. Rejected because existing frontend API patterns use REST plus PocketBase realtime, and REST is easier to authorize and test.

## Decision: Reuse persisted result stream for realtime delivery

**Decision**: Persist high-cadence probe results to the existing `network_probe_results` collection and let the existing PocketBase realtime subscription deliver them to the chart.

**Rationale**: The 004 implementation already listens to realtime `network_probe_results` and filters live data by browser session. Reusing that stream avoids adding a second data plane and keeps failure states consistent.

**Alternatives considered**:

- Send custom websocket messages directly to the browser. Rejected because it duplicates authorization and chart integration.
- Keep high-cadence samples in memory only. Rejected because the current browser receives probe updates through persisted collection realtime events.

## Decision: Use separate rendering modes for `1 分钟` and historical ranges

**Decision**: Treat node detail latency range selection as an explicit rendering-mode switch. `1 分钟` uses a live realtime mode that starts with an empty observation window and appends fresh samples from the active live session. `30 分钟` and longer ranges use a historical range mode that loads persisted probe results for the selected time window and applies range-appropriate density/downsampling.

**Rationale**: The user explicitly wants `1 分钟` to behave like the CPU live chart, but also wants `30 分钟` and wider ranges to remain available and readable. Reusing the `1 分钟` live point model for wider ranges makes the chart crowded, causes range options to drift, and visually conflicts with the rest of the detail page.

**Alternatives considered**:

- Use the live session model for all ranges. Rejected because it over-samples wider ranges and makes historical charts unreadable.
- Remove `30 分钟` to simplify the selector. Rejected because `30 分钟` is a required range.
- Show all raw historical points without density control. Rejected because high-cadence samples can create overcrowded paths when viewed over 30 minutes or more.

## Decision: Keep `30 分钟` as the first historical latency range

**Decision**: The node detail `线路检测` range selector must include `1 分钟`, `30 分钟`, `1 小时`, `12 小时`, `24 小时`, `1 周`, and `30 天` where the surrounding page supports those ranges. Only `1 分钟` maps to live realtime mode; `30 分钟` is the shortest historical mode.

**Rationale**: This aligns with the user's latest clarification and gives operators an immediate non-live comparison window without forcing a jump to `1 小时`.

**Alternatives considered**:

- Use only the inherited page-wide range options without latency-specific validation. Rejected because previous changes already lost `30 分钟`, so this requirement needs an explicit contract and test.
- Add a separate latency-only selector. Rejected because the user asked for the latency chart to follow the node detail page-wide range rather than having its own selector.

## Decision: Guard live cadence with timeout and line-type constraints

**Decision**: Live cadence applies only to enabled latency-capable lines assigned to the viewed node, primarily `tcping` and `icmp_ping`. A probe whose timeout cannot fit within the cadence should use a bounded live timeout or be skipped with a safe failure rather than blocking the cadence loop.

**Rationale**: A 1-second cadence cannot remain smooth if each measurement is allowed to block for 5 seconds. Boundaries are needed to keep the live worker responsive and avoid duplicate stale samples.

**Alternatives considered**:

- Run all probe types including HTTP GET in high-cadence mode. Rejected because the latency chart groups only latency-capable lines and HTTP checks can be heavier.
- Allow overlapping runs for the same assignment. Rejected because overlapping samples can duplicate work and distort the chart.
