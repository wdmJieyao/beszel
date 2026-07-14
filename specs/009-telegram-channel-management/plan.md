# Implementation Plan: Telegram Channel Management Improvements

**Branch**: `main` (working tree) | **Date**: 2026-07-10 | **Spec**: [spec.md](./spec.md)

**Input**: Feature specification from `/specs/009-telegram-channel-management/spec.md`

## Summary

Separate Telegram chat identity and authorization from notification routing.
The existing unique `telegram_destinations` record remains the channel-level
resource for a Chat ID, while a new child policy collection owns node and alert
category scopes. Existing destination scopes are backfilled into one default
policy without deleting legacy fields. Delivery evaluates all enabled policies
for a channel with OR semantics and sends at most one message per alert.

The same increment fixes Bot verification result parsing and returns staged
connectivity/menu results; makes channel deletion visible and confirmed; adds
explicit all-node and selected-node scope modes; applies scopes to both roles;
and explains administrator/read-only permissions and load-average categories.
The notifications backup section moves to section version 2 so channels and
policies round-trip while version 1 inline scopes remain importable.

## Technical Context

**Language/Version**: Go 1.26.3 for the hub/backend and migrations; React 19,
TypeScript 5.9, and Vite 7 for `internal/site`. No `beszel-agent` changes.

**Primary Dependencies**: PocketBase 0.36.8 collections, migrations, and route
handlers; existing `internal/hub/telegram_*` transport/store/delivery/menu
packages; Go `net/http` and JSON support for Telegram Bot API calls; existing
configuration backup helpers and YAML section-version handling; React settings
components, Valibot, Lingui, Radix Alert Dialog/Checkbox/Select/Tooltip,
lucide-react, Nanostores, and Biome.

**Storage**: Existing `telegram_settings` and `telegram_destinations`
collections remain. Add `telegram_notification_policies` related to one
destination. Existing destination `node_scope` and `alert_level_scope` fields
remain during the compatibility window and are copied to a default child
policy. Configuration backup notifications section version advances from 1 to
2.

**Testing**: Test-first focused Go coverage for Telegram response decoding,
channel/policy validation, migration/backfill, scope matching and de-duplication,
role behavior, deletion, backup v1/v2 compatibility, and REST handlers. Frontend
unit coverage for Bot test payload/result presentation, node-scope modes,
select-all/clear/search helpers, role descriptions, load labels, and channel/
policy payloads. Final gates: `gofmt` on touched Go files,
`go test -tags=testing ./...`, `go vet -tags=testing ./...`,
`golangci-lint run --build-tags testing`,
`npm --prefix ./internal/site run test:unit`,
`npm --prefix ./internal/site run check`, and
`npm --prefix ./internal/site run build`.

**Target Platform**: Authenticated Beszel hub settings UI and outbound Telegram
Bot API integration on the existing Linux/container deployment targets.

**Project Type**: Cross-boundary hub backend, PocketBase schema migration,
REST API, configuration backup compatibility, and authenticated React settings
workflow.

**Performance Goals**: Channel listing and policy editing remain responsive for
500 systems, 100 channels, and 20 policies per channel. Alert routing loads the
channel/policy inventory in bounded queries and performs in-memory matching,
without one query per policy. Search/filter feedback in the node selector should
appear within 100 ms for 500 nodes.

**Constraints**: Chat ID stays unique at channel level. Empty legacy node scope
means all current and future nodes. Selected-node mode cannot persist an empty
selection. Overlapping policies must never duplicate an alert to one Chat ID.
Administrator and read-only policies both honor node/category scopes. Only an
authorized private administrator chat may use privileged Bot commands.
Read-only content remains sanitized. Bot Tokens and backup secrets must not be
returned or logged. Production migration must be additive and rollback-aware.

**Scale/Scope**: One panel-wide Bot integration; up to 100 Telegram channels,
20 policies per channel, and 500 nodes in the selector. Includes Telegram
settings, destination/channel management, policy routing, notifications backup,
and related UI. Excludes agent changes, alert-definition editing, historical
telemetry, public dashboard behavior, and non-Telegram notification channels.

## Constitution Check

*GATE: Passed before research and re-checked after design.*

- **Architecture/Stack**: PASS. Backend work remains in Go under the current
  hub/migration boundaries; frontend work remains React + TypeScript + Vite in
  `internal/site`.
- **Unit Tests**: PASS WITH TASK REQUIREMENT. Each migration, routing, API, and
  UI behavior has a focused test seam and tests must precede implementation.
- **Quality Gates**: PASS. Go formatting, full tests, vet, configured Go lint,
  frontend unit tests, Biome, and production build commands are named above.
- **RESTful API Contracts**: PASS. Channels and nested policies use resource
  paths and standard methods/status codes. Bot verification remains an explicit
  integration action on the existing settings resource and is documented as a
  compatibility exception.
- **Incremental Delivery**: PASS. Work divides into Bot verification, additive
  schema/backfill, channel-policy APIs, delivery matching, backup compatibility,
  and frontend UX slices.
- **Release/GHCR Verification**: PASS. A GitHub push that publishes deployable
  code is incomplete until the `Make docker images` workflow succeeds and
  `ghcr.io/wdmjieyao/beszel:edge` is pullable at the expected digest. No agent
  image behavior changes are planned.

## Project Structure

### Documentation (this feature)

```text
specs/009-telegram-channel-management/
├── spec.md
├── plan.md
├── research.md
├── data-model.md
├── quickstart.md
├── contracts/
│   ├── telegram-channel-api.md
│   └── telegram-backup-v2.md
└── tasks.md                 # generated later by /speckit-tasks
```

### Source Code (repository root)

```text
internal/migrations/
└── add_telegram_notification_policies.go

internal/hub/
├── api.go
├── telegram_settings.go
├── telegram_transport.go
├── telegram_destinations.go
├── telegram_store.go
├── telegram_delivery.go
├── telegram_types.go
├── telegram_*_test.go
├── config_backup_types.go
├── config_backup_schema.go
├── config_backup_sources.go
├── config_backup_restore.go
└── config_backup_*_test.go

internal/site/src/
├── components/routes/settings/
│   ├── telegram-destinations.tsx
│   ├── telegram-utils.ts
│   └── telegram-utils.test.ts
├── components/ui/
├── lib/api.ts
└── types.d.ts
```

**Structure Decision**: Keep the existing destination collection and API name
as the channel compatibility boundary instead of renaming production records.
Add one focused child-policy collection and nested REST resource. Reuse current
settings components and hub Telegram modules; do not introduce a parallel
service or frontend state system.

## Migration And Delivery Sequence

1. Add the policy collection without removing or renaming destination fields.
2. Backfill exactly one `默认规则` policy per existing destination, mapping
   empty `node_scope` to `scope_mode=all` and non-empty scope to
   `scope_mode=selected`.
3. Add policy-aware reads and writes while preserving old destination payloads
   as compatibility operations on the default policy.
4. Switch delivery matching to load channels and policies, apply both scopes to
   both roles, OR policy matches per channel, and enqueue one send per channel.
5. Upgrade notifications backup section export to v2 and retain v1 restore.
6. Ship the channel/policy settings UI, staged Bot verification, clear role/load
   copy, and confirmed deletion.
7. After the compatibility period, removal of legacy scope fields may be
   considered in a separate feature; this plan does not remove them.

## Phase 0: Research

Research decisions are recorded in [research.md](./research.md). All technical
unknowns are resolved; no `NEEDS CLARIFICATION` markers remain.

## Phase 1: Design And Contracts

- Data ownership, validation, migration, and transitions:
  [data-model.md](./data-model.md)
- Channel, policy, deletion, and Bot verification API:
  [telegram-channel-api.md](./contracts/telegram-channel-api.md)
- Notifications backup section v2 compatibility:
  [telegram-backup-v2.md](./contracts/telegram-backup-v2.md)
- Runnable acceptance and quality checks: [quickstart.md](./quickstart.md)

## Post-Design Constitution Check

- **Architecture/Stack**: PASS. One additive PocketBase collection deepens the
  existing Telegram module without changing stack or adding a service.
- **Unit Tests**: PASS. The design exposes deterministic seams for response
  decoding, policy matching/de-duplication, migration, API validation, backup,
  and frontend pure helpers.
- **Quality Gates**: PASS. All required Go/frontend commands remain applicable.
- **RESTful API Contracts**: PASS. Nested policies have stable resource URLs;
  compatibility behavior and errors are explicit.
- **Incremental Delivery**: PASS. Additive schema ships before runtime cutover,
  and no destructive migration is required.
- **Release/GHCR Verification**: PASS. Hub image inputs change, so any push must
  wait for and verify the GHCR hub image workflow and tag.

## Complexity Tracking

No constitution violations.
