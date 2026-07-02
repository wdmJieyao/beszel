# Feature Specification: Latency Sample Cadence

**Feature Branch**: `[005-latency-sample-cadence]`

**Created**: 2026-07-02

**Status**: Draft

**Input**: User description: "还是不对，切换成1分钟之后，三个测速节点之间还是断断续续的，甚至连线有的都是断开的，我觉得当时间切换到1分钟之后，节点测速的探测时间你要缩短一下，调整为2s一次，这样才能画出连续的波形图。"

## Clarifications

### Session 2026-07-02

- Q: Should longer latency chart ranges use the live one-minute drawing model or range-appropriate historical drawing? → A: `1分钟` uses realtime high-cadence drawing; `30分钟` and longer ranges use historical range drawing with range-appropriate density/downsampling.

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Smooth One-Minute Latency Lines (Priority: P1)

As an operator watching a node detail page in the `1 分钟` range, I want each configured latency line to receive fresh measurements often enough to form a continuous-looking waveform, so the chart behaves like a live monitoring view rather than a few disconnected points.

**Why this priority**: The current `1 分钟` chart can start correctly but still fails the practical monitoring goal because sparse measurements make the three configured lines appear broken or jumpy.

**Independent Test**: Open a node detail page with three enabled latency lines, switch the page-wide range to `1 分钟`, observe the chart for 60 seconds, and confirm each available line receives frequent fresh samples and renders as connected waveform segments.

**Acceptance Scenarios**:

1. **Given** a node has three enabled latency lines, **When** the operator switches the node detail page to `1 分钟`, **Then** fresh measurements for those lines are requested at an approximately 1-second cadence while the live observation remains active.
2. **Given** the `1 分钟` view has been open for at least 20 seconds, **When** the configured targets are reachable, **Then** each enabled latency line has enough fresh samples to render connected line segments instead of only isolated points.
3. **Given** the `1 分钟` view has been open for 60 seconds, **When** the configured targets remain reachable, **Then** the chart shows a continuous-looking waveform for each configured latency line.

---

### User Story 2 - Preserve Existing Longer-Range Behavior (Priority: P2)

As an operator switching between time ranges, I want the higher sampling cadence to apply only to the live `1 分钟` observation, so longer historical ranges and public dashboard views remain stable and easy to compare.

**Why this priority**: Increasing sample frequency globally would create unnecessary noise and load. The complaint is specifically about the live `1 分钟` waveform.

**Independent Test**: Compare the node detail chart in `1 分钟`, `30 分钟`, and `1 小时`; verify `1 分钟` produces dense fresh samples while `30 分钟` and longer ranges remain available and present historical data at range-appropriate visual density.

**Acceptance Scenarios**:

1. **Given** the operator opens the range selector, **When** latency charts are available on the node detail page, **Then** `30 分钟` remains an available selectable range.
2. **Given** the operator switches away from `1 分钟` to `30 分钟` or a longer range, **When** the longer range loads, **Then** it continues to show historical latency data without adopting the 1-second live cadence or one-minute visual density as its display expectation.
3. **Given** the operator selects `1 小时`, `12 小时`, `24 小时`, `1 周`, or `30 天`, **When** the latency chart renders, **Then** the waveform is drawn with range-appropriate density/downsampling so the chart remains readable rather than overcrowded.
4. **Given** the public dashboard is viewed, **When** no explicit live `1 分钟` interaction is active there, **Then** public latency charts keep their existing historical-range behavior.

---

### User Story 3 - Handle Failures Without Misleading Lines (Priority: P3)

As an operator diagnosing latency, I want failed or delayed measurements to be represented honestly, so the chart does not fabricate continuous latency values when a target is unavailable.

**Why this priority**: A smooth waveform is useful only if it reflects real measurements. Failure handling must remain accurate.

**Independent Test**: Use one reachable target and one failing target during `1 分钟` observation; verify the reachable target produces dense connected samples while the failing target shows an appropriate failure or gap state.

**Acceptance Scenarios**:

1. **Given** a configured latency target fails during `1 分钟` observation, **When** fresh failed measurements are recorded, **Then** the chart preserves the failure indication and does not draw a fake successful latency segment.
2. **Given** a measurement takes longer than the requested cadence, **When** the next cycle occurs, **Then** the system avoids presenting duplicated or stale values as new measurements.

### Edge Cases

- If only one of three configured latency lines is reachable, the reachable line should become smooth while unreachable lines retain failure or empty state behavior.
- If a target intermittently fails, successful samples should connect only across valid measured points that do not misrepresent failures.
- If the operator switches repeatedly between `1 分钟` and longer ranges, each `1 分钟` session should resume the higher live cadence and the chart should still start from the current live observation window.
- If many latency lines are configured for one node, the live cadence should remain bounded enough that normal node monitoring remains responsive.
- If no latency lines are configured, the existing empty state remains appropriate.
- If the node is offline or cannot run measurements, the chart should show the existing unavailable/failure state instead of a stale smooth line.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: When a node detail page is in `1 分钟` live latency observation, the system MUST produce fresh latency measurements for enabled lines at an approximately 1-second cadence.
- **FR-002**: The `1 分钟` latency chart MUST use the higher-cadence fresh measurements to render connected waveform segments for reachable targets.
- **FR-003**: The `1 分钟` latency chart MUST continue to start from the current live observation window and MUST NOT backfill pre-switch historical latency points.
- **FR-004**: All configured and enabled latency lines MUST remain visible in the `1 分钟` legend even if some lines have not yet produced fresh high-cadence samples.
- **FR-005**: Switching away from `1 分钟` MUST stop treating the chart as a high-cadence live waveform view for that page interaction.
- **FR-006**: The node detail latency range selector MUST include `30 分钟` alongside the other supported ranges.
- **FR-007**: Longer ranges such as `30 分钟`, `1 小时`, `12 小时`, `24 小时`, `1 周`, and `30 天` MUST continue to show historical latency data without requiring 1-second visual density.
- **FR-008**: Longer latency ranges MUST use range-appropriate historical drawing density or downsampling so waveforms remain readable and do not appear as overcrowded one-minute realtime traces.
- **FR-009**: Public dashboard latency charts MUST keep their existing historical behavior unless a separate public live `1 分钟` interaction is explicitly added later.
- **FR-010**: Failed measurements MUST remain distinguishable from successful latency measurements and MUST NOT be converted into artificial successful line segments.
- **FR-011**: The higher-cadence behavior MUST avoid duplicating stale measurements as if they were new fresh samples.
- **FR-012**: The chart MUST keep Chinese labels, legends, empty states, and failure text for latency views.
- **FR-013**: Changed behavior MUST be covered by focused unit tests.
- **FR-014**: If HTTP APIs are added or changed, they MUST be RESTful or explicitly justify an existing non-REST contract.

### Key Entities *(include if feature involves data)*

- **Live Latency Cadence**: The expected measurement frequency while a node detail latency chart is actively observed in the `1 分钟` range.
- **High-Cadence Latency Sample**: A fresh measurement produced during the active `1 分钟` observation window and eligible to appear in the live waveform.
- **Reachable Latency Line**: A configured latency line whose target returns successful fresh measurements during the observation window.
- **Failed Latency Measurement**: A fresh measurement attempt that did not produce a successful latency value and should be shown as failure or gap state.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: In validation with three reachable latency lines, each line receives at least 40 fresh successful samples during a 60-second `1 分钟` observation.
- **SC-002**: In validation with three reachable latency lines, at least 90% of adjacent successful samples for each line are no more than 4 seconds apart during active `1 分钟` observation.
- **SC-003**: After 20 seconds in `1 分钟`, each reachable configured line renders as a connected segment rather than only isolated points.
- **SC-004**: Switching from `1 分钟` to `30 分钟` or `1 小时` still shows historical latency data for the selected range, and `30 分钟` remains selectable.
- **SC-005**: In `30 分钟` and longer ranges, latency waveforms render with range-appropriate density/downsampling and do not visually retain the overcrowded one-minute realtime plotting model.
- **SC-006**: Failed targets do not produce artificial successful latency segments in validation.
- **SC-007**: Operators can visually distinguish all configured latency lines and see a smoother live waveform in `1 分钟` without additional explanation.

## Assumptions

- The higher cadence is needed for the signed-in node detail page `1 分钟` live view, matching the user's complaint about that interaction.
- Longer historical ranges and public dashboard views remain outside the high-cadence live sampling change unless later requested, but their existing selectable ranges and historical readability must be preserved.
- The 1-second cadence is an active live-observation target with reasonable tolerance for network delays, target failures, and node load.
- Existing probe configuration and assignment concepts remain the source of which latency lines should be measured.
- Existing project architecture and quality requirements remain in force for any implementation work needed to satisfy this behavior.
