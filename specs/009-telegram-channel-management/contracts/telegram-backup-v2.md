# Contract: Notifications Backup Section Version 2

The global backup version remains unchanged. Only the `notifications` entry in
`meta.sectionVersions` advances to `"2"`.

## Version 2 Export Shape

```yaml
meta:
  backupVersion: "1"
  sections: [notifications]
  sectionVersions:
    notifications: "2"
notifications:
  telegram:
    settings:
      enabled: true
      pollingEnabled: true
      botUsername: beszel_alert_bot
      botToken:
        redacted: true
    destinations:
      - stableId: channel_01
        name: Operations
        chatId: "-100123456"
        chatType: group
        role: read_only
        enabled: true
    policies:
      - stableId: policy_01
        destinationStableId: channel_01
        name: Production status
        enabled: true
        nodeScopeMode: selected
        nodeScope: [system_01, system_02]
        alertLevelScope: [status]
```

Secret encryption/redaction rules remain unchanged.

## Version 1 Restore Compatibility

A version 1 destination with inline `nodeScope` and `alertLevelScope` creates or
updates its channel and one deterministic default policy:

```yaml
destinations:
  - stableId: channel_01
    chatId: "-100123456"
    nodeScope: []
    alertLevelScope: [status]
```

Mapping:

- Empty `nodeScope` -> `nodeScopeMode: all`
- Non-empty `nodeScope` -> `nodeScopeMode: selected`
- Inline alert scope -> default policy alert scope
- Destination enabled -> both channel and default policy enabled on initial import

## Preview And Merge Rules

- Channels match by stable ID first and detect Chat ID conflicts explicitly.
- Policies match by stable ID and must reference a channel present in the backup
  or already available on the target.
- Target-only channels and policies are preserved.
- Unknown systems in selected scopes are conflicts, not silently dropped.
- Deleting by absence is prohibited.
- A version 2 notifications section is restored transactionally so channel and
  policy changes do not partially diverge.
- A newer unsupported notifications section is warned and skipped under the
  existing section-version rules.

## Export/Restore Acceptance

- Exported v2 data contains no canonical policy scopes on channel records.
- Restoring v2 reproduces all channel and policy fields.
- Restoring v1 produces one policy per legacy destination without duplicates.
- Repeating the same restore is idempotent.
- Redacted Bot Tokens preserve the target Token.

Executable compatibility fixtures are kept at:

- `internal/hub/testdata/telegram_backup/notifications-v1.yml`
- `internal/hub/testdata/telegram_backup/notifications-v2.yml`
