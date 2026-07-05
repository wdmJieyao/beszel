# Tasks: Public Probe Visibility and Refresh Commands

**Input**: Design documents from `/specs/007-public-probe-visibility/`

**Prerequisites**: plan.md, spec.md, research.md, data-model.md, contracts/, quickstart.md

**Tests**: Unit tests are REQUIRED for every behavior change by the project constitution. Add focused Go and frontend regression tests before implementation tasks where practical.

**Organization**: Tasks are grouped by user story to enable independent implementation and testing of each story.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel because it touches different files or has no dependency on incomplete tasks
- **[Story]**: User story label for story-specific tasks only
- Every task includes an exact file path

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Confirm affected surfaces and current public-visibility ownership before changing behavior.

- [X] T001 Inspect current public dashboard settings API, public filtering, and visibility record helpers in `internal/hub/public_status.go`
- [X] T002 [P] Inspect current probe serialization and legacy `public_visible` handling in `internal/hub/network_probes.go`
- [X] T003 [P] Inspect current public dashboard settings UI and payload flow in `internal/site/src/components/routes/settings/public-status.tsx`, `internal/site/src/lib/api.ts`, and `internal/site/src/types.d.ts`
- [X] T004 [P] Inspect current probe settings public toggle and copy text in `internal/site/src/components/routes/settings/network-probes.tsx`
- [X] T005 [P] Inspect generated Docker run command builders and copy surfaces in `internal/site/src/components/install-dropdowns.tsx`, `internal/site/src/components/add-system.tsx`, and `internal/site/src/components/routes/settings/tokens-fingerprints.tsx`

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Add the shared persistence/API shape needed by all user stories.

**CRITICAL**: No user story work can begin until this phase is complete.

- [X] T006 Add failing Go regression tests for public system probe selection serialization and validation in `internal/hub/public_status_test.go`
- [X] T007 [P] Add failing Go migration regression tests for preserving existing public probe/system visibility without widening exposure in `internal/hub/public_status_test.go`
- [X] T008 [P] Add failing frontend unit tests for public system probe selection payloads and empty-default behavior in `internal/site/src/components/routes/settings/public-status.test.ts`
- [X] T009 [P] Add failing frontend unit tests for Docker run refresh command generation in `internal/site/src/components/install-dropdowns.test.ts`
- [X] T010 Create PocketBase migration adding per-VPS public probe selection to `public_system_visibility` and seeding legacy visibility in `internal/migrations/add_public_probe_visibility.go`
- [X] T011 Extend public system visibility models, normalization, and persistence to include `publicProbeIds` in `internal/hub/public_status.go`
- [X] T012 Update authenticated public systems API request/response types to carry `publicProbeIds` in `internal/site/src/types.d.ts` and `internal/site/src/lib/api.ts`

**Checkpoint**: Public system visibility rows can store per-VPS selected probe IDs and the admin API/types support reading and updating them.

---

## Phase 3: User Story 1 - Configure Public Probe Visibility Per VPS (Priority: P1) 🎯 MVP

**Goal**: Administrators manage anonymous latency-line visibility per VPS from the public dashboard settings page only.

**Independent Test**: Enable a VPS for public display, select a custom set of probe lines in public dashboard settings, open `/` anonymously, and confirm only the selected lines appear for that VPS while the normal VPS edit page does not duplicate the control.

### Tests for User Story 1

- [X] T013 [P] [US1] Add Go regression test proving anonymous public status only emits probe summaries selected for a specific public VPS in `internal/hub/public_status_test.go`
- [X] T014 [P] [US1] Add Go regression test proving unselected probe names, targets, and series remain hidden for a public VPS in `internal/hub/public_status_test.go`
- [X] T015 [P] [US1] Add frontend unit test proving public status settings support zero, one, and many selected probe IDs plus select-all behavior in `internal/site/src/components/routes/settings/public-status.test.ts`
- [X] T016 [P] [US1] Add frontend unit test proving probe settings no longer expose the public visibility toggle in `internal/site/src/components/routes/settings/network-probes-utils.test.ts`

### Implementation for User Story 1

- [X] T017 [US1] Update `GET /api/beszel/public/systems` and `PATCH /api/beszel/public/systems/{systemId}` to read and write `publicProbeIds` in `internal/hub/public_status.go`
- [X] T018 [US1] Update anonymous public dashboard probe filtering to require per-VPS selection plus effective coverage in `internal/hub/public_status.go`
- [X] T019 [US1] Update public settings UI to edit per-VPS probe selection, default new public VPS rows to empty selection, and provide a select-all action in `internal/site/src/components/routes/settings/public-status.tsx`
- [X] T020 [US1] Remove the probe-level public visibility toggle from the probe settings experience in `internal/site/src/components/routes/settings/network-probes.tsx`
- [X] T021 [US1] Update admin/public API helpers and local state shaping for `publicProbeIds` in `internal/site/src/lib/api.ts` and `internal/site/src/types.d.ts`

**Checkpoint**: User Story 1 is fully functional and testable as the MVP.

---

## Phase 4: User Story 2 - Preserve Existing Production Visibility During Upgrade (Priority: P1)

**Goal**: Production upgrades preserve exactly the previously visible public probe/system combinations without widening exposure.

**Independent Test**: Start from production-like data with public VPS rows and legacy-public probes, run the migration, and verify the resulting per-VPS `publicProbeIds` preserve previous visibility exactly and remain stable on rerun.

### Tests for User Story 2

- [X] T022 [P] [US2] Add Go regression test proving migration seeds only previously visible probe/system pairs and does not copy one public probe to every public VPS in `internal/hub/public_status_test.go`
- [X] T023 [P] [US2] Add Go regression test proving migration reruns are idempotent and do not duplicate or reset seeded `publicProbeIds` in `internal/hub/public_status_test.go`
- [X] T024 [P] [US2] Add Go regression test proving newly public VPS rows created after migration default to empty `publicProbeIds` in `internal/hub/public_status_test.go`

### Implementation for User Story 2

- [X] T025 [US2] Implement migration seeding logic from legacy probe-level public visibility into per-VPS selection in `internal/migrations/add_public_probe_visibility.go`
- [X] T026 [US2] Add runtime compatibility handling so legacy `public_visible` is treated as migration seed/compatibility input rather than primary anonymous filter ownership in `internal/hub/public_status.go` and `internal/hub/network_probes.go`
- [X] T027 [US2] Ensure public settings row creation/update preserves empty defaults for newly public VPS rows in `internal/hub/public_status.go`
- [X] T028 [US2] Update public dashboard quickstart validation notes to cover production-style migration and new-public-VPS defaults in `specs/007-public-probe-visibility/quickstart.md`

**Checkpoint**: Existing deployments upgrade safely without widening public probe exposure.

---

## Phase 5: User Story 3 - Generate Refresh-Safe Docker Run Commands (Priority: P2)

**Goal**: All generated Docker run commands refresh containers and images predictably when rerun.

**Independent Test**: Copy any generated Docker run command, run it on a Docker host with an existing same-name container and cached image, and confirm the command removes the old container, refreshes the image, and starts the replacement container.

### Tests for User Story 3

- [X] T029 [P] [US3] Add frontend unit test proving generated Docker run commands remove old containers, remove old images best-effort, pull new images, and preserve runtime args in `internal/site/src/components/install-dropdowns.test.ts`
- [X] T030 [P] [US3] Add frontend unit test proving all Docker run copy surfaces reuse the shared refresh command builder in `internal/site/src/components/install-dropdowns.test.ts`

### Implementation for User Story 3

- [X] T031 [US3] Refactor Docker run command generation into a shared refresh-safe builder in `internal/site/src/components/install-dropdowns.tsx`
- [X] T032 [US3] Update agent install dropdown copy actions to use the refreshed Docker run command helper in `internal/site/src/components/add-system.tsx` and `internal/site/src/components/routes/settings/tokens-fingerprints.tsx`
- [X] T033 [US3] Ensure any current or future panel Docker run surface can call the same helper by keeping the builder generic in `internal/site/src/components/install-dropdowns.tsx`

**Checkpoint**: Generated Docker run commands are rerunnable refresh commands across all UI copy surfaces.

---

## Phase 6: Polish & Cross-Cutting Concerns

**Purpose**: Validate end-to-end behavior and clean up any remaining dual-ownership assumptions.

- [X] T034 [P] Search for remaining probe-level public visibility UI strings/usages that should no longer control anonymous exposure in `internal/site/src` and `internal/hub`
- [X] T035 [P] Search for remaining anonymous public dashboard probe filters that still rely on probe-level `public_visible` alone in `internal/hub`
- [X] T036 [P] Update translated copy strings impacted by removal of the probe-level public toggle and addition of per-VPS selection controls in `internal/site/src/locales/zh-CN/zh-CN.po` and other locale catalogs touched by extraction/build
- [ ] T037 Run backend unit tests with `go test -tags=testing ./...` from `go.mod`
- [ ] T038 Run Go lint/static checks with `golangci-lint run` from `go.mod`
- [X] T039 Run frontend unit tests with `npm --prefix ./internal/site run test:unit` from `internal/site/package.json`
- [X] T040 Run frontend Biome/static checks with `npm --prefix ./internal/site run check` from `internal/site/package.json`
- [X] T041 Run frontend production build with `npm --prefix ./internal/site run build` from `internal/site/package.json`
- [X] T042 Review REST API compatibility against `specs/007-public-probe-visibility/contracts/public-system-api.md`
- [ ] T043 Run the end-to-end validation scenarios in `specs/007-public-probe-visibility/quickstart.md`
- [ ] T044 If pushed to GitHub or Docker image inputs changed, wait for the `Make docker images` GitHub Actions workflow and verify expected GHCR `edge` tags before reporting success

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies.
- **Foundational (Phase 2)**: Depends on Setup completion and blocks all user stories.
- **User Story 1 (Phase 3)**: Depends on Foundational completion and is the MVP.
- **User Story 2 (Phase 4)**: Depends on Foundational completion and should follow alongside US1 because migration safety relies on the same persistence/API shape.
- **User Story 3 (Phase 5)**: Depends on Foundational completion only for shared verification flow; it is otherwise independent from the public visibility migration.
- **Polish (Phase 6)**: Depends on all desired user stories being complete.

### User Story Dependencies

- **US1 (P1)**: No dependency on other user stories after foundation.
- **US2 (P1)**: Depends on the persistence/API work from the foundation and validates migration-safe behavior for US1.
- **US3 (P2)**: No dependency on US1/US2 code paths; can proceed after foundation in parallel with backend visibility work.

### Within Each User Story

- Regression tests should be written first and fail before implementation when practical.
- Backend storage/API changes precede frontend settings changes that depend on them.
- Public filtering logic should be updated before anonymous-dashboard validation.
- Docker run builder refactor should precede copy-surface rewiring.

### Parallel Opportunities

- Setup inspection tasks T002-T005 can run in parallel.
- Foundational test tasks T007-T009 can run in parallel before implementation.
- US1 test tasks T013-T016 can run in parallel.
- US2 regression tests T022-T024 can run in parallel.
- US3 command-generation tests T029-T030 can run in parallel.
- Polish searches T034-T036 can run in parallel before final quality gates.

---

## Parallel Example: User Story 1

```bash
Task: "T013 Add public-system selected-probe filtering test in internal/hub/public_status_test.go"
Task: "T014 Add hidden-unselected-probe regression test in internal/hub/public_status_test.go"
Task: "T015 Add public status settings selection payload test in internal/site/src/components/routes/settings/public-status.test.ts"
Task: "T016 Add probe settings toggle-removal test in internal/site/src/components/routes/settings/network-probes-utils.test.ts"
```

## Parallel Example: User Story 2

```bash
Task: "T022 Add migration seeding regression test in internal/hub/public_status_test.go"
Task: "T023 Add idempotent rerun migration test in internal/hub/public_status_test.go"
Task: "T024 Add newly-public-empty-default regression test in internal/hub/public_status_test.go"
```

## Parallel Example: User Story 3

```bash
Task: "T029 Add refresh-safe docker run builder test in internal/site/src/components/install-dropdowns.test.ts"
Task: "T030 Add shared copy-surface Docker run helper test in internal/site/src/components/install-dropdowns.test.ts"
```

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Complete Phase 1 and Phase 2.
2. Complete US1 tasks T013-T021.
3. Run the US1-focused backend/frontend tests.
4. Validate public per-VPS selection behavior from `quickstart.md`.

### Incremental Delivery

1. Deliver US1 to move public probe ownership to per-VPS settings.
2. Deliver US2 to make the production migration safe and idempotent.
3. Deliver US3 to make generated Docker run commands refresh-safe.
4. Finish polish, quality gates, contract review, and GHCR verification.

### Parallel Team Strategy

1. Complete Setup and Foundational work together.
2. After foundation:
   - Developer A: US1 backend/frontend public visibility ownership
   - Developer B: US2 migration and compatibility safety
   - Developer C: US3 Docker run command refresh helper
3. Rejoin for polish and final verification.
