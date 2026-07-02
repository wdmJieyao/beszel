# Research: Probe Chart and Public Dashboard Fixes

## Decision: Render latency-capable probe results as grouped multi-series charts

**Rationale**: The user expectation is Komari-like comparison: configured
TCPing lines for a node are compared in one chart. A card per TCPing item makes
comparison harder and caused the node detail page to become fragile.

**Alternatives considered**:

- One chart/card per probe: rejected because it does not match the requested
  product model and scales poorly.
- One table only: rejected because latency trend visualization is core to the
  feature.

## Decision: Keep existing probe storage and transform results for charting

**Rationale**: Existing probe definitions, assignments, and results already
store probe ID, executing system, timestamp, success, latency, and error. The
fix should reshape those records into chart series rather than introduce a new
time-series model.

**Alternatives considered**:

- New aggregation collection: rejected as unnecessary for the current bug fix.
- Client-only direct collection reads: rejected because public and authenticated
  read rules need sanitized, consistent shaping.

## Decision: Treat runtime TCPing failures as result data, not request failures

**Rationale**: Probe execution failures are expected monitoring data. A timeout
or connection refusal should create a failed point with a failure category
instead of breaking the page or the API response.

**Alternatives considered**:

- Return HTTP errors for unreachable targets: rejected because one bad target
  would block other results.
- Hide failed points: rejected because failures are the signal users need.

## Decision: Improve target validation before execution

**Rationale**: Many TCPing failures are configuration issues. Clear validation
for `host:port` before save prevents avoidable runtime failures and makes
remaining runtime failures more meaningful.

**Alternatives considered**:

- Allow any string and rely on agent errors: rejected because users see generic
  detection failures.
- Restrict to IP only: rejected because hostnames are normal TCPing targets.

## Decision: Public metrics must use the latest available system report

**Rationale**: The public dashboard should show CPU, memory, disk, and
freshness for reporting public nodes. Blank values indicate the read model is
not reading or mapping the existing latest system data correctly.

**Alternatives considered**:

- Hide missing metric rows entirely: rejected because it looks like the page is
  incomplete and does not distinguish unavailable data.
- Expose full private system records publicly: rejected for privacy and
  security.

## Decision: Refresh public dashboard data with lightweight polling

**Rationale**: Anonymous visitors do not have authenticated realtime
subscriptions. Periodic public-status refresh gives dynamic freshness and
metric updates without broadening realtime access.

**Alternatives considered**:

- Require manual browser refresh: rejected because the user explicitly expects
  dynamic update behavior.
- Anonymous realtime subscriptions: rejected for scope and security risk.
