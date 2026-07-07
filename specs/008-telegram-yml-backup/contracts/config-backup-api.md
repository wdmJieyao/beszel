# Contract: Configuration Backup API

All routes are under `/api/beszel` and require administrator access. The legacy
`GET /config-yaml` endpoint remains available for current system-only
configuration export; this feature adds a new backup resource with merge restore
semantics.

## POST `/config-backups/exports`

Creates a downloadable YML backup document from current panel configuration.

**Request**

```json
{
  "includeSecrets": true,
  "encryptionCredential": "operator-provided-passphrase",
  "sections": [
    "systems",
    "alerts",
    "notifications",
    "publicStatus",
    "networkProbes"
  ]
}
```

`includeSecrets=true` requires `encryptionCredential`. If `includeSecrets=false`,
sensitive values must be omitted or redacted and never returned as plaintext.

**Response 200**

```json
{
  "filename": "beszel-config-20260706-120000.yml",
  "contentType": "application/x-yaml",
  "backupVersion": "1",
  "warnings": [],
  "content": "meta:\n  backupVersion: \"1\"\n..."
}
```

**Responses**

- `200 OK`: export generated.
- `400 Bad Request`: invalid section name or missing encryption credential.
- `403 Forbidden`: user is not an administrator.
- `500 Internal Server Error`: export failed.

## POST `/config-backups/validations`

Parses a YML backup and returns a restore preview. Does not mutate data.

**Request**

```json
{
  "content": "meta:\n  backupVersion: \"1\"\n...",
  "decryptionCredential": "operator-provided-passphrase"
}
```

`decryptionCredential` is required only when encrypted sensitive values are
present and the caller wants validation to include decrypted secret sections.

**Response 200**

```json
{
  "previewId": "sha256:abc123",
  "mode": "merge",
  "backupMeta": {
    "backupVersion": "1",
    "sourceVersion": "0.18.7",
    "createdAt": "2026-07-06T12:00:00Z"
  },
  "summary": {
    "create": 2,
    "update": 4,
    "preserve": 1,
    "skip": 0,
    "conflict": 0,
    "error": 0
  },
  "items": [
    {
      "section": "systems",
      "stableId": "system_01",
      "displayName": "edge-node",
      "action": "update",
      "reason": "stable identifier matched existing system"
    }
  ],
  "warnings": []
}
```

**Responses**

- `200 OK`: preview generated.
- `400 Bad Request`: invalid YML, unsupported backup version, or missing
  decryption credential for encrypted sections.
- `409 Conflict`: backup is parseable but contains unresolved conflicts.

## POST `/config-backups/restores`

Applies a previously previewed backup using merge restore.

**Request**

```json
{
  "content": "meta:\n  backupVersion: \"1\"\n...",
  "previewId": "sha256:abc123",
  "mode": "merge",
  "decryptionCredential": "operator-provided-passphrase"
}
```

`previewId` must match the content and options used in validation.

**Response 200**

```json
{
  "mode": "merge",
  "applied": {
    "created": 2,
    "updated": 4,
    "preserved": 1,
    "skipped": 0
  },
  "warnings": []
}
```

**Responses**

- `200 OK`: restore applied.
- `400 Bad Request`: invalid request or preview mismatch.
- `409 Conflict`: unresolved conflicts; apply rejected.
- `422 Unprocessable Entity`: decryption failed, references missing, or a
  section failed validation.

## Merge restore rules

- Match existing records by stable identifiers stored in the backup.
- Use display names and emails for preview readability and conflict reporting.
- Create missing records when required references can be resolved.
- Update matched records only for supported configuration fields.
- Preserve target-only records by default.
- Do not delete systems, probes, alerts, destinations, or public visibility rows
  in this feature.
