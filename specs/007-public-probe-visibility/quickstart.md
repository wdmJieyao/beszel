# Quickstart: Public Probe Visibility and Refresh Commands

## Prerequisites

- A local Beszel hub build with the feature implemented.
- At least two test systems, with at least one enabled for the public dashboard.
- A set of configured latency probes covering at least one public VPS.
- An administrator account that can edit public dashboard settings.
- Docker available on a test host for validating generated commands.

## Quality Gates

Run backend tests:

```bash
go test -tags=testing ./...
```

Run Go linter:

```bash
golangci-lint run
```

Run frontend tests and checks:

```bash
npm --prefix ./internal/site run test:unit
npm --prefix ./internal/site run check
npm --prefix ./internal/site run build
```

## Validation Scenario 1: Migration Preserves Existing Public Output

1. Start from data where one or more public VPS instances and legacy-public
   probes already exist.
2. Upgrade the hub with the feature enabled.
3. Confirm each existing public VPS has a `publicProbeIds` selection seeded in
   `public_system_visibility`.
4. Open Settings -> 公共看板.
5. Inspect the selected probe controls for each already public VPS.
6. Open `/` anonymously.

Expected outcome:

- Each public VPS keeps only the probe lines it showed before upgrade.
- No additional public VPS/probe combinations appear.
- Non-public systems still do not appear.
- Re-running the migration leaves existing `publicProbeIds` unchanged.

## Validation Scenario 2: Newly Public VPS Starts Empty

1. Choose a system that is not currently public.
2. Enable it in Settings -> 公共看板.
3. Inspect its public probe visibility controls.
4. Open `/` anonymously.
5. Use the select-all action.
6. Reload `/` anonymously.

Expected outcome:

- The newly public VPS starts with no selected probe lines.
- `GET /api/beszel/public/systems` returns `publicProbeIds: []` for the newly
  public VPS until an administrator explicitly selects probes.
- The public dashboard initially shows the VPS without public probe series.
- One action selects all available probe lines for that VPS.
- After selecting all, the public dashboard shows the newly selected probe
  summaries only for that VPS.

## Validation Scenario 3: Unselected Probes Stay Hidden

1. Make two public VPS instances visible.
2. Select one probe for the first VPS and a different probe for the second VPS.
3. Open `/` anonymously.

Expected outcome:

- Each VPS shows only the probe lines explicitly selected for it.
- Unselected probe names, targets, latest result metadata, and series are not
  present in the anonymous response or UI for the other VPS.

## Validation Scenario 4: Docker Run Refresh Command

1. Open the UI surface that exposes Copy docker run.
2. Copy the generated command.
3. On a Docker host, run the command once.
4. Publish or point at a newer image tag in a test environment.
5. Run the same generated command again.

Expected outcome:

- The command removes any existing container with the same name.
- Missing container or image cleanup does not stop execution.
- The command pulls the image before starting.
- The new container starts with the expected runtime configuration.

## Release Verification

If this feature is pushed to `main`, wait for the `Make docker images` workflow
to complete successfully and verify the expected GHCR tags:

- `ghcr.io/wdmjieyao/beszel:edge`
- `ghcr.io/wdmjieyao/beszel-agent:edge`
- any additional generated Docker run image variants touched by the change
