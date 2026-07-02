# Research: Public Chart Time Ranges

## Decision: Add `30m` as a Public Chart Range

**Decision**: Introduce a 30-minute chart range for public latency and resource
charts and make it the default selected range for the newly added public charts.

**Rationale**: The existing node detail chart ranges are `1m`, `1h`, `12h`,
`24h`, `1w`, and `30d`. The user specifically wants public charts to focus on
the latest 30 minutes by default while keeping the existing range style. Adding
`30m` avoids overloading `1h` semantics and gives public charts the requested
default without changing the existing node-detail default behavior.

**Alternatives considered**:

- Reuse `1h` and filter client-side to 30 minutes: rejected because selector
  semantics would not match the visible chart.
- Replace existing ranges with a smaller public-only list: rejected because the
  user chose to reuse the existing range style.
- Use only a fixed 30-minute view: rejected because the spec requires user
  range selection.

## Decision: Keep Public Status as the Range-Aware Resource

**Decision**: Continue using the public status resource and add/define the
existing optional `range` query behavior for public chart data.

**Rationale**: Public dashboard data is already composed and sanitized in the
public status read model. Keeping range-aware history in that resource preserves
anonymous access, avoids new route duplication, and centralizes sanitization for
probe names, target addresses, and current system metrics.

**Alternatives considered**:

- Add separate public latency and resource endpoints: rejected for this slice
  because it would duplicate visibility and sanitization logic.
- Fetch all data once and filter in the browser: rejected because longer ranges
  would increase payload size and expose unnecessary data to public clients.

## Decision: Use Server-Side Range Filtering with Client-Side Display Filtering

**Decision**: The hub should return only enough timestamped data for the
requested range, while the frontend also filters plotted points to the selected
range before rendering.

**Rationale**: Server-side filtering limits public payload size and makes
selected range behavior testable at the API contract level. Client-side
filtering handles latest summary values, refresh windows, and sparse data safely
without relying on exact database record intervals.

**Alternatives considered**:

- Server-only filtering: rejected because latest summary values may need to be
  merged into visible resource series.
- Client-only filtering: rejected because public payloads for 7-day and 30-day
  ranges would be unnecessarily large.

## Decision: Reuse Existing Chart-Time Formatting Patterns

**Decision**: Extend the existing chart time metadata and axis tick generation
approach for public charts, with hour-minute-second labels for the new public
chart axes.

**Rationale**: Existing system charts already centralize range offsets, tick
counts, and label formatting. Reusing the pattern keeps range behavior
consistent and avoids a parallel chart-time model. Public chart x-axes need
visible time labels, unlike the current public latency chart's hidden x-axis.

**Alternatives considered**:

- Hard-code public chart ticks in each component: rejected because it would
  fragment range behavior and increase maintenance.
- Use date-only labels for long ranges: rejected for the requested public
  behavior because the spec asks for hour-minute-second labels on these newly
  added charts. For very long selected ranges, density is reduced rather than
  changing the label format.

## Decision: Refresh Public Chart Data Every 20 Seconds

**Decision**: Public chart data refreshes on a 20-second cadence while
preserving the selected range.

**Rationale**: The current public page has recently used shorter refreshes for
summary responsiveness, but this feature explicitly requires 20-second refresh
for the new chart data. A single public status reload can refresh both current
summary and selected range data without forcing a page reload.

**Alternatives considered**:

- Keep 6-second refresh for all public data: rejected because it conflicts with
  the feature requirement for chart refresh cadence.
- Refresh only after range changes: rejected because public trend charts would
  become stale during long viewing sessions.

## Decision: Keep Range State Per Public Node

**Decision**: Store selected chart range independently per public node card for
the current browser session.

**Rationale**: The spec requires changing the selected range for one node not to
unexpectedly change another node's selected range. Per-node state is enough for
the public dashboard and avoids adding persistence requirements.

**Alternatives considered**:

- One global public chart range: rejected because it violates the clarified
  independent-node behavior.
- Persist public range choices across browser sessions: deferred because the
  spec does not require persistence and existing anonymous public pages should
  remain lightweight.
