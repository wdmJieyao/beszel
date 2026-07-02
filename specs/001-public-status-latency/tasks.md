---
description: "Task list for public status page and network probe trends"
---

# Tasks: Public Status Page and Network Probe Trends

**Input**: Design documents from `/specs/001-public-status-latency/`

**Prerequisites**: plan.md, spec.md, research.md, data-model.md, contracts/

**Tests**: Unit tests are required by the project constitution for every behavior change. Test tasks appear before implementation tasks in each phase.

**Organization**: Tasks are grouped by user story to enable independent implementation and testing of each story.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (US1, US2, US3)
- Include exact file paths in descriptions

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Confirm local tooling and reserve feature file locations.

- [X] T001 Confirm Go, golangci-lint, and frontend package-manager commands from `Makefile` and `internal/site/package.json`
- [X] T002 [P] Create placeholder feature files `internal/hub/public_status.go`, `internal/hub/network_probes.go`, `agent/network_probe.go`, and `internal/site/src/components/routes/public-status.tsx`
- [X] T003 [P] Add TypeScript probe/public-status type placeholders in `internal/site/src/types.d.ts`
- [X] T004 Decide and document frontend unit-test harness or justified temporary test gap in `specs/001-public-status-latency/quickstart.md`

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Add shared schema, protocol, validation, and scheduling foundations required by all user stories.

**CRITICAL**: No user story work can begin until this phase is complete.

- [X] T005 Add PocketBase migration for `public_system_visibility`, `network_probes`, `network_probe_assignments`, and `network_probe_results` in `internal/migrations/add_public_status_probes.go`
- [X] T006 Add collection auth rules for new public/probe collections in `internal/hub/collections.go`
- [X] T007 [P] Add public visibility and network probe domain structs/constants in `internal/hub/network_probes.go`
- [X] T008 [P] Add `RunNetworkProbe` action and probe request/response structs in `internal/common/common-ws.go`
- [X] T009 [P] Add unit tests for probe validation rules in `internal/hub/network_probes_test.go`
- [X] T010 [P] Add unit tests for public status sanitization rules in `internal/hub/public_status_test.go`
- [X] T011 Implement probe validation helpers for `tcping`, `icmp_ping`, and `http_get` targets in `internal/hub/network_probes.go`
- [X] T012 Implement shared public status sanitization helpers in `internal/hub/public_status.go`
- [X] T013 Register REST route groups for `/api/beszel/public/*` and `/api/beszel/network-probes*` in `internal/hub/api.go`
- [X] T014 Register agent `RunNetworkProbe` handler skeleton in `agent/handlers.go`
- [X] T015 Add network probe result averaging/retention decision hooks or explicit no-aggregation path in `internal/records/records.go`

**Checkpoint**: Database schema, protocol types, validation helpers, and route skeletons exist.

---

## Phase 3: User Story 1 - View Public VPS Status (Priority: P1) MVP

**Goal**: Anonymous visitors can open `/` and see only public systems with safe minimal metrics.

**Independent Test**: Configure one public system and one private system, open `/` signed out, and verify only the public system appears with allowed fields.

### Tests for User Story 1

- [X] T016 [P] [US1] Add Go test for anonymous public status filtering in `internal/hub/public_status_test.go`
- [X] T017 [P] [US1] Add Go test rejecting private fields from public payloads in `internal/hub/public_status_test.go`
- [X] T018 [P] [US1] Add frontend public payload mapping test or documented frontend test gap in `internal/site/src/components/routes/public-status.test.ts`

### Implementation for User Story 1

- [X] T019 [US1] Implement public status read model from systems and visibility records in `internal/hub/public_status.go`
- [X] T020 [US1] Implement `GET /api/beszel/public/status` handler in `internal/hub/public_status.go`
- [X] T021 [US1] Wire public status handler into `/api/beszel/public/status` route in `internal/hub/api.go`
- [X] T022 [US1] Add public route key and path for `/public` in `internal/site/src/components/router.tsx`
- [X] T023 [US1] Implement anonymous public status data fetcher in `internal/site/src/lib/api.ts`
- [X] T024 [US1] Implement public status page empty/loading/error states in `internal/site/src/components/routes/public-status.tsx`
- [X] T025 [US1] Render public system cards/table with name, online state, freshness, CPU, memory, and disk summaries in `internal/site/src/components/routes/public-status.tsx`
- [X] T026 [US1] Ensure authenticated app navigation does not require login for `/public` in `internal/site/src/main.tsx`
- [X] T027 [US1] Add public status route smoke coverage or manual validation note to `specs/001-public-status-latency/quickstart.md`

**Checkpoint**: User Story 1 is functional and independently testable as the MVP.

---

## Phase 4: User Story 2 - Manage Public Visibility and Display Scope (Priority: P2)

**Goal**: Administrators can opt systems into or out of public visibility and control safe display scope.

**Independent Test**: As admin, enable and disable public visibility for a system and verify anonymous public output follows the change.

### Tests for User Story 2

- [X] T028 [P] [US2] Add Go tests for public visibility update authorization in `internal/hub/public_status_test.go`
- [X] T029 [P] [US2] Add Go tests for public visibility validation and defaults in `internal/hub/public_status_test.go`
- [X] T030 [P] [US2] Add frontend settings form mapping test or documented frontend test gap in `internal/site/src/components/routes/settings/public-status.test.ts`

### Implementation for User Story 2

- [X] T031 [US2] Implement `GET /api/beszel/public/systems` admin handler in `internal/hub/public_status.go`
- [X] T032 [US2] Implement `PATCH /api/beszel/public/systems/{systemId}` admin handler in `internal/hub/public_status.go`
- [X] T033 [US2] Enforce readonly-user rejection for public visibility updates in `internal/hub/api.go`
- [X] T034 [US2] Add public visibility fields to frontend types in `internal/site/src/types.d.ts`
- [X] T035 [US2] Implement public visibility settings API helpers in `internal/site/src/lib/api.ts`
- [X] T036 [US2] Add public visibility settings route or tab entry in `internal/site/src/components/routes/settings/layout.tsx`
- [X] T037 [US2] Implement public visibility settings UI in `internal/site/src/components/routes/settings/public-status.tsx`
- [X] T038 [US2] Add safe public-name and metric toggle controls in `internal/site/src/components/routes/settings/public-status.tsx`
- [X] T039 [US2] Refresh anonymous public status after admin visibility changes in `internal/site/src/components/routes/settings/public-status.tsx`

**Checkpoint**: Admins can control which systems appear publicly without exposing private systems.

---

## Phase 5: User Story 3 - View Configurable Network Probe Trends (Priority: P3)

**Goal**: Users can configure lines/observation points executed by selected nodes and view reachability/latency trends, including public-safe charts.

**Independent Test**: Configure TCPing, ICMP Ping, and HTTP GET lines, wait for results, and verify authenticated and public-safe trend views without requiring primary-flow agent binding.

### Tests for User Story 3

- [X] T040 [P] [US3] Add agent TCPing unit tests in `agent/network_probe_test.go`
- [X] T041 [P] [US3] Add agent HTTP GET unit tests in `agent/network_probe_test.go`
- [X] T042 [P] [US3] Add agent ICMP unsupported/permission failure unit tests in `agent/network_probe_test.go`
- [X] T043 [P] [US3] Add hub probe CRUD and assignment tests in `internal/hub/network_probes_test.go`
- [X] T044 [P] [US3] Add hub probe result persistence and failure handling tests in `internal/hub/network_probes_test.go`
- [X] T045 [P] [US3] Add public probe visibility filtering tests in `internal/hub/public_status_test.go`
- [X] T046 [P] [US3] Add frontend probe chart mapping test or documented frontend test gap in `internal/site/src/components/charts/network-probe-chart.test.ts`

### Implementation for User Story 3

- [X] T047 [US3] Implement TCPing, ICMP Ping, and HTTP GET probe execution in `agent/network_probe.go`
- [X] T048 [US3] Implement `RunNetworkProbe` agent handler response mapping in `agent/handlers.go`
- [X] T049 [US3] Add hub-side agent probe request helper in `internal/hub/ws/handlers.go`
- [X] T050 [US3] Implement probe scheduler/orchestrator for enabled assignments in `internal/hub/network_probes.go`
- [X] T051 [US3] Persist successful and failed probe results in `network_probe_results` from `internal/hub/network_probes.go`
- [X] T052 [US3] Implement `GET /api/beszel/network-probes` in `internal/hub/network_probes.go`
- [X] T053 [US3] Implement `POST /api/beszel/network-probes` in `internal/hub/network_probes.go`
- [X] T054 [US3] Implement `PATCH /api/beszel/network-probes/{probeId}` in `internal/hub/network_probes.go`
- [X] T055 [US3] Implement `DELETE /api/beszel/network-probes/{probeId}` in `internal/hub/network_probes.go`
- [X] T056 [US3] Implement `GET /api/beszel/network-probes/{probeId}/results` in `internal/hub/network_probes.go`
- [X] T057 [US3] Add network probe frontend types in `internal/site/src/types.d.ts`
- [X] T058 [US3] Implement network probe API helpers in `internal/site/src/lib/api.ts`
- [X] T059 [US3] Implement network probe admin list and editor in `internal/site/src/components/routes/settings/network-probes.tsx`
- [X] T060 [US3] Implement probe assignment selector for agent systems in `internal/site/src/components/routes/settings/network-probes.tsx`
- [X] T061 [US3] Implement probe visibility toggle defaulting to public-visible in `internal/site/src/components/routes/settings/network-probes.tsx`
- [X] T062 [US3] Implement probe chart data loader in `internal/site/src/components/routes/system/use-network-probe-data.ts`
- [X] T063 [US3] Implement reusable network probe chart component in `internal/site/src/components/charts/network-probe-chart.tsx`
- [X] T064 [US3] Add authenticated probe trend section to system page in `internal/site/src/components/routes/system.tsx`
- [X] T065 [US3] Add public-safe probe summaries and charts to public page in `internal/site/src/components/routes/public-status.tsx`
- [X] T066 [US3] Add offline/unsupported agent messaging for assigned probes in `internal/site/src/components/routes/settings/network-probes.tsx`

**Checkpoint**: Network probes run from agents, persist results, and render authenticated/public-safe trends.

---

## Phase 6: Rework After Product Review

**Purpose**: Align the already-built slice with the clarified product baseline: `/` is the anonymous public dashboard, settings use Chinese product language, static public settings have actionable controls, and execution-node binding is hidden unless advanced mode is used.

### Tests for Rework

- [X] T067 [P] [US1] Add or update route/authentication test coverage proving anonymous `/` renders the public dashboard and authenticated users can still reach the app shell in `internal/site/src/main.tsx`
- [X] T068 [P] [US1] Add Go/API smoke coverage or documented validation for anonymous `GET /api/beszel/public/status` backing `/` in `internal/hub/public_status_test.go`
- [X] T069 [P] [US2] Add or update frontend mapping validation notes for Chinese public-dashboard settings controls in `internal/site/src/components/routes/settings/public-status.tsx`
- [X] T070 [P] [US3] Add or update frontend mapping validation notes for automatic versus advanced execution mode in `internal/site/src/components/routes/settings/network-probes.tsx`

### Implementation for Rework

- [X] T071 [US1] Change anonymous home route behavior so `/` renders the public dashboard without login while preserving authenticated dashboard access in `internal/site/src/main.tsx` and `internal/site/src/components/router.tsx`
- [X] T072 [US1] Keep any legacy `/public` route as an optional alias or redirect to `/` in `internal/site/src/components/router.tsx`
- [X] T073 [US1] Update public dashboard UI text to Chinese product language and include a clear admin login entry in `internal/site/src/components/routes/public-status.tsx`
- [X] T074 [US2] Replace the empty/static public settings experience with actionable controls: visible node selection, metric toggles, latency chart toggle, and preview state in `internal/site/src/components/routes/settings/public-status.tsx`
- [X] T075 [US2] Ensure node visibility can be managed from the public-dashboard settings page without requiring edits to existing system edit modules in `internal/site/src/components/routes/settings/public-status.tsx`
- [X] T076 [US3] Rename network-probe settings UI labels to Chinese product terms such as 线路, 观测点, 检测目标, 检测间隔, 公开展示, and 高级设置 in `internal/site/src/components/routes/settings/network-probes.tsx`
- [X] T077 [US3] Hide explicit agent/execution-node binding in the default probe flow and expose it only inside an advanced settings section in `internal/site/src/components/routes/settings/network-probes.tsx`
- [X] T078 [US3] Add automatic execution-mode request/response handling as needed in `internal/site/src/lib/api.ts`, `internal/site/src/types.d.ts`, and `internal/hub/network_probes.go`
- [X] T079 [US3] Update public and authenticated probe charts to use line/observation-point labels rather than internal agent wording in `internal/site/src/components/charts/network-probe-chart.tsx` and `internal/site/src/components/routes/public-status.tsx`
- [X] T080 [US2] Update user-facing guide to describe `/` as the public dashboard and remove references to separate public links in `supplemental/guides/public-status-probes.md`
- [X] T081 Update Docker compose documentation or defaults if route behavior changes require no extra setup in `docker-compose.yml`

**Checkpoint**: The Docker-verifiable UI matches product review feedback.

---

## Phase 7: Polish & Cross-Cutting Concerns

**Purpose**: Validate security, performance, docs, and final quality gates across all stories.

- [X] T082 [P] Update public/probe contract examples after implementation in `specs/001-public-status-latency/contracts/rest-api.md`
- [X] T083 [P] Update agent probe contract examples after implementation in `specs/001-public-status-latency/contracts/agent-probe.md`
- [X] T084 [P] Add user-facing documentation draft for public page and probes in `supplemental/guides/public-status-probes.md`
- [X] T085 Review anonymous public payloads for private field leaks in `internal/hub/public_status.go`
- [X] T086 Review probe target validation and SSRF-sensitive behavior in `internal/hub/network_probes.go`
- [ ] T087 Run backend tests with `go test -tags=testing ./...`
- [ ] T088 Run Go lint with `golangci-lint run`
- [X] T089 Run frontend Biome checks with `bun run --cwd ./internal/site check`
- [ ] T090 Validate quickstart scenarios in `specs/001-public-status-latency/quickstart.md`
- [ ] T091 Confirm all REST endpoints match `specs/001-public-status-latency/contracts/rest-api.md`
- [ ] T092 Confirm agent websocket payloads match `specs/001-public-status-latency/contracts/agent-probe.md`

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies.
- **Foundational (Phase 2)**: Depends on Setup; blocks all user stories.
- **US1 Anonymous Home Dashboard (Phase 3)**: Depends on Foundational.
- **US2 Visibility Management (Phase 4)**: Depends on Foundational; can proceed after US1 backend read model exists, but is independently testable through admin APIs.
- **US3 Network Probe Trends (Phase 5)**: Depends on Foundational; public probe display task T065 depends on US1 public page component.
- **Rework (Phase 6)**: Depends on product review feedback and completed initial implementation.
- **Polish (Phase 7)**: Depends on desired stories and rework being complete.

### User Story Dependencies

- **US1 (P1)**: MVP and first deliverable; `/` must be the anonymous dashboard.
- **US2 (P2)**: Builds admin controls for visibility; complements US1.
- **US3 (P3)**: Uses shared probe schema/protocol; public chart rendering depends on US1 home dashboard existing.

### Within Each User Story

- Tests before implementation when practical.
- Backend validation before handlers.
- Handlers before frontend API helpers.
- Frontend API helpers before UI components.
- Story checkpoint before moving to next priority.

---

## Parallel Opportunities

- Setup placeholders T002 and frontend types T003 can run in parallel.
- Foundational tests T009 and T010 can run in parallel with protocol/schema work after T005-T008 are drafted.
- US1 tests T016-T018 can run in parallel.
- US2 tests T028-T030 can run in parallel.
- US3 tests T040-T046 can run in parallel because they cover distinct agent, hub, and frontend files.
- US3 frontend tasks T057-T063 can run in parallel with backend endpoint tasks T052-T056 once contracts are stable.
- Rework tests T067-T070 can run in parallel before route/UI edits.
- Polish verification T087-T092 runs after rework completes.

## Parallel Example: User Story 3

```bash
Task: "T040 [P] [US3] Add agent TCPing unit tests in agent/network_probe_test.go"
Task: "T043 [P] [US3] Add hub probe CRUD and assignment tests in internal/hub/network_probes_test.go"
Task: "T046 [P] [US3] Add frontend probe chart mapping test or documented frontend test gap in internal/site/src/components/charts/network-probe-chart.test.ts"
```

## Implementation Strategy

### MVP First (US1 Only)

1. Complete Phase 1 and Phase 2.
2. Complete Phase 3.
3. Validate anonymous `/` shows only public systems and safe minimal metrics.
4. Stop and review public data exposure before adding admin and probe features.

### Incremental Delivery

1. Deliver US1 anonymous home dashboard.
2. Add US2 admin controls for public visibility.
3. Add US3 probe configuration, execution-node probing, trend charts, and public probe display.
4. Apply Phase 6 product-review rework.
5. Run Phase 7 verification.

### Risk Notes

- Public data leaks are the primary risk; keep all anonymous payloads custom and sanitized.
- Probe targets may introduce network/security concerns; validate target formats and document SSRF-sensitive behavior.
- ICMP may require platform permissions; unsupported or denied ICMP must return a failed result rather than crash.
