# Agent Contract: Probe Failure Classification

## Transport

Continue using the existing hub-to-agent websocket request/response protocol
and `RunNetworkProbe` action. This is an existing non-REST internal transport.

## Response Payload Additions

The agent or hub may add a normalized failure category to probe results.

```json
{
  "probeId": "probe_id",
  "type": "tcping",
  "target": "example.com:443",
  "success": false,
  "latencyMs": null,
  "packetLossPercent": null,
  "httpStatus": null,
  "error": "connection refused",
  "failureCategory": "connection_refused",
  "checkedAt": "2026-06-29T12:00:00Z"
}
```

## Failure Categories

- `invalid_target`: target format is invalid before execution.
- `dns_failure`: target host cannot be resolved.
- `timeout`: probe timed out.
- `connection_refused`: TCP connection was refused.
- `target_unreachable`: network path or target is unreachable.
- `execution_node_unavailable`: hub cannot execute because the node is offline or has no compatible probe transport.
- `unsupported`: execution node does not support the requested probe type.
- `unknown_failure`: fallback for safe but unclassified errors.

## Rules

- Agent-side runtime failures must return result objects instead of crashing.
- Hub-side offline or unsupported execution failures may be classified by the hub.
- Error strings must be safe for display to authenticated users and must not include credentials or local paths.
- Failed results must be persisted so charts can show outages and gaps.
