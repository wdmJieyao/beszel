# Feature Specification: Global Probe Binding Regression Fix

**Feature Branch**: `006-global-probe-binding`

**Created**: 2026-07-04

**Status**: Draft

**Input**: User description: "$speckit-specify 有一个及其严重的bug，当我先添加一个机器，然后再添加测速节点时候，他能正常绑定没问题，但是当我再添加一个新的机器的时候，测速节点就没有自动绑定了，之前我不是已经说过了嘛？测速节点如果没有选择一个固定的机器，那么就是全机器通用的 现在不对"

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Global Probes Apply to Future Machines (Priority: P1)

An administrator creates a latency probe without choosing a fixed machine. Later,
when the administrator adds another machine, the existing probe automatically
becomes active for the new machine because the probe is global.

**Why this priority**: This is the reported severe regression. A global probe
that only covers machines present at creation time silently misses new machines
and gives the administrator incomplete latency coverage.

**Independent Test**: Add one machine, create a latency probe without selecting
a fixed machine, then add a second machine. Verify both machines become eligible
to run and display the same global probe without manually editing the probe.

**Acceptance Scenarios**:

1. **Given** one machine exists and a latency probe is created without a fixed
   machine selection, **When** a second machine is added later, **Then** the
   second machine automatically receives coverage from that probe.
2. **Given** a global latency probe already has results from existing machines,
   **When** a new machine is added, **Then** the new machine starts with no
   historical results for that probe but begins collecting new results when it
   becomes capable of running checks.
3. **Given** multiple global latency probes exist, **When** a new machine is
   added, **Then** the machine becomes covered by every enabled global probe.

---

### User Story 2 - Fixed-Machine Probes Stay Scoped (Priority: P2)

An administrator creates a latency probe and explicitly chooses one or more
fixed machines. When new machines are added later, those scoped probes remain
limited to the selected machines and do not unexpectedly expand.

**Why this priority**: Fixing global behavior must not break deliberate scoped
probe behavior. Administrators need both broad default coverage and explicit
targeting.

**Independent Test**: Create one global probe and one fixed-machine probe. Add a
new machine. Verify the new machine gets the global probe but does not get the
fixed-machine probe.

**Acceptance Scenarios**:

1. **Given** a probe was created with a fixed machine selection, **When** a new
   machine is added, **Then** that probe does not automatically apply to the new
   machine.
2. **Given** a probe is changed from fixed-machine scope to global scope,
   **When** the change is saved, **Then** all current machines and future
   machines become covered by the probe.
3. **Given** a probe is changed from global scope to fixed-machine scope,
   **When** the change is saved, **Then** only the selected machines remain
   covered going forward.

---

### User Story 3 - Coverage Is Understandable to Administrators (Priority: P3)

An administrator reviewing latency probe settings can tell whether a probe is
global or limited to fixed machines, so the behavior around newly added machines
is predictable.

**Why this priority**: The bug is partly caused by an ambiguous mental model.
The product must make the difference between "all machines" and "selected
machines" clear enough that administrators can verify coverage.

**Independent Test**: Open the probe settings list or editor and verify each
probe clearly communicates whether it applies to all machines or only selected
machines.

**Acceptance Scenarios**:

1. **Given** a probe has no fixed machine selection, **When** an administrator
   reviews its settings, **Then** it is clearly represented as applying to all
   machines.
2. **Given** a probe has fixed machine selections, **When** an administrator
   reviews its settings, **Then** it is clearly represented as applying only to
   those machines.
3. **Given** a machine is added after global probes already exist, **When** the
   administrator views the new machine's latency section after checks run,
   **Then** the applicable global probe lines are visible without additional
   probe editing.

### Edge Cases

- A new machine is added while one or more global probes are disabled.
- A new machine is added while it is offline or has not yet reported enough
  information to execute checks.
- A global probe is created before any machines exist.
- A fixed-machine probe has selected machines that are later deleted or hidden.
- A probe changes scope while checks are running.
- A machine is added and then quickly removed before its first scheduled check.
- A global probe is hidden from the public dashboard but remains active for
  authenticated views.
- A private machine receives global probe coverage but must not expose private
  results on the public dashboard unless public visibility rules allow it.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: A latency probe with no fixed machine selection MUST be treated as
  a global probe that applies to every eligible current machine and every
  eligible machine added in the future.
- **FR-002**: Adding a new eligible machine MUST make all enabled global probes
  applicable to that machine without requiring the administrator to edit or
  re-save each probe.
- **FR-003**: A global probe created before any machines exist MUST apply to
  machines that are added later.
- **FR-004**: A latency probe with one or more fixed machine selections MUST
  remain limited to those selected machines unless the administrator changes the
  probe scope.
- **FR-005**: Changing a probe from fixed-machine scope to global scope MUST
  make it apply to all eligible current machines and future machines.
- **FR-006**: Changing a probe from global scope to fixed-machine scope MUST
  stop applying it to unselected machines going forward while preserving
  historical results as historical data.
- **FR-007**: Disabled probes MUST NOT start running on newly added machines
  until they are enabled.
- **FR-008**: Probe coverage for public views MUST continue to respect existing
  public visibility controls for machines and probes.
- **FR-009**: Probe coverage for authenticated views MUST not expose machines
  that the viewer is not allowed to see.
- **FR-010**: The probe settings experience MUST clearly distinguish global
  probes from fixed-machine probes.
- **FR-011**: New machines covered by global probes MUST show an understandable
  no-history or pending state until their first results exist.
- **FR-012**: Changed behavior MUST be covered by focused unit tests.
- **FR-013**: If HTTP APIs are added or changed, they MUST be RESTful or
  explicitly justify an existing non-REST contract.

### Key Entities *(include if feature involves data)*

- **Latency Probe**: Administrator-defined check target with a display label,
  enabled state, public display setting, and coverage scope.
- **Coverage Scope**: Whether a latency probe applies globally to all eligible
  machines or only to explicitly selected fixed machines.
- **Machine**: Agent-backed monitored system that may execute latency probes and
  report results.
- **Probe Result**: Timestamped outcome from one machine executing one latency
  probe, including success, failure, and latency where available.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: In acceptance testing, a global probe created before a second
  machine is added appears for the second machine without any manual probe edit
  in 100% of attempts.
- **SC-002**: In acceptance testing, a fixed-machine probe does not appear for a
  newly added unselected machine in 100% of attempts.
- **SC-003**: A global probe created when zero machines exist applies to the
  first machine added later during acceptance testing.
- **SC-004**: Administrators can identify whether a probe is global or
  fixed-machine scoped within 5 seconds when reviewing probe settings.
- **SC-005**: Public dashboard checks confirm that private machines and probes
  hidden from public display remain hidden even when global probe coverage
  exists.

## Assumptions

- "没有选择一个固定的机器" means the administrator intentionally leaves machine
  selection empty, and the intended meaning is "all eligible machines" rather
  than "no machines."
- "机器" refers to monitored agent-backed systems that are eligible to execute
  latency probes.
- Historical results remain tied to the machine that produced them; changing
  scope affects future coverage and does not rewrite history.
- Existing public visibility rules continue to decide what anonymous visitors
  can see.
- Existing probe types and chart behavior remain in scope only as needed to
  verify that global coverage produces visible results.
