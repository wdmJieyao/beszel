# Contract: Network Probe API Coverage Scope

## List Probes

`GET /api/beszel/network-probes`

### Response

```json
{
  "probes": [
    {
      "id": "probe_123",
      "name": "广东电信",
      "type": "tcping",
      "target": "gd-ct-v4.ip.zstaticcdn.com:80",
      "intervalSeconds": 20,
      "timeoutSeconds": 5,
      "enabled": true,
      "publicVisible": true,
      "scope": "global",
      "systems": []
    }
  ]
}
```

### Rules

- `scope = "global"` means the probe applies to all eligible current and future
  systems.
- For global probes, `systems` is the fixed selection list and MUST be empty.
- `scope = "fixed"` means the probe applies only to `systems`.
- Clients must not infer global coverage from a changing all-system list.

## Create Probe

`POST /api/beszel/network-probes`

### Request

```json
{
  "name": "广东电信",
  "type": "tcping",
  "target": "gd-ct-v4.ip.zstaticcdn.com:80",
  "intervalSeconds": 20,
  "timeoutSeconds": 5,
  "enabled": true,
  "publicVisible": true,
  "scope": "global",
  "systems": []
}
```

### Rules

- If `scope` is omitted and `systems` is empty or omitted, the server treats the
  probe as `global`.
- If `scope` is omitted and `systems` contains one or more IDs, the server treats
  the probe as `fixed`.
- If `scope = "global"`, the server ignores any `systems` values for fixed
  coverage and returns `systems: []`.
- If `scope = "fixed"`, `systems` must contain authorized system IDs.
- Invalid system IDs or unauthorized system IDs return a client-visible
  validation error.

## Update Probe

`PATCH /api/beszel/network-probes/{probeId}`

### Request: fixed to global

```json
{
  "scope": "global",
  "systems": []
}
```

### Request: global to fixed

```json
{
  "scope": "fixed",
  "systems": ["system_123"]
}
```

### Rules

- Updating from fixed to global makes current and future eligible systems
  covered.
- Updating from global to fixed limits future coverage to the selected systems.
- Historical result rows are not rewritten by scope changes.

## Get Results

`GET /api/beszel/network-probes/{probeId}/results?system={systemId}&range=30m`

### Rules

- If the requested system is effectively covered by the probe and the viewer is
  authorized, return result points for that probe/system pair.
- If the system is covered but has no results, return an empty `series`.
- If the system is not covered or not visible to the viewer, return the same
  safe not-found/empty behavior used by existing authorization rules.
