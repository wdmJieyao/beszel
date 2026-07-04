# Tasks: Global Probe Binding Regression Fix

**Input**: Design documents from `/specs/006-global-probe-binding/`

**Prerequisites**: plan.md, spec.md, research.md, data-model.md, contracts/, quickstart.md

**Tests**: Unit tests are REQUIRED for every behavior change by the project constitution. Write focused regression tests before implementation tasks where practical.

**Organization**: Tasks are grouped by user story so each story can be implemented and tested independently after the shared foundation is complete.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel because it touches different files or does not depend on incomplete tasks
- **[Story]**: User story label for story phases only
- Every task includes an exact file path

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Confirm affected surfaces and current snapshot-binding behavior before changing code.

- [X] T001 Inspect current probe schema, assignment writes, scheduler, live-session resolution, and result APIs in `internal/hub/network_probes.go`
- [X] T002 [P] Inspect public dashboard probe summary filtering and public visibility behavior in `internal/hub/public_status.go`
- [X] T003 [P] Inspect migration patterns for PocketBase collection field changes in `internal/migrations/add_public_status_probes.go`
- [X] T004 [P] Inspect frontend probe settings payload creation and current all-system snapshot behavior in `internal/site/src/components/routes/settings/network-probes.tsx`
- [X] T005 [P] Inspect frontend node-detail probe filtering and grouping behavior in `internal/site/src/components/routes/system/use-network-probe-data.ts`
- [X] T006 [P] Inspect frontend API/types shape for network probes in `internal/site/src/lib/api.ts` and `internal/site/src/types.d.ts`

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Add durable coverage scope and shared resolver behavior that every user story depends on.

**CRITICAL**: No user story work can begin until this phase is complete.

- [X] T007 Add failing Go regression tests for scope defaulting, legacy all-system classification, and response normalization in `internal/hub/network_probes_test.go`
- [X] T008 Add failing Go regression tests for effective probe coverage deduplication and disabled-probe exclusion in `internal/hub/network_probes_test.go`
- [X] T009 Create PocketBase migration adding `scope` to `network_probes` and classifying existing probes in `internal/migrations/add_network_probe_scope.go`
- [X] T010 Add `global` and `fixed` scope constants, request/response fields, and config mapping for network probes in `internal/hub/network_probes.go`
- [X] T011 Implement probe scope normalization for create, patch, and legacy records in `internal/hub/network_probes.go`
- [X] T012 Implement shared effective coverage resolver for global and fixed probe/system pairs in `internal/hub/network_probes.go`
- [X] T013 Update network probe list/create/patch API serialization so global probes return `scope: "global"` and `systems: []` in `internal/hub/network_probes.go`
- [X] T014 Update TypeScript network probe types to include `scope: "global" | "fixed"` in `internal/site/src/types.d.ts`
- [X] T015 Update network probe API client request/response handling for the additive `scope` field in `internal/site/src/lib/api.ts`

**Checkpoint**: Durable scope exists and shared resolver semantics are available for scheduler, live checks, API consumers, and public summaries.

---

## Phase 3: User Story 1 - Global Probes Apply to Future Machines (Priority: P1) MVP

**Goal**: A probe saved without fixed machine selection remains global and automatically covers machines added later.

**Independent Test**: Add one machine, create a probe with no fixed machine selection, add a second machine, and verify both machines are eligible for the same global probe without editing the probe.

### Tests for User Story 1

- [X] T016 [P] [US1] Add Go scheduler regression test proving a newly added system is due for an existing global probe without an assignment row in `internal/hub/network_probes_test.go`
- [X] T017 [P] [US1] Add Go live-session regression test proving global TCPing/ICMP probes are returned for a system without an assignment row in `internal/hub/network_probes_test.go`
- [X] T018 [P] [US1] Add frontend unit test proving global probes are treated as assigned to the active system in `internal/site/src/components/routes/system/use-network-probe-data.test.ts`
- [X] T019 [P] [US1] Add frontend settings payload test proving empty/all-machine selection sends `scope: "global"` and `systems: []` in `internal/site/src/components/routes/settings/network-probes-utils.test.ts`

### Implementation for User Story 1

- [X] T020 [US1] Update scheduled probe execution to resolve global probes dynamically for all eligible systems in `internal/hub/network_probes.go`
- [X] T021 [US1] Update latest-result lookup and due-check calculation to work for generated global probe/system pairs in `internal/hub/network_probes.go`
- [X] T022 [US1] Update live network probe assignment resolution to include global TCPing and ICMP probes for the active system in `internal/hub/network_probes.go`
- [X] T023 [US1] Update frontend node-detail probe filtering to include `probe.scope === "global"` for accessible systems in `internal/site/src/components/routes/system/use-network-probe-data.ts`
- [X] T024 [US1] Update frontend settings submit logic so the all-machine/default choice persists global scope instead of expanding to `systems.map(...)` in `internal/site/src/components/routes/settings/network-probes.tsx`
- [X] T025 [US1] Update node-detail grouping to render global probes with pending/no-history state until first results arrive in `internal/site/src/components/routes/system/network-probe-groups.ts`

**Checkpoint**: User Story 1 is fully functional and testable as the MVP.

---

## Phase 4: User Story 2 - Fixed-Machine Probes Stay Scoped (Priority: P2)

**Goal**: Explicit fixed-machine probes remain limited to selected machines while scope transitions behave predictably.

**Independent Test**: Create one global probe and one fixed-machine probe, add a new machine, and verify the new machine gets only the global probe.

### Tests for User Story 2

- [X] T026 [P] [US2] Add Go regression test proving a fixed probe does not cover a later unselected system in `internal/hub/network_probes_test.go`
- [X] T027 [P] [US2] Add Go regression test for fixed-to-global and global-to-fixed patch transitions preserving historical results in `internal/hub/network_probes_test.go`
- [X] T028 [P] [US2] Add frontend unit test proving fixed probes still require `systems.includes(activeSystemId)` in `internal/site/src/components/routes/system/use-network-probe-data.test.ts`
- [X] T029 [P] [US2] Add frontend settings payload test proving explicit selected machines send `scope: "fixed"` with selected system IDs in `internal/site/src/components/routes/settings/network-probes-utils.test.ts`

### Implementation for User Story 2

- [X] T030 [US2] Update assignment replacement so fixed scope writes selected assignments and global scope deletes or ignores assignment rows in `internal/hub/network_probes.go`
- [X] T031 [US2] Update create and patch validation so `ensureSystemsVisibleToAuth` validates system IDs only for fixed scope in `internal/hub/network_probes.go`
- [X] T032 [US2] Update patch merge semantics so omitted `scope` and omitted `systems` do not accidentally flip a probe between global and fixed in `internal/hub/network_probes.go`
- [X] T033 [US2] Update frontend edit state so global-to-fixed and fixed-to-global transitions preserve the administrator's explicit choice in `internal/site/src/components/routes/settings/network-probes.tsx`
- [X] T034 [US2] Verify historical result queries remain keyed by `probe_id` and `system_id` without rewriting result rows in `internal/hub/network_probes.go`

**Checkpoint**: User Stories 1 and 2 both work independently and together.

---

## Phase 5: User Story 3 - Coverage Is Understandable to Administrators (Priority: P3)

**Goal**: Administrators can clearly tell whether each probe applies to all machines or only fixed machines, and public views remain safe.

**Independent Test**: Open probe settings and verify each probe clearly shows either all-machine/global coverage or fixed-machine coverage; open the public dashboard and verify only public-safe global results appear.

### Tests for User Story 3

- [X] T035 [P] [US3] Add Go public dashboard regression test proving public-visible global probes appear for public systems in `internal/hub/public_status_test.go`
- [X] T036 [P] [US3] Add Go public dashboard regression test proving global probes do not expose private systems or hidden probes in `internal/hub/public_status_test.go`
- [X] T037 [P] [US3] Add frontend settings display test for global and fixed scope labels in `internal/site/src/components/routes/settings/network-probes-utils.test.ts`

### Implementation for User Story 3

- [X] T038 [US3] Update probe settings table/editor labels to show `全部可用节点` for global probes and fixed-node counts/names for fixed probes in `internal/site/src/components/routes/settings/network-probes.tsx`
- [X] T039 [US3] Update public dashboard probe summary resolution to include global probe coverage for public-visible systems in `internal/hub/public_status.go`
- [X] T040 [US3] Update public dashboard filtering to exclude private systems, non-public probes, and unauthorized global coverage in `internal/hub/public_status.go`
- [X] T041 [US3] Update frontend node-detail and public-facing display copy for global pending/no-history probe states in `internal/site/src/components/routes/system/network-probe-groups.ts`

**Checkpoint**: All user stories are independently functional and administrator-facing behavior is understandable.

---

## Phase 6: Polish & Cross-Cutting Concerns

**Purpose**: Validate the feature end to end and remove old snapshot-assignment assumptions.

- [X] T042 [P] Search for remaining snapshot-all behavior and assignment-only filters such as `systems.map((system) => system.id)` and `probe.systems.includes` in `internal/site/src`
- [X] T043 [P] Search for remaining assignment-only network probe queries that should use effective coverage in `internal/hub`
- [X] T044 [P] Update quickstart validation notes if implementation details change in `specs/006-global-probe-binding/quickstart.md`
- [ ] T045 Run backend unit tests with `go test -tags=testing ./...` using `go.mod`
- [X] T046 Run Go lint/static checks with `golangci-lint run` when available using `go.mod`
- [X] T047 Run frontend unit tests with `npm --prefix ./internal/site run test:unit` using `internal/site/package.json`
- [X] T048 Run frontend Biome/static checks with `npm --prefix ./internal/site run check` using `internal/site/biome.json`
- [X] T049 Run frontend production build with `npm --prefix ./internal/site run build` using `internal/site/vite.config.ts`
- [X] T050 Review REST API compatibility against `specs/006-global-probe-binding/contracts/network-probe-api.md`
- [X] T051 If pushed to GitHub or Docker image inputs changed, wait for the `Make docker images` GitHub Actions workflow in `.github/workflows/docker-images.yml` and verify expected GHCR tags before reporting success in `specs/006-global-probe-binding/quickstart.md`

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies.
- **Foundational (Phase 2)**: Depends on Setup completion and blocks all user stories.
- **User Story 1 (Phase 3)**: Depends on Foundational completion and is the MVP.
- **User Story 2 (Phase 4)**: Depends on Foundational completion; can run after or alongside US1 but must preserve US1 behavior.
- **User Story 3 (Phase 5)**: Depends on Foundational completion; public summary tasks also depend on the effective coverage resolver from Phase 2.
- **Polish (Phase 6)**: Depends on all desired user stories being complete.

### User Story Dependencies

- **US1 (P1)**: No dependency on US2 or US3 after foundation.
- **US2 (P2)**: No dependency on US3; validates fixed-scope behavior alongside global behavior.
- **US3 (P3)**: Depends on shared scope fields and resolver; otherwise independently testable.

### Within Each User Story

- Tests should be written first and should fail before implementation when practical.
- Backend resolver/schema work must precede scheduler/live/public behavior that consumes it.
- Frontend type/API updates must precede component changes that read or write `scope`.
- A story is complete only after its tests and independent validation criteria pass.

### Parallel Opportunities

- Setup inspection tasks T002-T006 can run in parallel.
- Foundational tests T007-T008 can run before implementation and in parallel with migration drafting T009.
- US1 test tasks T016-T019 can run in parallel.
- US2 test tasks T026-T029 can run in parallel.
- US3 public backend tests T035-T036 and frontend display test T037 can run in parallel.
- Polish searches T042-T044 can run in parallel before final quality gates.

---

## Parallel Example: User Story 1

```bash
# Backend regression tests:
Task: "T016 Add Go scheduler regression test in internal/hub/network_probes_test.go"
Task: "T017 Add Go live-session regression test in internal/hub/network_probes_test.go"

# Frontend regression tests:
Task: "T018 Add node-detail filtering test in internal/site/src/components/routes/system/use-network-probe-data.test.ts"
Task: "T019 Add settings payload test in internal/site/src/components/routes/settings/network-probes-utils.test.ts"
```

## Parallel Example: User Story 2

```bash
Task: "T026 Add fixed-probe future-system regression test in internal/hub/network_probes_test.go"
Task: "T028 Add frontend fixed-probe filtering test in internal/site/src/components/routes/system/use-network-probe-data.test.ts"
Task: "T029 Add frontend fixed-scope payload test in internal/site/src/components/routes/settings/network-probes-utils.test.ts"
```

## Parallel Example: User Story 3

```bash
Task: "T035 Add public-visible global probe test in internal/hub/public_status_test.go"
Task: "T036 Add private/hidden public safety test in internal/hub/public_status_test.go"
Task: "T037 Add frontend settings scope label test in internal/site/src/components/routes/settings/network-probes-utils.test.ts"
```

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Complete Phase 1 and Phase 2.
2. Complete US1 tasks T016-T025.
3. Run the US1-focused backend and frontend tests.
4. Manually validate the quickstart global future-machine scenario.

### Incremental Delivery

1. Deliver US1 to fix the severe regression for global probes covering future machines.
2. Deliver US2 to protect explicit fixed-machine behavior and scope transitions.
3. Deliver US3 to make scope understandable in settings and preserve public dashboard safety.
4. Complete Polish tasks and all quality gates.

### Final Validation

1. `go test -tags=testing ./...`
2. `golangci-lint run` when available
3. `npm --prefix ./internal/site run test:unit`
4. `npm --prefix ./internal/site run check`
5. `npm --prefix ./internal/site run build`
6. If pushed to GitHub or image inputs changed, wait for GHCR workflow success and verify expected image tags before reporting completion.
