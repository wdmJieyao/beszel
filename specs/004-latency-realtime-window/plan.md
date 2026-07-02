# Implementation Plan: Latency Realtime Window

**Branch**: `` | **Date**: 2026-07-01 | **Spec**: [spec.md](./spec.md)

**Input**: Feature specification from `/specs/004-latency-realtime-window/spec.md`

**Note**: This template is filled in by the `/speckit-plan` command. See `.specify/templates/plan-template.md` for the execution workflow.

## Summary

Make the node-detail线路检测 chart's `1 分钟` range behave like the existing CPU 使用率 `1 分钟` range: switching into the range starts a fresh live observation session with no pre-switch plotted latency history, keeps all configured latency lines visible in the legend, and appends only realtime probe result events received by the browser after the session begins. Longer latency ranges keep their existing historical behavior, and public dashboard charts keep their existing historical-range behavior.

The technical approach is to adjust the node-detail network probe data hook and chart preparation so `1m` initializes an empty session, loads configured probes for legend structure only, skips historical result plotting, and appends results only from the active browser realtime subscription. Non-`1m` ranges continue to load historical probe results. The visual chart should reuse the existing chart-time domain/tick behavior used by node-detail charts where practical, without adding new storage, agent protocol, or HTTP API requirements.

## Technical Context

**Language/Version**: React 19 + TypeScript 5.9 + Vite 7 in `internal/site`; no backend Go change expected for the live-window semantics unless tests expose a scheduling defect.

**Primary Dependencies**: Existing PocketBase client/realtime subscription, existing `network_probe_results` collection, existing chart range store, Recharts, Radix Select, Lingui, Tailwind CSS, Biome.

**Storage**: Existing PocketBase collections only: `network_probes`, `network_probe_assignments`, `network_probe_results`, `systems`. No schema migration or new persisted model is planned.

**Testing**: Focused frontend unit coverage for live-session initialization and realtime-event-only merge logic where practical; Playwright/Docker validation for switching to `1 分钟`, empty initial state, legend completeness, and post-switch live plotting. Backend tests are not required unless backend scheduling or API behavior changes.

**Target Platform**: Beszel hub web UI in modern desktop and mobile browsers; signed-in node detail page is the primary target.

**Project Type**: `internal/site` frontend behavior change for node detail charts; no new public API contract.

**Performance Goals**: Range switch renders the chart frame and all configured legend entries within 1 second on a normal node detail page; realtime probe events append without full page reload; 1-minute view holds only points received during the active browser live session and needed for the visible minute.

**Constraints**: `1 分钟` latency mode must not plot any result obtained from historical fetches or previous live sessions; longer ranges must retain historical backfill; public dashboard 30-minute historical behavior must not regress; chart labels and legend text remain readable in normal desktop and mobile widths.

**Scale/Scope**: One node detail page with any number of configured latency lines; validation must cover at least three latency lines with staggered reporting times.

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

- **Architecture/Stack**: PASS. Work stays in `internal/site` React + TypeScript + Vite and reuses existing PocketBase realtime and chart components. No backend architecture change is planned.
- **Unit Tests**: PASS WITH TASK REQUIREMENT. Plan requires focused tests for empty live-session initialization and realtime-event-only merge behavior; Playwright/Docker validation covers visual behavior that is hard to assert with the current frontend test setup.
- **Quality Gates**: PASS. Required commands are `npm --prefix ./internal/site run check` and `npm --prefix ./internal/site run build`; Go test/lint gates are required only if implementation touches Go.
- **RESTful API Contracts**: PASS. No new or changed HTTP API is planned. Existing realtime subscription behavior is an existing PocketBase contract.
- **Incremental Delivery**: PASS. Work slices into live-session data semantics, chart rendering behavior, and end-to-end validation.

## Project Structure

### Documentation (this feature)

```text
specs/004-latency-realtime-window/
├── plan.md
├── research.md
├── data-model.md
├── quickstart.md
├── contracts/
│   └── latency-live-window.md
└── tasks.md
```

### Source Code (repository root)

```text
internal/site/src/
├── components/charts/
│   └── network-probe-chart.tsx
├── components/routes/system/
│   └── use-network-probe-data.ts
├── lib/
│   └── utils.ts
└── types.d.ts

specs/004-latency-realtime-window/
├── contracts/
└── quickstart.md
```

**Structure Decision**: Keep the feature in the existing frontend hook and chart component that already own线路检测 data and rendering. Avoid backend, storage, or agent protocol changes unless implementation reveals the probe results are not being emitted frequently enough to support live plotting. The node-detail CPU 使用率 `1 分钟` realtime path is the reference behavior for reset and append semantics.

## Complexity Tracking

No constitution violations.

## Phase 0: Research

See [research.md](./research.md). All planning unknowns resolved.

## Phase 1: Design and Contracts

See:

- [data-model.md](./data-model.md)
- [contracts/latency-live-window.md](./contracts/latency-live-window.md)
- [quickstart.md](./quickstart.md)

## Post-Design Constitution Check

- **Architecture/Stack**: PASS. Design stays in `internal/site` and uses existing node-detail range and realtime concepts.
- **Unit Tests**: PASS WITH TASK REQUIREMENT. Data-model and quickstart identify testable live-session boundaries; tasks must put unit tests before implementation.
- **Quality Gates**: PASS. Frontend Biome check and production build are required; Go gates are conditional on Go changes.
- **RESTful API Contracts**: PASS. No HTTP API changes are introduced; UI/realtime behavior contract is documented.
- **Incremental Delivery**: PASS. The feature can be validated by switching only the node-detail latency chart to live-session semantics while leaving other ranges untouched.
