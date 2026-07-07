# Quickstart: Validate Telegram Notifications and YML Configuration Backup

This guide defines end-to-end validation scenarios for implementation. It is not
a task list and does not include implementation code.

## Prerequisites

- A local Beszel hub test environment.
- A Telegram bot token for manual live testing, or a fake Telegram transport for
  automated tests.
- At least two systems, one alert definition, public dashboard settings, and
  several network probes configured in the test environment.
- Administrator login for backup export/import and Telegram settings.

## Rollout scope

This feature is a panel-side update. Telegram notification delivery, Telegram
menu polling, and configuration backup/restore all run in the hub and use
existing hub collections. Existing `beszel-agent` deployments can keep running
unchanged; no agent image, agent command, agent websocket, or metric reporting
change is required for this feature.

Before deploying a build from this feature, verify that:

- `/api/beszel/agent-connect` is still registered as the unauthenticated agent
  websocket route.
- Generated agent install commands still use the current configured agent image
  policy unless a later feature explicitly changes agent code.
- Existing agents continue to report metrics after the hub is updated.

## Automated verification commands

Run these before marking implementation complete:

```sh
gofmt -w <touched-go-files>
go test -tags=testing ./...
golangci-lint run --build-tags testing
npm --prefix ./internal/site run test:unit
npm --prefix ./internal/site run check
npm --prefix ./internal/site run build
```

If `golangci-lint` is not installed locally, this repository can be checked with
the same project Go toolchain without adding a dependency:

```sh
GOTOOLCHAIN=go1.26.3 go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@latest run --build-tags testing
```

If existing unrelated repository failures block full commands, capture the exact
failing package/file and run the narrow feature tests that prove this feature's
behavior.

## Scenario 1: Telegram admin destination

1. Configure Telegram bot settings with a valid token.
2. Add an `admin` destination using a private chat ID.
3. Send a test Telegram message.
4. Trigger a representative alert.
5. Open the bot menu from the admin chat and request status overview, alert
   summary, node list, and a single node summary.

Expected:

- Test message is delivered.
- Alert message includes node, alert type, severity/state, timestamp, and a
  safe panel link where appropriate.
- Admin menu returns current panel state.
- Token is never shown in UI responses, API responses, logs, or toast messages.

## Scenario 2: Telegram read-only destination

1. Add a `read_only` destination for a group or channel.
2. Scope the destination to one system and selected alert levels.
3. Trigger one matching alert and one out-of-scope alert.
4. Attempt a privileged bot menu command from the read-only chat.

Expected:

- Matching alert produces a non-sensitive summary.
- Out-of-scope alert is not delivered.
- Privileged menu action returns no monitoring data.
- Message content does not include internal hosts, tokens, probe targets,
  webhook URLs, raw metrics, or backup details.

## Scenario 3: Encrypted YML export

1. Configure systems, alerts, quiet hours, email/webhook notifications,
   Telegram settings/destinations, public dashboard visibility, and network
   probes.
2. Export a full backup with secrets included and an encryption credential.
3. Inspect the YML.

Expected:

- YML includes metadata, version, and all supported configuration sections.
- Sensitive values are represented as encrypted envelopes.
- No plaintext bot token, webhook URL credential, or agent token appears in the
  file.

## Scenario 4: Merge restore preview

1. Import the encrypted backup into a test instance with some matching stable
   IDs and some target-only records.
2. Provide the correct decryption credential.
3. Run validation/preview without applying.

Expected:

- Preview shows create, update, preserve, skip, conflict, and error counts.
- Target-only records are marked `preserve`.
- Renamed records with matching stable IDs are marked `update`, not `create`.
- Missing references or invalid sections block apply with clear conflicts.

## Scenario 5: Merge restore apply

1. Apply a previously reviewed preview.
2. Re-open systems, notifications, public dashboard settings, and network probe
   settings.
3. Export again and compare supported configuration semantics.

Expected:

- Missing configuration was created.
- Matched configuration was updated.
- Target-only configuration still exists.
- No runtime histories were imported.
- Existing agents continue reporting without any agent update.

## Scenario 6: Wrong decryption credential

1. Import an encrypted backup with an incorrect credential.
2. Attempt preview and apply.

Expected:

- Preview reports decryption failure for encrypted sections.
- Apply is blocked.
- Existing configuration remains unchanged.

## Scenario 7: Legacy config.yml compatibility

1. Use the existing `GET /api/beszel/config-yaml` or settings page export.
2. Confirm it still produces the legacy system-only YAML.
3. Confirm the new backup export is available separately and documents merge
   restore semantics.

Expected:

- Existing system-only behavior is preserved.
- Operators can distinguish legacy `config.yml` sync from the new full
  configuration backup flow.

## Push/GHCR verification

If the completed implementation is pushed to GitHub `main` or affects Docker
image inputs, wait for the `Make docker images` workflow to finish and verify
the expected GHCR tags, including `ghcr.io/wdmjieyao/beszel:edge`, before
reporting the push as complete.
