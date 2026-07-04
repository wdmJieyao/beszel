# Contract: Effective Probe Coverage Resolution

## Background Scheduler

### Input

- All enabled probes.
- All eligible systems.
- Fixed assignment rows for fixed probes.
- Latest result per probe/system pair.

### Output

Effective probe/system pairs due for execution.

### Rules

- Global enabled probe: include every eligible system.
- Fixed enabled probe: include every enabled assignment system.
- Disabled probe: include no systems.
- A probe/system pair is due when no result exists or the latest result is older
  than the probe interval.
- The same probe/system pair must not be emitted twice in one scheduler cycle.

## Node Detail Live Session

### Input

- Active live session system ID.
- All probes effectively covering that system.

### Output

Enabled latency-capable probe/system pairs for one-minute live execution.

### Rules

- Global TCPing and ICMP probes cover the active system even when there is no
  assignment row for that system.
- Fixed TCPing and ICMP probes cover the active system only when an enabled
  assignment exists.
- HTTP probes are excluded from live latency execution.
- Existing live-session coalescing and overlap prevention remain unchanged.

## Authenticated Node Detail Probe List

### Input

- Authenticated viewer.
- Current system ID.
- Probe list response.

### Output

Probes shown for the current system.

### Rules

- Global probes are shown for any system the viewer may access.
- Fixed probes are shown only for systems in the fixed `systems` list.
- A shown probe with no results displays a pending/no-history state.

## Public Dashboard Probe Summary

### Input

- Public-visible systems.
- Public-visible probes.
- Effective probe coverage.
- Latest and range results.

### Output

Public-safe probe summaries per public system.

### Rules

- Public summaries include global probes for public systems when the probe is
  public-visible.
- Public summaries include fixed probes only for selected public systems.
- Private systems, private-only fields, hidden probes, and unauthorized data are
  never exposed.
