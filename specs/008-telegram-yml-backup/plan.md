# Implementation Plan: Telegram Notifications and YML Configuration Backup

**Branch**: `` | **Date**: 2026-07-10 | **Spec**: [spec.md](./spec.md)

**Input**: Feature specification from `/specs/008-telegram-yml-backup/spec.md`

**Note**: This template is filled in by the `/speckit-plan` command. See `.specify/templates/plan-template.md` for the execution workflow.

## Summary

Add first-class Telegram notification support and a versioned YML configuration
backup/restore workflow. Telegram support extends the existing alert delivery
path with panel-managed chat ID allowlist entries, role-based destinations
(`admin` and `read_only`), scoped read-only alert summaries, test delivery, and
a limited administrator bot menu. The bot menu is handled by the hub through the
Telegram Bot API and does not require agent changes.

The YML work introduces a new configuration backup flow rather than expanding
the existing destructive `config.yml` startup sync. Exports include all supported
panel-managed configuration sections, encrypt sensitive values when included,
and imports use previewed merge restore: create missing records, update records
matched by stable identifiers, preserve target-only records, and reject unsafe
or incompatible sections before applying.

Node-detail latency chart realtime refresh fixes may ship in the same release
batch, but they remain specified, planned, and accepted under the existing
latency-monitoring feature specs rather than this Telegram/YML feature.

## Technical Context

**Language/Version**: Go 1.26.3 for hub/backend, migrations, and alert delivery;
React 19 + TypeScript 5.9 + Vite 7 in `internal/site`. No baseline
`beszel-agent` source changes.

**Primary Dependencies**: PocketBase collections/migrations and route handlers,
existing `internal/alerts` delivery pipeline, existing `internal/hub/config`
YAML export/sync package as a reference, `gopkg.in/yaml.v3`, existing
`golang.org/x/crypto` dependency for authenticated backup encryption, Go
`net/http` for Telegram Bot API calls, existing React settings components,
Valibot, Lingui, Radix UI, lucide-react, and Biome.

**Storage**: Existing PocketBase collections `systems`, `fingerprints`,
`alerts`, `quiet_hours`, `user_settings`, `public_system_visibility`,
`network_probes`, `network_probe_assignments`, and supporting user records are
read by backup export/restore. New PocketBase collections are planned for
Telegram bot settings and Telegram destinations. Runtime telemetry collections
such as `system_stats`, `alerts_history`, and `network_probe_results` are out of
backup scope.

**Testing**: Focused Go unit tests for Telegram destination validation,
role/scope filtering, test delivery plumbing with fake Telegram transport,
backup export schema, encryption/decryption failure cases, merge restore
preview/apply, stable-ID matching, and legacy `config.yml` compatibility.
Frontend unit tests with `npm --prefix ./internal/site run test:unit` for
settings payloads, validation helpers, backup preview decisions, and sensitive
export UI state. Final gates: `gofmt` on touched Go files,
`go test -tags=testing ./...`, `golangci-lint run --build-tags testing`,
`npm --prefix ./internal/site run test:unit`,
`npm --prefix ./internal/site run check`, and
`npm --prefix ./internal/site run build`.

**Target Platform**: Beszel hub web service and authenticated settings UI.
Telegram integration requires outbound HTTPS from the hub to Telegram. No
inbound public Telegram webhook is required for the baseline plan.

**Project Type**: Cross-boundary Beszel hub/backend API, PocketBase migration,
alert delivery integration, YML configuration import/export service, and
`internal/site` frontend settings workflows.

**Performance Goals**: Telegram alert delivery should normally complete within
30 seconds when Telegram is available. Bot menu requests should return compact
responses and avoid long-running database scans. Backup export/preview should
complete within 10 seconds for representative small-to-medium deployments
(hundreds of systems/probes/alerts), and restore apply should run transactionally
per section where supported.

**Constraints**: Telegram authorization is panel-managed by chat ID allowlist.
Read-only Telegram destinations can receive only non-sensitive alert summaries
within configured node and alert-level scope. Sensitive YML fields must never be
exported as plaintext; full restorable exports require encryption and matching
decryption credentials on import. Import defaults to merge restore and must not
delete target-only configuration. Existing `config.yml` startup sync remains
system-only and keeps its current semantics; the new backup flow must not inherit
its destructive delete-by-absence behavior. Baseline implementation must not
require `beszel-agent` updates. Unrelated latency-chart refresh fixes, even when
co-released, are out of scope for this plan and must stay owned by the latency
feature line.

**Scale/Scope**: Applies to all configured systems, alert definitions, quiet
hours, user notification preferences, public dashboard visibility, public probe
visibility selections, network probe definitions and assignments, Telegram bot
settings, and Telegram destinations. Excludes users/passwords/auth providers,
historical metrics, alert history, probe result history, logs, generated
charts, and unrelated node detail latency chart realtime refresh behavior.

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

- **Architecture/Stack**: PASS. Backend work stays in Go under existing hub,
  alerts, config, and migration boundaries. Frontend work stays in
  `internal/site` React + TypeScript + Vite.
- **Unit Tests**: PASS WITH TASK REQUIREMENT. Focused tests are required for
  Telegram auth/scope, delivery plumbing, encrypted backup export/import,
  preview/apply decisions, and frontend validation helpers.
- **Quality Gates**: PASS. Required commands are `gofmt` on touched Go files,
  `go test -tags=testing ./...`, `golangci-lint run --build-tags testing`,
  `npm --prefix ./internal/site run test:unit`,
  `npm --prefix ./internal/site run check`, and
  `npm --prefix ./internal/site run build`.
- **RESTful API Contracts**: PASS. New hub routes are resource-oriented:
  Telegram settings/destinations and configuration backup export/validation/
  restore resources. Telegram Bot API polling is an external integration, not a
  public inbound API.
- **Incremental Delivery**: PASS. Work slices into Telegram schema/settings,
  delivery integration, bot menu worker, backup schema/encryption, export,
  validation/preview, restore apply, frontend settings, and final verification.
- **Release/GHCR Verification**: PASS. If implementation is pushed to `main` or
  otherwise used for deployable images, the `Make docker images` workflow and
  expected GHCR tags (`ghcr.io/wdmjieyao/beszel:edge` and relevant tags) must be
  verified before reporting push success.

## Project Structure

### Documentation (this feature)

```text
specs/008-telegram-yml-backup/
├── plan.md
├── research.md
├── data-model.md
├── quickstart.md
├── contracts/
│   ├── config-backup-api.md
│   ├── telegram-api.md
│   └── yml-backup-schema.md
└── tasks.md
```

### Source Code (repository root)

```text
internal/alerts/
├── alerts.go                    # extend delivery to Telegram destination service
├── alerts_api.go                # keep existing shoutrrr test route; add/route Telegram tests through hub service if needed
└── *_test.go                    # delivery, scope, and regression tests

internal/hub/
├── api.go                       # register Telegram and config backup routes
├── telegram_*.go                # Telegram settings, destinations, worker, menu, transport
├── config_backup_*.go           # export, encryption, validation, preview, merge restore
├── public_status.go             # exported config source for public visibility data
├── network_probes.go            # exported config source for probe definitions/assignments
└── *_test.go                    # route, export/import, preview, restore tests

internal/hub/config/
└── config.go                    # retain legacy system-only config.yml behavior; share safe helpers where useful

internal/migrations/
└── add_telegram_notifications.go # Telegram settings/destination collections

internal/site/src/
├── components/routes/settings/
│   ├── notifications.tsx        # Telegram channel management in notifications settings
│   ├── config-yaml.tsx          # evolve into configuration backup export/import UI
│   └── *.{test,ts}              # validation and payload helpers
├── lib/api.ts
└── types.d.ts
```

**Structure Decision**: Keep Telegram alert delivery close to
`internal/alerts`, but store and expose Telegram configuration from
`internal/hub` because it is panel-managed data and needs authenticated REST
routes. Add a dedicated `internal/hub/config_backup`-style implementation file
set rather than expanding `internal/hub/config/config.go`; the existing
`config.yml` code is system-only and currently deletes systems absent from the
file, which conflicts with the new merge-restore contract. Frontend changes
remain inside the existing settings area.

## Complexity Tracking

No constitution violations.

## Phase 0: Research

See [research.md](./research.md). Planning unknowns are resolved.

## Phase 1: Design and Contracts

See:

- [data-model.md](./data-model.md)
- [contracts/telegram-api.md](./contracts/telegram-api.md)
- [contracts/config-backup-api.md](./contracts/config-backup-api.md)
- [contracts/yml-backup-schema.md](./contracts/yml-backup-schema.md)
- [quickstart.md](./quickstart.md)

## Post-Design Constitution Check

- **Architecture/Stack**: PASS. Design preserves Go hub/backend ownership,
  keeps frontend work in `internal/site`, and explicitly excludes baseline
  agent changes.
- **Unit Tests**: PASS WITH TASK REQUIREMENT. Design identifies focused backend
  and frontend tests for each changed behavior before implementation.
- **Quality Gates**: PASS. Quickstart and later tasks must run Go tests/lint and
  frontend unit/Biome/build gates.
- **RESTful API Contracts**: PASS. Contracts document resource paths, methods,
  status behavior, schemas, auth, and compatibility.
- **Incremental Delivery**: PASS. Telegram and backup work can ship as separate
  testable slices, with backup export usable before restore apply.
- **Release/GHCR Verification**: PASS. Any main push or Docker-affecting change
  remains incomplete until GHCR workflow and image tag verification finishes.
