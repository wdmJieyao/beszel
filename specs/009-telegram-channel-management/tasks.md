# Tasks: Telegram Channel Management Improvements

**Input**: Design documents from `/specs/009-telegram-channel-management/`

**Prerequisites**: plan.md, spec.md, research.md, data-model.md, contracts/, quickstart.md

**Tests**: Unit and integration tests are REQUIRED and MUST be written before implementation for every behavior change.

**Organization**: Tasks are grouped by user story so each increment has an explicit independent acceptance path.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel because it uses different files and has no dependency on another incomplete task
- **[Story]**: Maps the task to a user story from spec.md
- Every task names the exact file or files it changes

## Phase 1: Setup (Shared Test Infrastructure)

**Purpose**: Establish a clean baseline and reusable fixtures without changing runtime behavior

- [X] T001 Run the existing Go and frontend unit baselines and record any pre-existing failures in `specs/009-telegram-channel-management/quickstart.md`
- [X] T002 [P] Add reusable Telegram channel and notification-policy record fixtures in `internal/hub/telegram_policy_test_helpers_test.go`
- [X] T003 [P] Add reusable frontend channel, policy, and 500-node fixtures in `internal/site/src/components/routes/settings/telegram-utils.test.ts`

---

## Phase 2: Foundational (Blocking Channel-Policy Model)

**Purpose**: Add the production-safe policy model and compatibility boundary required by every story

**CRITICAL**: Complete this phase before user-story implementation.

### Tests First

- [X] T004 [P] Add failing migration tests for additive policy collection creation, one-policy-per-destination backfill, empty/all mapping, selected-node mapping, and idempotent restart behavior in `internal/hub/telegram_policy_migration_test.go`
- [X] T005 [P] Add failing policy store and validation tests for parent relation, per-channel name uniqueness, scope modes, valid node references, category vocabulary, and cascade behavior in `internal/hub/telegram_policies_test.go`

### Foundation Implementation

- [X] T006 Implement the additive `telegram_notification_policies` collection and idempotent legacy destination backfill in `internal/migrations/add_telegram_notification_policies.go`
- [X] T007 Define channel/policy constants, records, request/response types, scope modes, and validation vocabulary in `internal/hub/telegram_types.go`
- [X] T008 Implement policy list/find/upsert/delete helpers and deterministic default-policy compatibility helpers in `internal/hub/telegram_store.go` and `internal/hub/telegram_policies.go`
- [X] T009 Add channel/policy and staged Bot-test frontend type declarations while retaining deprecated inline scope fields in `internal/site/src/types.d.ts`
- [X] T010 Run the focused migration/store suite and record the passing command in `specs/009-telegram-channel-management/quickstart.md`

**Checkpoint**: Existing destinations are preserved and each has exactly one canonical default policy.

---

## Phase 3: User Story 1 - Reliably Manage Bot And Channels (Priority: P1) MVP

**Goal**: Make valid Bot verification reliable and make channel deletion visible, confirmed, and safe.

**Independent Test**: Verify a valid saved or unsaved Token with separate credential/menu results, then cancel and confirm deletion of a channel while proving Bot settings and unrelated channels remain.

### Tests for User Story 1

- [X] T011 [P] [US1] Add a failing transport regression test for Telegram methods returning scalar boolean success envelopes in `internal/hub/telegram_transport_test.go`
- [X] T012 [P] [US1] Add failing handler tests for saved-versus-unsaved Token selection, staged credential/menu outcomes, sanitization, and no implicit Token save in `internal/hub/telegram_settings_test.go`
- [X] T013 [P] [US1] Add failing deletion tests for channel-policy cascade, missing channel, retained Bot settings, retained sibling channels, and concurrent stale deletion in `internal/hub/telegram_destinations_test.go`
- [X] T014 [P] [US1] Add failing frontend tests for staged Bot result formatting, entered-versus-saved Token test payloads, masked Chat ID confirmation text, and delete state updates in `internal/site/src/components/routes/settings/telegram-utils.test.ts`

### Implementation for User Story 1

- [X] T015 [US1] Make ignored Telegram API results accept scalar/object/array success values while still validating `ok` and sanitized errors in `internal/hub/telegram_transport.go`
- [X] T016 [US1] Return staged credential and command-menu verification results without persisting an entered Token in `internal/hub/telegram_settings.go` and `internal/hub/telegram_types.go`
- [X] T017 [US1] Make channel deletion transactionally remove child policies only and return stable not-found/error behavior in `internal/hub/telegram_destinations.go`
- [X] T018 [P] [US1] Update staged Bot-test and delete response clients in `internal/site/src/lib/api.ts` and `internal/site/src/types.d.ts`
- [X] T019 [US1] Display credential and command-menu verification stages with actionable sanitized errors in `internal/site/src/components/routes/settings/telegram-destinations.tsx`
- [X] T020 [US1] Replace the undiscoverable trash-only action with labelled tooltip/accessible text and a channel/policy-count confirmation dialog in `internal/site/src/components/routes/settings/telegram-destinations.tsx`
- [X] T021 [US1] Run the US1 focused tests and execute the Bot verification and deletion scenarios from `specs/009-telegram-channel-management/quickstart.md`

**Checkpoint**: US1 works independently as the MVP; a valid Bot no longer reports the boolean-result parsing error and channel deletion is explicit and safe.

---

## Phase 4: User Story 2 - Multiple Delivery Policies Per Chat (Priority: P1)

**Goal**: Keep one channel per Chat ID while allowing multiple named routing policies with one delivery per matching alert.

**Independent Test**: Create two policies for one Chat ID, trigger overlapping and non-overlapping alerts, verify one delivery per alert, delete one policy, and verify channel authorization plus sibling policies remain.

### Tests for User Story 2

- [X] T022 [P] [US2] Add failing REST contract tests for channel uniqueness conflict metadata, nested policy CRUD, ownership checks, validation status codes, and legacy inline-scope compatibility in `internal/hub/telegram_policies_api_test.go`
- [X] T023 [P] [US2] Add failing delivery tests for OR matching, disabled policies, overlap de-duplication, zero-policy channels, channel mute/enable state, and one send per Chat ID in `internal/hub/telegram_delivery_test.go`
- [X] T024 [P] [US2] Add failing backup tests for notifications section v2 export/preview/restore, v1 inline-scope conversion, target-only policy preservation, rollback, and idempotency in `internal/hub/config_backup_notifications_test.go`
- [X] T025 [P] [US2] Add failing frontend tests for channel and multi-policy payload normalization, validation errors, default-policy compatibility, and duplicate Chat ID conflict handling in `internal/site/src/components/routes/settings/telegram-utils.test.ts`

### Implementation for User Story 2

- [X] T026 [US2] Implement nested policy list/create/update/delete handlers with parent ownership validation in `internal/hub/telegram_policies.go`
- [X] T027 [US2] Register administrator-only nested policy routes and preserve existing destination routes in `internal/hub/api.go`
- [X] T028 [US2] Return `409 Conflict` with the existing destination ID for duplicate Chat IDs and map legacy inline scopes to the default policy transactionally in `internal/hub/telegram_destinations.go`
- [X] T029 [US2] Load channel policies in bounded queries, apply OR semantics, group matches by channel, and enqueue at most one delivery per Chat ID in `internal/hub/telegram_delivery.go` and `internal/hub/telegram_store.go`
- [X] T030 [US2] Add notifications section v2 channel/policy DTOs, section-specific versioning, v2 export, v1/v2 preview, and transactional restore in `internal/hub/config_backup_types.go`, `internal/hub/config_backup_schema.go`, `internal/hub/config_backup_sources.go`, and `internal/hub/config_backup_restore.go`
- [X] T031 [US2] Include channel and policy preserve/create/update/conflict decisions in backup preview accounting in `internal/hub/config_backup_restore.go`
- [X] T032 [P] [US2] Add nested policy API clients and compatibility response handling in `internal/site/src/lib/api.ts` and `internal/site/src/types.d.ts`
- [X] T033 [US2] Refactor the Telegram settings surface into channel-level fields and a per-channel policy list/editor in `internal/site/src/components/routes/settings/telegram-destinations.tsx`
- [X] T034 [US2] Add visible create/edit/enable/delete policy actions and ensure deleting one policy keeps its channel and siblings in `internal/site/src/components/routes/settings/telegram-destinations.tsx`
- [X] T035 [US2] Run US2 backend/frontend tests and execute the overlapping-policy plus backup v1/v2 scenarios in `specs/009-telegram-channel-management/quickstart.md`

**Checkpoint**: One Chat ID has one authorization/health record and any number of independently managed policies, with no duplicate alert delivery.

---

## Phase 5: User Story 3 - Understand And Select Node Scope (Priority: P2)

**Goal**: Make all-node/future-node behavior explicit and make selected-node management efficient for large node inventories.

**Independent Test**: Save/reload both scope modes, verify future nodes join all-node policies, use search/select-all/clear with 500 nodes, and prove selected-empty cannot silently become all-node.

### Tests for User Story 3

- [X] T036 [P] [US3] Add failing backend tests for explicit all/selected persistence, future-node matching, selected-empty rejection, unknown node rejection, and compatibility mapping in `internal/hub/telegram_policies_test.go`
- [X] T037 [P] [US3] Add failing frontend tests for mode switching, select-all-current, clear-all, selected count, search results, 500-node behavior, and legacy empty-scope mapping in `internal/site/src/components/routes/settings/telegram-utils.test.ts`

### Implementation for User Story 3

- [X] T038 [US3] Enforce canonical all/selected node-scope rules and future-node matching in `internal/hub/telegram_types.go`, `internal/hub/telegram_store.go`, and `internal/hub/telegram_delivery.go`
- [X] T039 [P] [US3] Implement pure node search, mode conversion, select-all, clear-all, and selected-count helpers in `internal/site/src/components/routes/settings/telegram-utils.ts`
- [X] T040 [US3] Build explicit “全部节点（包含未来新增）/指定节点” controls with search and stable scroll constraints in `internal/site/src/components/routes/settings/telegram-destinations.tsx`
- [X] T041 [US3] Add selected-mode validation, select-all-current, clear-all, and visible selected count without ambiguous empty saves in `internal/site/src/components/routes/settings/telegram-destinations.tsx`
- [X] T042 [US3] Run US3 focused tests and execute the 500-node and future-node scenarios in `specs/009-telegram-channel-management/quickstart.md`

**Checkpoint**: Administrators can tell whether future nodes are covered and can manage 500 current nodes without individual repetitive selection.

---

## Phase 6: User Story 4 - Understand Roles And Alert Categories (Priority: P2)

**Goal**: Make role permissions and load-average meanings explicit while applying policy scopes consistently to both roles.

**Independent Test**: Compare role behavior for matching/non-matching alerts, verify private-admin menu authorization and read-only sanitization, and verify all load labels/help against CPU-core reference wording.

### Tests for User Story 4

- [X] T043 [P] [US4] Add failing backend tests proving node/category scopes apply to both roles, read-only content stays sanitized, and privileged commands require an enabled private admin channel in `internal/hub/telegram_delivery_test.go` and `internal/hub/telegram_authorization_test.go`
- [X] T044 [P] [US4] Add failing frontend tests for Chinese role labels/descriptions, chat-type menu warnings, effective scope controls, and 1/5/15-minute load-average copy in `internal/site/src/components/routes/settings/telegram-utils.test.ts`

### Implementation for User Story 4

- [X] T045 [US4] Apply node and alert-category policy filtering before role-specific message formatting for every channel in `internal/hub/telegram_delivery.go`
- [X] T046 [US4] Keep privileged menu authorization private-admin-only while allowing non-private admin channels to receive scoped full-detail notifications in `internal/hub/telegram_readonly.go` and `internal/hub/telegram_menu.go`
- [X] T047 [P] [US4] Define reusable Chinese role descriptions, chat-type capability text, and load-average labels/help in `internal/site/src/components/routes/settings/telegram-utils.ts`
- [X] T048 [US4] Render role differences next to the selector, warn when a chat type has no menu capability, and ensure all visible scope controls are effective in `internal/site/src/components/routes/settings/telegram-destinations.tsx`
- [X] T049 [US4] Replace ambiguous load labels with “系统负载（1/5/15 分钟均值）” and logical-core guidance, then synchronize catalogs under `internal/site/src/locales/`
- [X] T050 [US4] Run US4 focused tests and execute role, sanitization, menu authorization, and load-copy scenarios in `specs/009-telegram-channel-management/quickstart.md`

**Checkpoint**: Both roles obey policies, their authority/content differences are visible, and load categories cannot be mistaken for percentages or alert durations.

---

## Phase 7: Polish And Cross-Cutting Verification

**Purpose**: Close concurrency, compatibility, security, quality, deployment, and release risks across all stories

- [X] T051 [P] Add rollback and retry regression coverage for partial channel/policy writes and concurrent edit/delete/test operations in `internal/hub/telegram_policies_api_test.go` and `internal/hub/telegram_destinations_test.go`
- [X] T052 [P] Add notifications backup compatibility fixtures and documented examples for v1/v2 in `specs/009-telegram-channel-management/contracts/telegram-backup-v2.md` and `internal/hub/testdata/telegram_backup/`
- [X] T053 Review implemented routes, schemas, status codes, deprecation fields, and authorization against `specs/009-telegram-channel-management/contracts/telegram-channel-api.md`
- [X] T054 Perform a security review for Token/error leakage, read-only sanitization, private-admin command authorization, cross-channel policy access, and destructive confirmation in `internal/hub/telegram_*.go` and `internal/site/src/components/routes/settings/telegram-destinations.tsx`
- [X] T055 Run `gofmt`, `go test -tags=testing ./...`, `go vet -tags=testing ./...`, and `golangci-lint run --build-tags testing`, recording exact results in `specs/009-telegram-channel-management/quickstart.md`
- [X] T056 Run `npm --prefix ./internal/site run test:unit`, `npm --prefix ./internal/site run check`, and `npm --prefix ./internal/site run build`, recording warnings/results in `specs/009-telegram-channel-management/quickstart.md`
- [X] T057 Build and start the local Hub with Docker, validate migration against a copied production-shaped data directory, and complete browser plus live Telegram scenarios from `specs/009-telegram-channel-management/quickstart.md`
- [X] T058 Remove temporary administrators, test channels/policies, copied data directories, debug logs, and temporary containers, documenting retained local demo state in `specs/009-telegram-channel-management/quickstart.md`
- [X] T059 If code is pushed, wait for the `Make docker images` GitHub Actions workflow and verify the expected `ghcr.io/wdmjieyao/beszel:edge` digest before reporting release success in `specs/009-telegram-channel-management/quickstart.md`

---

## Dependencies And Execution Order

### Phase Dependencies

- **Phase 1 Setup**: No dependencies.
- **Phase 2 Foundation**: Depends on Phase 1 and blocks all stories.
- **US1 (Phase 3)**: Starts after Foundation and is the recommended MVP.
- **US2 (Phase 4)**: Starts after Foundation; completes the multi-policy domain required by US3 and the policy portions of US4.
- **US3 (Phase 5)**: Depends on US2 policy CRUD, but its backend/frontend helpers can be developed in parallel after Foundation.
- **US4 (Phase 6)**: Depends on US2 policy matching; copy/helper work can run in parallel with US3.
- **Phase 7 Polish**: Depends on all selected stories.

### User Story Dependency Graph

```text
Setup -> Foundation -> US1 (MVP)
                    -> US2 -> US3
                           -> US4
US1 + US2 + US3 + US4 -> Polish
```

### Within Each Story

- Write the listed tests first and observe the intended failure.
- Implement storage/model behavior before handlers.
- Implement handlers/contracts before frontend integration.
- Run the story checkpoint before proceeding.
- Do not mark a task complete solely because adjacent tests pass.

## Parallel Opportunities

### Foundation

```text
T004 migration/backfill tests || T005 policy validation/store tests
```

### User Story 1

```text
T011 transport tests || T012 settings tests || T013 deletion tests || T014 frontend tests
T018 frontend API/types can proceed after the staged response contract is fixed
```

### User Story 2

```text
T022 API tests || T023 delivery tests || T024 backup tests || T025 frontend tests
T032 frontend API/types can proceed while backend delivery and backup implementation continues
```

### User Stories 3 And 4

```text
T036 backend scope tests || T037 frontend scope tests
T043 backend role tests || T044 frontend role/load tests
US3 frontend helpers and US4 copy helpers can proceed in parallel after US2 contracts stabilize
```

## Implementation Strategy

### MVP First

1. Complete Setup and Foundation.
2. Complete US1 Test Bot and deletion behavior.
3. Stop and validate the MVP independently.
4. Continue to US2 only after valid Bot verification and safe deletion are demonstrated.

### Incremental Delivery

1. **US1**: Reliable Bot test and visible safe deletion.
2. **US2**: Unique channels with multiple policies, de-duplication, and backup v2.
3. **US3**: Explicit all/selected node scope at 500-node scale.
4. **US4**: Clear role permissions and load-average semantics.
5. **Polish**: Full migration, browser, Docker, security, and release verification.

### Suggested Parallel Team Split

- **Backend A**: migration, policy store/API, delivery matching.
- **Backend B**: Bot verification, deletion, backup v2 compatibility.
- **Frontend**: channel/policy editor, node scope, role/load copy after contracts stabilize.
- **Reviewer**: migration/backup security and duplicate-delivery review before Docker validation.

## Notes

- `[P]` means different files or isolated test surfaces; do not parallelize tasks that edit the same file without coordination.
- Existing dirty-worktree changes must not be reverted or overwritten.
- Agent source and agent container behavior are out of scope.
- Every task remains unchecked until its implementation and stated verification pass.
- A Git push is not release completion when GHCR publication is required.

## Phase 8: Convergence

**Purpose**: Close independently verified gaps between the completed implementation and the feature contracts

- [X] T060 [CRITICAL] Replace all newly added hardcoded Telegram settings, policy, health, role, and load-average UI copy with Lingui messages, synchronize every catalog under `internal/site/src/locales/`, and add focused localization regression coverage per Constitution I and T049 (contradicts)
- [X] T061 Return `400 Bad Request` for malformed `/telegram/settings/test` request bodies instead of silently testing the saved Token, with focused handler regression coverage in `internal/hub/telegram_settings_test.go`, per FR-005 and FR-023 (partial)
- [X] T062 Map duplicate notification-policy names within one channel to a predictable `409 Conflict` response and cover create/update conflicts in `internal/hub/telegram_policies_api_test.go` per FR-023 and `contracts/telegram-channel-api.md` (partial)
- [X] T063 Clear the deleted channel's open policy destination, policy list, draft, and search state after successful deletion, while preserving them on failure, with frontend regression coverage per FR-003 and US1/AC5 (partial)
- [X] T064 Detect Telegram Chat ID collisions, unknown selected systems, policy-parent ownership mismatches, and duplicate policy names during notifications v2 backup preview, and add preview/apply compatibility tests per SC-008 and `contracts/telegram-backup-v2.md` (partial)
- [X] T065 Add true concurrent channel edit/delete/test and policy write/delete regression tests, including deterministic acceptable outcomes and retry behavior where applicable, in `internal/hub/telegram_policies_api_test.go` and `internal/hub/telegram_destinations_test.go` per FR-021 and T051 (partial)
- [X] T066 Implement true field-preserving partial `PATCH` semantics for Telegram channels and nested notification policies, including optional-field DTOs, validation after merge, conflict/not-found behavior, and focused API tests per FR-023 and Constitution IV (contradicts)

## Phase 9: Convergence

**Purpose**: Close second-pass contract and regression-coverage gaps after the Lingui and REST hardening work

- [X] T067 [CRITICAL] Restore focused frontend regression coverage for Lingui-managed role labels/descriptions, chat-type menu warnings, effective scope controls, 1/5/15-minute load-average labels/help, and their Simplified Chinese catalog entries per Constitution II, FR-014, FR-018, FR-019, FR-020, T044, and T060 (contradicts)
- [X] T068 [CRITICAL] Execute the notifications v1/v2 compatibility fixtures through preview and apply, then assert every Telegram channel/policy field, target-only preservation, redacted Token behavior, idempotency, and transactional rollback per Constitution II, SC-008, T024, T052, and `contracts/telegram-backup-v2.md` (contradicts)
- [X] T069 Add failing channel and policy PATCH tests for explicit `null` on every non-nullable field, then distinguish omission from null so invalid nulls return `400` while `muteUntil: null` still clears the mute per FR-023 and T066 (partial)
- [X] T070 Add concurrent duplicate policy create/rename regression tests and map PocketBase composite unique-index validation failures to the same stable `409 Conflict` response as preflight name conflicts per FR-008 and FR-023 (partial)
