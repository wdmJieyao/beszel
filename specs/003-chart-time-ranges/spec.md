# Feature Specification: Public Chart Time Ranges

**Feature Branch**: `[003-chart-time-ranges]`

**Created**: 2026-06-30

**Status**: Draft

**Input**: User description: "线路延迟只展示最近30分钟之内的就好了，不用太长。横坐标还是要展示以下时分秒即可，维度可以自行计算，但是我们更新的时候还是20S刷一次。新增的CPU 使用率、内存使用率、磁盘使用率展示时间也是一样的可以做成默认展示30分钟之内的，横坐标单位也是自己计算一下。对于我们新增的所有折线图，都新增如节点详情页面-CPU 使用率-更多里面，可以筛选时间范围，动态变化横坐标，你先参考研究以下"

## Clarifications

### Session 2026-06-30

- Q: Which time range options should the new public chart range selector offer? → A: Reuse the existing node detail chart range style and add 30 minutes as the default option.

## User Scenarios & Testing *(mandatory)*

### User Story 1 - View Recent Public Trends (Priority: P1)

As an anonymous public dashboard visitor, I want latency and resource charts to focus on the most recent 30 minutes by default, so I can quickly understand current node behavior without scanning old data.

**Why this priority**: The public dashboard's primary value is current status visibility. Long default ranges make recent movement harder to read and clutter the chart.

**Independent Test**: Open the public dashboard for a node with latency and resource history, confirm each new chart defaults to the most recent 30 minutes, shows only points from that window, and refreshes the visible data without a page reload.

**Acceptance Scenarios**:

1. **Given** a public node has more than 30 minutes of latency results, **When** a visitor opens the public dashboard, **Then** the latency chart displays only the latest 30-minute window by default.
2. **Given** a public node has more than 30 minutes of CPU, memory, and disk history, **When** a visitor opens the resource trend dialog, **Then** each resource chart displays only the latest 30-minute window by default.
3. **Given** new measurements arrive while the public dashboard remains open, **When** the dashboard refreshes, **Then** the charts update their visible 30-minute window while preserving the selected range.

---

### User Story 2 - Read Time from Chart Axes (Priority: P2)

As a dashboard visitor, I want chart horizontal axes to show readable hour-minute-second labels, so I can understand when visible latency and resource changes happened.

**Why this priority**: A recent-only chart still needs time context. Time labels make trend interpretation possible without exposing sensitive target details.

**Independent Test**: Open the public latency chart and resource trend dialog across default and alternate ranges, confirm horizontal labels use hour-minute-second format and adjust density so labels remain readable.

**Acceptance Scenarios**:

1. **Given** a chart is showing the default 30-minute range, **When** the chart renders, **Then** the horizontal axis displays readable time labels in hour-minute-second format.
2. **Given** a user changes the visible time range, **When** the chart redraws, **Then** the horizontal axis recalculates label spacing and remains readable.
3. **Given** the visible range has sparse data, **When** the chart renders, **Then** the axis still communicates the time window without overlapping labels.

---

### User Story 3 - Change Chart Time Range (Priority: P3)

As a dashboard visitor, I want the newly added public latency and resource charts to offer the same kind of time range filtering pattern used by existing node detail charts, so I can inspect recent or longer-term trends when needed.

**Why this priority**: Defaulting to 30 minutes solves the main readability problem, while range selection provides continuity with the existing product behavior.

**Independent Test**: On the public dashboard, use the chart range control for latency and resource charts, verify that the selected range changes the plotted data and time-axis labels for each chart without affecting unrelated nodes.

**Acceptance Scenarios**:

1. **Given** a public latency chart is visible, **When** a visitor selects a different time range, **Then** the chart shows data for that selected range and updates its horizontal axis.
2. **Given** the resource trend dialog is open, **When** a visitor selects a different time range, **Then** all visible CPU, memory, and disk charts use the selected range consistently.
3. **Given** a visitor changes a chart range for one node, **When** another node card is viewed, **Then** the other node keeps its own default or selected range independently.

### Edge Cases

- If fewer than 30 minutes of data exists, charts show all available data in the default window and clearly preserve the current time scale.
- If a selected range has no successful latency points, the chart shows the existing empty or failure state without breaking the resource charts.
- If refresh fails, the current chart data and selected range remain visible until the next successful refresh.
- If time labels would overlap in a narrow viewport, label density is reduced rather than wrapping or colliding.
- If the latest data point is newer than stored history, the chart includes the latest visible metric so top-line values and trend endings remain consistent.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: Public latency charts MUST default to a 30-minute visible time range.
- **FR-002**: Public CPU, memory, and disk trend charts MUST default to a 30-minute visible time range.
- **FR-003**: Public latency charts MUST refresh visible data every 20 seconds while preserving the selected time range.
- **FR-004**: Public CPU, memory, and disk trend charts MUST refresh visible data every 20 seconds while preserving the selected time range.
- **FR-005**: Newly added public charts MUST show horizontal time labels in hour-minute-second format.
- **FR-006**: Newly added public charts MUST automatically adjust horizontal label density so labels do not wrap or overlap in normal desktop and mobile dashboard widths.
- **FR-007**: Newly added public charts MUST provide a time range selector consistent with the existing node detail chart range selection pattern, using the same existing range style and adding 30 minutes as the default option.
- **FR-008**: The default selected range for newly added public charts MUST be 30 minutes.
- **FR-009**: Changing the selected range MUST update the plotted data and horizontal axis without requiring a page reload.
- **FR-010**: Changing the selected range for one node MUST NOT unexpectedly change another node's selected range.
- **FR-011**: Resource trend charts MUST keep the latest displayed values consistent with the node's current public CPU, memory, and disk summary values when those values are available.
- **FR-012**: Chart range changes MUST avoid exposing hidden probe target addresses or other non-public target metadata.
- **FR-013**: Changed behavior MUST be covered by focused unit tests.
- **FR-014**: If HTTP APIs are added or changed, they MUST be RESTful or explicitly justify an existing non-REST contract.

### API Contracts *(include if HTTP API changes)*

- **Resources**: Public dashboard status data for systems, public latency probe series, and public resource metric history.
- **Methods**: Any new or changed public data retrieval behavior must use read-only retrieval semantics and must not require authentication for public dashboard data.
- **Status Codes**: Successful public chart data retrieval returns a success response; invalid range requests return a clear client error; unavailable internal data returns a user-safe response that does not expose private details.
- **Schemas**: Responses must include enough timestamped points to render the selected public chart range, while retaining existing public sanitization rules for probe names and targets.
- **Compatibility**: Existing public dashboard clients that do not request a specific range continue to receive data suitable for the default 30-minute view.

### Key Entities *(include if feature involves data)*

- **Chart Time Range**: The visitor-selected visible duration for a chart, defaulting to 30 minutes and used to filter plotted points and axis labels.
- **Public Latency Series**: Sanitized timestamped latency results for a public node and probe line.
- **Public Resource Series**: Timestamped CPU, memory, or disk usage values for a public node.
- **Chart Axis Label**: A generated time label that communicates visible timestamps without clutter or overlap.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: 100% of newly added public latency and resource charts open with a 30-minute default range.
- **SC-002**: Public chart data refreshes within 20 seconds of each scheduled refresh opportunity when the service is reachable.
- **SC-003**: In desktop and mobile validation, horizontal time labels remain readable without wrapping for default and selected ranges.
- **SC-004**: A visitor can change the time range for latency and resource charts in under 3 interactions per chart area.
- **SC-005**: In validation with more than 30 minutes of history, default charts exclude points older than the selected 30-minute window.
- **SC-006**: Existing public sanitization behavior is preserved: public charts do not reveal hidden probe target addresses.

## Assumptions

- The selectable range options follow the existing node detail chart range style: 1 minute, 1 hour, 12 hours, 24 hours, 1 week, and 30 days where applicable, with 30 minutes added as the default option for these new public charts.
- Range selection state is scoped to the public dashboard view and does not need to persist across browser sessions unless the existing chart pattern already provides persistence.
- Default public chart refresh for this feature is 20 seconds, even if other public summary fields use a different refresh cadence.
- The public dashboard remains available without login, and range controls do not add new permission requirements.
- Existing hidden target sanitization remains mandatory for public latency charts.
