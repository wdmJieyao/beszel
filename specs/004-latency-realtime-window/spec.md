# Feature Specification: Latency Realtime Window

**Feature Branch**: `[004-latency-realtime-window]`

**Created**: 2026-07-01

**Status**: Draft

**Input**: User description: "还是不对 人家刚切到1分钟的时候都是空白的，你这刚切到1分钟的时候直接就不知道从哪里开始 完全跟我说的让你参考cpu的完全不一样，你再好好研究一下需求先"

## Clarifications

### Session 2026-07-01

- Q: For node detail 1 分钟线路检测, which source should define fresh points after switching? → A: Use only realtime events received by the browser after entering 1 分钟, matching CPU 使用率 behavior.

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Start One-Minute Latency View Empty (Priority: P1)

As a signed-in operator viewing a node detail page, I want the线路检测 chart to behave like the existing CPU 使用率 chart when I switch to the 1 分钟 range, so I can watch fresh latency measurements arrive from the moment I start observing instead of seeing a pre-filled historical line.

**Why this priority**: The user explicitly expects the 1 分钟 latency view to match the familiar real-time CPU chart. Showing historical latency immediately after switching makes the chart appear to start from an arbitrary point and undermines trust in the real-time view.

**Independent Test**: Open a node detail page with three configured latency lines, switch the page-wide range selector from 1 小时 to 1 分钟, and confirm the线路检测 chart initially has no plotted latency line while retaining chart structure and legends.

**Acceptance Scenarios**:

1. **Given** a node has existing latency history from before the current page interaction, **When** the operator switches the node detail page range to 1 分钟, **Then** the线路检测 chart starts with no plotted historical latency line from before that switch.
2. **Given** the线路检测 chart has just entered 1 分钟 mode, **When** no new latency result has arrived after the switch, **Then** the chart remains visually empty rather than drawing an old line.
3. **Given** the chart is in 1 分钟 mode, **When** new latency results arrive after the switch, **Then** the plotted line begins from those new results and extends forward over time.

---

### User Story 2 - Preserve All Configured Latency Lines (Priority: P1)

As an operator comparing multiple configured latency lines for a node, I want all configured latency lines to remain represented in the 1 分钟 chart, so I can confirm which lines have started reporting and which are still waiting for fresh data.

**Why this priority**: The current complaint includes seeing fewer plotted lines than configured. The operator must not lose track of configured lines simply because their latest result has not arrived at the same instant as the others.

**Independent Test**: Configure three latency lines for one node, switch to 1 分钟, and verify that the chart legend lists all three lines immediately while each line begins plotting only after fresh post-switch samples arrive.

**Acceptance Scenarios**:

1. **Given** three latency lines are configured and enabled for a node, **When** the operator switches to 1 分钟, **Then** the线路检测 legend displays all three lines immediately.
2. **Given** the three lines report at staggered times, **When** only one or two lines have fresh post-switch samples, **Then** the chart still shows all three legend entries and does not imply the missing line was removed.
3. **Given** each configured line has received enough fresh post-switch samples to form a line, **When** the chart refreshes, **Then** all three lines are plotted in the same 1 分钟 view.

---

### User Story 3 - Match CPU One-Minute Time Behavior (Priority: P2)

As an operator familiar with the CPU 使用率 chart, I want the线路检测 1 分钟 view to use the same real-time time-window behavior, so the two charts feel consistent when I compare resource usage and latency at the same moment.

**Why this priority**: Consistent time-window behavior reduces interpretation errors. A latency chart that uses a different 1 分钟 meaning from CPU creates confusion even if the data is technically correct.

**Independent Test**: On the same node detail page, switch to 1 分钟 and compare the CPU 使用率 and线路检测 charts. Confirm both begin from the current observation window and advance as fresh data arrives, rather than one showing history and the other starting live.

**Acceptance Scenarios**:

1. **Given** the node detail page is switched to 1 分钟, **When** the CPU 使用率 chart begins its live one-minute view, **Then** the线路检测 chart follows the same start-empty and append-fresh-results behavior.
2. **Given** fresh data arrives over time, **When** the operator keeps the page open, **Then** the线路检测 x-axis window advances in a way that matches the CPU 1 分钟 experience.
3. **Given** the operator switches away from 1 分钟 and later returns to 1 分钟, **When** the range switch occurs again, **Then** the线路检测 chart begins a new live observation window instead of reusing old points from the previous 1 分钟 session.

### Edge Cases

- If a latency line does not report during the current 1 分钟 observation session, its legend entry remains visible but no stale line is drawn.
- If only one fresh point exists for a latency line after switching to 1 分钟, the chart may show a minimal point state but MUST NOT draw a long historical segment from before the switch.
- If a probe reports later than the others, it appears when fresh data arrives rather than being hidden or removed from the configured line list.
- If the operator switches ranges repeatedly, each new 1 分钟 session resets the live plotted latency data for that session.
- If a fresh latency result fails, the chart preserves the existing failure/empty state behavior without drawing stale successful history.
- If no latency lines are configured, the existing empty state remains appropriate.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: On the node detail page, switching the page-wide chart range to 1 分钟 MUST start the线路检测 chart as a live observation window with no plotted latency points from before the switch moment.
- **FR-002**: In 1 分钟 mode, the线路检测 chart MUST append only realtime latency result events received by the browser after the current 1 分钟 session begins.
- **FR-003**: In 1 分钟 mode, the线路检测 chart MUST NOT backfill historical latency results simply because they occurred within the previous clock minute.
- **FR-004**: In 1 分钟 mode, switching away from and back to 1 分钟 MUST create a new live observation session for线路检测.
- **FR-005**: The 1 分钟线路检测 view MUST keep all configured and enabled latency lines visible in the legend immediately after switching, even before fresh samples arrive.
- **FR-006**: A configured latency line MUST begin plotting only after it receives fresh results for the current 1 分钟 session.
- **FR-007**: Once a configured latency line has enough fresh points in the current 1 分钟 session, it MUST render as a normal line in the shared线路检测 chart.
- **FR-008**: The线路检测 x-axis in 1 分钟 mode MUST communicate the same live observation window behavior as the CPU 使用率 1 分钟 chart.
- **FR-009**: The线路检测 chart MUST avoid drawing long line segments that visually imply data existed before the current 1 分钟 session.
- **FR-010**: Longer ranges such as 30 分钟, 1 小时, 12 小时, 24 小时, 1 周, and 30 天 MUST continue to show historical latency data for their selected windows.
- **FR-011**: Public dashboard latency charts MUST keep the existing historical-range behavior unless the public dashboard has an explicit 1 分钟 live mode in the same interaction pattern.
- **FR-012**: Changed behavior MUST be covered by focused unit tests.
- **FR-013**: If HTTP APIs are added or changed, they MUST be RESTful or explicitly justify an existing non-REST contract.

### Key Entities *(include if feature involves data)*

- **Latency Live Session**: A temporary observation window that begins when an operator selects 1 分钟 on a node detail page and contains only fresh latency results received after that selection.
- **Configured Latency Line**: An enabled latency check assigned to the current node, represented in the chart legend regardless of whether it has fresh results in the current live session.
- **Fresh Latency Result**: A latency result delivered through the browser realtime subscription after the current 1 分钟 live session begins.
- **Historical Latency Result**: A latency result recorded before the current 1 分钟 live session begins; excluded from the 1 分钟 live chart but available in longer historical ranges.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: In validation, 100% of node detail 1 分钟线路检测 sessions start without plotting pre-switch latency history.
- **SC-002**: In validation with three configured latency lines, the 1 分钟线路检测 legend shows all three lines immediately after the switch.
- **SC-003**: In validation with three configured latency lines, all three lines appear in the chart after each has received sufficient fresh post-switch samples.
- **SC-004**: In validation, no 1 分钟线路检测 line segment begins before the selected 1 分钟 session start.
- **SC-005**: In validation, switching from 1 分钟 to another range and back to 1 分钟 resets the live线路检测 plotted points within one user interaction cycle.
- **SC-006**: Operators comparing CPU 使用率 and线路检测 in 1 分钟 mode can identify both as live observation views without needing additional explanation.

## Assumptions

- The primary scope is the signed-in node detail page because the user explicitly compared线路检测 with the node detail CPU 使用率 chart.
- The public dashboard keeps its existing default 30-minute historical behavior unless a separate public live 1 分钟 interaction is requested later.
- "空白" means no plotted historical latency line; the chart title, axes, legend, and empty or waiting state may still be visible.
- A line may need more than one fresh successful realtime latency event before it visually appears as a connected line.
- Existing probe configuration, probe execution, and longer-range history storage remain in scope only as dependencies; this feature changes the 1 分钟 display semantics.
