# Contract: Public System Probe Visibility API

## List Public System Settings

`GET /api/beszel/public/systems`

### Response

```json
{
  "systems": [
    {
      "id": "system_123",
      "name": "node-a",
      "status": "up",
      "publicEnabled": true,
      "publicName": "Tokyo A",
      "showCpu": true,
      "showMemory": true,
      "showDisk": true,
      "publicProbeIds": ["probe_1", "probe_2"]
    }
  ]
}
```

### Rules

- `publicProbeIds` is the source-of-truth selection for anonymous probe display
  on that system.
- A public-enabled system may return an empty `publicProbeIds` list.
- The response must not imply that all probes are public for that system unless
  every visible probe ID is explicitly listed.

## Update Public System Settings

`PATCH /api/beszel/public/systems/{systemId}`

### Request

```json
{
  "publicEnabled": true,
  "publicName": "Tokyo A",
  "showCpu": true,
  "showMemory": true,
  "showDisk": true,
  "publicProbeIds": ["probe_1", "probe_2"]
}
```

### Rules

- Replacing `publicProbeIds` updates only the selected probe list for the
  addressed system.
- Invalid or unauthorized probe IDs must return a client-visible validation
  error or be rejected consistently with existing settings update behavior.
- Probe IDs that do not effectively cover the addressed system must not become
  anonymously visible.
- Omitting `publicProbeIds` from a patch must not clear selection implicitly
  unless the request semantics explicitly replace the full resource.

## Anonymous Public Dashboard Response

`GET /api/beszel/public/status`

### Probe Visibility Rules

- A system appears only when `publicEnabled = true`.
- A probe summary appears for a system only when all are true:
  - the system is public-enabled,
  - the probe ID is selected in `publicProbeIds`,
  - the probe effectively covers the system,
  - the probe still exists and is eligible for summary generation.
- Unselected probe names, targets, latest results, and series must not be
  emitted for that system.

## Migration Compatibility

- Existing deployments seed `publicProbeIds` from legacy public probe behavior.
- Migration preserves only previously visible probe/system combinations.
- Migration must be idempotent and must not widen exposure on re-run.
