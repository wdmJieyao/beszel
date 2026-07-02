# Implementation Plan: Probe Chart and Public Dashboard Fixes

**Branch**: `` | **Date**: 2026-06-29 | **Spec**: [spec.md](./spec.md)

**Input**: Feature specification from `/specs/002-probe-public-fixes/spec.md`

**Note**: This template is filled in by the `/speckit-plan` command. See `.specify/templates/plan-template.md` for the execution workflow.

## Summary

Fix the product and runtime issues found during Docker validation of the public
dashboard and network probe feature. Node detail pages must stay usable after
TCPing checks are added, latency-capable probe results must be grouped into
combined multi-series charts, TCPing failures must carry actionable failure
reasons, and the anonymous public dashboard must show and refresh public CPU,
memory, disk, and freshness values for reporting nodes.

## Technical Context

**Language/Version**: Go 1.24+ module for hub/agent/backend changes; React 19 + TypeScript 5.9 + Vite 7 in `internal/site`

**Primary Dependencies**: Existing PocketBase collections/routes, hub websocket request manager, CBOR agent protocol, React, nanostores router, PocketBase JS client, Recharts, Lingui, Biome

**Storage**: Existing PocketBase system records, network probe definitions, assignments, and results; no new database engine

**Testing**: Focused Go tests for public metric extraction, TCPing validation/failure classification, probe result grouping inputs, plus frontend build/Biome checks; frontend has no dedicated unit-test runner, so pure mapping coverage is planned where practical and manual Docker quickstart validates UI behavior

**Target Platform**: Beszel hub web service and Beszel agents on supported platforms; authenticated admins/users; anonymous public-dashboard visitors

**Project Type**: Beszel hub/agent/backend package, `internal/site` frontend, REST API contract, internal agent websocket contract

**Performance Goals**: Node detail page remains interactive within 5 seconds with multiple probe series; public dashboard displays available metrics within 5 seconds and refreshes without login; failed probes do not block successful probe series

**Constraints**: Public responses must remain sanitized; combined charts must use existing Recharts/UI primitives; TCPing remains execution-node/agent based; backend stays Go; frontend stays React/TypeScript/Vite and passes Biome

**Scale/Scope**: One node detail probe section, public dashboard metric/freshness behavior, TCPing validation/failure diagnostics, existing probe result APIs extended additively if needed

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

- **Architecture/Stack**: PASS. Backend remains Go/PocketBase/hub/agent. Frontend remains `internal/site` React + TypeScript + Vite with existing chart primitives.
- **Unit Tests**: PASS WITH TASK REQUIREMENT. Go tests are required for TCPing validation/failure mapping and public metric extraction. Frontend unit-test gap remains documented; UI behavior is verified through build, Biome, and Docker quickstart scenarios.
- **Quality Gates**: PASS. Required gates are `go test -tags=testing ./...`, `golangci-lint run`, and `npm --prefix ./internal/site run check` or package-manager equivalent.
- **RESTful API Contracts**: PASS. Any HTTP API changes are additive to existing resource endpoints. Internal agent probe execution remains websocket/CBOR and is justified by existing architecture.
- **Incremental Delivery**: PASS. Work slices into node detail chart stability, TCPing diagnostics, and public dashboard metrics/freshness.

## Project Structure

### Documentation (this feature)

```text
specs/002-probe-public-fixes/
├── plan.md              # This file (/speckit-plan command output)
├── research.md          # Phase 0 output (/speckit-plan command)
├── data-model.md        # Phase 1 output (/speckit-plan command)
├── quickstart.md        # Phase 1 output (/speckit-plan command)
├── contracts/           # Phase 1 output (/speckit-plan command)
└── tasks.md             # Phase 2 output (/speckit-tasks command - NOT created by /speckit-plan)
```

### Source Code (repository root)
```text
agent/
├── network_probe*.go            # TCPing execution and failure classification
└── handlers.go                  # probe action response mapping

internal/
├── hub/
│   ├── network_probes*.go       # probe validation, results, grouping inputs
│   └── public_status*.go        # public metric/freshness composition
├── common/common-ws.go          # probe result fields if extended additively
└── hub/ws/handlers.go           # websocket probe request handling

internal/site/src/
├── components/routes/system.tsx
├── components/routes/system/use-network-probe-data.ts
├── components/charts/network-probe-chart.tsx
├── components/routes/public-status.tsx
├── components/routes/settings/network-probes.tsx
├── lib/api.ts
└── types.d.ts

specs/002-probe-public-fixes/
├── contracts/
└── quickstart.md
```

**Structure Decision**: Keep the previous feature's package boundaries. Fix the
read models, validation, and chart composition in place rather than adding new
storage or a new frontend charting stack.

## Complexity Tracking

No constitution violations.

## Phase 0: Research

See [research.md](./research.md). All planning unknowns resolved.

## Phase 1: Design and Contracts

See:

- [data-model.md](./data-model.md)
- [contracts/rest-api.md](./contracts/rest-api.md)
- [contracts/agent-probe.md](./contracts/agent-probe.md)
- [quickstart.md](./quickstart.md)

## Post-Design Constitution Check

- **Architecture/Stack**: PASS. Design keeps Go backend/agent and React/TypeScript/Vite frontend.
- **Unit Tests**: PASS WITH TASK REQUIREMENT. Data model and contracts identify focused Go test surfaces; frontend harness gap remains explicit.
- **Quality Gates**: PASS. Quickstart lists backend tests, Go lint, frontend Biome, frontend build, and Docker validation.
- **RESTful API Contracts**: PASS. HTTP changes are additive resource schema changes; agent websocket remains documented separately.
- **Incremental Delivery**: PASS. Each user story can be validated independently.
