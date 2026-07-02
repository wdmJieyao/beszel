---
description: "Task list for probe chart and public dashboard fixes"
---

# Tasks: Probe Chart and Public Dashboard Fixes

**Input**: Design documents from `/specs/002-probe-public-fixes/`

**Prerequisites**: plan.md, spec.md, research.md, data-model.md, contracts/

**Tests**: Unit tests are required by the project constitution for every behavior change. Test tasks appear before implementation tasks in each phase.

**Organization**: Tasks are grouped by user story to enable independent implementation and testing.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (US1, US2, US3)
- Include exact file paths in descriptions

## Phase 1: Setup (Shared Investigation)

**Purpose**: Confirm current failure points and affected files before changing behavior.

- [X] T001 Confirm current Docker data state and reproduce node detail page failure after TCPing creation using `docker compose logs` and browser/API validation against `http://127.0.0.1:8090`
- [X] T002 [P] Inspect current probe result payloads and assignments from `internal/hub/network_probes.go` and `internal/site/src/components/routes/system/use-network-probe-data.ts`
- [X] T003 [P] Inspect public status metric composition in `internal/hub/public_status.go` and compare it with latest system record fields used by `internal/site/src/components/routes/home.tsx`
- [X] T004 [P] Confirm required quality commands from `specs/002-probe-public-fixes/quickstart.md`, `Makefile`, and `internal/site/package.json`

---

## Phase 2: Foundational (Shared Probe/Public Data Helpers)

**Purpose**: Add shared helpers and types needed by all stories.

**CRITICAL**: No story implementation should begin until this phase is complete.

- [X] T005 [P] Add or update TypeScript types for grouped probe chart series, failure categories, and unavailable public metrics in `internal/site/src/types.d.ts`
- [X] T006 [P] Add Go failure-category constants and safe error-label mapping helpers in `internal/hub/network_probes.go`
- [X] T007 [P] Add public metric extraction helper tests in `internal/hub/public_status_test.go`
- [X] T008 [P] Add TCPing target/failure classification tests in `internal/hub/network_probes_test.go`
- [X] T009 Implement public metric extraction helpers in `internal/hub/public_status.go`
- [X] T010 Implement TCPing failure-category normalization helpers in `internal/hub/network_probes.go`

**Checkpoint**: Shared types and backend helpers exist for story work.

---

## Phase 3: User Story 1 - View Stable Aggregated TCPing Trends (Priority: P1) MVP

**Goal**: Node detail pages load after TCPing checks are configured and show multiple TCPing series in one combined chart.

**Independent Test**: Configure at least two TCPing targets for a connected local node, open that node's detail page, and verify both TCPing lines appear in one chart without crashing the page.

### Tests for User Story 1

- [X] T011 [P] [US1] Add frontend data-shaping test note or pure helper test for grouping probe results by type/system in `internal/site/src/components/routes/system/use-network-probe-data.ts`
- [X] T012 [P] [US1] Add Go result serialization coverage for probe result points with failed and empty history cases in `internal/hub/network_probes_test.go`

### Implementation for User Story 1

- [X] T013 [US1] Refactor `useNetworkProbeData` to return grouped latency series and reachability-only series separately in `internal/site/src/components/routes/system/use-network-probe-data.ts`
- [X] T014 [US1] Update `NetworkProbeChart` to support multiple named series in one chart without crashing on empty, failed, or mixed points in `internal/site/src/components/charts/network-probe-chart.tsx`
- [X] T015 [US1] Replace per-probe TCPing frames with one grouped latency section in the node detail page in `internal/site/src/components/routes/system.tsx`
- [X] T016 [US1] Ensure HTTP GET or reachability-only checks render as compact status summaries, not standalone TCPing-style latency cards, in `internal/site/src/components/routes/system.tsx`
- [X] T017 [US1] Add clear pending/no-results state for newly configured probes in `internal/site/src/components/charts/network-probe-chart.tsx`

**Checkpoint**: US1 works independently: node detail page stays usable and combined TCPing chart renders.

---

## Phase 4: User Story 2 - Diagnose TCPing Failures (Priority: P2)

**Goal**: TCPing failures show actionable reasons while successful checks continue to display.

**Independent Test**: Create one reachable TCPing target and one invalid/unreachable TCPing target, wait for checks, and verify validation/runtime failure reasons are visible without blocking successful latency series.

### Tests for User Story 2

- [X] T018 [P] [US2] Add Go validation tests for invalid TCPing target formats in `internal/hub/network_probes_test.go`
- [X] T019 [P] [US2] Add agent TCPing failure classification tests for DNS failure, timeout, and connection refused where practical in `agent/network_probe_test.go`
- [X] T020 [P] [US2] Add hub offline/unsupported execution-node failure classification tests in `internal/hub/network_probes_test.go`

### Implementation for User Story 2

- [X] T021 [US2] Tighten TCPing target validation and return clear `host:port` validation messages in `internal/hub/network_probes.go`
- [X] T022 [US2] Preserve normalized failure categories in probe result responses and persistence in `internal/hub/network_probes.go`
- [X] T023 [US2] Add failure category field to websocket probe result structs if needed in `internal/common/common-ws.go`
- [X] T024 [US2] Map agent TCPing runtime errors to safe failure categories in `agent/network_probe.go`
- [X] T025 [US2] Render failure category labels in settings/results/chart UI in `internal/site/src/components/charts/network-probe-chart.tsx` and `internal/site/src/components/routes/settings/network-probes.tsx`
- [X] T026 [US2] Ensure failed TCPing points do not remove or hide successful series in `internal/site/src/components/routes/system/use-network-probe-data.ts`

**Checkpoint**: US2 works independently: TCPing save/runtime failures are diagnosable and non-blocking.

---

## Phase 5: User Story 3 - View Complete Public Metrics and Freshness (Priority: P3)

**Goal**: Public dashboard shows CPU, memory, disk, and freshness for public connected nodes and refreshes them over time.

**Independent Test**: Enable the local connected node for the public dashboard, open `/` anonymously, and verify CPU/memory/disk/freshness appear and update after new reports.

### Tests for User Story 3

- [X] T027 [P] [US3] Add Go tests for public CPU/memory/disk extraction from latest system info in `internal/hub/public_status_test.go`
- [X] T028 [P] [US3] Add Go tests for explicit unavailable metric state without private field leaks in `internal/hub/public_status_test.go`
- [X] T029 [P] [US3] Add frontend refresh behavior validation note or helper test for public dashboard polling in `internal/site/src/components/routes/public-status.tsx`

### Implementation for User Story 3

- [X] T030 [US3] Fix public status read model to populate CPU, memory, disk, and freshness from latest available system report in `internal/hub/public_status.go`
- [X] T031 [US3] Extend public status response with explicit unavailable metric state if needed in `internal/hub/public_status.go` and `internal/site/src/types.d.ts`
- [X] T032 [US3] Update public dashboard metric rendering to show unavailable states instead of blank-looking values in `internal/site/src/components/routes/public-status.tsx`
- [X] T033 [US3] Add lightweight anonymous polling for public dashboard data without clearing existing data on refresh errors in `internal/site/src/components/routes/public-status.tsx`
- [X] T034 [US3] Preserve public response sanitization while adding metrics and freshness in `internal/hub/public_status.go`

**Checkpoint**: US3 works independently: public dashboard metrics and freshness display and refresh.

---

## Phase 6: Polish & Cross-Cutting Concerns

**Purpose**: Final validation, docs, and deployment checks across all stories.

- [X] T035 [P] Update REST contract examples after implementation in `specs/002-probe-public-fixes/contracts/rest-api.md`
- [X] T036 [P] Update agent probe contract examples after implementation in `specs/002-probe-public-fixes/contracts/agent-probe.md`
- [X] T037 [P] Update quickstart validation notes in `specs/002-probe-public-fixes/quickstart.md`
- [X] T038 Review node detail UI for dense, non-overlapping chart layout in `internal/site/src/components/routes/system.tsx`
- [X] T039 Review public dashboard payload for private field leaks in `internal/hub/public_status.go`
- [X] T040 Run focused Go tests with `/usr/local/go/bin/go test -tags=testing ./internal/hub ./agent ./internal/common ./internal/hub/ws`
- [X] T041 Run full Go tests with `/usr/local/go/bin/go test -tags=testing ./...`
- [ ] T042 Run Go lint with `golangci-lint run`
- [X] T043 Run frontend Biome check with `npm --prefix ./internal/site run check`
- [X] T044 Run frontend production build with `npm --prefix ./internal/site run build`
- [X] T045 Rebuild and start Docker validation service with `docker compose up -d --build`
- [X] T046 Validate Docker UI scenarios from `specs/002-probe-public-fixes/quickstart.md` against `http://127.0.0.1:8090`

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies.
- **Foundational (Phase 2)**: Depends on Setup; blocks all user stories.
- **US1 Aggregated TCPing Trends (Phase 3)**: Depends on Foundational and is MVP.
- **US2 TCPing Diagnostics (Phase 4)**: Depends on Foundational; can proceed after or alongside US1 backend helpers but final UI integrates with US1 chart components.
- **US3 Public Metrics/Freshness (Phase 5)**: Depends on Foundational; independent from US1/US2 except final Docker validation.
- **Polish (Phase 6)**: Depends on desired stories being complete.

### User Story Dependencies

- **US1 (P1)**: Highest priority; fixes node detail crash and chart model.
- **US2 (P2)**: Adds failure diagnostics; uses shared failure helpers and chart/status rendering.
- **US3 (P3)**: Fixes public dashboard metrics and refresh; independent of probe chart aggregation.

### Within Each User Story

- Tests before implementation where practical.
- Backend helpers before API response changes.
- API/types before frontend rendering.
- Chart/data shaping before node detail layout changes.
- Story checkpoint before moving to final validation.

### Parallel Opportunities

- Setup investigation tasks T002-T004 can run in parallel.
- Foundational tests T007-T008 can run in parallel with TypeScript/Go type preparation T005-T006.
- US1 tests T011-T012 can run in parallel.
- US2 tests T018-T020 can run in parallel.
- US3 tests T027-T029 can run in parallel.
- US3 backend public status work can proceed independently from US1/US2 chart work after Phase 2.
- Documentation updates T035-T037 can run in parallel before final verification.

## Parallel Example: User Story 1

```bash
Task: "T011 [P] [US1] Add frontend data-shaping test note or pure helper test for grouping probe results by type/system in internal/site/src/components/routes/system/use-network-probe-data.ts"
Task: "T012 [P] [US1] Add Go result serialization coverage for probe result points with failed and empty history cases in internal/hub/network_probes_test.go"
```

## Parallel Example: User Story 3

```bash
Task: "T027 [P] [US3] Add Go tests for public CPU/memory/disk extraction from latest system info in internal/hub/public_status_test.go"
Task: "T029 [P] [US3] Add frontend refresh behavior validation note or helper test for public dashboard polling in internal/site/src/components/routes/public-status.tsx"
```

## Implementation Strategy

### MVP First (US1 Only)

1. Complete Phase 1 and Phase 2.
2. Complete Phase 3.
3. Validate a node detail page with multiple TCPing checks renders one combined chart.
4. Stop and verify the page no longer crashes before adding diagnostics or public-dashboard fixes.

### Incremental Delivery

1. Deliver US1 chart stability and combined TCPing view.
2. Add US2 TCPing validation and failure diagnostics.
3. Add US3 public dashboard metric/freshness repair.
4. Run final quickstart and Docker validation.

### Risk Notes

- The node detail crash likely comes from frontend chart/data assumptions around empty or failed probe series; keep result shaping defensive.
- Public metric fixes must not expose private system fields.
- TCPing failure categories should be useful but safe; raw low-level errors may need normalization.
- Docker validation uses existing local data, so stale test records may affect manual observations unless noted.
