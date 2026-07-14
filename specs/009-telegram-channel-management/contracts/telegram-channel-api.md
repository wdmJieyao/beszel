# Contract: Telegram Channel And Policy API

All paths are under `/api/beszel`. All operations require an authenticated
administrator. Bot Tokens are accepted only in request bodies and never returned.

## Bot Settings Verification

### POST `/telegram/settings/test`

Tests an entered Token when `botToken` is non-empty; otherwise tests the saved
Token. This action does not persist the entered Token.

**Request**

```json
{
  "botToken": "optional-unsaved-token"
}
```

**Response 200: all stages successful**

```json
{
  "ok": true,
  "botUsername": "beszel_alert_bot",
  "stages": {
    "credentials": { "ok": true, "error": "" },
    "commandMenu": { "ok": true, "error": "" }
  }
}
```

**Response 200: credentials valid, menu initialization failed**

```json
{
  "ok": false,
  "botUsername": "beszel_alert_bot",
  "stages": {
    "credentials": { "ok": true, "error": "" },
    "commandMenu": { "ok": false, "error": "sanitized upstream error" }
  }
}
```

Invalid request shape or Token format returns `400`. Authentication failures use
`401`/`403`. Upstream errors never expose the Token.

## Telegram Channels

The existing destination resource becomes the channel compatibility resource.

### GET `/telegram/destinations`

Returns channels with policy count and non-sensitive health state.

```json
{
  "destinations": [
    {
      "id": "channel_01",
      "name": "Operations",
      "chatId": "-100123456",
      "chatType": "group",
      "role": "read_only",
      "enabled": true,
      "muteUntil": null,
      "lastTestAt": "2026-07-10T08:00:00Z",
      "lastDeliveryAt": "2026-07-10T08:05:00Z",
      "lastError": "",
      "policyCount": 2,
      "nodeScope": [],
      "alertLevelScope": []
    }
  ]
}
```

The final two scope fields are deprecated compatibility projections of the
default policy and are omitted after the compatibility window.

### POST `/telegram/destinations`

Creates one unique channel. A new client sends channel fields only. A legacy
request containing inline scope fields creates a default policy transactionally.

Duplicate Chat ID returns:

```json
{
  "status": 409,
  "message": "Telegram channel already exists",
  "data": { "existingDestinationId": "channel_01" }
}
```

### PATCH `/telegram/destinations/{destinationId}`

Updates channel name, Chat ID, chat type, role, enabled state, or mute state.
Legacy inline scopes update the default policy for compatibility.

### DELETE `/telegram/destinations/{destinationId}`

Deletes the channel and all child policies. Returns `204 No Content`. A missing
channel returns `404`. Bot settings and other channels are never deleted.

### POST `/telegram/destinations/{destinationId}/test`

Sends one test message and updates channel test health. Existing response shape
is retained.

## Notification Policies

### GET `/telegram/destinations/{destinationId}/policies`

```json
{
  "policies": [
    {
      "id": "policy_01",
      "destinationId": "channel_01",
      "name": "生产节点状态",
      "enabled": true,
      "nodeScopeMode": "selected",
      "nodeScope": ["system_01", "system_02"],
      "alertLevelScope": ["status", "cpu"]
    }
  ]
}
```

### POST `/telegram/destinations/{destinationId}/policies`

Creates a policy and returns `201 Created`. Validation rules:

- Parent channel must exist.
- Name is required and unique within the channel.
- Mode is `all` or `selected`.
- `all` requires empty node scope.
- `selected` requires one or more valid systems.
- Alert categories must use the supported vocabulary.

### PATCH `/telegram/destinations/{destinationId}/policies/{policyId}`

Updates name, enabled state, node mode/scope, or alert categories. The policy
must belong to the path channel; otherwise return `404` without leaking another
channel's policy.

### DELETE `/telegram/destinations/{destinationId}/policies/{policyId}`

Deletes only the policy and returns `204 No Content`. The parent channel and
sibling policies remain.

## Role And Matching Contract

- All enabled policies apply node and alert-category filtering for both roles.
- Multiple matching policies produce one delivery per channel.
- `admin` receives complete alert content; `read_only` receives sanitized content.
- Privileged Bot commands require an enabled private `admin` channel.
- Group/channel `admin` records may receive full content but cannot execute
  privileged commands; the UI must state this before save.
- A channel with no enabled policies receives no alerts but can still be tested.
