# Research: Telegram Notifications and YML Configuration Backup

## Decision: Use direct Telegram Bot API for first-class Telegram support

**Rationale**: Beszel already supports generic Shoutrrr webhook URLs, including
Telegram-style delivery, but the requested bot menu requires receiving Telegram
commands and callback actions. A dedicated Telegram integration can share the
existing alert pipeline for message generation while using direct Telegram Bot
API calls for send, long polling, command menus, and callback responses.

**Alternatives considered**:

- Shoutrrr only: good for send-only alerts, insufficient for bot menu and
  role-aware chat allowlist.
- Telegram webhook: efficient, but requires the hub to be reachable by
  Telegram. Many self-hosted Beszel deployments are private, so it should not be
  the baseline.
- Third-party Go bot library: may simplify bot details but adds a new
  dependency. The required Bot API surface is small enough for Go `net/http`.

## Decision: Use hub-side long polling for Telegram menu events

**Rationale**: Long polling needs only outbound HTTPS from the hub and works
for private deployments. It also keeps the baseline feature panel-only and
avoids agent updates or public webhook setup.

**Alternatives considered**:

- Inbound webhook route: lower polling overhead, but needs public URL and TLS
  routing. Can be considered later as an optional advanced mode.
- Browser-driven polling: unreliable because menus must work when no browser is
  open.

## Decision: Panel-managed chat ID allowlist with roles

**Rationale**: The user explicitly chose manual chat ID allowlist. Roles let one
bot serve administrators with menu actions and read-only destinations with
scoped notification-only delivery. This avoids accidental disclosure to anyone
who discovers the bot username.

**Alternatives considered**:

- One-time Telegram binding code: convenient, but not what the user chose.
- Open `/start` requests with approval queue: more UX work and a larger attack
  surface.
- Notification-only Telegram channels: too limited because administrator menus
  are in scope.

## Decision: Read-only Telegram destinations receive scoped non-sensitive summaries

**Rationale**: Read-only users, groups, and channels should receive operational
signals without internal metadata. Node and alert-level scopes prevent noisy or
overbroad delivery, while message sanitization prevents leaking probe targets,
internal hosts, tokens, raw metrics, or admin links.

**Alternatives considered**:

- Same content as administrators: simpler but violates least privilege.
- Public-dashboard-only scope: safer but too restrictive for private operations.
- All alerts with no filtering: simpler but too noisy for shared channels.

## Decision: New config backup workflow, keep legacy `config.yml` separate

**Rationale**: Existing `internal/hub/config/config.go` syncs systems from
`config.yml` on restart and deletes systems absent from the file. The new
feature requires full panel configuration export and previewed merge restore
with no default deletion. Keeping a separate backup workflow avoids breaking
existing `config.yml` users and avoids importing destructive semantics into the
new feature.

**Alternatives considered**:

- Extend current `config.yml`: conflicts with merge restore and would make the
  legacy startup behavior riskier.
- Create database backup dumps: too broad and includes runtime data, auth data,
  and implementation details outside the requested configuration scope.

## Decision: Versioned YML schema with stable identifiers and user-readable refs

**Rationale**: Stable identifiers prevent renamed systems/probes/destinations
from being recreated accidentally. Names and emails remain in the document for
operator readability and conflict reporting. User accounts are referenced but
not created or restored as authentication records.

**Alternatives considered**:

- Name-only matching: readable but breaks on rename and can collide.
- Always create new records: avoids overwrites but duplicates configuration and
  breaks relationships.
- Per-section ad hoc matching only: useful for edge cases, but stable IDs
  should remain the primary contract.

## Decision: Encrypt sensitive values in export; never write plaintext secrets

**Rationale**: The user expects restorable backups without exposing bot tokens,
agent tokens, webhook URLs, or other secrets in plaintext. Authenticated
encryption with an operator-provided credential lets the export remain portable
without depending on the source instance database key.

**Alternatives considered**:

- Redact all secrets: safest but not fully restorable.
- Plaintext with warning: too easy to leak.
- Separate secret file: more operational complexity and still needs encryption.

## Decision: Merge restore by default

**Rationale**: Production restore should not remove target-only systems,
probes, or notification channels just because they are absent from a backup.
The preview must show creates, updates, preserved target-only records, skips,
conflicts, encrypted sections, and failures before apply.

**Alternatives considered**:

- Mirror restore: useful for strict disaster recovery, but deletion risk is too
  high for default behavior.
- Create-only restore: safer but cannot repair or update existing config.

## Decision: Baseline feature is hub-only and does not require agent changes

**Rationale**: Telegram notifications and configuration backup operate on hub
state, alert records, and panel settings. Existing agents can continue reporting
unchanged.

**Alternatives considered**:

- Agent-assisted backup: unnecessary because the backed-up data is hub
  configuration, not host runtime state.
- Agent-side Telegram: wrong ownership; notifications are hub decisions.

## Decision: Keep unrelated latency refresh fixes in the latency feature line

**Rationale**: The user may choose to release node-detail latency chart realtime
refresh fixes together with Telegram/YML work, but those fixes change a separate
behavioral surface, have their own acceptance logic, and already belong to the
existing latency-related specs. Keeping the feature boundary explicit prevents
plan/tasks drift and avoids mixing unrelated regression work into Telegram/YML
contracts.

**Alternatives considered**:

- Move latency refresh fixes into this feature: would blur scope and make the
  resulting plan and validation artifacts harder to reason about.
- Ignore co-release intent completely: would keep scope pure, but would not
  document the operator expectation that separate features may still ship in one
  release batch.
