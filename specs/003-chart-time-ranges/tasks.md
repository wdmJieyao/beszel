# Tasks: Public Chart Time Ranges

**Input**: Design documents from `/specs/003-chart-time-ranges/`

**Prerequisites**: plan.md, spec.md, research.md, data-model.md, contracts/public-status-range.md, quickstart.md

**Tests**: Unit tests are REQUIRED for every behavior change by the project constitution. Include tests before implementation tasks. Integration or contract tests are added when the feature touches workflows, APIs, persistence, or component boundaries.

**Organization**: Tasks are grouped by user story to enable independent implementation and testing of each story.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2, US3)
- Include exact file paths in descriptions

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Confirm current project constraints and locate reusable chart/range patterns.

- [x] T001 Confirm current public status route, chart-time metadata, and public dashboard component entry points in `internal/hub/public_status.go`, `internal/site/src/lib/utils.ts`, and `internal/site/src/components/routes/public-status.tsx`
- [x] T002 [P] Confirm existing node detail range selector behavior in `internal/site/src/components/charts/chart-time-select.tsx` and `internal/site/src/components/routes/system/chart-data.ts`
- [x] T003 [P] Confirm quality commands and current baseline for `go test -tags=testing ./internal/hub ./internal/common ./internal/hub/ws`, `npm --prefix ./internal/site run check`, and `npm --prefix ./internal/site run build`

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Shared public chart range definitions and API typing needed by all stories.

**⚠️ CRITICAL**: No user story work can begin until this phase is complete.

- [x] T004 Add focused public chart range parsing tests for default `30m`, valid existing ranges, and invalid explicit ranges in `internal/hub/public_status_test.go`
- [x] T005 Implement public chart range parsing, defaulting, stats bucket selection, and range cutoff helpers in `internal/hub/public_status.go`
- [x] T006 Add `30m` chart range metadata and public chart range typing in `internal/site/src/lib/utils.ts` and `internal/site/src/types.d.ts`
- [x] T007 Update public status API helper typing to accept the public chart range values in `internal/site/src/lib/api.ts`
- [x] T008 [P] Document any intentional frontend unit-test harness gap for pure chart helper tests in `specs/003-chart-time-ranges/quickstart.md` if no runnable frontend test command exists

**Checkpoint**: Public range concepts are defined in backend, frontend, and API helper layers.

---

## Phase 3: User Story 1 - View Recent Public Trends (Priority: P1) 🎯 MVP

**Goal**: Public latency and resource charts default to the latest 30-minute window, refresh every 20 seconds, and keep latest resource summary values aligned with resource trend endings.

**Independent Test**: Open the public dashboard for a node with more than 30 minutes of data and confirm the visible latency chart and resource dialog charts default to 30 minutes, exclude older points, and refresh without page reload.

### Tests for User Story 1 ⚠️

- [x] T009 [P] [US1] Add backend tests for range-aware public metric history filtering and latest metric merging in `internal/hub/public_status_test.go`
- [x] T010 [P] [US1] Add backend tests for range-aware public latency series filtering and sanitized latest probe status in `internal/hub/public_status_test.go`
- [x] T011 [P] [US1] Add Playwright smoke-check notes for default 30-minute public latency/resource behavior in `specs/003-chart-time-ranges/quickstart.md`

### Implementation for User Story 1

- [x] T012 [US1] Apply parsed range filtering to public resource history queries in `internal/hub/public_status.go`
- [x] T013 [US1] Apply parsed range filtering to public latency probe result queries while preserving sanitized latest status in `internal/hub/public_status.go`
- [x] T014 [US1] Change public dashboard chart data refresh cadence to 20 seconds while preserving current data on refresh failure in `internal/site/src/components/routes/public-status.tsx`
- [x] T015 [US1] Filter public latency and resource points to the selected 30-minute default window before rendering in `internal/site/src/components/routes/public-status.tsx` and `internal/site/src/components/charts/network-probe-chart.tsx`
- [x] T016 [US1] Keep public CPU, memory, and disk summary values aligned with the latest visible resource trend point in `internal/site/src/components/routes/public-status.tsx`
- [x] T017 [US1] Verify public responses for default and `range=30m` do not expose target hostnames, IPs, ports, or raw target labels in `internal/hub/public_status.go`

**Checkpoint**: User Story 1 is functional and independently testable as the MVP.

---

## Phase 4: User Story 2 - Read Time from Chart Axes (Priority: P2)

**Goal**: Newly added public charts display readable hour-minute-second x-axis labels with dynamic label density for default and selected ranges.

**Independent Test**: Open the public latency chart and resource trend dialog, confirm the x-axis is visible, labels use hour-minute-second format, and labels do not overlap or wrap at desktop and mobile widths.

### Tests for User Story 2 ⚠️

- [x] T018 [P] [US2] Add unit-testable helper coverage or documented validation for public chart tick/domain generation in `internal/site/src/lib/utils.ts` or `specs/003-chart-time-ranges/quickstart.md`
- [x] T019 [P] [US2] Add Playwright smoke-check notes for x-axis labels and non-wrapping ticks in `specs/003-chart-time-ranges/quickstart.md`

### Implementation for User Story 2

- [x] T020 [US2] Implement public chart domain and tick generation for `30m`, `1m`, `1h`, `12h`, `24h`, `1w`, and `30d` in `internal/site/src/lib/utils.ts`
- [x] T021 [US2] Add visible x-axis labels to the public latency chart using hour-minute-second formatting in `internal/site/src/components/charts/network-probe-chart.tsx`
- [x] T022 [US2] Add visible x-axis labels to public CPU, memory, and disk resource charts using hour-minute-second formatting in `internal/site/src/components/routes/public-status.tsx`
- [x] T023 [US2] Tune chart margins, tick counts, and responsive behavior to prevent wrapping or overlap in `internal/site/src/components/charts/network-probe-chart.tsx` and `internal/site/src/components/routes/public-status.tsx`

**Checkpoint**: User Stories 1 and 2 both work independently.

---

## Phase 5: User Story 3 - Change Chart Time Range (Priority: P3)

**Goal**: Public latency and resource charts expose a range selector consistent with existing node-detail chart controls, default to 30 minutes, and keep selected ranges independent per node.

**Independent Test**: Use the public chart range controls to select multiple ranges and confirm plotted data and x-axis labels update without page reload, with one node's selection not changing another node unexpectedly.

### Tests for User Story 3 ⚠️

- [x] T024 [P] [US3] Add backend tests for explicit `range=1m`, `range=1h`, `range=12h`, `range=24h`, `range=1w`, and `range=30d` public status responses in `internal/hub/public_status_test.go`
- [x] T025 [P] [US3] Add backend contract test for invalid explicit range response behavior in `internal/hub/public_status_test.go`
- [x] T026 [P] [US3] Add Playwright smoke-check notes for range selector interactions and per-node range independence in `specs/003-chart-time-ranges/quickstart.md`

### Implementation for User Story 3

- [x] T027 [US3] Add reusable public chart range selector UI consistent with existing chart selector patterns in `internal/site/src/components/charts/chart-time-select.tsx` or a new colocated component under `internal/site/src/components/charts/`
- [x] T028 [US3] Add per-node public latency range state and pass selected range to API reload/rendering in `internal/site/src/components/routes/public-status.tsx`
- [x] T029 [US3] Add resource trend dialog range selector that updates CPU, memory, and disk charts consistently in `internal/site/src/components/routes/public-status.tsx`
- [x] T030 [US3] Wire explicit `range` query requests for public status reloads in `internal/site/src/lib/api.ts` and `internal/site/src/components/routes/public-status.tsx`
- [x] T031 [US3] Preserve selected range on 20-second refresh and refresh failures in `internal/site/src/components/routes/public-status.tsx`
- [x] T032 [US3] Ensure selected range changes do not expose hidden probe target metadata in rendered public chart labels or API response handling in `internal/site/src/components/routes/public-status.tsx`

**Checkpoint**: All user stories are independently functional.

---

## Phase 6: Polish & Cross-Cutting Concerns

**Purpose**: Final verification, compatibility review, and cleanup across all stories.

- [x] T033 [P] Review `specs/003-chart-time-ranges/contracts/public-status-range.md` against implemented public status behavior and update only if behavior intentionally differs
- [x] T034 [P] Review `specs/003-chart-time-ranges/quickstart.md` for final validation commands and expected outcomes
- [x] T035 Run focused backend tests with `go test -tags=testing ./internal/hub ./internal/common ./internal/hub/ws`
- [x] T036 Run frontend static checks with `npm --prefix ./internal/site run check` and document pre-existing unrelated diagnostics if any
- [x] T037 Run frontend production build with `npm --prefix ./internal/site run build`
- [x] T038 Run Go lint/static checks with `golangci-lint run` if the tool is available in the environment
- [x] T039 Run Docker validation with `docker compose up -d --build`
- [x] T040 Run API quickstart checks for default, `range=30m`, longer ranges, invalid range behavior, and target sanitization from `specs/003-chart-time-ranges/quickstart.md`
- [x] T041 Run Playwright browser validation for default 30-minute charts, range selection, 20-second refresh, x-axis labels, resource dialog charts, and per-node range independence

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies - can start immediately
- **Foundational (Phase 2)**: Depends on Setup completion - BLOCKS all user stories
- **User Story 1 (Phase 3)**: Depends on Foundational - MVP
- **User Story 2 (Phase 4)**: Depends on Foundational and can be developed after or alongside US1 once chart data exists, but final validation needs US1 data filtering
- **User Story 3 (Phase 5)**: Depends on Foundational and integrates with US1/US2 chart data and axes
- **Polish (Phase 6)**: Depends on all desired user stories being complete

### User Story Dependencies

- **User Story 1 (P1)**: Can start after Foundational - no dependency on US2 or US3
- **User Story 2 (P2)**: Can start after Foundational, but uses the chart data structures delivered by US1 for full end-to-end validation
- **User Story 3 (P3)**: Can start after Foundational, but range controls are most useful after US1 filtering and US2 axes are in place

### Within Each User Story

- Tests MUST be written before implementation when practical
- Backend range/read-model tests before backend implementation
- Shared range helpers before chart component integration
- API/helper typing before frontend API calls
- Core implementation before Playwright/Docker validation

### Parallel Opportunities

- T002 and T003 can run in parallel during setup
- T004 and T008 can run in parallel with T006/T007 because they touch different files
- T009, T010, and T011 can run in parallel for US1
- T018 and T019 can run in parallel for US2
- T024, T025, and T026 can run in parallel for US3
- Documentation review tasks T033 and T034 can run in parallel

---

## Parallel Example: User Story 1

```bash
# Backend tests and quickstart updates can be started together:
Task: "Add backend tests for range-aware public metric history filtering and latest metric merging in internal/hub/public_status_test.go"
Task: "Add backend tests for range-aware public latency series filtering and sanitized latest probe status in internal/hub/public_status_test.go"
Task: "Add Playwright smoke-check notes for default 30-minute public latency/resource behavior in specs/003-chart-time-ranges/quickstart.md"
```

---

## Parallel Example: User Story 2

```bash
# Axis helper validation and browser validation notes can proceed together:
Task: "Add unit-testable helper coverage or documented validation for public chart tick/domain generation in internal/site/src/lib/utils.ts or specs/003-chart-time-ranges/quickstart.md"
Task: "Add Playwright smoke-check notes for x-axis labels and non-wrapping ticks in specs/003-chart-time-ranges/quickstart.md"
```

---

## Parallel Example: User Story 3

```bash
# Backend range contract tests and range selector validation notes can proceed together:
Task: "Add backend tests for explicit range public status responses in internal/hub/public_status_test.go"
Task: "Add backend contract test for invalid explicit range response behavior in internal/hub/public_status_test.go"
Task: "Add Playwright smoke-check notes for range selector interactions and per-node range independence in specs/003-chart-time-ranges/quickstart.md"
```

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Complete Phase 1: Setup
2. Complete Phase 2: Foundational
3. Complete Phase 3: User Story 1
4. Stop and validate default 30-minute public latency/resource behavior
5. Deploy/demo if ready

### Incremental Delivery

1. Complete Setup + Foundational
2. Add User Story 1 -> default 30-minute data filtering and 20-second refresh
3. Add User Story 2 -> visible hour-minute-second x-axis labels
4. Add User Story 3 -> range selector and per-node range independence
5. Run Polish validation and Docker/Playwright checks

### Parallel Team Strategy

With multiple developers:

1. Complete Setup + Foundational together
2. Backend-focused worker implements range-aware public status tests and read model
3. Frontend-focused worker implements chart-time metadata, x-axes, and range selector UI
4. Verification-focused worker maintains quickstart/Playwright validation and checks public sanitization

## Notes

- [P] tasks = different files, no dependencies
- [US1], [US2], and [US3] labels map to user stories in `spec.md`
- Keep public chart data sanitized; do not reintroduce probe target labels or addresses
- Prefer existing chart and range patterns over introducing new charting abstractions
- Stop at each checkpoint to validate the story independently
