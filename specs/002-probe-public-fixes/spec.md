# Feature Specification: Probe Chart and Public Dashboard Fixes

**Feature Branch**: `002-probe-public-fixes`

**Created**: 2026-06-29

**Status**: Draft

**Input**: User description: "本地测试节点验证后需要优化：节点详情页添加 tcping 检测后打不开；tcping 不应该一个检测一个单独显示框，应参考 Komari 的呈现方式，把配置的检测目标由对应节点 agent 定期探测，并将同一类检测结果聚合到一个折线图里；添加 tcping 节点后提示检测失败；公共看板中公开节点的 CPU/内存/磁盘信息为空，最后上报时间不会动态刷新。"

## User Scenarios & Testing *(mandatory)*

### User Story 1 - View Stable Aggregated TCPing Trends (Priority: P1)

An authenticated user opens a node detail page after configuring TCPing checks and sees a stable latency trend section. Multiple configured TCPing targets or observation lines are shown together in a single comparative chart, rather than as one standalone card per TCPing item.

**Why this priority**: The node detail page currently becomes unusable after adding TCPing. Restoring the detail page and correcting the chart model is the highest-impact regression.

**Independent Test**: Configure at least two TCPing targets for a connected local node, open that node's detail page, and verify the page loads and shows both TCPing series in one combined trend chart.

**Acceptance Scenarios**:

1. **Given** a connected node has at least one TCPing check configured, **When** the user opens the node detail page, **Then** the page loads without crashing or becoming blank.
2. **Given** a node has multiple TCPing checks configured, **When** the user views the latency section, **Then** the checks are displayed as multiple series in one combined折线图.
3. **Given** TCPing, ICMP Ping, and HTTP GET checks all exist, **When** the user views the node detail page, **Then** latency-capable checks are grouped coherently and reachability-only checks do not create confusing standalone TCPing cards.
4. **Given** no results exist yet for a newly configured TCPing check, **When** the user opens the node detail page, **Then** the chart area shows a clear pending/empty state instead of breaking the page.

---

### User Story 2 - Diagnose TCPing Failures (Priority: P2)

An administrator creates a TCPing check and can tell whether a failure is caused by target format, target reachability, offline execution node, unsupported execution capability, or timeout.

**Why this priority**: A generic "检测失败" message is not enough to distinguish configuration mistakes from real network latency or reachability problems.

**Independent Test**: Create one reachable TCPing target and one unreachable or invalid TCPing target, wait for checks, and verify the UI records success for the reachable target and a useful failure reason for the failed target.

**Acceptance Scenarios**:

1. **Given** a TCPing target uses an invalid format, **When** the user saves it, **Then** the UI blocks the save or shows a validation message explaining the expected `host:port` format.
2. **Given** a TCPing target is syntactically valid but unreachable, **When** the check runs, **Then** the result is recorded as failed with a user-visible reason such as timeout, connection refused, DNS failure, or execution node unavailable.
3. **Given** one TCPing target fails and another succeeds, **When** results are displayed, **Then** the failed target does not prevent successful targets from showing their latest latency.
4. **Given** an execution node is offline, **When** its assigned checks are due, **Then** the UI distinguishes node-unavailable failure from target-unreachable failure.

---

### User Story 3 - View Complete Public Metrics and Freshness (Priority: P3)

An anonymous visitor opens the public dashboard and sees non-empty CPU, memory, disk, and freshness information for public nodes as soon as those metrics are available, with values updating over time.

**Why this priority**: The public dashboard's core value is visible node health. Blank metrics and stale freshness make the page appear broken even when the node is reporting.

**Independent Test**: Add a connected local node, enable it for the public dashboard, open `/` anonymously, and verify CPU, memory, disk, and last-report time appear and update after new reports.

**Acceptance Scenarios**:

1. **Given** a public node has reported CPU, memory, and disk values, **When** an anonymous visitor opens `/`, **Then** those values are shown instead of blanks.
2. **Given** a public node sends a newer report, **When** the public dashboard refreshes or polls, **Then** the displayed metrics and last-report time update without requiring a full browser restart.
3. **Given** a metric is genuinely unavailable, **When** the public dashboard renders, **Then** the missing value is shown as an explicit unavailable state rather than a blank-looking value.
4. **Given** a node is public but temporarily stale, **When** an anonymous visitor views it, **Then** the freshness state makes staleness understandable without exposing private details.

### Edge Cases

- A TCPing check has no history yet.
- Multiple checks share the same display name.
- A check changes target after historical data already exists.
- One execution node has results while another has no results for the same check.
- Latency values include spikes, timeouts, or gaps.
- A public node has latest status but missing one of CPU, memory, or disk values.
- Public dashboard polling happens while visibility settings change.
- A user opens the node detail page while checks are still running.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: Node detail pages MUST remain usable after TCPing checks are created.
- **FR-002**: TCPing latency results for a node MUST be presented as comparable series in a combined trend chart, not as one standalone display frame per TCPing item.
- **FR-003**: The combined chart MUST clearly identify each series by configured line/target label.
- **FR-004**: Newly configured checks with no results MUST show a pending or empty state without breaking the page.
- **FR-005**: TCPing target validation MUST clearly communicate the expected `host:port` format before or during save.
- **FR-006**: TCPing failures MUST preserve a user-visible failure category sufficient to distinguish invalid target, DNS failure, timeout, connection refused, target unreachable, and execution node unavailable where that information is available.
- **FR-007**: A failing TCPing check MUST NOT prevent other configured checks or other series in the same chart from displaying.
- **FR-008**: Public dashboard system summaries MUST show CPU, memory, disk, and freshness values when those values are available from the latest node report.
- **FR-009**: Public dashboard system summaries MUST refresh metrics and freshness over time without requiring users to log in.
- **FR-010**: Missing public metrics MUST render as an explicit unavailable state rather than visually empty content.
- **FR-011**: Public dashboard refresh behavior MUST continue to exclude private systems and private fields.
- **FR-012**: Changed behavior MUST be covered by focused unit tests.
- **FR-013**: If HTTP APIs are added or changed, they MUST be RESTful or explicitly justify an existing non-REST contract.

### API Contracts *(include if HTTP API changes)*

- **Resources**: Existing public dashboard status and network probe result resources may be extended if needed to provide grouped series, freshness, or failure categories.
- **Methods**: Existing read operations remain read-only; any configuration changes continue to use resource update semantics.
- **Status Codes**: Invalid TCPing configuration must return a client-visible validation failure, while runtime probe failures must be represented as result data rather than endpoint failure.
- **Schemas**: Probe result schemas must carry enough label, timestamp, success, latency, and failure-category data to build a combined chart. Public status schemas must carry available CPU, memory, disk, and freshness data.
- **Compatibility**: Existing clients that consume current probe or public status data should continue to work; new fields should be additive where possible.

### Key Entities *(include if feature involves data)*

- **Latency Chart Series**: A named line in a combined chart, derived from one configured check and one executing node/path, with timestamped latency or failure points.
- **Probe Failure Category**: User-facing classification of why a check failed, such as validation error, timeout, DNS failure, connection refused, target unreachable, or execution node unavailable.
- **Public Metric Summary**: Public-safe CPU, memory, disk, and freshness values for a node enabled on the anonymous dashboard.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Opening a node detail page with at least two TCPing checks succeeds 100% of the time during acceptance testing.
- **SC-002**: Two or more TCPing checks for the same node are visible in one combined trend chart during acceptance testing.
- **SC-003**: A reachable TCPing target records at least one successful latency result within two configured intervals when its execution node is online.
- **SC-004**: A failed TCPing target displays a meaningful failure reason in the UI during acceptance testing.
- **SC-005**: A public connected node shows CPU, memory, disk, and last-report freshness on `/` within 5 seconds of dashboard load when data is available.
- **SC-006**: Public dashboard metrics update after a newer node report without requiring logout or login.

## Assumptions

- "Komari-like" chart behavior means multiple configured latency series are compared in one chart for a node or target context.
- TCPing remains agent/execution-node based; hub-only probing is not introduced by this optimization.
- Public dashboard remains anonymous and must not expose host, port, tokens, user IDs, or other private operational details.
- Existing public visibility controls remain the source of which nodes appear on `/`.
- The first optimization target is TCPing because that is the user-verified failure path; ICMP and HTTP should not regress.
