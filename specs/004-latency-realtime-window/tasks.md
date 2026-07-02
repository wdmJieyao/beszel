# Tasks: Latency Realtime Window

**Input**: Design documents from `/specs/004-latency-realtime-window/`

**Prerequisites**: [plan.md](./plan.md), [spec.md](./spec.md), [research.md](./research.md), [data-model.md](./data-model.md), [contracts/latency-live-window.md](./contracts/latency-live-window.md), [quickstart.md](./quickstart.md)

**Tests**: Unit tests are required by the project constitution. This feature is frontend-only unless implementation discovers a backend scheduling defect.

**Organization**: Tasks are grouped by user story so each story can be implemented and validated independently.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel because it touches different files or only reads code.
- **[Story]**: Maps to the user story from [spec.md](./spec.md).
- Every task includes concrete file paths.

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Confirm the existing frontend surface and add the minimum unit-test entry point needed for this behavior change.

- [X] T001 Confirm the frontend unit-test command strategy in `internal/site/package.json` and keep Biome as the required static quality gate.
- [X] T002 [P] Review the CPU `1m` realtime reset behavior in `internal/site/src/components/routes/system/use-system-data.ts` and note the exact reset/append pattern in `specs/004-latency-realtime-window/quickstart.md`.
- [X] T003 [P] Review current latency hook and chart responsibilities in `internal/site/src/components/routes/system/use-network-probe-data.ts` and `internal/site/src/components/charts/network-probe-chart.tsx`.
- [X] T004 Add a `test:unit` script for frontend unit tests in `internal/site/package.json` using the repository's existing frontend toolchain.

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Create a small pure-logic boundary so the live-session semantics can be tested before UI wiring changes.

**Critical**: No user story implementation should begin until this phase is complete.

- [X] T005 Create `internal/site/src/components/routes/system/network-probe-live-session.ts` with exported types and placeholder function signatures for live-session initialization, realtime event normalization, and per-probe result merging.
- [X] T006 [P] Create `internal/site/src/components/routes/system/network-probe-live-session.test.ts` with shared fixtures for three configured latency probes, pre-session history, realtime success events, realtime failure events, and staggered probe arrivals.

**Checkpoint**: Testable pure-logic boundary exists; user story tasks can now add failing tests before implementation.

---

## Phase 3: User Story 1 - Start One-Minute Latency View Empty (Priority: P1) MVP

**Goal**: Switching node-detail 线路检测 to `1 分钟` starts empty and appends only realtime events received by the browser after entering the current session.

**Independent Test**: Open a node detail page with existing latency history, switch from `1 小时` to `1 分钟`, and verify the chart has no plotted historical latency line while fresh realtime samples later append.

### Tests for User Story 1

- [X] T007 [P] [US1] Add failing unit tests for empty `1m` session initialization and historical-result exclusion in `internal/site/src/components/routes/system/network-probe-live-session.test.ts`.
- [X] T008 [P] [US1] Add failing unit tests for appending only active-browser realtime result events after entering `1m` in `internal/site/src/components/routes/system/network-probe-live-session.test.ts`.
- [X] T009 [P] [US1] Add failing unit tests for clearing plotted `1m` results when leaving and re-entering `1m` in `internal/site/src/components/routes/system/network-probe-live-session.test.ts`.

### Implementation for User Story 1

- [X] T010 [US1] Implement live-session initialization and reset helpers in `internal/site/src/components/routes/system/network-probe-live-session.ts`.
- [X] T011 [US1] Implement realtime-event normalization, duplicate replacement, delete handling, and failure preservation in `internal/site/src/components/routes/system/network-probe-live-session.ts`.
- [X] T012 [US1] Update `useNetworkProbeData` in `internal/site/src/components/routes/system/use-network-probe-data.ts` so `range === "1m"` loads assigned probes for legend structure but does not call `getNetworkProbeResults` for plotted data.
- [X] T013 [US1] Update `useNetworkProbeData` in `internal/site/src/components/routes/system/use-network-probe-data.ts` so entering `1m` clears plotted results, subscribes to `network_probe_results`, and appends only events received by the active browser session.
- [X] T014 [US1] Preserve existing non-`1m` historical fetch and merge behavior in `internal/site/src/components/routes/system/use-network-probe-data.ts`.

**Checkpoint**: User Story 1 should pass unit tests and be independently verifiable by switching to `1 分钟` with pre-existing latency history.

---

## Phase 4: User Story 2 - Preserve All Configured Latency Lines (Priority: P1)

**Goal**: All configured and enabled latency lines remain visible in the `1 分钟` legend immediately, even when only some lines have fresh session points.

**Independent Test**: Configure three latency lines, switch to `1 分钟`, and verify all three legend labels appear immediately while plotted lines appear only after each line receives fresh realtime samples.

### Tests for User Story 2

- [X] T015 [P] [US2] Add failing unit tests for `groupNetworkProbeData` preserving three configured latency series with zero live points in `internal/site/src/components/routes/system/use-network-probe-data.test.ts`.
- [X] T016 [P] [US2] Add failing unit tests for staggered realtime arrivals preserving missing-line legend entries in `internal/site/src/components/routes/system/use-network-probe-data.test.ts`.

### Implementation for User Story 2

- [X] T017 [US2] Export or move pure grouping helpers needed by tests from `internal/site/src/components/routes/system/use-network-probe-data.ts` without changing public hook behavior.
- [X] T018 [US2] Update `groupNetworkProbeData` in `internal/site/src/components/routes/system/use-network-probe-data.ts` so configured latency probes always produce series entries even when their `points` array is empty.
- [X] T019 [US2] Update empty and no-latency rendering in `internal/site/src/components/charts/network-probe-chart.tsx` so the legend controls remain visible when configured series exist but no fresh points have arrived.
- [X] T020 [US2] Ensure hidden-series state pruning in `internal/site/src/components/charts/network-probe-chart.tsx` keeps all configured series selectable when their point arrays are empty.

**Checkpoint**: User Story 2 should pass unit tests and show three legend entries immediately after switching to `1 分钟`.

---

## Phase 5: User Story 3 - Match CPU One-Minute Time Behavior (Priority: P2)

**Goal**: The latency `1 分钟` chart uses CPU-like live-window timing so it starts blank, advances with current time, and does not stretch a short session into a misleading long line.

**Independent Test**: On the same node detail page, switch to `1 分钟` and compare CPU 使用率 and 线路检测; both should start from the current live observation and advance as fresh data arrives.

### Tests for User Story 3

- [X] T021 [P] [US3] Add failing unit tests for latency chart `1m` time-domain behavior using an injectable clock helper in `internal/site/src/components/charts/network-probe-chart-data.test.ts`.
- [X] T022 [P] [US3] Add failing unit tests that one fresh point produces only point state and two fresh points produce a short current-session segment in `internal/site/src/components/charts/network-probe-chart-data.test.ts`.

### Implementation for User Story 3

- [X] T023 [US3] Extract testable time-domain and row calculation from `NetworkProbeChart` into `internal/site/src/components/charts/network-probe-chart-data.ts` with an injectable current time for tests.
- [X] T024 [US3] Update `NetworkProbeChart` in `internal/site/src/components/charts/network-probe-chart.tsx` to consume `internal/site/src/components/charts/network-probe-chart-data.ts` for CPU-like rolling current-time domain and stable ticks when `range === "1m"`.
- [X] T025 [US3] Update `NetworkProbeChart` in `internal/site/src/components/charts/network-probe-chart.tsx` so single-point and two-point live sessions use the prepared per-series rows and do not draw a long pre-session-looking segment.
- [X] T026 [US3] Verify `线路检测` labels, tooltip latency units, and empty-state text remain Chinese in `internal/site/src/components/charts/network-probe-chart.tsx`.

**Checkpoint**: User Story 3 should pass unit tests and visually match CPU 使用率 `1 分钟` live-window semantics.

---

## Phase 6: Polish & Cross-Cutting Concerns

**Purpose**: Verify quality gates, no API scope drift, and browser behavior under the local deployment.

- [X] T027 [P] Update validation notes for the realtime-only behavior in `specs/004-latency-realtime-window/quickstart.md`.
- [X] T028 Run frontend unit tests through `internal/site/package.json` and fix failures in `internal/site/src/components/routes/system/network-probe-live-session.ts`, `internal/site/src/components/routes/system/use-network-probe-data.ts`, or `internal/site/src/components/charts/network-probe-chart.tsx`.
- [X] T029 Run `npm --prefix ./internal/site run check` and fix Biome issues in `internal/site/src/components/routes/system/use-network-probe-data.ts`, `internal/site/src/components/routes/system/network-probe-live-session.ts`, and `internal/site/src/components/charts/network-probe-chart.tsx`.
- [X] T030 Run `npm --prefix ./internal/site run build` and fix production build issues in `internal/site/package.json` or touched `internal/site/src/` files.
- [X] T031 Confirm no Go backend, agent, storage, or HTTP API files were changed for this feature by reviewing `git diff -- agent internal migrations pb_hooks` and record any exception in `specs/004-latency-realtime-window/quickstart.md`.
- [X] T032 Run `docker compose up -d --build beszel beszel-agent` using `docker-compose.yml`, then verify containers are healthy before browser validation.
- [X] T033 Use Playwright or an equivalent browser runner against `http://127.0.0.1:8090` to validate the scenarios in `specs/004-latency-realtime-window/quickstart.md`.

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies.
- **Foundational (Phase 2)**: Depends on Setup; blocks all user stories.
- **User Story 1 (Phase 3)**: Depends on Foundational and delivers the MVP.
- **User Story 2 (Phase 4)**: Depends on Foundational; can be implemented after or alongside US1, but final verification should include US1 behavior.
- **User Story 3 (Phase 5)**: Depends on Foundational; should run after US1 because it validates chart behavior for the live data produced by US1.
- **Polish (Phase 6)**: Depends on desired user stories being complete.

### User Story Dependencies

- **US1 Start One-Minute Latency View Empty (P1)**: MVP; no dependency on other user stories after Foundation.
- **US2 Preserve All Configured Latency Lines (P1)**: Independent legend/group behavior after Foundation; final UX pairs with US1.
- **US3 Match CPU One-Minute Time Behavior (P2)**: Depends on US1 live-session semantics for meaningful visual validation.

### Within Each User Story

- Write failing tests before implementation tasks.
- Implement pure live-session logic before hook wiring.
- Implement hook data semantics before chart visual adjustments.
- Validate each checkpoint before moving to the next story.

### Parallel Opportunities

- T002 and T003 can run in parallel.
- T006 can run in parallel after T005 signatures are known.
- T007, T008, and T009 can run in parallel because they add distinct test cases in the same test file after T006 establishes fixtures.
- T015 and T016 can run in parallel because they target separate grouping behaviors.
- T021 and T022 can run in parallel because they target separate chart behaviors.
- T027 can run in parallel with code polishing once implementation behavior is stable.

---

## Parallel Example: User Story 1

```bash
# Add US1 tests in parallel after fixtures exist:
Task: "T007 Add failing unit tests for empty `1m` session initialization in internal/site/src/components/routes/system/network-probe-live-session.test.ts"
Task: "T008 Add failing unit tests for appending only browser realtime events in internal/site/src/components/routes/system/network-probe-live-session.test.ts"
Task: "T009 Add failing unit tests for clearing plotted results when re-entering `1m` in internal/site/src/components/routes/system/network-probe-live-session.test.ts"
```

## Parallel Example: User Story 2

```bash
# Add US2 tests in parallel:
Task: "T015 Add failing unit tests for preserving three configured latency series in internal/site/src/components/routes/system/use-network-probe-data.test.ts"
Task: "T016 Add failing unit tests for staggered realtime arrivals preserving legend entries in internal/site/src/components/routes/system/use-network-probe-data.test.ts"
```

## Parallel Example: User Story 3

```bash
# Add US3 tests in parallel:
Task: "T021 Add failing unit tests for latency chart `1m` time-domain behavior in internal/site/src/components/charts/network-probe-chart.test.tsx"
Task: "T022 Add failing unit tests for one-point and two-point live-session rendering in internal/site/src/components/charts/network-probe-chart.test.tsx"
```

---

## Implementation Strategy

### MVP First (US1 Only)

1. Complete Phase 1 and Phase 2.
2. Complete Phase 3 for US1.
3. Stop and validate: switching node-detail 线路检测 to `1 分钟` starts empty and appends only browser realtime events.

### Incremental Delivery

1. Setup + Foundation: establish unit-test entry point and pure live-session logic boundary.
2. US1: fix the rejected backfill behavior.
3. US2: keep all configured line legends visible while points arrive later.
4. US3: align visual timing and short-session rendering with CPU 使用率.
5. Polish: run unit tests, Biome check, production build, Docker deployment, and browser validation.

### Parallel Team Strategy

1. One agent handles unit-test harness and live-session fixtures.
2. One agent handles hook data semantics for US1.
3. One agent handles grouping/legend chart behavior for US2.
4. One agent handles chart time-domain rendering for US3 after US1 semantics are available.

## Notes

- Do not add backend, agent, database, or HTTP API changes unless validation proves fresh probe results are not being emitted often enough.
- Do not use historical fetches or `created >= switch time` queries to populate node-detail `1 分钟` plotted latency data.
- Keep public dashboard historical behavior unchanged.
- Keep UI text in Chinese for chart labels, tooltip units, and empty states.
