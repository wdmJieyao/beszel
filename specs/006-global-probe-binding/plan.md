# Implementation Plan: Global Probe Binding Regression Fix

**Branch**: `` | **Date**: 2026-07-04 | **Spec**: [spec.md](./spec.md)

**Input**: Feature specification from `/specs/006-global-probe-binding/spec.md`

## Summary

Fix the latency probe coverage model so a probe with no fixed machine selection
means "all eligible machines" as an ongoing rule, not a one-time expansion to
the machines that existed when the probe was saved. The current settings UI
converts an empty selection into the current system list before saving, and the
hub persists only `network_probe_assignments`; newly added systems therefore do
not inherit existing all-node probes.

The implementation approach adds an explicit probe coverage scope, keeps fixed
assignments for scoped probes, and resolves global probes dynamically for
scheduling, live node-detail checks, authenticated chart loading, and public
dashboard summaries. Existing all-node probes should be migrated or interpreted
as global when their assignment set covered all systems at the time of upgrade.

## Technical Context

**Language/Version**: Go 1.24 module for hub/backend changes; React 19 +
TypeScript 5.9 + Vite 7 in `internal/site`

**Primary Dependencies**: PocketBase collections/migrations, existing Beszel hub
network probe APIs, existing hub-to-agent `RunNetworkProbe` websocket transport,
React hooks/components, PocketBase JS client, Biome

**Storage**: Existing PocketBase collections `network_probes`,
`network_probe_assignments`, and `network_probe_results`. Add persistent probe
coverage metadata to `network_probes`; keep `network_probe_assignments` for
fixed-machine scope and per-system execution records where needed.

**Testing**: Go unit tests with `go test -tags=testing ./...` for scope
normalization, migration/backfill classification, scheduler assignment
resolution, live assignment resolution, auth visibility, and public filtering.
Frontend unit tests with `npm --prefix ./internal/site run test:unit` for
settings payload behavior and system detail grouping. Frontend quality gates:
`npm --prefix ./internal/site run check` and `npm --prefix ./internal/site run
build`.

**Target Platform**: Beszel hub web service, connected Beszel agents, authenticated
settings/node-detail pages, and anonymous public dashboard.

**Project Type**: Cross-boundary Beszel hub/backend API and `internal/site`
frontend behavior. Agent probe execution logic is reused unchanged.

**Performance Goals**: Resolving global probe coverage must remain bounded for
normal Beszel deployments and must not create duplicate results for the same
probe/system interval. Adding a new system should make global probes visible and
eligible on the next probe/list refresh without manual probe editing.

**Constraints**: Empty fixed-machine selection means global coverage. Fixed
selection remains fixed. Disabled probes do not execute for new systems. Public
views still obey public system/probe visibility. Existing historical results
remain tied to the producing system. Existing agent protocol remains unchanged.

**Scale/Scope**: Applies to all configurable network probe types and all
eligible systems visible to the authenticated owner/admin. Validation covers
zero-system, one-system, newly added system, fixed scope, global scope, disabled
probe, node-detail live, scheduled background, and public dashboard paths.

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

- **Architecture/Stack**: PASS. Backend work stays in Go and existing
  `internal/hub`/migration boundaries; frontend work stays in
  `internal/site` React + TypeScript + Vite.
- **Unit Tests**: PASS WITH TASK REQUIREMENT. Focused Go tests are required for
  coverage resolution and scheduler/live/public behavior; frontend tests are
  required for settings payload and grouping semantics.
- **Quality Gates**: PASS. Required commands are `go test -tags=testing ./...`,
  `golangci-lint run` when available, `npm --prefix ./internal/site run
  test:unit`, `npm --prefix ./internal/site run check`, and `npm --prefix
  ./internal/site run build`.
- **RESTful API Contracts**: PASS. Existing `/api/beszel/network-probes`
  resources remain resource-oriented. Any schema additions are additive except
  the corrected meaning of empty system coverage.
- **Incremental Delivery**: PASS. Work slices into data migration/schema,
  backend coverage resolver, API response/payload updates, frontend settings,
  display consumers, and validation.
- **Release/GHCR Verification**: PASS. If implementation is pushed to `main`,
  the `Make docker images` workflow must complete successfully and expected
  GHCR `edge` tags must be verified before reporting the push as complete.

## Project Structure

### Documentation (this feature)

```text
specs/006-global-probe-binding/
├── plan.md
├── research.md
├── data-model.md
├── quickstart.md
├── contracts/
│   ├── network-probe-api.md
│   └── coverage-resolution.md
└── tasks.md
```

### Source Code (repository root)

```text
internal/hub/
├── network_probes.go              # probe API, coverage scope, scheduler/live resolution
├── network_probes_test.go         # coverage resolver and API/scheduler tests
├── public_status.go               # public probe summaries include global coverage
└── public_status_test.go          # public visibility with global probes

internal/migrations/
└── add_network_probe_scope.go      # add coverage scope to network_probes and classify old probes

internal/site/src/
├── types.d.ts
├── lib/api.ts
├── components/routes/settings/network-probes.tsx
├── components/routes/system/use-network-probe-data.ts
├── components/routes/system/network-probe-groups.ts
└── components/routes/system/use-network-probe-data.test.ts
```

**Structure Decision**: Keep assignments as the storage mechanism for fixed
machine scope. Add explicit coverage scope to the probe record so global
coverage is stable across future system additions. Centralize backend scope
resolution in `internal/hub/network_probes.go` so scheduled checks, live checks,
API list responses, and public summaries cannot diverge.

## Complexity Tracking

No constitution violations.

## Phase 0: Research

See [research.md](./research.md). All planning unknowns resolved.

## Phase 1: Design and Contracts

See:

- [data-model.md](./data-model.md)
- [contracts/network-probe-api.md](./contracts/network-probe-api.md)
- [contracts/coverage-resolution.md](./contracts/coverage-resolution.md)
- [quickstart.md](./quickstart.md)

## Post-Design Constitution Check

- **Architecture/Stack**: PASS. Design keeps backend changes in Go hub/migration
  code and frontend changes in existing React/TypeScript settings/detail code.
- **Unit Tests**: PASS WITH TASK REQUIREMENT. The design identifies concrete
  Go and frontend unit coverage before implementation.
- **Quality Gates**: PASS. Quickstart lists Go tests, frontend unit tests,
  Biome check, build, and browser/manual validation.
- **RESTful API Contracts**: PASS. The network probe API remains resource
  oriented; response/request schema additions are documented in contracts.
- **Incremental Delivery**: PASS. Scope persistence, resolver behavior, API,
  frontend, and validation can be implemented independently.
- **Release/GHCR Verification**: PASS. Any later `main` push triggers the
  documented GHCR workflow verification requirement.
