# Research: Public Probe Visibility and Refresh Commands

## Decision 1: Store per-VPS public probe selection on `public_system_visibility`

### Decision

Extend the existing `public_system_visibility` collection with a multi-value
relation/list of selected probe IDs for anonymous display.

### Rationale

- The public dashboard already has exactly one visibility row per system.
- The new requirement is system-owned visibility, not probe-owned visibility.
- A single-row-per-system model keeps the `GET/PATCH /api/beszel/public/systems`
  contract straightforward and avoids a second join table just for UI state.
- Migration can seed selected probe IDs in-place during the same collection
  upgrade that introduces the new field.

### Alternatives considered

- Create a new `public_system_probe_visibility` collection:
  more flexible but adds extra queries, uniqueness rules, and admin API
  complexity without clear value for this scope.
- Keep using probe-level `public_visible` plus client filtering:
  wrong ownership model and cannot express different visibility per public VPS.

## Decision 2: Treat probe-level `public_visible` as migration seed, not runtime source of truth

### Decision

Use legacy `network_probes.public_visible` only to seed the new per-VPS probe
selection during migration and compatibility checks. After migration, anonymous
public dashboard output is determined by per-VPS selected probe IDs combined
with effective probe coverage.

### Rationale

- The user explicitly wants public visibility managed from the public dashboard
  VPS settings page.
- Keeping the legacy field as a runtime filter would create dual ownership and
  contradictory states.
- Migration needs a source to preserve existing production visibility without
  exposing more than before.

### Alternatives considered

- Delete `public_visible` immediately:
  risky for deployed data and harder to roll through a compatibility window.
- Require both legacy flag and per-VPS selection at runtime:
  creates unnecessary hidden coupling and makes operator behavior harder to
  predict.

## Decision 3: Seed only previously visible probe/VPS pairs during migration

### Decision

For each public VPS, seed selected probe IDs only for probes that were both
legacy-public and actually visible for that VPS before the upgrade. Do not
expand one public probe to every public VPS.

### Rationale

- This matches the clarified requirement and prevents accidental disclosure.
- It preserves the visible public dashboard shape that operators already rely
  on in production.
- It makes the migration idempotent: once a row has selected probe IDs, reruns
  can preserve them instead of recomputing a wider set.

### Alternatives considered

- Seed every legacy-public probe onto every public VPS:
  simpler migration logic, but violates the requirement not to widen exposure.
- Seed nothing and require manual reconfiguration:
  safer than widening, but breaks existing production dashboards on upgrade.

## Decision 4: Make public dashboard settings the only editing surface

### Decision

Public probe visibility is edited only in
`internal/site/src/components/routes/settings/public-status.tsx`. The probe
settings screen removes the public toggle and no normal VPS edit surface
duplicates the per-VPS public selection control.

### Rationale

- One owner avoids contradictory edits.
- The user explicitly chose the public dashboard settings page as the control
  point.
- This keeps “what is public” in the same place as “which VPS is public” and
  “which metrics are public.”

### Alternatives considered

- Dual-edit from public settings and probe settings:
  synchronization complexity with no user benefit.
- Keep the control only in probe settings:
  conflicts directly with the requested workflow.

## Decision 5: Refresh Docker run commands with container cleanup, image cleanup, pull, and run

### Decision

Generated Docker run commands should follow this semantic sequence:

1. Remove any existing container with the target container name, best effort.
2. Remove the old local image for that command, best effort.
3. Pull the target image.
4. Start a new container with the requested options.

### Rationale

- Removing only the image is not enough when a same-name container already
  exists.
- Pulling before run guarantees the latest published image is used instead of a
  cached local image.
- Best-effort cleanup keeps the command rerunnable on fresh hosts.

### Alternatives considered

- Use `docker run --pull always` only:
  still fails when the container name already exists.
- Remove only the image:
  does not solve same-name container conflicts.
- Stop and start existing containers in place:
  harder to express in one portable copied command and less predictable across
  changed options.

## Decision 6: Scope Docker command changes to generated `docker run` helpers, not compose samples

### Decision

This feature changes generated Docker run copy actions in the product UI. The
supplemental compose files remain out of scope for this request.

### Rationale

- The requirement explicitly targets “产生docker run 命令的地方”.
- The current generated `docker run` surface lives in
  `internal/site/src/components/install-dropdowns.tsx`.
- Compose files already support pull/recreate through different workflows and do
  not share the same UX or one-liner constraints.

### Alternatives considered

- Rewrite all compose examples too:
  extra churn not asked for by this feature.
- Ignore future panel run generators:
  acceptable for current code, but the helper should still be written so any
  generated panel run surface can reuse the same semantics.
