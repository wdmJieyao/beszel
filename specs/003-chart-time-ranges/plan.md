# Implementation Plan: Public Chart Time Ranges

**Branch**: `` | **Date**: 2026-06-30 | **Spec**: [spec.md](./spec.md)

**Input**: Feature specification from `/specs/003-chart-time-ranges/spec.md`

**Note**: This template is filled in by the `/speckit-plan` command. See `.specify/templates/plan-template.md` for the execution workflow.

## Summary

Add first-class time range behavior to the public dashboard's newly added
latency and resource charts. Public charts default to the latest 30 minutes,
refresh plotted data every 20 seconds, show readable hour-minute-second x-axis
labels, and expose a range selector consistent with existing node-detail chart
controls while keeping public probe target data sanitized.

The technical approach is to reuse the existing chart-time concepts and
Recharts primitives, extend public chart range handling with a `30m` default,
return range-appropriate public history from the existing public status
resource, and keep all changes inside the current Go hub plus React/TypeScript
frontend architecture.

## Technical Context

**Language/Version**: Go 1.24+ for hub/API changes; React 19 + TypeScript 5.9 + Vite 7 in `internal/site`

**Primary Dependencies**: PocketBase records and request handlers, existing public status route, existing network probe result storage, React, Recharts, Radix UI select/dialog primitives, Lingui, Tailwind CSS, Biome

**Storage**: Existing PocketBase collections only: `system_stats`, `network_probe_results`, `network_probes`, `network_probe_assignments`, `systems`, and `public_system_visibility`

**Testing**: Focused Go tests for public range parsing/filtering and sanitized public responses; frontend build plus focused pure helper tests if a test harness is present or introduced; Playwright/Docker quickstart validation for chart range UX

**Target Platform**: Beszel hub web service and anonymous public dashboard visitors on desktop and mobile browsers

**Project Type**: Beszel hub backend package, `internal/site` frontend, additive REST-style public API query contract

**Performance Goals**: Public dashboard renders within 5 seconds for visible systems; range changes update chart data without full page reload; 20-second chart refresh does not expose hidden probe target metadata or overload the hub for normal public-dashboard use

**Constraints**: Public data remains sanitized; default public chart range is 30 minutes; chart refresh for newly added public charts is 20 seconds; x-axis labels use hour-minute-second format and must not wrap or overlap in validated desktop/mobile widths; frontend stays React/TypeScript/Vite and backend stays Go

**Scale/Scope**: Public dashboard cards for all public systems; latency series for public-visible TCPing/ICMP probes; CPU/memory/disk public history in the resource trend dialog; no new storage engine or agent protocol changes

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

- **Architecture/Stack**: PASS. Backend work stays in Go hub request/read-model code; frontend work stays in `internal/site` React + TypeScript + Vite with existing Recharts/Radix patterns.
- **Unit Tests**: PASS WITH TASK REQUIREMENT. Plan requires Go tests for public range parsing/filtering and frontend helper coverage where practical; Playwright/Docker validation covers chart rendering and interaction.
- **Quality Gates**: PASS. Required commands are `go test -tags=testing ./internal/hub ./internal/common ./internal/hub/ws`, `golangci-lint run` when available for touched Go code, `npm --prefix ./internal/site run check`, and `npm --prefix ./internal/site run build`.
- **RESTful API Contracts**: PASS. Existing public status resource remains read-only; optional `range` query parameter is additive and preserves anonymous public access.
- **Incremental Delivery**: PASS. Work slices into range contract/read model, shared chart-time helpers, public latency chart UX, public resource chart UX, and verification.

## Project Structure

### Documentation (this feature)

```text
specs/003-chart-time-ranges/
├── plan.md
├── research.md
├── data-model.md
├── quickstart.md
├── contracts/
│   └── public-status-range.md
└── tasks.md
```

### Source Code (repository root)

```text
internal/
├── hub/
│   ├── public_status.go          # public range parsing and range-aware history/probe queries
│   └── public_status_test.go     # focused tests for public range behavior and sanitization
└── site/src/
    ├── components/charts/
    │   ├── network-probe-chart.tsx       # range-aware x-axis labels and selector support
    │   └── chart-time-select.tsx         # reuse/adapt existing selector pattern where appropriate
    ├── components/routes/public-status.tsx # public card/dialog chart range state and 20s refresh
    ├── lib/
    │   ├── api.ts                         # range query typing for public status
    │   └── utils.ts                       # chart-time metadata extended with 30m/public use
    └── types.d.ts                         # public range/type additions

specs/003-chart-time-ranges/
├── contracts/
└── quickstart.md
```

**Structure Decision**: Keep the feature inside the existing public dashboard,
chart components, API helper, and Go public status read model. Do not add a new
frontend charting library, storage layer, or agent-side behavior.

## Complexity Tracking

No constitution violations.

## Phase 0: Research

See [research.md](./research.md). All planning unknowns resolved.

## Phase 1: Design and Contracts

See:

- [data-model.md](./data-model.md)
- [contracts/public-status-range.md](./contracts/public-status-range.md)
- [quickstart.md](./quickstart.md)

## Post-Design Constitution Check

- **Architecture/Stack**: PASS. Design extends existing Go public status read model and `internal/site` React chart components.
- **Unit Tests**: PASS WITH TASK REQUIREMENT. Go range/read-model tests and frontend helper or interaction validation are explicitly required in downstream tasks.
- **Quality Gates**: PASS. Quickstart names backend tests, frontend check/build, and Docker/Playwright validation.
- **RESTful API Contracts**: PASS. `GET /api/beszel/public/status?range=...` is additive, read-only, and backward-compatible.
- **Incremental Delivery**: PASS. Each story can be delivered and validated independently: default 30m filtering, readable axes, then range selection.
