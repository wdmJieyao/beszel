# Data Model: Telegram Channels And Notification Policies

## Telegram Bot Settings

Existing singleton integration settings remain unchanged.

| Field | Meaning | Validation |
|-------|---------|------------|
| enabled | Enables Telegram delivery | Boolean |
| polling_enabled | Enables Bot command polling | Boolean |
| bot_token_encrypted | Protected Bot credential | Never returned or logged |
| bot_username | Last verified Bot username | Non-sensitive display value |
| last_poll_offset | Last consumed Telegram update | Non-negative integer |
| last_error | Sanitized runtime error | Maximum existing field length |

Bot verification returns stage results but does not persist an unsaved Token.

## Telegram Channel (`telegram_destinations`)

One record represents one Telegram Chat ID and owns authorization, health, and
mute state.

| Field | Meaning | Validation |
|-------|---------|------------|
| id | Stable channel identifier | Existing PocketBase ID |
| user | Optional related panel user | Existing user when set |
| name | Administrator-facing channel name | Required, trimmed |
| chat_id | Telegram chat identity | Required numeric string; globally unique |
| chat_type | private/group/supergroup/channel/unknown | Existing allowed values |
| role | admin/read_only | Required |
| enabled | Channel delivery and authorization switch | Boolean |
| mute_until | Channel-wide notification pause | Optional UTC timestamp |
| last_test_at | Last successful test | Runtime metadata |
| last_delivery_at | Last successful delivery | Runtime metadata |
| last_error | Sanitized latest channel error | Runtime metadata |
| node_scope | Legacy compatibility field | Retained, not canonical |
| alert_level_scope | Legacy compatibility field | Retained, not canonical |

### Channel Rules

- Chat ID uniqueness remains enforced.
- An enabled `admin` channel receives full matched message content.
- An enabled `read_only` channel receives sanitized matched message content.
- Privileged commands require `admin` plus `private` chat type.
- Deleting a channel deletes all related policies in the same operation.

## Notification Policy (`telegram_notification_policies`)

One record is one named routing template belonging to exactly one channel.

| Field | Meaning | Validation |
|-------|---------|------------|
| id | Stable policy identifier | PocketBase ID |
| destination | Parent Telegram channel | Required relation, cascade delete |
| name | Policy/template name | Required; unique within parent channel |
| enabled | Whether this policy participates in matching | Boolean |
| node_scope_mode | Dynamic all nodes or selected nodes | `all` or `selected` |
| node_scope | Explicit selected system IDs | Empty for all; non-empty for selected |
| alert_level_scope | Allowed alert categories | Empty means all categories |
| created/updated | Audit timestamps | Automatic |

### Policy Validation

- `node_scope_mode=all` requires an empty canonical node list.
- `node_scope_mode=selected` requires at least one valid system ID.
- Every alert category must belong to the supported vocabulary.
- Both roles use the same policy matching rules.
- A channel may own multiple policies, but policy names are unique within it.

### Match Semantics

For each enabled channel:

1. Ignore disabled policies.
2. A policy node-matches if mode is `all` or the alert system is selected.
3. A policy category-matches if its category list is empty or contains the
   alert category.
4. The channel matches if any policy matches both dimensions.
5. Deliver at most once to that channel for the alert.

## Migration Mapping

For every existing destination without a policy:

| Existing value | New default policy value |
|----------------|--------------------------|
| destination ID | New policy relation points to destination |
| destination name | Policy name `默认规则` (with deterministic conflict suffix if needed) |
| empty node_scope | node_scope_mode `all`, empty node_scope |
| non-empty node_scope | node_scope_mode `selected`, copied node IDs |
| alert_level_scope | Copied category IDs |
| destination enabled | Default policy enabled |

The destination remains enabled and retains all legacy values. Backfill is
idempotent: rerunning it must not create another default policy.

## Backup Transfer Objects

Notifications section version 2 represents:

- Bot settings (unchanged secret handling)
- User email/webhook settings (unchanged)
- Telegram channels without canonical scope fields
- Telegram notification policies with stable channel references

Version 1 destinations with inline scopes map to one channel and one default
policy during restore. Version 2 target-only channels and policies are preserved
under existing merge semantics.

## State Transitions

### Bot Verification

`untested -> credential verified -> menu initialized`

- Credential failure stops before menu initialization.
- Menu failure preserves successful credential identity in the response and
  reports the menu stage separately.
- Verification never saves an entered Token.

### Channel

`enabled <-> disabled`, `unmuted <-> muted`, `healthy <-> error`, `exists -> deleted`

Channel deletion is terminal and cascades policies.

### Policy

`enabled <-> disabled`, `all nodes <-> selected nodes`, `exists -> deleted`

Deleting the last policy leaves the channel available for testing and private
admin commands but it receives no alert deliveries until another policy exists.
