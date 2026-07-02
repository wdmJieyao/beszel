# Feature Specification: Public Status Page and Network Probe Trends

**Feature Branch**: `001-public-status-latency`

**Created**: 2026-06-28

**Status**: Draft

**Input**: User description: "添加如 nezha 面板一样的公开页，无需登录也能看到部分 VPS 的状态；增加三网延迟检查的功能，可以检测网络延迟走势。"

## Clarifications

### Session 2026-06-28

- Q: What metrics should the public page expose by default? → A: Minimal public metrics: name, online state, data freshness, CPU/memory/disk summary percentages, plus the new network latency trend chart.
- Q: Should network checks be fixed to three built-in routes? → A: No. "Three-network" is shorthand for configurable network probes; administrators define probe names, targets, visibility, and coverage without built-in targets or a fixed count.
- Q: Which network probe types should be supported? → A: Support TCPing, ICMP Ping, and HTTP GET probes; TCPing and ICMP Ping provide latency trend data, while HTTP GET provides service reachability and response status.
- Q: Where should network probes be executed from? → A: Probes are executed by the selected agent nodes themselves, not centrally by the hub, so each result represents that node's own network path to the target.
- Q: How should network probe charts appear on the public page? → A: Network probes are visible on the public page by default when their executing agent node is public, and administrators can manually disable public display per probe.
- Q: What should anonymous visitors see at the home route? → A: Anonymous visitors opening `/` should land directly on the public dashboard; there is no separate public-link landing page.
- Q: What language should the feature UI use? → A: The feature UI should use Chinese product language such as 公共看板, 线路, 观测点, and 执行节点.
- Q: How should the settings experience feel? → A: Settings must expose real controls and preview state for public visibility and probe configuration; empty static pages are not acceptable.

## User Scenarios & Testing *(mandatory)*

### User Story 1 - View Public VPS Status (Priority: P1)

A visitor opens `/` without signing in and sees the public dashboard with only
the VPS systems that an administrator has chosen to expose.

**Why this priority**: This is the primary user-facing capability and must be
safe before any additional public monitoring data is exposed.

**Independent Test**: Configure one system as public and another as private,
open `/` in a signed-out session, and verify only the public system appears
with allowed status fields.

**Acceptance Scenarios**:

1. **Given** at least one system is marked public, **When** an anonymous visitor opens `/`, **Then** the visitor sees the public system's display name, online/offline state, data freshness, CPU/memory/disk summary percentages, and the public network latency trend chart when latency display is enabled.
2. **Given** a system is not marked public, **When** an anonymous visitor opens `/`, **Then** that system and its metrics are not shown.
3. **Given** no systems are marked public, **When** an anonymous visitor opens `/`, **Then** the page shows an empty state rather than exposing private inventory.
4. **Given** an administrator changes a system's public visibility, **When** the anonymous home page is refreshed, **Then** the visibility change is reflected.

---

### User Story 2 - Manage Public Visibility and Display Scope (Priority: P2)

An administrator selects which VPS systems appear publicly and controls the
safe display scope so visitors can see service health without seeing private
operational details.

**Why this priority**: Public access requires an explicit administrative
control path to prevent accidental data exposure.

**Independent Test**: As an administrator, enable and disable public visibility
for specific systems and verify the anonymous home page follows those
settings.

**Acceptance Scenarios**:

1. **Given** an administrator views public dashboard settings, **When** public visibility is enabled for a system, **Then** that system becomes eligible for the anonymous home page.
2. **Given** a public system is later disabled from public visibility, **When** an anonymous visitor reloads `/`, **Then** the system no longer appears.
3. **Given** a public status field is considered sensitive, **When** the anonymous home page is rendered, **Then** that field is omitted for anonymous visitors.

---

### User Story 3 - View Configurable Network Probe Trends (Priority: P3)

An authenticated user views configurable network probe checks expressed as
线路/观测点 and can understand each execution node's current latency, packet
loss, reachability, HTTP response status, and historical trend changes to
selected targets.

**Why this priority**: Latency trends add diagnostic value after the public
page and visibility controls are established.

**Independent Test**: Configure multiple network probes, allow multiple check
intervals to complete, and verify the trend view shows recent and historical
reachability and latency values from each configured execution node to each
configured probe target.

**Acceptance Scenarios**:

1. **Given** network probes are assigned to configured execution nodes, **When** checks complete over time, **Then** the user can view each execution node's current reachability, latency trends for TCPing and ICMP Ping probes, and response status for HTTP GET probes.
2. **Given** a probe target is unreachable, **When** a check runs, **Then** the trend records the failure state without blocking other configured probes.
3. **Given** historical latency exists, **When** the user changes the viewed time range, **Then** the chart and summary update to the selected range.
4. **Given** a system is public and a network probe has not been manually hidden from the public page, **When** an anonymous visitor views that system, **Then** only public-safe probe summaries and latency trends are shown.
5. **Given** an administrator manually disables public display for a network probe, **When** an anonymous visitor views a public system that executes that probe, **Then** that probe's summary and chart are not shown.

### Edge Cases

- The anonymous home page is requested while the hub has no available systems.
- A public system stops reporting data or reports stale data.
- A system changes from public to private while an anonymous visitor has the page open.
- One or more network probes fail while others continue to report.
- Latency values spike, time out, or become temporarily unavailable.
- Historical latency data is missing for a newly enabled system.
- An administrator configures many probes or changes which systems a probe covers.
- An execution node is offline or unable to execute an assigned probe.
- The anonymous home page receives heavy anonymous traffic.
- Public display names or labels contain user-entered content.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: System MUST render the public dashboard at `/` for anonymous visitors without requiring login.
- **FR-002**: System MUST only display systems explicitly enabled for public visibility.
- **FR-003**: System MUST exclude private systems and private-only fields from anonymous responses and views.
- **FR-004**: Administrators MUST be able to enable or disable public visibility per system.
- **FR-005**: Public status entries MUST include only the default public metrics: display name, current availability state, data freshness, CPU summary percentage, memory summary percentage, disk summary percentage, and the new network latency trend chart when latency display is enabled.
- **FR-006**: Public status entries MUST avoid exposing secrets, internal addresses, credentials, tokens, owner-only identifiers, or administrative controls.
- **FR-007**: System MUST handle the anonymous home page with no public systems by showing a safe empty state.
- **FR-008**: System MUST let administrators create configurable network probes with a display name, probe type, target, visibility, check interval, and execution-node assignment, while keeping agent binding out of the default flow.
- **FR-009**: System MUST NOT include built-in network probe targets or limit the feature to exactly three probes.
- **FR-010**: System MUST support TCPing, ICMP Ping, and HTTP GET probe types.
- **FR-011**: Network probe results MUST record timestamp, probe label, probe type, target identity, executing node, success or failure, latency value when available, HTTP response status when available, and packet-loss or timeout state when available.
- **FR-012**: Users MUST be able to view current and historical reachability for each configured network probe, plus latency trends for TCPing and ICMP Ping probes.
- **FR-013**: Latency trend views MUST support at least recent, daily, and longer-range time windows.
- **FR-014**: A failed network probe or offline execution node MUST NOT prevent recording results for other configured probes or execution nodes.
- **FR-015**: Network probe summaries and latency charts MUST be visible on the anonymous home page by default when their executing node is public.
- **FR-016**: Administrators MUST be able to manually disable anonymous public display per network probe.
- **FR-017**: Existing authenticated monitoring views MUST continue to show private details unless an administrator changes visibility settings.
- **FR-018**: Public views MUST present user-entered display names and labels safely without allowing page content injection.
- **FR-019**: Public dashboard, visibility settings, and network probe screens MUST use Chinese product language and must not read like untranslated internal tooling.
- **FR-020**: Public dashboard and settings pages for this feature MUST show actionable controls and/or preview state instead of empty static placeholders.

### Key Entities *(include if feature involves data)*

- **Public System Visibility**: Per-system setting that determines whether the system appears on the anonymous home page.
- **Public Status Summary**: Sanitized status record for anonymous visitors, including display name, availability, freshness, and selected health indicators.
- **Network Probe**: A named check target configured by an administrator, including type, target, visibility, interval, and execution nodes that run it periodically.
- **Network Probe Result**: Timestamped measurement from an executing node to a probe target, including type, success, latency where applicable, HTTP response status where applicable, packet loss or timeout, and failure state.
- **Latency Trend Window**: A selected time range used to summarize and chart historical latency results.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: An anonymous visitor can open `/` and understand visible system availability within 5 seconds on a normal broadband connection.
- **SC-002**: When one system is public and one is private, 100% of anonymous home page views show only the public system.
- **SC-003**: Administrators can change a system's public visibility and see `/` reflect the change after refresh.
- **SC-004**: For enabled systems, users can compare latency trends from selected execution nodes to configured network probe targets for at least the most recent 24 hours.
- **SC-005**: If one network probe fails, the remaining probes continue to show their latest status and trend data.
- **SC-006**: Public views expose no private addresses, credentials, owner-only identifiers, or administrative actions during acceptance testing.

## Assumptions

- "Nezha-like public page" means the anonymous home dashboard showing selected server health, not full administrative access.
- Public visibility is opt-in per system; existing systems remain private by default.
- The anonymous home page displays a safe subset of metrics rather than every metric available to authenticated users.
- "Three-network" is treated as a common shorthand for network reachability and latency monitoring; the product supports administrator-defined probes rather than a fixed count or built-in targets.
- Latency trend history follows the project's existing data retention approach unless planning identifies a required retention change.
- Authenticated users retain access to richer private system detail than anonymous visitors.
- Anonymous users land on the public dashboard at `/`; there is no separate public-link route to discover.
