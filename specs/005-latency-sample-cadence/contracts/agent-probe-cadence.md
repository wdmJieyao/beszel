# Contract: High-Cadence Agent Probe Execution

## Transport

Use the existing hub-to-agent websocket request/response action:

- `RunNetworkProbe`

No new agent action is required for this feature.

## Live Cadence Semantics

When the hub has at least one active live latency session for a system:

1. The hub selects enabled latency-capable probe assignments for that system.
2. The hub sends `RunNetworkProbe` requests to the system's connected agent at approximately 1-second cadence.
3. The hub persists each returned result to `network_probe_results`.
4. PocketBase realtime delivers created result records to subscribed browsers.

The current implementation coalesces viewers per system and runs eligible assignments concurrently while preventing overlap for the same assignment.

Live cadence is only entered from the authenticated node detail `1 分钟` range. Selecting `30 分钟` or any longer range must not change the agent execution cadence; those ranges read persisted historical results.

## Request Payload

Same as the existing network probe request:

```json
{
  "probeId": "probe_id",
  "type": "tcping",
  "target": "example.com:443",
  "timeoutSeconds": 1
}
```

**Live cadence notes**

- `timeoutSeconds` should be bounded so one failed check does not block the next cadence window.
- If the configured timeout is longer than the live cadence budget, the live request should use a smaller safe timeout for the high-cadence run without permanently mutating probe configuration.
- The current live timeout bound is 1 second for latency probes in the 1-second cadence loop.

## Response Payload

Same as the existing network probe response:

```json
{
  "probeId": "probe_id",
  "type": "tcping",
  "target": "example.com:443",
  "success": true,
  "latencyMs": 8.2,
  "packetLossPercent": 0,
  "httpStatus": null,
  "error": "",
  "failureCategory": "",
  "checkedAt": "2026-07-02T08:00:02Z"
}
```

## Failure Rules

- Timeout, DNS, connection, permission, unsupported, and offline failures must persist as failed result samples.
- Failed samples must not be converted into successful latency points.
- One failed assignment must not block other assignments for the same system.
- The hub should avoid overlapping live executions for the same assignment if the previous execution has not completed.

## Compatibility

- Agents that already support `RunNetworkProbe` can participate without protocol changes.
- Agents that do not support the action remain represented as execution-node unavailable failures, matching existing behavior.
- Normal configured probe intervals continue to run outside live sessions.
- Historical chart ranges continue to use persisted results and do not require any agent-side range awareness.
