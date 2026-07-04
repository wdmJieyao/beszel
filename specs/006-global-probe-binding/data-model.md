# Data Model: Global Probe Binding Regression Fix

## Latency Probe

Administrator-defined network check target.

### Fields

- `id`: stable probe identifier.
- `name`: display label.
- `type`: `tcping`, `icmp_ping`, or `http_get`.
- `target`: configured target string.
- `interval_seconds`: normal scheduled cadence.
- `timeout_seconds`: timeout for scheduled checks.
- `enabled`: whether scheduled/live checks may execute.
- `public_visible`: whether public summaries may include this probe when system
  visibility allows it.
- `scope`: `global` or `fixed`.

### Relationships

- Has zero or more `Fixed Probe Assignment` records when `scope = fixed`.
- Has many `Probe Result` records through `probe_id`.

### Validation Rules

- `scope = global`: fixed assignment list is ignored for effective coverage and
  the API returns an empty fixed `systems` list.
- `scope = fixed`: at least one selected system should be present when saving
  from the settings UI.
- Disabled probes do not execute even when global.

### State Transitions

- `fixed -> global`: delete or ignore old fixed assignment rows for effective
  coverage; all eligible current and future systems become covered.
- `global -> fixed`: create fixed assignment rows for the selected systems;
  unselected systems stop receiving new results.
- `enabled -> disabled`: stop scheduling all effective coverage pairs.
- `disabled -> enabled`: resume scheduling effective coverage pairs according
  to scope.

## Fixed Probe Assignment

Persistent mapping from one probe to one selected machine for fixed scope.

### Fields

- `id`
- `probe_id`
- `system_id`
- `enabled`

### Relationships

- Belongs to one `Latency Probe`.
- Belongs to one `Machine`.

### Validation Rules

- `(probe_id, system_id)` remains unique.
- Assignment rows for global probes are not used to decide effective coverage.
- Assignment rows for deleted systems may be removed by existing cascade
  behavior or ignored by effective coverage resolution.

## Effective Probe Coverage

Computed concept used by scheduling, live checks, chart loading, and public
summaries.

### Fields

- `probe_id`
- `system_id`
- `scope_source`: `global` or `fixed`
- `enabled`: derived from probe enabled state and assignment state where fixed.

### Relationships

- Combines `Latency Probe`, `Machine`, and optionally `Fixed Probe Assignment`.
- Produces `Probe Result` rows when checks execute.

### Validation Rules

- Global coverage includes every eligible system visible to the operation.
- Fixed coverage includes only enabled assignment rows.
- A probe/system pair must appear at most once per resolution call.
- Public resolution must additionally require public system visibility and
  public probe visibility.
- Authenticated resolution must require user/admin authorization for the system.

## Machine

Agent-backed monitored system that can execute probes.

### Fields Relevant to This Feature

- `id`
- `name`
- `status`
- `users`
- public visibility settings from existing public status configuration

### State Transitions

- `created`: becomes covered by all enabled global probes immediately for
  settings/listing purposes and at the next scheduled/live check opportunity for
  results.
- `private -> public`: public summaries may include global probe results if the
  probe is public-visible.
- `public -> private`: public summaries must stop showing this machine and its
  probe results.
- `removed`: historical results remain historical unless removed by existing
  cascade/retention behavior.

## Probe Result

Timestamped outcome from one machine executing one probe.

### Fields

- `probe_id`
- `system_id`
- `type`
- `target`
- `success`
- `latency_ms`
- `packet_loss_percent`
- `http_status`
- `failure_category`
- `error`
- `created`
- `bucket`

### Validation Rules

- Results are always tied to the executing system.
- Scope changes do not rewrite historical results.
- New systems covered by global probes start with no historical results and
  show pending/no-history state until new results arrive.
