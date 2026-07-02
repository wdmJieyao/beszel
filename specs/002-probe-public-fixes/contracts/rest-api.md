# REST API Contract: Probe Chart and Public Dashboard Fixes

## Existing Resource: GET `/api/beszel/network-probes/{probeId}/results`

Returns probe history for authenticated views. The existing endpoint may be
extended additively to include failure categories and safe labels.

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
      "created": "2026-06-29T12:00:00Z",
      "success": false,
      "latencyMs": null,
      "failureCategory": "timeout",
      "error": "连接超时",
      "packetLossPercent": null,
      "httpStatus": null
    }
  ]
}
```

**Rules**

- Runtime probe failures return `200` with failed result points.
- Save-time validation errors return a client error from create/update probe endpoints.
- `failureCategory` is additive and optional for older stored results.
- Result payloads must not expose credentials, local filesystem paths, or private connection internals.

## Existing Resource: GET `/api/beszel/network-probes`

Lists configured probes and assignments for authenticated views. The frontend
uses this list to build combined node detail charts.

**Authentication**: Required.

**Rules**

- Response must include enough stable probe metadata to label each chart series.
- Multiple probes assigned to the same node may be grouped by type in the UI.
- Existing fields remain backward compatible.

## Existing Resource: POST/PATCH `/api/beszel/network-probes`

Creates or updates probe definitions.

**Validation Rules**

- TCPing targets must be validated as `host:port`.
- Invalid TCPing targets return a client validation error and do not create a runtime failed result.
- Valid but unreachable TCPing targets are allowed and become runtime result failures.

**Error Response Example**

```json
{
  "message": "TCPing 目标必须使用 host:port 格式",
  "code": 400,
  "data": {
    "target": {
      "code": "invalid_target",
      "message": "TCPing 目标必须使用 host:port 格式"
    }
  }
}
```

## Existing Resource: GET `/api/beszel/public/status`

Returns the anonymous public dashboard payload used by `/`.

**Authentication**: Not required.

**200 Response**

```json
{
  "generatedAt": "2026-06-29T12:00:00Z",
  "systems": [
    {
      "id": "system_public_id",
      "name": "Local Test Node",
      "status": "up",
      "freshness": "2026-06-29T11:59:55Z",
      "metrics": {
        "cpuPercent": 12,
        "memoryPercent": 42,
        "diskPercent": 58,
        "unavailable": []
      },
      "probes": []
    }
  ]
}
```

**Rules**

- CPU, memory, disk, and freshness are included when available from the latest public-safe system report.
- Missing values are explicit, for example by omission plus an `unavailable` list or equivalent UI state.
- Anonymous responses remain sanitized and include only public-enabled systems.
- Public dashboard clients may poll this endpoint to refresh metrics and freshness.

## Compatibility

- All changes should be additive where possible.
- Existing public dashboard and probe clients that ignore new fields must continue to work.
- Internal agent websocket contracts remain documented separately in `agent-probe.md`.
