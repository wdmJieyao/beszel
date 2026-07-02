# Quickstart: Probe Chart and Public Dashboard Fixes

## Prerequisites

- A development hub with the public dashboard/probe feature enabled.
- At least one connected local test node.
- Frontend dependencies installed for `internal/site`.
- Docker available for final validation when using the compose setup.

## Scenario 1: Node Detail Page with TCPing

1. Sign in as an administrator.
2. Create at least two TCPing checks with valid targets such as `example.com:443` and `cloudflare.com:443`.
3. Ensure the local test node is included as an execution node.
4. Wait for at least two check intervals.
5. Open the local node detail page.

**Expected**

- The node detail page opens normally.
- TCPing results appear in one combined latency chart.
- Each configured TCPing line is distinguishable by label.
- Missing or pending results do not break the page.

## Scenario 2: TCPing Failure Diagnostics

1. Try to create a TCPing check with an invalid target such as `example.com`.
2. Create a valid but unreachable TCPing target such as a closed test port.
3. Wait for checks to run.

**Expected**

- Invalid target save is blocked or returns a clear `host:port` validation message.
- Unreachable target records failed results.
- Failure display distinguishes validation, timeout, connection refused, DNS, and execution-node unavailable when available.
- Other successful TCPing lines remain visible.

## Scenario 3: Public Dashboard Metrics

1. Add or use a connected local test node.
2. Enable it in Settings -> 公共看板.
3. Open `/` in an anonymous browser session.
4. Wait for a newer node report or refresh interval.

**Expected**

- CPU, memory, disk, and last-report freshness are visible when available.
- Missing values are shown as unavailable rather than blank.
- Freshness updates after newer reports without logging in.
- Private host, port, token, and user information remain hidden.

## Verification Commands

Run focused backend tests:

```bash
/usr/local/go/bin/go test -tags=testing ./internal/hub ./agent ./internal/common ./internal/hub/ws
```

Run full backend tests:

```bash
/usr/local/go/bin/go test -tags=testing ./...
```

Run Go lint:

```bash
golangci-lint run
```

Run frontend checks and build:

```bash
npm --prefix ./internal/site run check
npm --prefix ./internal/site run build
```

Rebuild Docker validation service:

```bash
docker compose up -d --build
```

Then validate:

```text
http://127.0.0.1:8090
```

## Contract References

- REST resources: [contracts/rest-api.md](./contracts/rest-api.md)
- Agent probe result contract: [contracts/agent-probe.md](./contracts/agent-probe.md)
- Data model: [data-model.md](./data-model.md)
