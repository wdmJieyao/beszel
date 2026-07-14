# Research: Telegram Channel Management Improvements

## Decision 1: Keep Chat ID Unique And Add Child Policies

**Decision**: Keep `telegram_destinations` as one channel per unique Chat ID and
add `telegram_notification_policies` as a one-to-many child collection.

**Rationale**: Chat identity, role, mute state, health, and menu authorization
are channel-level concerns. Node and alert scopes are repeatable routing rules.
Direct duplicate destination rows would make command authorization and delivery
health ambiguous and would require de-duplication after every send candidate.

**Alternatives considered**:

- Allow duplicate destination rows: rejected because overlapping rows can send
  duplicates and disagree on role/mute state.
- Keep one combined policy per Chat ID: rejected because it does not satisfy the
  confirmed multi-template requirement.

## Decision 2: Additive Backfill With Legacy Fields Retained

**Decision**: Add the policy collection and backfill one default policy for each
existing destination. Keep legacy destination scope fields during this feature.

**Rationale**: The production database already contains destination records.
An additive migration avoids data loss, permits controlled compatibility reads,
and leaves a downgrade path. Empty legacy node scope maps to dynamic all-node
mode; non-empty scope maps to selected-node mode.

**Alternatives considered**:

- Move fields and remove them in one migration: rejected as unnecessarily
  destructive and difficult to roll back.
- Interpret destinations without backfill at runtime: rejected because it
  creates permanent dual routing paths and inconsistent backup behavior.

## Decision 3: Match Policies With OR Semantics And Send Once Per Channel

**Decision**: A channel matches when any enabled child policy matches both the
alert node and alert category. Delivery sends one message to that channel even
if several policies match.

**Rationale**: Policies represent alternative reasons for a channel to receive
an alert, not separate copies. Grouping candidates by destination before send is
deterministic and supports the no-duplicate requirement.

**Alternatives considered**:

- Send once per matching policy: rejected because overlap creates duplicates.
- Merge all policies into one stored scope: rejected because independent policy
  naming, editing, and deletion would be lost.

## Decision 4: Persist Explicit Node Scope Mode

**Decision**: Policies store `node_scope_mode` as `all` or `selected` in addition
to the selected node IDs. Selected mode requires at least one node.

**Rationale**: Empty arrays previously meant all nodes, making “selected but
empty” impossible to distinguish. An explicit mode makes future-node behavior
testable and prevents ambiguous saves.

**Alternatives considered**:

- Infer mode only from an empty array: rejected because clearing selected nodes
  would silently broaden delivery to all nodes.
- Materialize every node ID for all-node mode: rejected because future nodes
  would not be included automatically.

## Decision 5: Apply Scopes To Both Roles

**Decision**: Node and alert-category scopes apply identically to administrator
and read-only policies. Role controls content detail and interactive authority.
Administrator messages retain full details; read-only messages are sanitized.
Privileged menu commands require an administrator destination in a private chat.

**Rationale**: A visible scope must have the same routing meaning regardless of
role. Separating routing from authorization removes the current silent no-op for
administrator alert scopes.

**Alternatives considered**:

- Preserve admin-always-receives-all behavior: rejected by the confirmed Q2
  choice and because it contradicts visible controls.
- Replace roles with many permission toggles: rejected as unnecessary scope and
  migration complexity for this feature.

## Decision 6: Return Staged Bot Verification

**Decision**: Bot verification reports credential/identity and command-menu
stages separately. Telegram success envelopes accept scalar, object, array, or
empty results when callers do not request a typed result.

**Rationale**: Telegram methods such as command initialization return a boolean
result, while identity retrieval returns an object. Treating all ignored results
as objects caused a valid Bot to fail verification. Staged output tells the
administrator whether messaging works even if menu registration fails.

**Alternatives considered**:

- Remove command initialization from testing: rejected because initialization
  still needs verification and actionable failure reporting.
- Treat any menu failure as an opaque Bot failure: rejected because it hides a
  healthy Token and working message path.

## Decision 7: Channel Delete Cascades Policies After Confirmation

**Decision**: Deleting a channel explicitly removes its child policies in the
same operation after a confirmation naming the channel and policy count.
Deleting one policy leaves the channel and sibling policies intact.

**Rationale**: Policies cannot operate without a Chat ID, and an explicit
channel delete should fully remove authorization and delivery for that chat.
The confirmation makes the impact visible.

**Alternatives considered**:

- Block channel deletion until policies are deleted manually: rejected as
  unnecessary friction after explicit confirmation.
- Orphan policies: rejected because it creates invalid routing state.

## Decision 8: Upgrade Only The Notifications Backup Section

**Decision**: Export notifications section version 2 with channel and policy
arrays. Restore accepts version 1 inline destination scopes by creating one
default policy and accepts version 2 directly.

**Rationale**: Section-level versions already exist, so unrelated backup
sections do not need a global version bump. Compatibility restore protects
existing production exports.

**Alternatives considered**:

- Increment the whole backup format: rejected because only notifications change.
- Stop accepting version 1: rejected because users may need to restore existing
  backups after upgrading.

## Decision 9: Keep Existing Destination URLs During Compatibility

**Decision**: Existing destination URLs remain channel resources. New policy
resources are nested beneath a destination. Legacy inline scope request fields
operate on the default policy during a documented compatibility period.

**Rationale**: This minimizes frontend and external-client disruption while
making new ownership explicit. Returning `409 Conflict` with the existing
destination ID is clearer than the current generic duplicate validation error.

**Alternatives considered**:

- Rename destinations to channels immediately: rejected because it adds broad
  compatibility churn without user value.

## Decision 10: Use Existing UI Primitives And Pure Selection Helpers

**Decision**: Use the existing confirmation dialog, checkbox, segmented/select,
tooltip, input, and list patterns. Keep mode/select-all/search transformations in
pure frontend helpers with unit tests.

**Rationale**: The repository already supplies accessible primitives and a Node
test harness. Pure helpers provide deterministic coverage without introducing a
new component or state library.

**Alternatives considered**:

- Add a new form or data-grid dependency: rejected because 500 nodes are within
  the capability of current searchable/scrollable patterns.
