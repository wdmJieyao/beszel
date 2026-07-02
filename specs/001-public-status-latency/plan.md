# Implementation Plan: Public Status Page and Network Probe Trends

**Branch**: `` | **Date**: 2026-06-28 | **Spec**: [spec.md](./spec.md)

**Input**: Feature specification from `/specs/001-public-status-latency/spec.md`

## Summary

Add a public dashboard that is the anonymous home route and exposes only
opt-in systems and minimal metrics, plus configurable network probes executed
by selected nodes. The hub stores probe definitions/results, serves sanitized
REST resources for public views, coordinates probe execution over the existing
agent request channel, and the frontend adds public/admin views using the
existing React + TypeScript + Vite stack with Chinese product language.

## Technical Context

**Language/Version**: Go 1.24 module for hub/agent/backend changes; React 19 + TypeScript 5.9 + Vite 7 in `internal/site`

**Primary Dependencies**: PocketBase core collections/routes, existing hub websocket request manager, CBOR agent protocol, React, nanostores router, PocketBase JS client, Recharts, Biome

**Storage**: PocketBase collections for public visibility settings, network probe definitions, and timestamped probe results; existing system records remain source of current status

**Testing**: `go test -tags=testing ./...`; focused Go tests for public filtering, probe validation, result handling, and agent probe execution; frontend checks with `bun run --cwd internal/site check` or package-manager equivalent; frontend unit harness is not currently present, so plan records a task to add or justify focused frontend testing before implementation

**Target Platform**: Beszel hub web service and Beszel agents across supported Linux/Windows environments; anonymous web visitors and authenticated admins/users

**Project Type**: Beszel hub/agent/backend package, `internal/site` frontend, REST API contract, agent websocket contract, PocketBase migration

**Performance Goals**: Anonymous home dashboard usable within 5 seconds on normal broadband; probe execution must not block normal system metric collection; failed/offline execution nodes must not block other probe results

**Constraints**: Public responses must exclude private hosts, internal addresses, tokens, owner-only IDs, and admin actions; network probes are configurable with no built-in targets and no fixed count; probes execute from selected nodes with advanced execution binding hidden by default; the home route must render the public dashboard for anonymous visitors; frontend must stay on existing stack and pass Biome

**Scale/Scope**: One public dashboard at `/`, per-system public opt-in, configurable probe definitions, timestamped probe result history, public-safe probe charts, authenticated management UI, REST endpoints, agent request/response action

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

- **Architecture/Stack**: PASS. Backend stays in Go/PocketBase and existing hub/agent package boundaries. Frontend stays in `internal/site` React + TypeScript + Vite.
- **Unit Tests**: PASS WITH TASK REQUIREMENT. Go unit tests are required for filtering, validation, contracts, and probe execution. Frontend unit coverage requires adding a minimal harness or documenting a justified gap in tasks.
- **Quality Gates**: PASS. Required gates are `go test -tags=testing ./...`, `golangci-lint run`, and `bun run --cwd internal/site check` or equivalent.
- **RESTful API Contracts**: PASS. Public and admin HTTP APIs are resource-oriented REST endpoints; agent probe execution remains on existing websocket/CBOR transport with an explicit contract.
- **Incremental Delivery**: PASS. Work splits into anonymous home dashboard read model, admin visibility controls, probe configuration/execution, probe history/charts, then public probe display.

## Project Structure

### Documentation (this feature)

```text
specs/001-public-status-latency/
├── plan.md
├── research.md
├── data-model.md
├── quickstart.md
├── contracts/
│   ├── rest-api.md
│   └── agent-probe.md
└── tasks.md
```

### Source Code (repository root)

```text
agent/
├── handlers.go                  # register/handle probe execution action
└── network_probe*.go            # probe execution helpers and tests

internal/common/
└── common-ws.go                 # probe request/response CBOR contract

internal/hub/
├── api.go                       # REST route registration
├── public_status*.go            # public status read model and handlers
├── network_probes*.go           # probe admin/result orchestration
├── collections.go               # collection auth rules
└── ws/handlers.go               # hub-side probe response handling

internal/migrations/
└── add_public_status_probes.go  # public/probe collections and indexes

internal/records/
└── records*.go                  # retention/aggregation for probe results if reused

internal/site/src/
├── components/router.tsx        # home/public route behavior
├── components/routes/public-status*.tsx
├── components/routes/settings/  # admin probe and visibility controls
├── components/charts/           # probe chart composition using existing chart primitives
├── lib/api.ts
└── types.d.ts
```

**Structure Decision**: Use existing hub/agent/frontend boundaries. Add custom
REST handlers for sanitized anonymous data rather than opening PocketBase
collection rules broadly. Persist probe configuration and results in PocketBase
collections, and use the existing websocket request manager for hub-to-agent
probe execution.

## Complexity Tracking

No constitution violations. Complexity is inherent because the feature crosses
public access control, execution-node probing, persisted time-series data, and frontend
charts; each piece maps to an existing subsystem instead of adding a new stack.

## Phase 0: Research

See [research.md](./research.md). All planning unknowns resolved.

## Phase 1: Design and Contracts

See:

- [data-model.md](./data-model.md)
- [contracts/rest-api.md](./contracts/rest-api.md)
- [contracts/agent-probe.md](./contracts/agent-probe.md)
- [quickstart.md](./quickstart.md)

## Post-Design Constitution Check

- **Architecture/Stack**: PASS. Artifacts use Go/PocketBase, existing CBOR websocket protocol, and current React/TypeScript/Vite frontend.
- **Unit Tests**: PASS WITH TASK REQUIREMENT. Data model and contracts identify testable units; tasks must include Go unit tests before implementation and frontend test harness decision.
- **Quality Gates**: PASS. Quickstart lists required verification commands.
- **RESTful API Contracts**: PASS. REST resources use `GET`, `POST`, `PATCH`, and `DELETE` with stable response schemas. Agent CBOR action is separately documented as non-REST internal transport.
- **Incremental Delivery**: PASS. Quickstart and contracts support staged validation by user story.
