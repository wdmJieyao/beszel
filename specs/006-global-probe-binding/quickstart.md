# Quickstart: Global Probe Binding Regression Fix

## Prerequisites

- A local Beszel hub build with the feature implemented.
- At least two test agents or test systems that can be added sequentially.
- A logged-in administrator account.

## Quality Gates

Run backend tests:

```bash
go test -tags=testing ./...
```

Run Go linter when available:

```bash
golangci-lint run
```

Run frontend tests and checks:

```bash
npm --prefix ./internal/site run test:unit
npm --prefix ./internal/site run check
npm --prefix ./internal/site run build
```

## Validation Scenario 1: Global Probe Covers Future Machine

1. Start with one machine in the hub.
2. Open Settings -> 线路检测.
3. Create a TCPing probe and leave execution node selection as 全部可用节点.
4. Confirm the probe settings show it as all-machine/global.
5. Add a second machine after the probe already exists.
6. Open the second machine detail page.

Expected outcome:

- The second machine shows the global probe in 线路检测 without editing the
  probe.
- Before first result, the chart shows a pending/no-history state.
- After checks run, the second machine records its own result points.

## Validation Scenario 2: Fixed Probe Does Not Expand

1. Create a second probe and select only the first machine as its execution node.
2. Add or use a second machine.
3. Open the second machine detail page.

Expected outcome:

- The fixed-machine probe does not appear for the second machine.
- The global probe still appears for the second machine.

## Validation Scenario 3: Scope Change

1. Edit the fixed-machine probe.
2. Change it to 全部可用节点.
3. Save.
4. Open each machine detail page.

Expected outcome:

- The probe appears for all current machines.
- A later newly added machine also receives the probe automatically.

## Validation Scenario 4: Public Dashboard Safety

1. Enable public display for one machine and keep another private.
2. Keep one global probe public-visible.
3. Open `/` anonymously.

Expected outcome:

- The public machine can show public-safe global probe summaries.
- The private machine and its probe results are not shown.
- Hidden probes remain hidden even if they are global.

## Release Verification

If this feature is pushed to `main`, wait for the `Make docker images` GitHub
Actions workflow to finish successfully and verify the expected GHCR `edge`
image tags before reporting the push as complete.
