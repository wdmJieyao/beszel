# Contract: YML Backup Schema

The backup document is YAML and must remain human-readable except for encrypted
secret envelopes. Field names are stable API contract names, not database column
dumps.

## Top-level shape

```yaml
meta:
  backupVersion: "1"
  sourceVersion: "0.18.7"
  createdAt: "2026-07-06T12:00:00Z"
  mode: "merge"
  sections:
    - systems
    - alerts
    - notifications
    - publicStatus
    - networkProbes

encryption:
  enabled: true
  algorithm: "xchacha20poly1305"
  kdf: "argon2id"
  salt: "base64..."

users:
  - stableId: "user_01"
    email: "admin@example.com"

systems:
  - stableId: "system_01"
    name: "edge-node"
    host: "10.0.0.10"
    port: 45876
    users:
      - email: "admin@example.com"
    token:
      encrypted: "base64..."

alerts:
  definitions:
    - stableId: "alert_01"
      systemStableId: "system_01"
      userEmail: "admin@example.com"
      name: "CPU"
      min: 1
      value: 80
  quietHours:
    - stableId: "quiet_01"
      userEmail: "admin@example.com"
      systemStableId: "system_01"
      type: "daily"
      start: "2000-01-01T22:00:00Z"
      end: "2000-01-01T23:00:00Z"

notifications:
  userSettings:
    - userEmail: "admin@example.com"
      emails:
        - "admin@example.com"
      webhooks:
        - encrypted: "base64..."
  telegram:
    settings:
      enabled: true
      pollingEnabled: true
      botToken:
        encrypted: "base64..."
    destinations:
      - stableId: "tgdest_01"
        name: "Ops channel"
        chatId: "-1001234567890"
        chatType: "channel"
        role: "read_only"
        enabled: true
        nodeScope:
          - "system_01"
        alertLevelScope:
          - "status"
          - "critical"

publicStatus:
  systems:
    - systemStableId: "system_01"
      publicEnabled: true
      publicName: "Edge"
      showCpu: true
      showMemory: true
      showDisk: true
      publicProbeStableIds:
        - "probe_01"

networkProbes:
  probes:
    - stableId: "probe_01"
      name: "Guangdong Telecom"
      type: "tcping"
      target: "example.com:443"
      intervalSeconds: 20
      timeoutSeconds: 3
      enabled: true
      scope: "global"
  assignments:
    - probeStableId: "probe_01"
      systemStableId: "system_01"
      enabled: true
```

## Required metadata

- `meta.backupVersion` is required.
- `meta.sourceVersion` is required.
- `meta.createdAt` is required and uses an ISO timestamp.
- `meta.mode` must be `merge` for this feature.
- `meta.sections` lists included sections.

## Stable identifiers

Every restorable item must include a `stableId`. During import, stable IDs are
the primary matching key. Display names, emails, and chat IDs are used for
preview readability, validation, and conflict reporting.

## Secret envelope

Sensitive values use object form instead of plaintext:

```yaml
botToken:
  encrypted: "base64..."
  nonce: "base64..."
  contentType: "telegram.botToken"
```

Secrets must not appear as raw strings in an encrypted export. If the operator
chooses an export without secrets, sensitive fields must be omitted or marked:

```yaml
botToken:
  redacted: true
```

## Compatibility rules

- Unknown top-level sections are reported and skipped.
- Newer incompatible section versions are rejected during preview.
- Missing optional sections are treated as empty.
- Missing referenced users require preview warnings or administrator mapping.
- Missing required references produce conflicts and block apply.
