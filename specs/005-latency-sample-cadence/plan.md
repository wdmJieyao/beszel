# Implementation Plan: Latency Sample Cadence

**Branch**: `` | **Date**: 2026-07-02 | **Spec**: [spec.md](./spec.md)

**Input**: Feature specification from `/specs/005-latency-sample-cadence/spec.md`

**Note**: This template is filled in by the `/speckit-plan` command. See `.specify/templates/plan-template.md` for the execution workflow.

## Summary

Make the node-detail `线路检测` chart use the same range semantics as the existing system metric charts: `1 分钟` is a fresh realtime view that starts empty and appends current samples, while `30 分钟` and longer ranges are historical views with range-appropriate density. The regression being addressed is twofold: the `1 分钟` latency chart still looked sparse or disconnected, and later fixes accidentally let the one-minute drawing model leak into wider ranges while removing the required `30 分钟` option.

The technical approach keeps the authenticated, short-lived live-observation session only for node detail `1 分钟`. When the frontend enters `1 分钟`, it creates/renews a live latency session for the current node. While active, the hub runs enabled latency probes at an approximately 1-second cadence through the existing hub-to-agent `RunNetworkProbe` websocket request path and persists results into the existing realtime result stream. Leaving `1 分钟`, closing the page, or missing heartbeats stops the high-cadence override. Selecting `30 分钟`, `1 小时`, or longer ranges must not create a live session; those ranges load historical latency results and apply range-aware bucketing/downsampling so the waveform remains readable instead of using a crowded one-minute trace.

## Technical Context

**Language/Version**: Go 1.24 module for hub/agent/backend changes; React 19 + TypeScript 5.9 + Vite 7 in `internal/site`

**Primary Dependencies**: PocketBase core routes/auth/realtime, existing hub websocket request manager, existing CBOR `RunNetworkProbe` agent action, React hooks, PocketBase JS client, Recharts, Biome

**Storage**: Existing PocketBase collections `network_probes`, `network_probe_assignments`, and `network_probe_results`. No new persisted collection is planned; live sessions are in-memory with TTL. Existing result persistence is reused so PocketBase realtime can deliver fresh samples to the browser.

**Testing**: Go unit tests with `go test -tags=testing ./...` for live-session state, cadence selection, due checks, auth/visibility, and scheduler behavior; existing frontend unit command `npm --prefix ./internal/site run test:unit` for hook/client lifecycle and chart range data-mode selection where practical; frontend quality gate `npm --prefix ./internal/site run check`; production build `npm --prefix ./internal/site run build`; Playwright/Docker validation for 60-second `1 分钟` waveform density plus `30 分钟`/longer historical readability.

**Target Platform**: Beszel hub web service, connected Beszel agents, and signed-in node detail page in modern browsers. The high-cadence mode is primarily for reachable TCPing/ICMP latency lines assigned to the viewed node.

**Project Type**: Cross-boundary Beszel hub/agent orchestration plus `internal/site` frontend behavior and REST API contract. Agent probe execution logic is reused; hub scheduling and frontend session lifecycle change.

**Performance Goals**: Active `1 分钟` observation produces samples approximately every 1 second per enabled latency line; validation with three reachable lines yields at least 40 successful samples per line in 60 seconds; active live mode must not block normal system metrics or permanent probe scheduling. Historical ranges must keep bounded chart point counts and avoid visually overcrowded one-minute density.

**Constraints**: The 1-second cadence is temporary and tied to active authenticated node-detail `1 分钟` sessions; configured probe interval validation remains for normal background probing; timeouts must not exceed the live cadence target; longer ranges and public dashboard must not create live sessions; the node detail latency range selector must include `30 分钟`; `30 分钟` and longer ranges must use historical fetch/rendering with range-appropriate density; failed probes remain failures and must not be converted into artificial successful line segments.

**Scale/Scope**: One or more active browsers may observe the same node; the hub should coalesce sessions per system so concurrent observers do not multiply probe execution. Validation covers one node with three reachable latency probes plus failure and range-switch cases.

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

- **Architecture/Stack**: PASS. Backend/hub work stays in Go and existing PocketBase/hub package boundaries. Agent transport reuses the current Go websocket/CBOR action. Frontend work stays in `internal/site` React + TypeScript + Vite.
- **Unit Tests**: PASS WITH TASK REQUIREMENT. Go tests are required for live-session lifecycle, cadence selection, scheduler coalescing, and REST auth. Frontend focused tests are required for start/renew/stop client lifecycle and latency range data-mode selection where practical.
- **Quality Gates**: PASS. Required gates are `go test -tags=testing ./...`, `npm --prefix ./internal/site run test:unit`, `npm --prefix ./internal/site run check`, and `npm --prefix ./internal/site run build`. `golangci-lint run` should be run if available in the environment.
- **RESTful API Contracts**: PASS. The new live-session HTTP contract is resource-oriented under systems and uses `POST`, `PATCH`, and `DELETE`; agent execution remains the existing non-REST websocket contract and is documented separately.
- **Incremental Delivery**: PASS. Work slices into live-session API/state, high-cadence scheduler override, frontend lifecycle, historical range rendering, and browser validation.

## Project Structure

### Documentation (this feature)

```text
specs/005-latency-sample-cadence/
├── plan.md
├── research.md
├── data-model.md
├── quickstart.md
├── contracts/
│   ├── live-session-api.md
│   ├── agent-probe-cadence.md
│   └── latency-chart-ranges.md
└── tasks.md
```

### Source Code (repository root)

```text
internal/hub/
├── api.go                         # register live-session REST routes
├── network_probes*.go             # normal probe config/results plus live cadence helpers/tests
├── network_probe_live*.go         # planned live-session manager and tests
└── ws/handlers.go                 # existing RunNetworkProbe request path reused

agent/
├── handlers.go                    # existing RunNetworkProbe handler reused
└── network_probe*.go              # probe timeout/failure behavior reused, tests extended if needed

internal/common/
└── common-ws.go                   # existing NetworkProbeRequest/Result contract reused

internal/site/src/
├── lib/api.ts                     # live-session client functions
├── types.d.ts                     # live-session response/request types
├── components/routes/system/
│   ├── use-network-probe-data.ts  # 1m live lifecycle; 30m+ historical fetch/rendering
│   └── network-probe-live-session*.ts
└── components/charts/
    └── network-probe-chart*.ts(x) # range-aware latency chart density and labels
```

**Structure Decision**: Keep normal configured probe scheduling intact and layer a hub-owned live-session manager beside the existing network probe orchestration. Do not add a new agent protocol for cadence; the hub already controls when a probe request is sent. Frontend responsibility is to select the correct data mode: signal active observation and append realtime result events only for `1 分钟`, but use historical result data and range-aware density for `30 分钟` and longer ranges.

## Complexity Tracking

No constitution violations. The cross-boundary complexity is required because the defect is caused by sparse measurements, not by chart rendering alone.

## Phase 0: Research

See [research.md](./research.md). All planning unknowns resolved.

## Phase 1: Design and Contracts

See:

- [data-model.md](./data-model.md)
- [contracts/live-session-api.md](./contracts/live-session-api.md)
- [contracts/agent-probe-cadence.md](./contracts/agent-probe-cadence.md)
- [contracts/latency-chart-ranges.md](./contracts/latency-chart-ranges.md)
- [quickstart.md](./quickstart.md)

## Post-Design Constitution Check

- **Architecture/Stack**: PASS. Design stays in Go/PocketBase/hub/agent boundaries and current React/TypeScript/Vite frontend.
- **Unit Tests**: PASS WITH TASK REQUIREMENT. Data model and contracts identify pure Go state and cadence helpers suitable for unit tests before implementation; frontend lifecycle and range-mode helpers should receive focused unit coverage.
- **Quality Gates**: PASS. Quickstart lists Go tests, frontend unit tests, Biome check, build, Docker, and Playwright validation.
- **RESTful API Contracts**: PASS. New HTTP endpoints are RESTful live-session resources; existing agent websocket execution remains documented as the internal non-REST transport.
- **Incremental Delivery**: PASS. The feature can be implemented and validated by first creating session state/API, then cadence execution, then frontend lifecycle, then historical range rendering, then waveform validation.
