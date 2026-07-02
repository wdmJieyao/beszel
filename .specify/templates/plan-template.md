# Implementation Plan: [FEATURE]

**Branch**: `[###-feature-name]` | **Date**: [DATE] | **Spec**: [link]

**Input**: Feature specification from `/specs/[###-feature-name]/spec.md`

**Note**: This template is filled in by the `/speckit-plan` command. See `.specify/templates/plan-template.md` for the execution workflow.

## Summary

[Extract from feature spec: primary requirement + technical approach from research]

## Technical Context

<!--
  ACTION REQUIRED: Replace the content in this section with the technical details
  for the project. The structure here is presented in advisory capacity to guide
  the iteration process.
-->

**Language/Version**: [Go version for backend/agent changes; React + TypeScript + Vite for `internal/site` frontend changes; or NEEDS CLARIFICATION]

**Primary Dependencies**: [e.g., FastAPI, UIKit, LLVM or NEEDS CLARIFICATION]

**Storage**: [if applicable, e.g., PostgreSQL, CoreData, files or N/A]

**Testing**: [Go unit tests with `go test -tags=testing ./...`; frontend unit test approach for `internal/site`; or NEEDS CLARIFICATION]

**Target Platform**: [e.g., Linux server, iOS 15+, WASM or NEEDS CLARIFICATION]

**Project Type**: [Beszel hub/agent/backend package, `internal/site` frontend, API contract, deployment artifact, or NEEDS CLARIFICATION]

**Performance Goals**: [domain-specific, e.g., 1000 req/s, 10k lines/sec, 60 fps or NEEDS CLARIFICATION]

**Constraints**: [domain-specific, e.g., <200ms p95, <100MB memory, offline-capable or NEEDS CLARIFICATION]

**Scale/Scope**: [domain-specific, e.g., 10k users, 1M LOC, 50 screens or NEEDS CLARIFICATION]

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

- **Architecture/Stack**: Backend changes stay in Go and follow existing hub/agent/internal package boundaries. Frontend changes stay on the current `internal/site` React + TypeScript + Vite stack unless a violation is documented.
- **Unit Tests**: Each behavior change has focused unit tests planned before implementation. Any test gap is justified with a follow-up task.
- **Quality Gates**: Required lint/static commands are named for touched areas, including Go linting and `internal/site` Biome checks.
- **RESTful API Contracts**: New or changed HTTP APIs use resource-oriented REST semantics, standard methods/status codes, and documented schemas. Non-REST contracts are justified by existing PocketBase, websocket, or agent transport architecture.
- **Incremental Delivery**: Work is sliced into independently testable stories or increments, with complexity and migration risks recorded.

## Project Structure

### Documentation (this feature)

```text
specs/[###-feature]/
├── plan.md              # This file (/speckit-plan command output)
├── research.md          # Phase 0 output (/speckit-plan command)
├── data-model.md        # Phase 1 output (/speckit-plan command)
├── quickstart.md        # Phase 1 output (/speckit-plan command)
├── contracts/           # Phase 1 output (/speckit-plan command)
└── tasks.md             # Phase 2 output (/speckit-tasks command - NOT created by /speckit-plan)
```

### Source Code (repository root)
<!--
  ACTION REQUIRED: Replace the placeholder tree below with the concrete layout
  for this feature. Delete unused options and expand the chosen structure with
  real paths (e.g., apps/admin, packages/something). The delivered plan must
  not include Option labels.
-->

```text
# [REMOVE IF UNUSED] Backend / agent Go changes
agent/
internal/
beszel.go

# [REMOVE IF UNUSED] Frontend changes
internal/site/src/
internal/site/package.json
internal/site/vite.config.ts

# [REMOVE IF UNUSED] Feature specs and contracts
specs/[###-feature]/
├── contracts/
└── quickstart.md
```

**Structure Decision**: [Document the selected structure and reference the real
directories captured above]

## Complexity Tracking

> **Fill ONLY if Constitution Check has violations that must be justified**

| Violation | Why Needed | Simpler Alternative Rejected Because |
|-----------|------------|-------------------------------------|
| [e.g., 4th project] | [current need] | [why 3 projects insufficient] |
| [e.g., Repository pattern] | [specific problem] | [why direct DB access insufficient] |
