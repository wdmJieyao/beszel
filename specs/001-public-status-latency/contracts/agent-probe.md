# Agent Contract: Network Probe Execution

## Transport

Use the existing hub-to-agent websocket request/response protocol defined in
`internal/common/common-ws.go`.

Add one action:

- `RunNetworkProbe`

The hub sends the request to a connected selected agent node. The agent
executes the probe and returns one result. The request must use the existing
request ID flow so timeouts and cancellation behave like other agent requests.

## Request Payload

```json
{
  "probeId": "probe_id",
  "type": "tcping",
  "target": "example.com:443",
  "timeoutSeconds": 5
}
```

**Fields**

- `probeId`: hub-side probe identifier used for correlating results.
- `type`: `tcping`, `icmp_ping`, or `http_get`.
- `target`: type-specific target string.
- `timeoutSeconds`: maximum time the agent should spend on the probe.

## Response Payload

```json
{
  "probeId": "probe_id",
  "type": "tcping",
  "target": "example.com:443",
  "success": true,
  "latencyMs": 42.3,
  "packetLossPercent": 0,
  "httpStatus": null,
  "error": "",
  "checkedAt": "2026-06-28T11:59:00Z"
}
```

**Fields**

- `probeId`: copied from request.
- `type`: copied from request.
- `target`: copied from request.
- `success`: whether the check succeeded.
- `latencyMs`: set for successful latency-capable checks when available.
- `packetLossPercent`: set for ICMP Ping when available.
- `httpStatus`: set for HTTP GET when a response is received.
- `error`: short failure category/message; no secrets or local paths.
- `checkedAt`: agent-side check completion timestamp.

## Probe Semantics

### TCPing

- Target format: `host:port`.
- Success means a TCP connection can be established before timeout.
- `latencyMs` measures connection establishment latency.

### ICMP Ping

- Target format: host or IP.
- Success means at least one ping response is received before timeout.
- `latencyMs` records measured round-trip latency when available.
- `packetLossPercent` records loss when multiple attempts are used.

### HTTP GET

- Target format: `http://` or `https://` URL.
- Success means the agent receives an HTTP response before timeout. The plan
  allows implementation to classify status code ranges in tasks.
- `httpStatus` records the response status when available.
- `latencyMs` may record request duration if implementation exposes it, but
  public latency trend charts focus on TCPing and ICMP Ping.

## Failure Rules

- Agent must return a result object for timeouts, DNS failures, connection
  failures, permission issues, and unsupported probe types.
- A failed probe must not terminate the agent process or block later requests.
- Hub records failures as probe results so charts can show outages.
- Offline agents are represented hub-side as missing/failed results; other
  agents and probes continue.

## Version Compatibility

- Agents that do not support `RunNetworkProbe` are treated as unavailable for
  assigned probe execution.
- Hub UI should surface unsupported/offline agent state to authenticated users.
