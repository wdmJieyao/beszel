# Contract: Telegram Notification and Bot Menu API

All routes are under `/api/beszel` and require authenticated Beszel users unless
explicitly noted. Write routes require an administrator or non-readonly user as
specified by implementation tasks; settings that affect global bot credentials
require administrator access.

## Telegram Bot Settings

### GET `/telegram/settings`

Returns non-sensitive Telegram integration state.

**Response 200**

```json
{
  "enabled": true,
  "pollingEnabled": true,
  "botUsername": "beszel_alert_bot",
  "hasToken": true,
  "lastError": "",
  "updated": "2026-07-06T12:00:00Z"
}
```

The bot token is never returned.

### PUT `/telegram/settings`

Creates or replaces Telegram bot settings.

**Request**

```json
{
  "enabled": true,
  "pollingEnabled": true,
  "botToken": "123456:token"
}
```

`botToken` is optional when updating non-token fields. If supplied, it is stored
only in protected/encrypted form.

**Responses**

- `200 OK`: settings saved.
- `400 Bad Request`: invalid token format or unsupported settings.
- `401 Unauthorized`: missing auth.
- `403 Forbidden`: not an administrator.

### POST `/telegram/settings/test`

Validates the bot token by calling Telegram and returns non-sensitive bot
identity details.

**Response 200**

```json
{
  "ok": true,
  "botUsername": "beszel_alert_bot"
}
```

Failures return `200` with `ok=false` for delivery-style test errors or `400`
for invalid request shape.

## Telegram Destinations

### GET `/telegram/destinations`

Lists configured chat ID allowlist entries.

**Response 200**

```json
{
  "destinations": [
    {
      "id": "tgdest_01",
      "name": "Ops Admin",
      "chatId": "123456789",
      "chatType": "private",
      "role": "admin",
      "enabled": true,
      "nodeScope": [],
      "alertLevelScope": [],
      "muteUntil": null,
      "lastTestAt": "2026-07-06T12:05:00Z",
      "lastDeliveryAt": "2026-07-06T12:10:00Z",
      "lastError": ""
    }
  ]
}
```

### POST `/telegram/destinations`

Creates a destination.

**Request**

```json
{
  "name": "Public channel",
  "chatId": "-1001234567890",
  "chatType": "channel",
  "role": "read_only",
  "enabled": true,
  "nodeScope": ["system_id_1", "system_id_2"],
  "alertLevelScope": ["status", "critical"]
}
```

**Responses**

- `201 Created`: destination created.
- `400 Bad Request`: invalid chat ID, role, scope, or duplicate chat ID.
- `403 Forbidden`: insufficient permissions.

### PATCH `/telegram/destinations/{destinationId}`

Updates destination metadata, role, scopes, enabled state, or mute state.

**Request**

```json
{
  "enabled": false,
  "muteUntil": "2026-07-06T13:00:00Z"
}
```

### DELETE `/telegram/destinations/{destinationId}`

Deletes a destination. Delete is allowed only for authorized operators and must
not delete alert definitions or user settings.

### POST `/telegram/destinations/{destinationId}/test`

Sends a test message to the destination.

**Response 200**

```json
{
  "ok": true,
  "sentAt": "2026-07-06T12:05:00Z"
}
```

**Failure response 200**

```json
{
  "ok": false,
  "error": "bot cannot post to this chat"
}
```

Errors must not expose bot tokens or raw Telegram credentials.

## Telegram Bot Menu Contract

The baseline integration uses outbound Telegram long polling. No public inbound
route is required.

### Supported commands/actions

- `/start` or `/help`: safe help text and binding status.
- `/status`: administrator status overview.
- `/alerts`: administrator alert summary.
- `/systems`: administrator node list.
- `/system <id-or-index>`: administrator node summary.
- `/mute`: administrator temporary notification mute.
- `/unmute`: administrator restore notifications.

### Authorization rules

- `admin` destination: may use bot menu actions.
- `read_only` destination: may receive scoped non-sensitive alert summaries but
  cannot use privileged menu actions.
- unknown chat ID: no monitoring data is returned.

### Sanitization rules for read-only notifications

Read-only summaries may include node display name, broad alert type, severity,
event time, and resolved/triggered state. They must not include tokens, internal
hostnames/IPs, public probe targets, webhook URLs, raw metric payloads, admin
links, or backup details.
