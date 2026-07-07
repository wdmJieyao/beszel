# Tasks: Telegram Notifications and YML Configuration Backup

**Input**: Design documents from `/specs/008-telegram-yml-backup/`

**Prerequisites**: plan.md, spec.md, research.md, data-model.md, contracts/, quickstart.md

**Tests**: Unit tests are REQUIRED for every behavior change by the project constitution. Test tasks are listed before implementation tasks in each user story.

**Organization**: Tasks are grouped by user story so Telegram notification delivery, Telegram bot menus, YML backup/restore, and agent compatibility can be implemented and validated independently.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel because it touches different files or depends only on completed earlier phases
- **[Story]**: User story mapping from `spec.md`
- Every task includes concrete repository file paths

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Confirm repository baseline and prepare shared implementation surfaces.

- [X] T001 Review the implementation plan and contracts in `specs/008-telegram-yml-backup/plan.md`
- [X] T002 [P] Inspect existing notification settings behavior in `internal/site/src/components/routes/settings/notifications.tsx`
- [X] T003 [P] Inspect existing alert delivery behavior in `internal/alerts/alerts.go`
- [X] T004 [P] Inspect existing system-only YAML export/sync behavior in `internal/hub/config/config.go`
- [X] T005 [P] Inspect current custom route registration in `internal/hub/api.go`

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Shared schema, types, and route foundations that block all user stories.

**CRITICAL**: Do not begin story implementation until this phase is complete.

- [X] T006 Create Telegram PocketBase collection migration for settings and destinations in `internal/migrations/add_telegram_notifications.go`
- [X] T007 Add backend Telegram data types and constants in `internal/hub/telegram_types.go`
- [X] T008 Add backend configuration backup data types and constants in `internal/hub/config_backup_types.go`
- [X] T009 [P] Add frontend Telegram and backup TypeScript types in `internal/site/src/types.d.ts`
- [X] T010 [P] Add frontend Telegram settings helper schemas in `internal/site/src/components/routes/settings/telegram-utils.ts`
- [X] T011 [P] Add frontend config backup helper schemas in `internal/site/src/components/routes/settings/config-backup-utils.ts`
- [X] T012 Add REST route registrations for Telegram and config backup resources in `internal/hub/api.go`
- [X] T013 Add shared backup secret envelope encryption/decryption helpers in `internal/hub/config_backup_crypto.go`
- [X] T014 Add shared Telegram HTTP transport interface with fake-test support in `internal/hub/telegram_transport.go`
- [X] T015 Add shared Telegram repository helpers for settings and destinations in `internal/hub/telegram_store.go`
- [X] T016 Add shared config backup repository helpers for systems, alerts, public status, and probes in `internal/hub/config_backup_sources.go`

**Checkpoint**: Shared migrations, types, routes, and helpers are ready for user-story work.

---

## Phase 3: User Story 1 - Telegram Bot Notification Channel (Priority: P1) MVP

**Goal**: Administrators can configure Telegram delivery, send a test message, and receive alert notifications through authorized destinations.

**Independent Test**: Configure bot settings and one admin destination, send a test message, trigger a representative alert, and confirm only the authorized destination receives non-secret content.

### Tests for User Story 1

- [X] T017 [P] [US1] Add backend tests for Telegram settings validation and secret redaction in `internal/hub/telegram_settings_test.go`
- [X] T018 [P] [US1] Add backend tests for Telegram destination create/update/delete validation in `internal/hub/telegram_destinations_test.go`
- [X] T019 [P] [US1] Add backend tests for Telegram test delivery using fake transport in `internal/hub/telegram_delivery_test.go`
- [X] T020 [P] [US1] Add backend tests for alert pipeline Telegram delivery integration in `internal/alerts/alerts_telegram_test.go`
- [X] T021 [P] [US1] Add frontend unit tests for Telegram settings and destination payload validation in `internal/site/src/components/routes/settings/telegram-utils.test.ts`

### Implementation for User Story 1

- [X] T022 [US1] Implement Telegram settings read/update/test handlers in `internal/hub/telegram_settings.go`
- [X] T023 [US1] Implement Telegram destination list/create/update/delete/test handlers in `internal/hub/telegram_destinations.go`
- [X] T024 [US1] Implement Telegram sendMessage delivery and sanitized error mapping in `internal/hub/telegram_delivery.go`
- [X] T025 [US1] Integrate Telegram destination delivery with existing alert sending in `internal/alerts/alerts.go`
- [X] T026 [US1] Add Telegram alert delivery adapter interface between alerts and hub in `internal/alerts/telegram_adapter.go`
- [X] T027 [US1] Add Telegram route auth and readonly restrictions in `internal/hub/api.go`
- [X] T028 [US1] Add Telegram settings UI section to notification settings in `internal/site/src/components/routes/settings/notifications.tsx`
- [X] T029 [US1] Add Telegram destination editor UI with role, chat ID, scope, and test-send controls in `internal/site/src/components/routes/settings/telegram-destinations.tsx`
- [X] T030 [US1] Add Telegram API client helper calls in `internal/site/src/lib/api.ts`
- [X] T031 [US1] Add Lingui extraction coverage for new Telegram settings text in `internal/site/src/locales/en/en.po`

**Checkpoint**: Telegram notification channel is usable without bot menu commands or YML backup restore.

---

## Phase 4: User Story 2 - Telegram Bot Menu and Authorized Queries (Priority: P2)

**Goal**: Authorized admin Telegram chats can query safe monitoring summaries and manage low-risk notification mute/restore actions; read-only destinations cannot access privileged actions.

**Independent Test**: Use fake Telegram updates for admin, read-only, and unknown chat IDs; verify menu responses match panel state and unauthorized requests disclose no monitoring data.

### Tests for User Story 2

- [X] T032 [P] [US2] Add backend tests for Telegram update polling offset and command parsing in `internal/hub/telegram_polling_test.go`
- [X] T033 [P] [US2] Add backend tests for admin menu action authorization in `internal/hub/telegram_menu_test.go`
- [X] T034 [P] [US2] Add backend tests for read-only and unknown chat rejection in `internal/hub/telegram_authorization_test.go`
- [X] T035 [P] [US2] Add backend tests for read-only scoped alert summary sanitization in `internal/hub/telegram_sanitization_test.go`

### Implementation for User Story 2

- [X] T036 [US2] Implement Telegram long polling worker lifecycle in `internal/hub/telegram_polling.go`
- [X] T037 [US2] Implement Telegram command and callback parser in `internal/hub/telegram_commands.go`
- [X] T038 [US2] Implement admin status overview and alert summary menu actions in `internal/hub/telegram_menu.go`
- [X] T039 [US2] Implement admin node list and node detail menu actions in `internal/hub/telegram_menu_systems.go`
- [X] T040 [US2] Implement notification mute and restore menu actions in `internal/hub/telegram_menu_notifications.go`
- [X] T041 [US2] Implement read-only destination scope filtering and non-sensitive summary formatting in `internal/hub/telegram_readonly.go`
- [X] T042 [US2] Wire Telegram polling startup/shutdown into hub lifecycle in `internal/hub/hub.go`
- [X] T043 [US2] Add Telegram menu status/help display in settings UI in `internal/site/src/components/routes/settings/telegram-destinations.tsx`
- [X] T044 [US2] Add non-sensitive logging for Telegram polling and menu failures in `internal/hub/telegram_polling.go`

**Checkpoint**: Telegram bot menu works for admin allowlist entries and refuses read-only/unknown privileged access.

---

## Phase 5: User Story 3 - YML Configuration Export and Restore (Priority: P3)

**Goal**: Administrators can export a versioned encrypted YML backup covering supported panel configuration and restore it through previewed merge semantics.

**Independent Test**: Export a representative configuration, verify no plaintext secrets, preview restore into a test instance, apply merge restore, and confirm target-only records are preserved.

### Tests for User Story 3

- [X] T045 [P] [US3] Add backend tests for backup schema metadata and section selection in `internal/hub/config_backup_export_test.go`
- [X] T046 [P] [US3] Add backend tests for encrypted secret envelopes and wrong-credential failures in `internal/hub/config_backup_crypto_test.go`
- [X] T047 [P] [US3] Add backend tests for systems and fingerprint token export/import decisions in `internal/hub/config_backup_systems_test.go`
- [X] T048 [P] [US3] Add backend tests for alerts and quiet-hours export/import decisions in `internal/hub/config_backup_alerts_test.go`
- [X] T049 [P] [US3] Add backend tests for public status and public probe visibility export/import decisions in `internal/hub/config_backup_public_status_test.go`
- [X] T050 [P] [US3] Add backend tests for network probe definition and assignment export/import decisions in `internal/hub/config_backup_network_probes_test.go`
- [X] T051 [P] [US3] Add backend tests for notification and Telegram destination export/import decisions in `internal/hub/config_backup_notifications_test.go`
- [X] T052 [P] [US3] Add backend tests for merge restore preview and stable-ID matching in `internal/hub/config_backup_restore_test.go`
- [X] T053 [P] [US3] Add frontend unit tests for config backup export/import validation helpers in `internal/site/src/components/routes/settings/config-backup-utils.test.ts`

### Implementation for User Story 3

- [X] T054 [US3] Implement versioned backup document marshaling and parsing in `internal/hub/config_backup_schema.go`
- [X] T055 [US3] Implement authenticated encryption and redaction behavior for sensitive values in `internal/hub/config_backup_crypto.go`
- [X] T056 [US3] Implement systems and fingerprint token backup source in `internal/hub/config_backup_sources.go`
- [X] T057 [US3] Implement alerts and quiet-hours backup source in `internal/hub/config_backup_sources.go`
- [X] T058 [US3] Implement user notification and Telegram backup source in `internal/hub/config_backup_sources.go`
- [X] T059 [US3] Implement public dashboard and public probe visibility backup source in `internal/hub/config_backup_sources.go`
- [X] T060 [US3] Implement network probe and assignment backup source in `internal/hub/config_backup_sources.go`
- [X] T061 [US3] Implement backup export handler for `POST /api/beszel/config-backups/exports` in `internal/hub/config_backup_api.go`
- [X] T062 [US3] Implement restore validation and preview handler for `POST /api/beszel/config-backups/validations` in `internal/hub/config_backup_api.go`
- [X] T063 [US3] Implement merge restore apply handler for `POST /api/beszel/config-backups/restores` in `internal/hub/config_backup_restore.go`
- [X] T064 [US3] Preserve legacy `GET /api/beszel/config-yaml` behavior while linking it to the new backup UI in `internal/hub/config/config.go`
- [X] T065 [US3] Replace settings YAML page with backup export/preview/restore UI in `internal/site/src/components/routes/settings/config-yaml.tsx`
- [X] T066 [US3] Add backup preview summary component for create/update/preserve/skip/conflict/error decisions in `internal/site/src/components/routes/settings/config-backup-preview.tsx`
- [X] T067 [US3] Add backup API client helper calls in `internal/site/src/lib/api.ts`
- [X] T068 [US3] Add TypeScript backup response/request types in `internal/site/src/types.d.ts`

**Checkpoint**: Full configuration backup export and merge restore are independently functional and preserve target-only records.

---

## Phase 6: User Story 4 - Preserve Agent Compatibility (Priority: P4)

**Goal**: Confirm the feature can roll out as a panel-only update and existing agents keep reporting.

**Independent Test**: Run panel with existing or simulated agents, exercise Telegram and backup workflows, and confirm no agent source or deployment command changes are required.

### Tests for User Story 4

- [X] T069 [P] [US4] Add backend regression test asserting backup/Telegram routes do not require agent transport changes in `internal/hub/agent_compatibility_test.go`
- [X] T070 [P] [US4] Add repository guard test or script check for unintended `agent/` source changes in `specs/008-telegram-yml-backup/quickstart.md`

### Implementation for User Story 4

- [X] T071 [US4] Document panel-only rollout and no-agent-update expectation in `specs/008-telegram-yml-backup/quickstart.md`
- [X] T072 [US4] Validate existing agent websocket and metric reporting paths remain untouched in `internal/hub/api.go`
- [X] T073 [US4] Validate generated install commands still point to existing images unless a later feature changes agent code in `internal/site/src/components/install-dropdowns-utils.ts`

**Checkpoint**: Panel-only deployment compatibility is explicitly verified.

---

## Phase 7: Polish & Cross-Cutting Concerns

**Purpose**: Final quality, documentation, localization, and release checks across all stories.

- [X] T074 [P] Run Go formatting on touched Go files from repository root `go.mod`
- [X] T075 Run focused backend feature tests for Telegram and backup packages from repository root `go.mod`
- [X] T076 Run full backend tests with `go test -tags=testing ./...` from repository root `go.mod`
- [X] T077 Run Go lint/static checks with `golangci-lint run --build-tags testing` from repository root `.golangci.yml`
- [X] T078 Run frontend unit tests with `npm --prefix ./internal/site run test:unit` using `internal/site/package.json`
- [X] T079 Run frontend Biome check with `npm --prefix ./internal/site run check` using `internal/site/package.json`
- [X] T080 Run frontend production build with `npm --prefix ./internal/site run build` using `internal/site/package.json`
- [X] T081 [P] Review REST API contract compatibility against `specs/008-telegram-yml-backup/contracts/telegram-api.md`
- [X] T082 [P] Review backup schema compatibility against `specs/008-telegram-yml-backup/contracts/yml-backup-schema.md`
- [X] T083 [P] Update operator validation notes in `specs/008-telegram-yml-backup/quickstart.md`
- [X] T084 If pushed to GitHub or Docker image inputs change, wait for `Make docker images` and verify GHCR tags before reporting success using `specs/008-telegram-yml-backup/quickstart.md`

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies.
- **Foundational (Phase 2)**: Depends on Phase 1 and blocks all user stories.
- **US1 Telegram notification channel (Phase 3)**: Depends on Phase 2.
- **US2 Telegram bot menu (Phase 4)**: Depends on Phase 2 and can begin after the Telegram settings/destination store from US1 exists; keep tests independent with fakes.
- **US3 YML backup/restore (Phase 5)**: Depends on Phase 2; can run in parallel with US1/US2 after shared schema and helpers exist.
- **US4 Agent compatibility (Phase 6)**: Depends on US1/US3 implementation surfaces enough to verify no agent changes.
- **Polish (Phase 7)**: Depends on all desired stories being complete.

### User Story Dependencies

- **US1 (P1)**: MVP. No dependency on US2/US3/US4 after foundation.
- **US2 (P2)**: Uses Telegram settings/destinations from foundation and US1; menu actions are independently testable with fake updates.
- **US3 (P3)**: Can proceed after foundation; includes Telegram config if US1 schema exists.
- **US4 (P4)**: Verification story; should run after implementation to prove panel-only rollout.

### Within Each User Story

- Tests first; verify they fail where practical.
- Data types and migrations before service code.
- Service code before route handlers.
- Route handlers before frontend API calls and UI.
- Story checkpoint before moving to the next priority story.

## Parallel Opportunities

- T002-T005 can run in parallel.
- T009-T011 can run in parallel after T006-T008 are understood.
- US1 test tasks T017-T021 can run in parallel.
- US2 test tasks T032-T035 can run in parallel.
- US3 test tasks T045-T053 can run in parallel by section.
- US3 implementation tasks T056-T060 can run in parallel after T054-T055.
- Final review tasks T081-T083 can run in parallel.

## Parallel Example: User Story 3

```text
Task: "Add backend tests for systems and fingerprint token export/import decisions in internal/hub/config_backup_systems_test.go"
Task: "Add backend tests for alerts and quiet-hours export/import decisions in internal/hub/config_backup_alerts_test.go"
Task: "Add backend tests for public status and public probe visibility export/import decisions in internal/hub/config_backup_public_status_test.go"
Task: "Add backend tests for network probe definition and assignment export/import decisions in internal/hub/config_backup_network_probes_test.go"
```

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Complete Phase 1 and Phase 2.
2. Complete Phase 3 only.
3. Validate Telegram settings, destination CRUD, test send, alert delivery, and no secret leakage.
4. Stop and demo Telegram notification delivery before adding bot menu or backup restore.

### Incremental Delivery

1. Foundation.
2. US1: Telegram notification delivery.
3. US2: Telegram bot menu and read-only scoping.
4. US3: YML backup export, preview, and merge restore.
5. US4: Panel-only deployment verification.
6. Polish and release verification.

### Parallel Team Strategy

After Phase 2, one agent can work US1 Telegram delivery, another can work US3 backup export sections, and another can prepare US2 menu tests using fake Telegram updates. Avoid editing the same files in parallel without coordination: `internal/hub/api.go`, `internal/site/src/lib/api.ts`, and `internal/site/src/types.d.ts`.

## Notes

- Baseline implementation must not modify `agent/` source files.
- Existing `config.yml` startup sync remains legacy system-only behavior and may still delete systems absent from that file; new backup restore must not copy that deletion behavior.
- Read-only Telegram destinations must never receive internal hosts, tokens, probe targets, webhook URLs, raw metric payloads, admin links, or backup details.
- If full verification commands fail due existing unrelated repository issues, record exact blockers and run narrow feature tests before reporting implementation status.
