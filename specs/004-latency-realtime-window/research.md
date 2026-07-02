# Research: Latency Realtime Window

## Decision: Treat node-detail `1 分钟` latency as a browser realtime observation session

**Rationale**: The existing CPU 使用率 `1 分钟` chart does not backfill stored history when the user enters that range. It subscribes to realtime metrics and, on the first realtime message, replaces prior chart data with the new point. The latency chart must follow the same user-facing semantic: selecting `1 分钟` starts empty and then draws from realtime result events received by the current browser session.

**Alternatives considered**:

- Show the most recent clock minute of stored latency results. Rejected because it immediately draws history and is the behavior the user rejected.
- Fetch records with `created >= switch time`. Rejected because the clarified requirement is to mirror CPU 使用率 by using browser-received realtime events after entering `1 分钟`, not a historical query window.
- Show stored results but visually fade or mark them. Rejected because the acceptance criteria require no pre-switch plotted line.
- Keep public and detail latency charts identical. Rejected because public dashboard charts have a separate historical dashboard purpose and the spec explicitly scopes live behavior to node detail.

## Decision: Preserve configured line legends independently from plotted points

**Rationale**: A line can be configured and valid before it has fresh samples in the current live session. Hiding it until samples arrive makes the user think configured probes disappeared. The legend should represent configured lines; plotted geometry should represent fresh samples.

**Alternatives considered**:

- Hide a line until it has at least two points. Rejected because it causes the "three configured, only two visible" complaint.
- Render stale points for missing lines. Rejected because stale plotted history violates the live-session requirement.
- Render a placeholder line at zero. Rejected because it implies measured latency that did not happen.

## Decision: Split `1m` behavior from historical range behavior in the data hook

**Rationale**: The current hook path fetches assigned probes and historical results for all ranges. For `1m`, the hook should fetch probe definitions for legend/group structure but not use historical results for plotting. Realtime events received by the active browser subscription after entering `1m` become the data source.

**Alternatives considered**:

- Change the results endpoint to accept a session start parameter. Rejected because no API change is needed and endpoint filtering would still encourage history-based `1m` rendering.
- Store live-session state globally. Rejected because a session is local to the page interaction and should reset on range switches.
- Reuse public chart range behavior. Rejected because public charts default to 30-minute historical analysis.

## Decision: Use current node-detail chart-time domain semantics for visual consistency

**Rationale**: CPU and other node-detail charts derive ticks and domains from the current chart range and current time. Latency 1-minute live rendering should use the same mental model so the chart starts empty and then extends within a moving 1-minute window as fresh results arrive.

**Alternatives considered**:

- Domain from first latency point to latest latency point. Rejected because the chart would visually stretch a short session to the full width, making a newly started session look mature.
- Domain from oldest available stored latency point. Rejected because it draws pre-session history.
- Fixed domain from exact switch time to switch time plus one minute. Rejected because existing CPU behavior is a rolling current-time window, not a frozen future window.

## Decision: Keep backend scheduling and probe execution out of scope

**Rationale**: The defect is the display semantics of `1 分钟` after switching ranges. Existing probe execution may continue to supply samples. Backend scheduling should only be revisited if validation shows no fresh post-switch results arrive often enough to satisfy live plotting.

**Alternatives considered**:

- Always change probe scheduling in this feature. Rejected as unnecessary scope and higher operational risk.
- Move probe execution into the browser. Rejected because probes must be executed from assigned agents, not the viewer's browser.
