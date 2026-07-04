<!--
Sync Impact Report
Version change: 1.0.1 -> 1.1.0
Modified principles:
- I. Existing Architecture and Stack (unchanged)
- II. Unit Tests Required (unchanged)
- III. Lint and Static Quality Gates (unchanged)
- IV. RESTful API Contracts (unchanged)
- V. Incremental, Maintainable Changes (unchanged)
Added sections:
- VI. Release and Registry Verification
Removed sections:
- None
Templates requiring updates:
- ✅ .specify/templates/plan-template.md
- ✅ .specify/templates/tasks-template.md
- ✅ .specify/templates/checklist-template.md
- ✅ .specify/templates/spec-template.md (reviewed; no change required)
- ✅ .specify/templates/commands/ (not present in this repository)
Follow-up TODOs:
- None
-->
# Beszel Constitution

## Core Principles

### I. Existing Architecture and Stack
All backend changes MUST follow the current Go architecture used by the hub,
agent, internal packages, and PocketBase integration. Backend plans MUST use Go
unless an approved constitution amendment explicitly changes that direction.

Frontend changes MUST follow the current `internal/site` stack: React,
TypeScript, Vite, Tailwind CSS, Lingui, and Biome. New frontend libraries,
state-management patterns, build tools, or styling systems MUST be justified in
the implementation plan before use.

Rationale: Beszel already has a production architecture. New work must reduce
maintenance cost by extending it instead of creating parallel stacks.

### II. Unit Tests Required
Every behavior change MUST include focused unit tests for the changed logic.
Backend tests MUST use the existing Go test structure and run with the project
test command. Frontend logic changes MUST include the closest practical unit
coverage for pure logic, hooks, utilities, or component behavior; if the current
frontend test harness is insufficient, the plan MUST either add one or justify a
temporary gap.

Bug fixes MUST include a regression test that fails before the fix when
practical. Test gaps MUST be listed in the plan and tasks with the reason they
cannot be closed in the same change.

Rationale: Monitoring, alerting, and API behavior must stay reliable across
platforms and deployment modes.

### III. Lint and Static Quality Gates
All changes MUST pass the repository quality gates before completion. Backend
changes MUST pass Go formatting, vetting where applicable, and the configured
Go linter. Frontend changes MUST pass the existing frontend lint/check command
in `internal/site`; the required frontend gate is Biome.

Plans and task lists MUST name the exact commands required for the touched
areas. Any skipped gate MUST be documented with the blocker and follow-up task.

Rationale: A single, explicit quality gate keeps automated and human review
aligned.

### IV. RESTful API Contracts
New or changed HTTP APIs MUST be RESTful unless an existing PocketBase endpoint,
websocket flow, or agent transport contract requires another pattern. RESTful
APIs MUST use resource-oriented paths, standard HTTP methods, predictable status
codes, and stable request/response schemas.

API changes MUST document compatibility, authentication and authorization,
validation behavior, error responses, and migration impact. Breaking API changes
MUST include a migration or compatibility plan.

Rationale: Beszel exposes API access and integrates multiple components; clear
HTTP contracts reduce coupling and client regressions.

### V. Incremental, Maintainable Changes
Feature work MUST be sliced into independently testable user stories or
implementation increments. Changes MUST prefer existing packages, components,
helpers, and conventions before adding abstractions. Cross-cutting refactors
MUST be justified by a current feature or defect, not speculative cleanup.

Plans MUST call out complexity, data migrations, platform-specific behavior,
and operational risks before implementation starts.

Rationale: Beszel supports many environments. Small, traceable changes make
review, release, and rollback safer.

### VI. Release and Registry Verification
Any push to GitHub that triggers GHCR image publishing, or is used to publish
deployable code, MUST NOT be reported as successful until the relevant GitHub
Actions workflow has completed and the GHCR image publication has succeeded.

For main-branch pushes, release tags, or any change that affects Docker images,
the operator MUST wait for the GHCR workflow result and verify the expected
image tags are available. If a push does not trigger GHCR publication, the
completion report MUST explicitly state that GHCR verification was not
applicable and include the command or workflow evidence used to determine that.

Rationale: Beszel deployments depend on container images. A Git push alone is
not enough when users will deploy from GHCR.

## Technology Baseline

Backend and agent code live primarily in Go under `agent/`, `internal/`, and
the root command entry points. Backend additions MUST preserve existing package
boundaries and reuse established transport, records, alerts, and hub patterns
where applicable.

Frontend code lives in `internal/site` and uses React with TypeScript, Vite,
Tailwind CSS, Lingui, Radix UI, lucide-react, Recharts, PocketBase client code,
and Biome. Frontend additions MUST use these tools unless the implementation
plan records a concrete reason to extend the stack.

Repository commands that plans MUST consider include `go test -tags=testing
./...`, `golangci-lint run`, and the relevant `internal/site` package scripts
such as `bun run lint`, `bun run check`, or equivalent package-manager commands
available in the working environment.

## Development Workflow

Specs MUST identify affected backend, frontend, API, data, and deployment
surfaces. Plans MUST include a Constitution Check that answers:

- Does the backend stay in Go and the existing architecture?
- Does the frontend stay on the current React/TypeScript/Vite stack?
- What unit tests are required for each changed behavior?
- Which lint/static quality commands must pass?
- Are HTTP APIs RESTful, or is a non-REST contract justified by existing
  architecture?
- If the work will be pushed to GitHub or changes Docker images, which GHCR
  workflow and image tags must be verified before reporting success?

Task lists MUST include unit-test tasks before implementation tasks for each
story or behavior change. Task lists MUST include final verification tasks for
tests, lint/static checks, GHCR publication when applicable, and API contract
review when APIs are touched.

## Governance

This constitution supersedes informal development preferences for Spec Kit
work in this repository. Specifications, plans, and tasks MUST satisfy the
Core Principles unless an explicit violation is recorded with rationale,
impact, and a follow-up path.

Amendments require updating this file, incrementing the version, recording the
change in the Sync Impact Report, and syncing dependent templates. Versioning
uses semantic versioning: MAJOR for incompatible governance or principle
redefinitions, MINOR for added principles or materially expanded requirements,
and PATCH for clarifications that do not change obligations.

All implementation reviews MUST verify constitution compliance before work is
marked complete. Any GitHub push that triggers GHCR publication is incomplete
until the workflow and image tags are verified. Unresolved compliance gaps MUST
remain visible in the plan or task list until closed.

**Version**: 1.1.0 | **Ratified**: 2026-06-28 | **Last Amended**: 2026-07-04
