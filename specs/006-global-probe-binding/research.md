# Research: Global Probe Binding Regression Fix

## Decision: Represent all-machine coverage as explicit probe scope

Add a durable coverage concept to `network_probes`: global coverage means the
probe applies to all eligible current and future systems; fixed coverage means
the probe applies only to records in `network_probe_assignments`.

**Rationale**: The current model stores only assignment rows. The settings UI
turns "all nodes" into the current system IDs, which converts a persistent rule
into a snapshot. A probe-level scope matches the administrator's intent and
removes the need to create assignment rows whenever a new system is added.

**Alternatives considered**:

- Continue creating assignment rows for all systems when saving. Rejected
  because it still needs a reliable hook for every future system and can drift.
- Create assignment rows from a systems create hook. Rejected because the same
  global-vs-fixed intent remains implicit and historical all-node probes remain
  ambiguous.
- Treat empty assignment list as global without adding scope. Rejected because
  it cannot distinguish a future valid "no selected fixed systems" state and
  makes migration/reporting ambiguous.

## Decision: Keep fixed assignments for scoped probes

For fixed-machine probes, continue using `network_probe_assignments` as the
authoritative selected-system list.

**Rationale**: The existing collection already carries probe/system uniqueness,
auth rules, and result lookup expectations. Keeping it avoids broad data model
churn and preserves existing fixed-machine semantics.

**Alternatives considered**:

- Store all selected systems directly on `network_probes`. Rejected because it
  duplicates relation behavior already modeled by assignments and would require
  broader collection rule changes.
- Replace assignments with generated virtual rows only. Rejected because fixed
  per-system state and existing queries/tests already depend on assignment
  records.

## Decision: Resolve global coverage dynamically for all consumers

The hub should expose helper behavior that returns effective probe/system pairs
for a system or all systems. Scheduled checks, one-minute live checks, node
detail loading, and public dashboard summaries should all use this effective
coverage instead of directly filtering assignment rows only.

**Rationale**: The bug can reappear if only one path is fixed. Node detail uses
`probe.systems.includes(systemId)`, scheduled checks iterate assignments, live
checks filter assignments by system, and public summaries query assignments by
system. Each path needs the same global rule.

**Alternatives considered**:

- Fix only the settings UI to keep `systems: []`. Rejected because backend
  scheduler and readers still ignore probes without assignments.
- Fix only the scheduler. Rejected because charts and public summaries would
  still fail to show configured global probes for new systems until results
  exist.

## Decision: Return coverage scope in the network probe API

Add an explicit response/request field such as `scope` with values `global` and
`fixed`. Keep `systems` as the fixed selection list. For global probes,
`systems` remains empty in the API to avoid encoding a moving list as a stable
selection.

**Rationale**: The frontend needs to distinguish "all systems" from "no fixed
systems" and display the correct Chinese label. Keeping `systems` as the fixed
selection avoids forcing clients to diff dynamic all-system responses.

**Alternatives considered**:

- Return all effective systems in `systems` for global probes. Rejected because
  it recreates the snapshot mental model and makes new-machine coverage look
  like fixed selections.
- Infer scope from `systems.length === 0` in the frontend only. Rejected because
  backend scheduling and public views also need the scope.

## Decision: Classify existing all-assigned probes as global during migration

For existing data, probes whose assignment set matches all current visible
systems should become global. Probes with a partial assignment set should become
fixed. Probes with no assignments should become global because the intended
default in the product text is "默认由所有可用节点定期检测".

**Rationale**: Current UI created all-node probes by saving every current system
ID. Without migration, most existing "all nodes" probes would remain fixed to
old machines and the user's reported bug would persist for old probes.

**Alternatives considered**:

- Mark every existing probe with assignments as fixed. Rejected because it
  preserves the regression for probes originally created through the all-node
  default.
- Mark every existing probe as global. Rejected because administrators may have
  intentionally selected a single fixed execution node.

## Decision: Do not change agent execution protocol

The hub continues to decide which agent executes which probe and sends the
existing `RunNetworkProbe` request to that agent.

**Rationale**: The defect is coverage resolution in the hub/UI, not probe
execution on the agent. Keeping the protocol stable reduces risk.

**Alternatives considered**:

- Push coverage scope to agents. Rejected because agents do not know the hub's
  system inventory or public/auth visibility rules.
