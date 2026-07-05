# Data Model: Public Probe Visibility and Refresh Commands

## Public System Visibility

One row per monitored system controlling what anonymous visitors may see.

### Fields

- `system_id`: stable relation to the monitored system.
- `public_enabled`: whether the system appears on the anonymous public dashboard.
- `public_name`: optional public-facing display name.
- `show_cpu`
- `show_memory`
- `show_disk`
- `public_probe_ids`: selected latency probe IDs allowed to appear for this
  system on the anonymous public dashboard.

### Relationships

- Belongs to one `System`.
- References zero or more `Latency Probe` rows through `public_probe_ids`.

### Validation Rules

- Only one visibility row exists per system.
- When `public_enabled = false`, `public_probe_ids` may remain stored but must
  not cause public exposure by themselves.
- A probe ID in `public_probe_ids` is only effective when the probe still
  exists and effectively covers the system.
- Newly public systems default to `public_probe_ids = []`.

### State Transitions

- `private -> public`: system becomes eligible for anonymous display, but probe
  selection starts empty after migration-era defaults.
- `public -> private`: system is removed from anonymous output, regardless of
  stored probe selection.
- `selection update`: replacing `public_probe_ids` changes only future anonymous
  rendering and does not rewrite historical probe results.

## Latency Probe

Configured network check target that may be eligible for authenticated and/or
anonymous display.

### Fields Relevant to This Feature

- `id`
- `name`
- `type`
- `target`
- `enabled`
- `scope`
- `systems`
- `public_visible` (legacy migration seed / compatibility input)

### Relationships

- Has effective coverage over zero or more `System` rows.
- May be selected by zero or more `Public System Visibility` rows.

### Validation Rules

- A probe may be selected publicly for a system only if its effective coverage
  includes that system.
- Legacy `public_visible` is not the primary anonymous display owner after
  migration.

## Effective Public Probe Selection

Computed concept used to decide which probe series may appear for a public VPS.

### Fields

- `system_id`
- `probe_id`
- `selected`: whether the probe ID is present in that system’s
  `public_probe_ids`.
- `covered`: whether the probe’s effective scope includes the system.
- `visible`: `public_enabled && selected && covered`

### Validation Rules

- Anonymous responses include probe metadata and series only when `visible` is
  true.
- Selection without coverage is ignored safely and should be removable from the
  settings experience.
- Legacy probe visibility does not make `visible` true by itself after
  migration.

## Legacy Probe Public Visibility

Compatibility-era concept representing the old probe-level public toggle.

### Fields

- `probe_id`
- `public_visible`

### Purpose

- Source for seeding per-VPS public selection during migration.
- Temporary compatibility data until all runtime public filtering is driven by
  `public_probe_ids`.

### Validation Rules

- Must not widen exposure during migration.
- Must not be the user-facing source of truth after this feature lands.

## Generated Docker Run Command

Copyable deployment command shown to administrators.

### Fields

- `image`
- `container_name`
- `cleanup_container`: best-effort remove existing container with that name.
- `cleanup_image`: best-effort remove local image before refresh.
- `pull`: explicit image pull step.
- `run_args`: container runtime flags and environment settings.

### Validation Rules

- Cleanup steps tolerate missing container/image.
- Pull and run failures remain visible to the operator.
- Command semantics are consistent across all generated Docker run variants.
