# Tasks: Latency Sample Cadence

**Input**: Design documents from `/specs/005-latency-sample-cadence/`

**Prerequisites**: plan.md, spec.md, research.md, data-model.md, contracts/, quickstart.md

**Tests**: Unit tests are REQUIRED for every behavior change by the project constitution. Tests are listed before implementation tasks in each story phase.

**Organization**: Tasks are grouped by user story so `1 分钟` realtime behavior, `30 分钟` and longer historical behavior, and failure handling can be implemented and validated independently.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel because it touches different files and has no dependency on incomplete tasks in the same phase.
- **[Story]**: Maps to a user story from [spec.md](./spec.md).
- Every task includes exact file paths.

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Reconfirm current implementation surfaces before changing behavior.

- [X] T001 Review current live latency session manager and scheduler entry points in `internal/hub/network_probe_live.go`.
- [X] T002 [P] Review current network probe result query and persistence flow in `internal/hub/network_probes.go`.
- [X] T003 [P] Review current node-detail latency data hook and live-session lifecycle in `internal/site/src/components/routes/system/use-network-probe-data.ts`.
- [X] T004 [P] Review current latency chart grouping and rendering helpers in `internal/site/src/components/charts/network-probe-chart-data.ts` and `internal/site/src/components/charts/network-probe-chart.tsx`.
- [X] T005 [P] Review supported chart range types and API clients in `internal/site/src/types.d.ts` and `internal/site/src/lib/api.ts`.

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Establish shared range-mode helpers and fixtures required by all stories.

**CRITICAL**: No user story work should begin until this phase is complete.

- [X] T006 Add or update latency range-mode helper tests covering `1m` as live and `30m`/longer as historical in `internal/site/src/components/routes/system/network-probe-live-cadence.test.ts`.
- [X] T007 Implement shared latency range-mode helpers for `1m`, `30m`, `1h`, `12h`, `24h`, `1w`, and `30d` in `internal/site/src/components/routes/system/network-probe-live-cadence.ts`.
- [X] T008 [P] Add chart range type coverage for `30m` on node detail latency paths in `internal/site/src/types.d.ts`.
- [X] T009 [P] Add reusable historical latency chart fixture data with high-cadence samples in `internal/site/src/components/charts/network-probe-chart-data.test.ts`.
- [X] T010 [P] Add backend live cadence fixture helpers for active sessions and assignments in `internal/hub/network_probe_live_test.go`.

**Checkpoint**: Foundation ready; user story implementation can begin.

---

## Phase 3: User Story 1 - Smooth One-Minute Latency Lines (Priority: P1) MVP

**Goal**: Switching node detail `线路检测` to `1 分钟` starts from an empty realtime window, creates an authenticated live session, and draws continuous-looking fresh samples for reachable lines.

**Independent Test**: Open a node detail page with three reachable latency lines, switch to `1 分钟`, observe for 60 seconds, and confirm each reachable line receives at least 40 fresh successful samples and renders connected waveform segments.

### Tests for User Story 1

- [X] T011 [P] [US1] Add Go unit tests proving live session cadence remains 1 second and coalesces multiple viewers per system in `internal/hub/network_probe_live_test.go`.
- [X] T012 [P] [US1] Add Go unit tests proving live cadence selects only enabled latency-capable assignments in `internal/hub/network_probes_test.go`.
- [X] T013 [P] [US1] Add frontend unit tests proving `1m` starts with empty latency data and appends only realtime samples in `internal/site/src/components/routes/system/use-network-probe-data.test.ts`.
- [X] T014 [P] [US1] Add frontend unit tests proving `1m` creates, renews, and ends live sessions through the client lifecycle in `internal/site/src/components/routes/system/network-probe-live-cadence.test.ts`.
- [X] T015 [P] [US1] Add chart data tests proving all configured latency line legends remain present before every line has a fresh point in `internal/site/src/components/charts/network-probe-chart-data.test.ts`.

### Implementation for User Story 1

- [X] T016 [US1] Verify and adjust live session TTL, heartbeat tolerance, and one-second cadence constants in `internal/hub/network_probe_live.go`.
- [X] T017 [US1] Verify and adjust live probe execution so active systems are coalesced and assignments do not run once per browser tab in `internal/hub/network_probe_live.go`.
- [X] T018 [US1] Verify and adjust live assignment selection and bounded timeout request construction in `internal/hub/network_probes.go`.
- [X] T019 [US1] Verify and adjust live result persistence so fresh samples enter the existing realtime result stream in `internal/hub/network_probes.go`.
- [X] T020 [US1] Ensure `createNetworkProbeLiveSession`, `renewNetworkProbeLiveSession`, and `endNetworkProbeLiveSession` match `contracts/live-session-api.md` in `internal/site/src/lib/api.ts`.
- [X] T021 [US1] Wire `useNetworkProbeData` so only `range === "1m"` starts live cadence and the data window is reset on entry in `internal/site/src/components/routes/system/use-network-probe-data.ts`.
- [X] T022 [US1] Ensure realtime append logic preserves connected valid line segments without fabricating pre-switch history in `internal/site/src/components/routes/system/network-probe-live-session.ts`.
- [X] T023 [US1] Tune latency chart curve rendering for `1m` so reachable lines appear as continuous waveforms after fresh samples arrive in `internal/site/src/components/charts/network-probe-chart.tsx`.

**Checkpoint**: User Story 1 is functional and independently testable as the MVP.

---

## Phase 4: User Story 2 - Preserve Existing Longer-Range Behavior (Priority: P2)

**Goal**: `30 分钟` remains selectable, and `30 分钟` plus longer ranges load historical data with range-appropriate density instead of reusing the one-minute realtime drawing model.

**Independent Test**: Compare the node detail chart in `1 分钟`, `30 分钟`, and `1 小时`; verify `1 分钟` produces dense fresh samples while `30 分钟` and longer ranges remain available and present historical data at readable density.

### Tests for User Story 2

- [X] T024 [P] [US2] Add frontend unit tests proving the node detail range options include `30m` with Chinese label `30 分钟` in `internal/site/src/components/routes/system/chart-data.test.ts`.
- [X] T025 [P] [US2] Add frontend unit tests proving `30m`, `1h`, `12h`, `24h`, `1w`, and `30d` do not call live session APIs in `internal/site/src/components/routes/system/network-probe-live-cadence.test.ts`.
- [X] T026 [P] [US2] Add frontend unit tests proving `30m` and longer ranges request historical probe results instead of starting from an empty realtime window in `internal/site/src/components/routes/system/use-network-probe-data.test.ts`.
- [X] T027 [P] [US2] Add chart data tests proving historical latency ranges downsample or bucket high-cadence samples to bounded readable point counts in `internal/site/src/components/charts/network-probe-chart-data.test.ts`.
- [X] T028 [P] [US2] Add Go unit tests proving historical result queries accept `30m` and return persisted results for the requested window in `internal/hub/network_probes_test.go`.
- [X] T029 [P] [US2] Add public dashboard regression tests proving public status ranges keep historical behavior and never create live sessions in `internal/hub/public_status_test.go`.

### Implementation for User Story 2

- [X] T030 [US2] Add `30m` to node detail chart range typing without removing existing ranges in `internal/site/src/types.d.ts`.
- [X] T031 [US2] Ensure page-wide range metadata exposes Chinese labels for `1 分钟`, `30 分钟`, `1 小时`, `12 小时`, `24 小时`, `1 周`, and `30 天` in `internal/site/src/lib/utils.ts`.
- [X] T032 [US2] Update `useNetworkProbeData` so `30m` and longer ranges call `getNetworkProbeResults` and never start or renew live sessions in `internal/site/src/components/routes/system/use-network-probe-data.ts`.
- [X] T033 [US2] Implement range-aware historical latency bucketing or thinning in `internal/site/src/components/charts/network-probe-chart-data.ts`.
- [X] T034 [US2] Update latency chart x-axis tick formatting so historical ranges use readable time ticks and do not keep one-minute density in `internal/site/src/components/charts/network-probe-chart.tsx`.
- [X] T035 [US2] Ensure `getNetworkProbeResults` accepts `30m` for authenticated node detail latency requests in `internal/site/src/lib/api.ts`.
- [X] T036 [US2] Ensure backend network probe result range parsing supports `30m` and longer historical windows without live-session state in `internal/hub/network_probes.go`.
- [X] T037 [US2] Ensure public dashboard latency charts continue using historical public status APIs only in `internal/site/src/components/routes/public-status.tsx`.

**Checkpoint**: User Stories 1 and 2 work independently; expanding time range no longer displays overcrowded one-minute traces and `30 分钟` remains available.

---

## Phase 5: User Story 3 - Handle Failures Without Misleading Lines (Priority: P3)

**Goal**: Failed, slow, offline, or unsupported checks remain failure samples or gaps and never become artificial successful latency line segments.

**Independent Test**: Configure one reachable line and one failing line, switch to `1 分钟`, observe for 30 seconds, and confirm the reachable line smooths while the failing line remains failure/gap state.

### Tests for User Story 3

- [X] T038 [P] [US3] Add Go unit tests for live cadence timeout, offline-agent, and unsupported-probe failed result persistence in `internal/hub/network_probes_test.go`.
- [X] T039 [P] [US3] Add Go unit tests preventing overlapping live executions for the same assignment in `internal/hub/network_probe_live_test.go`.
- [X] T040 [P] [US3] Add frontend chart data tests ensuring failed live and historical samples remain null latency values in `internal/site/src/components/charts/network-probe-chart-data.test.ts`.
- [X] T041 [P] [US3] Add frontend rendering tests or focused assertions for Chinese failure labels in `internal/site/src/components/charts/network-probe-chart-data.test.ts`.

### Implementation for User Story 3

- [X] T042 [US3] Verify and adjust per-assignment in-flight guards for live cadence execution in `internal/hub/network_probe_live.go`.
- [X] T043 [US3] Persist live cadence timeout, offline, and unsupported failures through existing failed result helpers in `internal/hub/network_probes.go`.
- [X] T044 [US3] Ensure chart data preparation keeps failed samples as null latency values in both realtime and historical modes in `internal/site/src/components/charts/network-probe-chart-data.ts`.
- [X] T045 [US3] Ensure Chinese failure labels and empty states remain visible for latency failures in `internal/site/src/components/charts/network-probe-chart.tsx`.

**Checkpoint**: All user stories should now be independently functional.

---

## Phase 6: Polish & Cross-Cutting Concerns

**Purpose**: Verification, documentation, and deployment validation across all stories.

- [X] T046 [P] Update validation notes for `1 分钟`, `30 分钟`, and longer historical ranges in `specs/005-latency-sample-cadence/quickstart.md`.
- [X] T047 [P] Review REST API compatibility and auth behavior against `specs/005-latency-sample-cadence/contracts/live-session-api.md`.
- [X] T048 [P] Review chart range behavior against `specs/005-latency-sample-cadence/contracts/latency-chart-ranges.md`.
- [X] T049 Run Go tests for backend/agent changes with `/usr/local/go/bin/go test -tags=testing ./...` from `specs/005-latency-sample-cadence/quickstart.md`.
- [X] T050 Run frontend unit tests with `npm --prefix ./internal/site run test:unit` from `specs/005-latency-sample-cadence/quickstart.md`.
- [X] T051 Run frontend Biome check with `npm --prefix ./internal/site run check` from `specs/005-latency-sample-cadence/quickstart.md`.
- [X] T052 Run frontend production build with `npm --prefix ./internal/site run build` from `specs/005-latency-sample-cadence/quickstart.md`.
- [X] T053 Run `golangci-lint run` if available and record any environment blocker in `specs/005-latency-sample-cadence/quickstart.md`.
- [X] T054 Run `docker compose up -d --build beszel beszel-agent` and confirm containers are healthy using `docker-compose.yml`.
- [X] T055 Use Playwright against `http://127.0.0.1:8090` to validate `1 分钟` 20-second and 60-second waveform density scenarios from `specs/005-latency-sample-cadence/quickstart.md`.
- [X] T056 Use Playwright against `http://127.0.0.1:8090` to validate `30 分钟` remains selectable and historical ranges are readable per `specs/005-latency-sample-cadence/quickstart.md`.
- [X] T057 Use Playwright or result inspection to verify leaving `1 分钟` stops live session renewal and restores normal cadence per `specs/005-latency-sample-cadence/quickstart.md`.

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies.
- **Foundational (Phase 2)**: Depends on Setup and blocks all user stories.
- **User Story 1 (Phase 3)**: Depends on Foundational and is the MVP.
- **User Story 2 (Phase 4)**: Depends on Foundational; can be implemented after US1 data-mode helpers exist and should not require backend live-session changes beyond US1.
- **User Story 3 (Phase 5)**: Depends on Foundational; safest after US1 live runner shape is stable.
- **Polish (Phase 6)**: Depends on selected user stories being complete.

### User Story Dependencies

- **US1 Smooth One-Minute Latency Lines**: Required MVP; no dependency on US2 or US3.
- **US2 Preserve Existing Longer-Range Behavior**: Depends on the shared range-mode helper from Phase 2 and validates separation from US1 live mode.
- **US3 Handle Failures Without Misleading Lines**: Depends on live runner and chart data paths used by US1 and US2.

### Within Each User Story

- Tests must be written before implementation when practical.
- Go state/service tests before Go scheduler or result-query implementation.
- Frontend range-mode tests before wiring `useNetworkProbeData`.
- Chart data tests before chart rendering changes.
- Browser validation after Docker deployment.

## Parallel Opportunities

- Setup review tasks T002, T003, T004, and T005 can run in parallel.
- Foundational fixture/type/helper tasks T008, T009, and T010 can run in parallel after T006 identifies expected helper behavior.
- US1 backend tests T011-T012 can run in parallel with frontend tests T013-T015.
- US2 tests T024-T029 can run in parallel once foundational range helpers exist.
- US3 tests T038-T041 can run in parallel after US1 live runner and US2 chart data paths are stable.
- Polish review tasks T046-T048 can run in parallel with local verification preparation.

## Parallel Example: User Story 1

```bash
Task: "T011 Add Go unit tests proving live session cadence remains 1 second and coalesces multiple viewers per system in internal/hub/network_probe_live_test.go"
Task: "T012 Add Go unit tests proving live cadence selects only enabled latency-capable assignments in internal/hub/network_probes_test.go"
Task: "T013 Add frontend unit tests proving 1m starts with empty latency data and appends only realtime samples in internal/site/src/components/routes/system/use-network-probe-data.test.ts"
Task: "T014 Add frontend unit tests proving 1m creates, renews, and ends live sessions through the client lifecycle in internal/site/src/components/routes/system/network-probe-live-cadence.test.ts"
```

## Parallel Example: User Story 2

```bash
Task: "T024 Add frontend unit tests proving the node detail range options include 30m with Chinese label 30 分钟 in internal/site/src/components/routes/system/chart-data.test.ts"
Task: "T026 Add frontend unit tests proving 30m and longer ranges request historical probe results instead of starting from an empty realtime window in internal/site/src/components/routes/system/use-network-probe-data.test.ts"
Task: "T027 Add chart data tests proving historical latency ranges downsample or bucket high-cadence samples to bounded readable point counts in internal/site/src/components/charts/network-probe-chart-data.test.ts"
Task: "T028 Add Go unit tests proving historical result queries accept 30m and return persisted results for the requested window in internal/hub/network_probes_test.go"
```

## Parallel Example: User Story 3

```bash
Task: "T038 Add Go unit tests for live cadence timeout, offline-agent, and unsupported-probe failed result persistence in internal/hub/network_probes_test.go"
Task: "T039 Add Go unit tests preventing overlapping live executions for the same assignment in internal/hub/network_probe_live_test.go"
Task: "T040 Add frontend chart data tests ensuring failed live and historical samples remain null latency values in internal/site/src/components/charts/network-probe-chart-data.test.ts"
```

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Complete Phase 1 and Phase 2.
2. Complete Phase 3 for US1.
3. Validate with one node and three reachable latency lines for 60 seconds.
4. Confirm each reachable line receives at least 40 fresh samples and renders connected waveform segments.

### Incremental Delivery

1. US1 delivers the high-cadence `1 分钟` live waveform.
2. US2 restores `30 分钟` and keeps historical ranges readable without live cadence.
3. US3 hardens failure, timeout, offline, and duplicate prevention behavior.
4. Phase 6 runs full quality gates and browser validation.

### Notes

- Do not lower the normal configured probe minimum interval to 1 second.
- Do not add a new agent websocket action unless implementation proves the existing `RunNetworkProbe` request path cannot satisfy the contract.
- Do not fabricate smooth lines in the browser by interpolation.
- Preserve existing Chinese labels and the `1 分钟` start-empty behavior.
