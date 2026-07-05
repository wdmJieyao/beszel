# Implementation Plan: Public Probe Visibility and Refresh Commands

**Branch**: `` | **Date**: 2026-07-05 | **Spec**: [spec.md](./spec.md)

**Input**: Feature specification from `/specs/007-public-probe-visibility/spec.md`

## Summary

Move public latency-line exposure from the probe-level `public_visible` toggle to
per-VPS public dashboard settings, while preserving the exact public visibility
that exists in production today. The implementation extends public system
visibility data with selected probe IDs, migrates current public dashboard
visibility into that new per-system selection without widening exposure, removes
the probe-level public toggle from the settings UI, and updates public dashboard
filtering to emit only the selected probes for each public VPS.

The same delivery also updates generated Docker run commands so repeated
deployment commands behave like refresh commands: remove any existing container,
remove the old image best-effort, pull the latest image, then start the new
container. The current generated Docker run surface is the agent install
dropdown; the command builder should be structured so any panel Docker run
surface uses the same refresh semantics when present.

## Technical Context

**Language/Version**: Go 1.26.3 for hub/backend and migrations; React 19 +
TypeScript 5.9 + Vite 7 in `internal/site`

**Primary Dependencies**: PocketBase collections and migrations, existing Beszel
hub public status and network probe APIs, React components in
`internal/site/src/components/routes/settings`, PocketBase JS client, Lingui,
Biome

**Storage**: Existing PocketBase collections `public_system_visibility`,
`network_probes`, `network_probe_assignments`, and `network_probe_results`.
Extend `public_system_visibility` with a multi-value relation/list of selected
probe IDs for public display. Retain probe-level `public_visible` as a legacy
migration source until runtime ownership fully moves to per-VPS selection.

**Testing**: Focused Go unit tests with `go test -tags=testing ./...` for
migration seeding, public API updates, and public dashboard filtering; frontend
unit tests with `npm --prefix ./internal/site run test:unit` for settings
payloads and Docker command builders. Frontend quality gates:
`npm --prefix ./internal/site run check` and `npm --prefix ./internal/site run
build`. Go lint/static check: `golangci-lint run`.

**Target Platform**: Beszel hub web service, authenticated settings UI in
`internal/site`, anonymous public dashboard, and copied Docker deployment
commands for Linux hosts running Docker

**Project Type**: Cross-boundary Beszel hub/backend API, PocketBase migration,
and `internal/site` frontend behavior

**Performance Goals**: Public dashboard filtering must remain bounded to the set
of public VPS rows and their selected probes, without leaking unselected probe
metadata. Migration should complete in one pass and be idempotent on restart.
Generated Docker commands must remain copy-pastable one-liners with no manual
pre-cleanup requirement.

**Constraints**: Production migration must not widen public exposure, must not
hide currently visible public charts, and must not duplicate per-VPS probe
selections. Public probe visibility can only be edited from public dashboard
settings. Generated Docker run refresh behavior must remove old containers and
images best-effort but still fail on real pull/start errors. Compose examples
are out of scope because the requirement targets generated `docker run`
commands.

**Scale/Scope**: Applies to all public VPS configuration rows, all configured
latency probes, anonymous public dashboard responses, and all generated Docker
run command variants shown by the product. Validation covers migrated
production-style data, newly public VPS defaults, per-VPS public filtering, and
repeatable Docker refresh commands.

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

- **Architecture/Stack**: PASS. Backend work stays in Go within
  `internal/hub` and `internal/migrations`; frontend work stays in
  `internal/site` React + TypeScript + Vite.
- **Unit Tests**: PASS WITH TASK REQUIREMENT. Focused Go tests are required for
  migration/public filtering and frontend tests are required for settings
  payloads and Docker command generation.
- **Quality Gates**: PASS. Required commands are `go test -tags=testing ./...`,
  `golangci-lint run`, `npm --prefix ./internal/site run test:unit`,
  `npm --prefix ./internal/site run check`, and
  `npm --prefix ./internal/site run build`.
- **RESTful API Contracts**: PASS. Existing `/api/beszel/public/systems` and
  `/api/beszel/public/status` resources remain resource-oriented; additive
  request/response fields document migration compatibility.
- **Incremental Delivery**: PASS. Work slices into migration/schema,
  backend public filtering, frontend settings ownership, Docker command refresh
  helpers, and final verification.
- **Release/GHCR Verification**: PASS. If implementation is pushed to `main`,
  the `Make docker images` workflow must complete successfully and expected GHCR
  `edge` tags must be verified before the push is reported complete.

## Project Structure

### Documentation (this feature)

```text
specs/007-public-probe-visibility/
├── plan.md
├── research.md
├── data-model.md
├── quickstart.md
├── contracts/
│   ├── docker-run-command.md
│   └── public-system-api.md
└── tasks.md
```

### Source Code (repository root)

```text
internal/hub/
├── public_status.go               # public system settings API and anonymous dashboard filtering
├── public_status_test.go          # migration and public filtering regression tests
├── network_probes.go              # legacy public_visible source and effective coverage lookups
└── network_probes_test.go         # compatibility and coverage tests if selection validation touches probes

internal/migrations/
├── add_public_status_probes.go    # existing public system visibility collection definition
└── add_public_probe_visibility.go # new migration to add per-VPS selected probe IDs and seed legacy visibility

internal/site/src/
├── components/install-dropdowns.tsx
├── components/routes/settings/public-status.tsx
├── components/routes/settings/network-probes.tsx
├── lib/api.ts
└── types.d.ts

AGENTS.md
```

**Structure Decision**: Store per-VPS public probe selection on the existing
`public_system_visibility` row so public settings stay single-row-per-system and
API ownership remains centered in `internal/hub/public_status.go`. Keep the
legacy probe-level visibility field only as migration seed/compatibility input
while moving all runtime public filtering and UI ownership to per-VPS
selection. Centralize Docker run refresh command assembly in
`internal/site/src/components/install-dropdowns.tsx` so all generated command
surfaces share the same semantics.

## Complexity Tracking

No constitution violations.

## Phase 0: Research

See [research.md](./research.md). All planning unknowns resolved.

## Phase 1: Design and Contracts

See:

- [data-model.md](./data-model.md)
- [contracts/public-system-api.md](./contracts/public-system-api.md)
- [contracts/docker-run-command.md](./contracts/docker-run-command.md)
- [quickstart.md](./quickstart.md)

## Post-Design Constitution Check

- **Architecture/Stack**: PASS. Design keeps backend changes in Go hub/migration
  files and frontend changes in existing React/TypeScript settings components.
- **Unit Tests**: PASS WITH TASK REQUIREMENT. The design names focused Go and
  frontend regression coverage for migration, public filtering, and command
  generation.
- **Quality Gates**: PASS. Quickstart lists backend tests, Go lint, frontend
  unit tests, Biome check, and build validation.
- **RESTful API Contracts**: PASS. Public systems and public status remain
  resource-based APIs with additive schema updates for per-VPS probe selection.
- **Incremental Delivery**: PASS. Migration, backend filtering, frontend
  ownership, command generation, and verification can be implemented as
  independent increments.
- **Release/GHCR Verification**: PASS. Any later `main` push still requires the
  `Make docker images` workflow and expected GHCR tags to complete successfully.
