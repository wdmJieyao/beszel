# Quickstart: Public Status Page and Network Probe Trends

## Prerequisites

- Go toolchain available for hub/agent tests.
- `golangci-lint` available for Go lint verification.
- Bun or npm available for `internal/site`.
- A development hub and at least one connected agent.

## Setup

1. Install frontend dependencies if needed:

   ```bash
   bun install --cwd ./internal/site
   ```

   If Bun is unavailable, use npm with the existing package files.

2. Run the development hub and agent as usual:

   ```bash
   make dev-hub
   make dev-agent
   ```

3. Open the web UI and sign in as an administrator.

## Scenario 1: Anonymous Public Status Page

1. Ensure at least two systems exist.
2. Enable public visibility for one system.
3. Leave another system private.
4. Open the home route `/` in a signed-out browser session.

**Expected**

- The public system appears on the home page.
- The private system does not appear.
- Visible metrics are limited to display name, online state, data freshness,
  CPU percentage, memory percentage, disk percentage, and public-visible probe
  charts.
- No private host, port, internal address, token, user ID, or admin action is
  visible.

## Scenario 2: Public Visibility Changes

1. While signed in as admin, disable public visibility for the public system.
2. Refresh the anonymous home page.

**Expected**

- The system disappears from the anonymous page.
- If no systems are public, the page shows a safe empty state.

## Scenario 3: Configure Network Probes

1. Create three probes:
   - TCPing target such as `example.com:443`
   - ICMP Ping target such as `1.1.1.1`
   - HTTP GET target such as `https://example.com`
2. Leave the default execution mode on automatic.
3. Add execution nodes only if advanced mode is required.
4. Keep public visibility enabled for the executing system.
5. Leave probe public visibility enabled.
6. Wait for multiple check intervals.

**Expected**

- Authenticated views show current reachability for all probe types.
- TCPing and ICMP Ping show latency trend data.
- HTTP GET shows service reachability and response status.
- Results are attributed to the executing system.

## Scenario 4: Public Probe Visibility

1. Confirm probe charts appear on the anonymous home page for a public system.
2. Disable public display for one probe.
3. Refresh the anonymous home page.

**Expected**

- The disabled probe no longer appears publicly.
- Other public-visible probes remain visible.
- Authenticated views can still show private details according to user access.

## Scenario 5: Failure Handling

1. Configure one unreachable TCPing or HTTP target.
2. Keep another reachable probe assigned to the same agent.
3. Wait for checks to run.

**Expected**

- The unreachable target records a failed result.
- The reachable probe continues to record successful results.
- The agent remains connected and normal system metrics continue.

## Verification Commands

Run targeted backend and agent tests:

```bash
/usr/local/go/bin/go test -tags=testing ./internal/hub ./internal/common ./internal/hub/ws ./internal/records ./internal/migrations ./agent -run 'TestRunNetworkProbe|TestHandlerRegistry|TestValidateNetworkProbeConfig|TestSanitizePublicSystem|TestValidatePublicVisibility|TestPublicStatus|TestNetworkProbe|TestApiRoutesAuthentication'
```

Run full backend tests:

```bash
/usr/local/go/bin/go test -tags=testing ./...
```

Run Go lint:

```bash
golangci-lint run
```

Run frontend quality checks:

```bash
npm --prefix ./internal/site run check
```

If Bun is available, the equivalent command is `bun run --cwd ./internal/site
check`.

Frontend unit test note: the current site package does not include a dedicated
unit-test runner. Frontend behavior for this slice is validated through typed
Vite production build, Biome, and the manual route scenarios above.

## Contract References

- REST resources: [contracts/rest-api.md](./contracts/rest-api.md)
- Agent probe action: [contracts/agent-probe.md](./contracts/agent-probe.md)
- Data model: [data-model.md](./data-model.md)
