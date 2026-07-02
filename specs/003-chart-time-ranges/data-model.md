# Data Model: Public Chart Time Ranges

## Chart Time Range

Represents the selected duration for a public chart.

**Fields**

- `value`: Stable range key. Allowed values are `30m`, `1m`, `1h`, `12h`,
  `24h`, `1w`, and `30d` where applicable.
- `label`: User-facing range label.
- `duration`: Time window counted back from the current time.
- `statsType`: Existing stats bucket used for resource history when requesting
  backend data.
- `expectedInterval`: Expected spacing between points for gap detection and
  validation.
- `tickCount`: Preferred maximum number of horizontal axis labels.
- `format`: Time label formatter for the x-axis.
- `isDefault`: `true` only for `30m` in the new public charts.

**Validation Rules**

- Unknown range values are rejected or fall back to `30m` only where the public
  contract explicitly allows fallback.
- `30m` uses minute-level public history.
- Long ranges must reduce label density rather than wrap labels.

## Public Dashboard Chart State

Represents transient chart UI state for one public node card.

**Fields**

- `systemId`: Public system identifier.
- `selectedRange`: Current `Chart Time Range`, default `30m`.
- `lastRefreshAt`: Last successful public chart data refresh time.
- `refreshError`: Whether the last refresh failed while retaining existing data.

**Relationships**

- One public node card owns one selected range for its latency chart area.
- The resource trend dialog for the same node uses the node's selected or
  dialog-selected public range consistently for CPU, memory, and disk charts.

**State Transitions**

- Initial load -> `selectedRange = 30m`.
- User selects a range -> chart data and axes redraw for that range.
- Refresh succeeds -> points update while selected range is preserved.
- Refresh fails -> prior points and selected range remain visible.

## Public Latency Series

Sanitized chart data for one public probe line on one public node.

**Fields**

- `probeId`: Public-safe probe identifier.
- `systemId`: Public system identifier.
- `label`: Public-safe display name for the line.
- `type`: Probe type, limited to public latency-capable probes.
- `points`: Ordered timestamped latency results inside the selected range.
- `latest`: Latest sanitized result used for status display.

**Validation Rules**

- Points must be ordered ascending by timestamp for chart rendering.
- Points outside the selected range are excluded from returned public chart
  data, except where the latest sanitized status is separately required.
- Hidden target hostnames, addresses, and ports are never included.
- Failed points preserve user-safe failure categories and messages only.

## Public Resource Series

Timestamped CPU, memory, and disk usage values for one public node.

**Fields**

- `systemId`: Public system identifier.
- `created`: Timestamp for the measurement.
- `cpuPercent`: Optional CPU usage percentage if public visibility allows CPU.
- `memoryPercent`: Optional memory usage percentage if public visibility allows memory.
- `diskPercent`: Optional disk usage percentage if public visibility allows disk.

**Validation Rules**

- Values are included only for metrics enabled in public visibility settings.
- Points outside the selected range are excluded from returned public history.
- The current public summary value is merged into the visible series when newer
  than stored history so the chart endpoint matches the top-line public metrics.

## Chart Axis Label

Generated visible x-axis label for a public chart.

**Fields**

- `timestamp`: Millisecond timestamp represented by the label.
- `text`: Hour-minute-second formatted display text.
- `range`: Range for which the label was generated.

**Validation Rules**

- Labels must not overlap or wrap in validated desktop/mobile widths.
- Label count is reduced for narrow widths or longer selected ranges.
- Labels are generated from the selected visible domain, not from hidden
  out-of-range points.
