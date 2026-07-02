# Contract: Live Latency Session API

## Scope

Authenticated node detail pages use this REST contract to tell the hub that a user is actively watching `线路检测` in the `1 分钟` range. The hub uses active sessions to temporarily run enabled latency probes at an approximately 1-second cadence.

Anonymous public dashboard pages do not use this contract. Node detail `30 分钟` and longer latency ranges also do not use this contract; they must load historical results instead.

## Create Session

`POST /api/beszel/systems/{systemId}/network-probe-live-sessions`

### Request Body

```json
{
  "range": "1m"
}
```

### Success Response

Status: `201 Created`

```json
{
  "sessionId": "live_session_id",
  "systemId": "system_id",
  "range": "1m",
  "cadenceSeconds": 1,
  "expiresAt": "2026-07-02T08:00:15Z"
}
```

### Behavior

- The caller must be authenticated and allowed to view `systemId`.
- Only `range = "1m"` is accepted.
- Creating a session starts or renews high-cadence execution for the system while at least one session remains active.
- Creating multiple sessions for the same system must not multiply probe execution cadence.
- The current frontend heartbeat interval is 5 seconds and the current session TTL is approximately 15 seconds.
- Selecting `30m`, `1h`, `12h`, `24h`, `1w`, or `30d` must not call this endpoint.

## Renew Session

`PATCH /api/beszel/systems/{systemId}/network-probe-live-sessions/{sessionId}`

### Request Body

```json
{
  "range": "1m"
}
```

### Success Response

Status: `200 OK`

```json
{
  "sessionId": "live_session_id",
  "systemId": "system_id",
  "range": "1m",
  "cadenceSeconds": 1,
  "expiresAt": "2026-07-02T08:00:20Z"
}
```

### Behavior

- Renewal extends `expiresAt` for an existing active session.
- The hub may return `404 Not Found` if the session has expired; the frontend should create a new session if it is still in `1 分钟`.
- Renewal does not reset the browser chart by itself; the chart reset is controlled by the frontend range transition.
- Successful renewals keep the same `sessionId` and extend the current live observation window.

## End Session

`DELETE /api/beszel/systems/{systemId}/network-probe-live-sessions/{sessionId}`

### Success Response

Status: `204 No Content`

### Behavior

- Ends the caller's live session if it exists.
- If other sessions remain active for the same system, high-cadence execution continues.
- The endpoint is best-effort; TTL cleanup still handles closed tabs or network loss.
- The current frontend sends `DELETE` when leaving `1 分钟`; otherwise the session expires naturally after the TTL window.

## Error Responses

### Unauthorized

Status: `401 Unauthorized`

The caller is not signed in.

### Forbidden or Not Found System

Status: `404 Not Found`

The system does not exist or the caller is not allowed to view it.

### Invalid Range

Status: `400 Bad Request`

```json
{
  "message": "Invalid live latency session request.",
  "data": {
    "range": "Only 1m is supported for live latency sessions."
  }
}
```

## Realtime Result Delivery

The API does not return chart samples. Samples continue to arrive through the existing `network_probe_results` realtime collection subscription.

Expected realtime record fields:

```json
{
  "id": "result_id",
  "probe": "probe_id",
  "system": "system_id",
  "created": "2026-07-02T08:00:02Z",
  "success": true,
  "latency_ms": 8.2,
  "packet_loss_percent": null,
  "http_status": null,
  "failure_category": "",
  "error": "",
  "type": "tcping",
  "target": "example.com:443"
}
```

## Compatibility

- Existing `GET /api/beszel/network-probes` and `GET /api/beszel/network-probes/{probeId}/results` contracts remain valid.
- Existing public status endpoints do not create live sessions.
- Existing clients that do not call this contract keep normal configured probe intervals.
- Historical latency range rendering must use existing result-query contracts and must not depend on live-session state.
