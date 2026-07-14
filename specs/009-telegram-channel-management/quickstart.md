# Quickstart: Validate Telegram Channel Management

## Prerequisites

- Go and frontend dependencies installed for this repository.
- A disposable copy of a data directory containing at least one existing
  Telegram destination for migration validation.
- A Telegram Bot and private test chat for the final live integration checks.
- Do not place Bot Tokens in commands committed to source control or test logs.

## 1. Focused Backend Feedback Loops

### Baseline (2026-07-11)

- `go test -tags=testing ./internal/hub ./internal/alerts ./internal/records`: PASS
- `npm --prefix ./internal/site run test:unit`: PASS, 55/55 tests
- No pre-existing failures were observed in the touched backend or frontend unit suites.

Run focused tests while implementing each slice:

```bash
go test -tags=testing ./internal/hub -run 'Telegram|ConfigBackup' -count=1
```

Expected: response decoding, migration, API, policy matching, de-duplication,
role filtering, deletion, and backup compatibility tests pass.

## 2. Frontend Unit Feedback Loop

```bash
npm --prefix ./internal/site run test:unit
```

Expected: Bot result presentation, all/selected node modes, select-all/clear,
search, role copy, load labels, and request payload helpers pass.

## 3. Production Data Migration Scenario

1. Start the new hub against a copy of an existing data directory.
2. Open every pre-existing Telegram channel.
3. Confirm each has exactly one default policy.
4. Confirm empty old node scope displays as “全部节点（包含未来新增）”.
5. Confirm non-empty old scopes preserve exactly the same nodes and categories.
6. Restart the hub and confirm no duplicate default policies are created.

Expected: existing Chat IDs, roles, enabled/mute state, and scope behavior are
preserved with no destructive schema changes.

## 4. Bot Verification Scenario

1. Test the saved Token with the Token field empty.
2. Test a valid unsaved Token and confirm it is not persisted.
3. Simulate or test command-menu failure separately from credential success.

Expected: credentials and command menu show distinct results. Telegram boolean
success responses are accepted and no Token appears in errors.

## 5. Multiple Policies And De-duplication

1. Create one channel for a Chat ID.
2. Add a status policy for node A.
3. Add a CPU policy for nodes A and B.
4. Trigger a CPU alert on A that is arranged to match overlapping policies.
5. Disable/delete one policy and repeat.

Expected: the channel receives one matching message, never one per policy.
Deleting a policy leaves the channel and sibling policy intact.

## 6. Node Scope And Role UX

1. Select all-node mode, add a new node, and confirm it is automatically covered.
2. Select selected-node mode, use search, select all current, clear all, and
   verify an empty selected scope cannot be saved ambiguously.
3. Compare Admin and Read-only descriptions.
4. Confirm scopes affect both roles.
5. Confirm Read-only messages omit links and addresses and cannot run commands.
6. Confirm privileged commands work only for an authorized private Admin chat.

## 7. Deletion UX

1. Locate the labelled delete action using mouse and keyboard.
2. Cancel confirmation and verify no data changes.
3. Confirm deletion and verify the channel and policies disappear.
4. Confirm Bot settings and other channels remain.

## 8. Load Category Copy

Verify the settings UI labels categories as system load averages over 1, 5, and
15 minutes and explains that values are absolute, not percentages or alert
durations, with logical CPU cores as the reference.

## 9. Backup Compatibility

1. Export notifications and verify section version 2 plus channels and policies.
2. Restore the export into a disposable target and compare all fields.
3. Restore a version 1 backup with inline destination scopes.
4. Restore both documents twice.

Expected: v2 round-trips; v1 creates one default policy; repeated restore is
idempotent; target-only records and redacted secrets remain preserved.

## 10. Full Quality Gates

```bash
gofmt -w <touched-go-files>
go test -tags=testing ./...
go vet -tags=testing ./...
golangci-lint run --build-tags testing
npm --prefix ./internal/site run test:unit
npm --prefix ./internal/site run check
npm --prefix ./internal/site run build
```

Document any pre-existing linter warning or unrelated race separately; do not
mark a new regression as complete.

## 11. Docker And Live Telegram Validation

Build and start the hub with a disposable data copy, verify the settings UI in a
browser, then use a private test chat to check menu authorization, test delivery,
policy routing, and deletion. Remove temporary accounts and test data afterward.

If the change is pushed, completion additionally requires the `Make docker
images` GitHub Actions workflow to succeed and the expected
`ghcr.io/wdmjieyao/beszel:edge` digest to be pullable. A Git push alone is not a
successful release.

## 12. Implementation Validation (2026-07-11)

Focused and full quality gates:

- `go test -tags=testing ./internal/hub -run 'Telegram|ConfigBackup' -count=1`: PASS.
- `go test -tags=testing ./...`: PASS.
- `go vet -tags=testing ./...`: PASS.
- Go 1.26.3 build of `golangci-lint v2.12.2 run --build-tags testing`: PASS, 0 issues.
- `npm --prefix ./internal/site run test:unit`: PASS, 63/63 tests.
- `npm --prefix ./internal/site run check`: PASS with 17 pre-existing warnings and no errors.
- `npm --prefix ./internal/site run build`: PASS. Existing Vite chunk-size and mixed static/dynamic import warnings remain.

Migration and runtime validation:

- Built `beszel-speckit009:test` from the repository Dockerfile and started an isolated Hub on port 8091.
- A SQLite snapshot of the local production-shaped data contained one legacy Telegram destination. Migration created exactly one `selected` default policy; after restart the policy count remained one.
- Browser validation created one channel, added two policies, deleted one policy while retaining the channel, cancelled channel deletion, then confirmed deletion and verified the channel disappeared.
- A temporary Hub on port 8092 used the production-shaped snapshot to test the saved Bot without exposing its Token. Credential verification and command-menu initialization both succeeded.
- Temporary administrators/data existed only in disposable storage. Test containers, volumes, image, snapshots, screenshots, and temporary linter binaries were removed. The existing port 8090 Hub and its data were not modified.

Release validation:

- No Git push was requested or performed in this implementation run, so T059 and GHCR digest verification were not applicable.
