# UI Contract: Node Detail Latency `1 分钟` Live Window

## Scope

This contract covers the signed-in node detail page's线路检测 chart when the page-wide chart range selector is set to `1 分钟`.

Public dashboard latency charts and longer historical node-detail ranges are outside this contract except for the requirement that they do not regress.

## Inputs

- Current node identifier.
- Current page-wide chart range.
- Set of enabled latency probes assigned to the current node.
- Realtime latency result events for `network_probe_results` received by the active browser session.
- Existing historical latency result fetches for non-`1 分钟` ranges.

## Behavior

### Entering `1 分钟`

When the user changes the page-wide range to `1 分钟`:

1. A new Latency Live Session begins.
2. The session records that the browser has entered a new active realtime observation session.
3. The chart keeps all configured latency lines in the legend.
4. The chart does not plot any pre-session historical latency points.
5. The chart may show an empty or waiting state inside the plotting area until fresh points arrive.

### Receiving Fresh Results

When a realtime latency result arrives:

1. Ignore it if it belongs to a different system.
2. Ignore it if there is no active `1 分钟` session for the current node.
3. Attach it to the matching configured latency line.
4. Plot the line only from fresh session points.
5. Preserve all other configured lines in the legend even if they still have no points.

### Leaving and Re-entering `1 分钟`

When the user switches away from `1 分钟`, the current live session is no longer active.

When the user later switches back to `1 分钟`, a new live session starts empty; it must not reuse plotted points from the previous live session.

### Longer Ranges

When the selected range is not `1 分钟`, the chart may use historical stored latency results for the selected range. Longer ranges must not inherit the "start empty" live-session behavior.

## Acceptance Checks

- Switching from `1 小时` to `1 分钟` with existing latency history produces no pre-switch plotted line.
- With three configured lines, the legend shows three line labels immediately after switching.
- If fresh results arrive for only two of three lines, two lines may plot and the third remains represented in the legend.
- After all three lines receive sufficient fresh results, the chart plots all three lines.
- Switching away and back resets plotted live-session data.
- Immediately after switching back to `1 分钟`, no previously received realtime points from the prior `1 分钟` session are reused.

## Non-Goals

- No new HTTP endpoint is required.
- No probe target metadata is exposed.
- No change is required to agent-side probe execution.
