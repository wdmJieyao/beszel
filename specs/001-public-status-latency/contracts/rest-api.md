# REST API Contract: Public Status Page and Network Probe Trends

## Anonymous Public Resources

### GET `/api/beszel/public/status`

Returns the public-safe dashboard payload used by `/`.

**Authentication**: Not required.

**Query Parameters**

- `range`: optional chart window. Accepted values are defined by implementation, but must include recent, daily, and longer-range windows.

**200 Response**

```json
{
  "generatedAt": "2026-06-28T12:00:00Z",
  "systems": [
    {
      "id": "system_public_id",
      "name": "Tokyo VPS",
      "status": "up",
      "freshness": "2026-06-28T11:59:30Z",
      "metrics": {
        "cpuPercent": 12.4,
        "memoryPercent": 48.9,
        "diskPercent": 61.2
      },
      "probes": [
        {
          "id": "probe_id",
          "name": "China Telecom TCP",
          "type": "tcping",
          "targetLabel": "example.com:443",
          "latest": {
            "success": true,
            "latencyMs": 42.3,
            "created": "2026-06-28T11:59:00Z"
          },
          "series": [
            {
              "created": "2026-06-28T11:55:00Z",
              "success": true,
              "latencyMs": 41.8
            }
          ]
        }
      ]
    }
  ]
}
```

**Rules**

- Response MUST include only public-enabled systems.
- Response MUST omit private host, port, user relationships, tokens, owner-only IDs, and admin actions.
- Probe summaries/charts are included only when the executing system is public and the probe has not been manually hidden from public display.
- Empty state is represented by `"systems": []`.

## Authenticated Admin Resources

### GET `/api/beszel/public/systems`

Lists systems and their public visibility settings for admin UI.

**Authentication**: Required; admin or non-readonly role as implementation policy decides.

**200 Response**

```json
{
  "systems": [
    {
      "id": "system_id",
      "name": "Tokyo VPS",
      "status": "up",
      "publicEnabled": false,
      "publicName": ""
    }
  ]
}
```

### PATCH `/api/beszel/public/systems/{systemId}`

Updates public visibility for one system.

**Authentication**: Required; must reject readonly users.

**Request**

```json
{
  "publicEnabled": true,
  "publicName": "Tokyo VPS",
  "showCpu": true,
  "showMemory": true,
  "showDisk": true
}
```

**200 Response**

```json
{
  "id": "system_id",
  "publicEnabled": true,
  "publicName": "Tokyo VPS",
  "showCpu": true,
  "showMemory": true,
  "showDisk": true
}
```

**Errors**

- `400`: invalid public name or metric settings.
- `401`: missing authentication.
- `403`: user cannot update public visibility.
- `404`: system not found or not visible to user.

## Network Probe Resources

### GET `/api/beszel/network-probes`

Lists configured probes and execution settings for authenticated users.

**Authentication**: Required.

**200 Response**

```json
{
  "probes": [
    {
      "id": "probe_id",
      "name": "China Telecom TCP",
      "type": "tcping",
      "target": "example.com:443",
      "intervalSeconds": 60,
      "timeoutSeconds": 5,
      "enabled": true,
      "publicVisible": true,
      "systems": ["system_id"],
      "executionMode": "automatic"
    }
  ]
}
```

### POST `/api/beszel/network-probes`

Creates a probe.

**Authentication**: Required; must reject readonly users.

**Request**

```json
{
  "name": "China Telecom TCP",
  "type": "tcping",
  "target": "example.com:443",
  "intervalSeconds": 60,
  "timeoutSeconds": 5,
  "enabled": true,
  "publicVisible": true,
      "systems": ["system_id"],
      "executionMode": "automatic"
    }
```

**201 Response**: created probe object.

**Validation**

- `type` must be `tcping`, `icmp_ping`, or `http_get`.
- `target` must match type-specific format.
- `intervalSeconds` must be greater than `timeoutSeconds`.
- `systems` must contain execution-capable systems visible to the user.
- `executionMode` defaults to `automatic`; `advanced` exposes explicit execution-node binding in the UI.

### PATCH `/api/beszel/network-probes/{probeId}`

Updates a probe and its assignments.

**Authentication**: Required; must reject readonly users.

**Request**: Same fields as create; partial updates are allowed.

**200 Response**: updated probe object.

### DELETE `/api/beszel/network-probes/{probeId}`

Deletes or disables a probe.

**Authentication**: Required; must reject readonly users.

**204 Response**: no content.

### GET `/api/beszel/network-probes/{probeId}/results`

Returns probe history for authenticated views.

**Authentication**: Required.

**Query Parameters**

- `system`: optional executing system ID.
- `range`: required chart window.

**200 Response**

```json
{
  "probeId": "probe_id",
  "series": [
    {
      "systemId": "system_id",
      "created": "2026-06-28T11:59:00Z",
      "success": true,
      "latencyMs": 42.3,
      "packetLossPercent": 0,
      "httpStatus": null
    }
  ]
}
```

## Compatibility

- Existing authenticated system, stats, alerts, and container endpoints remain unchanged.
- Anonymous public endpoints never return full `systems` records.
- Internal agent websocket contracts are documented separately in `agent-probe.md`.
