# Contract: Public Status Range Query

## Resource

`GET /api/beszel/public/status`

Returns sanitized public dashboard data for systems marked as publicly visible.
The route remains anonymous and read-only.

## Query Parameters

| Name | Required | Values | Default | Description |
|------|----------|--------|---------|-------------|
| `range` | No | `30m`, `1m`, `1h`, `12h`, `24h`, `1w`, `30d` | `30m` for public chart data | Selects the visible chart history range returned for public latency and resource series. |

## Success Response

Status: `200 OK`

```json
{
  "generatedAt": "2026-06-30T12:20:00Z",
  "systems": [
    {
      "id": "system_id",
      "name": "public system name",
      "status": "up",
      "freshness": "2026-06-30T12:19:50Z",
      "updated": "2026-06-30T12:19:50Z",
      "metrics": {
        "cpuPercent": 12.3,
        "memoryPercent": 45.6,
        "diskPercent": 67.8,
        "unavailable": []
      },
      "history": [
        {
          "created": "2026-06-30T11:50:00Z",
          "cpuPercent": 10.1,
          "memoryPercent": 45.1,
          "diskPercent": 67.5
        }
      ],
      "probes": [
        {
          "id": "probe_id",
          "name": "public line name",
          "type": "tcping",
          "latest": {
            "success": true,
            "latencyMs": 14.2,
            "created": "2026-06-30T12:19:40Z"
          },
          "series": [
            {
              "created": "2026-06-30T11:50:00Z",
              "success": true,
              "latencyMs": 15.2
            }
          ]
        }
      ]
    }
  ]
}
```

## Range Semantics

- If `range` is omitted, public chart data uses `30m`.
- Returned `history` points are ordered ascending by `created`.
- Returned latency `series` points are ordered ascending by `created`.
- Returned chart points SHOULD be limited to the requested range from the time
  of the request, while `latest` may still represent the latest sanitized probe
  status.
- For `30m`, public resource history uses minute-level records.
- For longer ranges, the server may use existing coarser stored record types
  that match the selected range.

## Error Responses

### Invalid Range

Status: `400 Bad Request`

```json
{
  "message": "Invalid public chart range.",
  "data": {
    "range": "unsupported"
  }
}
```

Implementation may choose to fall back to `30m` only if the response remains
predictable and tests document that behavior. The preferred contract is a clear
client error for unsupported explicit range values.

## Compatibility

- Existing clients that call `/api/beszel/public/status` without `range` remain
  compatible and receive data suitable for the default public chart view.
- Response fields are additive-compatible with the existing public dashboard
  schema.
- Public probe target hostnames, IPs, ports, and raw target labels remain
  excluded from this response.

## Security and Privacy

- No authentication is required.
- Only systems with public visibility enabled are returned.
- Only metrics allowed by each system's public visibility settings are returned.
- Probe target metadata remains hidden even when a longer chart range is
  selected.
