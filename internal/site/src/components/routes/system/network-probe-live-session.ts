export type LiveProbeRange = "1m" | "30m" | "1h" | "12h" | "24h" | "1w" | "30d"

export type LiveProbeResultPoint = {
	systemId: string
	created: string
	success: boolean
	latencyMs?: number
	packetLossPercent?: number
	httpStatus?: number
	error?: string
	failureCategory?:
		| "invalid_target"
		| "dns_failure"
		| "timeout"
		| "connection_refused"
		| "target_unreachable"
		| "execution_node_unavailable"
		| "unsupported"
		| "unknown_failure"
	type?: "tcping" | "icmp_ping" | "http_get"
	target?: string
}

export type ProbeResultsByProbeId = Record<string, { series: LiveProbeResultPoint[] }>

export type NetworkProbeRealtimeEvent = {
	action: string
	record: Record<string, unknown>
}

export function initialLiveProbeResults(): ProbeResultsByProbeId {
	return {}
}

export function shouldUseLiveProbeSession(range: LiveProbeRange) {
	return range === "1m"
}

export function applyLiveProbeResultEvent(
	current: ProbeResultsByProbeId,
	event: NetworkProbeRealtimeEvent
): ProbeResultsByProbeId {
	const probeId = normalizeOptionalString(event.record.probe)
	const systemId = normalizeOptionalString(event.record.system)
	const createdValue = normalizeOptionalString(event.record.created)
	if (!probeId || !systemId || !createdValue) {
		return current
	}

	if (event.action === "delete") {
		if (!current[probeId]) {
			return current
		}
		const next = { ...current }
		delete next[probeId]
		return next
	}

	const createdTime = new Date(createdValue).getTime()
	if (!Number.isFinite(createdTime)) {
		return current
	}

	const nextPoint: LiveProbeResultPoint = {
		systemId,
		created: new Date(createdTime).toISOString(),
		success: event.record.success === true,
		latencyMs: normalizeOptionalNumber(event.record.latency_ms),
		packetLossPercent: normalizeOptionalNumber(event.record.packet_loss_percent),
		httpStatus: normalizeOptionalNumber(event.record.http_status),
		error: normalizeOptionalString(event.record.error),
		failureCategory: normalizeFailureCategory(event.record.failure_category),
		type: normalizeProbeType(event.record.type),
		target: normalizeOptionalString(event.record.target),
	}
	const existing = current[probeId]?.series ?? []
	const nextSeries = upsertProbePoint(existing, nextPoint)
	if (nextSeries === existing) {
		return current
	}
	return {
		...current,
		[probeId]: { series: nextSeries },
	}
}

function upsertProbePoint(existing: LiveProbeResultPoint[], nextPoint: LiveProbeResultPoint) {
	const index = existing.findIndex((point) => point.created === nextPoint.created)
	if (index === -1) {
		return [...existing, nextPoint].sort(compareByCreated)
	}
	const prevPoint = existing[index]
	if (
		prevPoint.success === nextPoint.success &&
		prevPoint.latencyMs === nextPoint.latencyMs &&
		prevPoint.packetLossPercent === nextPoint.packetLossPercent &&
		prevPoint.httpStatus === nextPoint.httpStatus &&
		prevPoint.error === nextPoint.error &&
		prevPoint.failureCategory === nextPoint.failureCategory &&
		prevPoint.type === nextPoint.type &&
		prevPoint.target === nextPoint.target
	) {
		return existing
	}
	const next = existing.slice()
	next[index] = nextPoint
	return next.sort(compareByCreated)
}

function compareByCreated(a: LiveProbeResultPoint, b: LiveProbeResultPoint) {
	return new Date(a.created).getTime() - new Date(b.created).getTime()
}

function normalizeOptionalNumber(value: unknown) {
	return typeof value === "number" && Number.isFinite(value) ? value : undefined
}

function normalizeOptionalString(value: unknown) {
	return typeof value === "string" && value.trim() ? value : undefined
}

function normalizeFailureCategory(value: unknown): LiveProbeResultPoint["failureCategory"] {
	const category = normalizeOptionalString(value)
	switch (category) {
		case "invalid_target":
		case "dns_failure":
		case "timeout":
		case "connection_refused":
		case "target_unreachable":
		case "execution_node_unavailable":
		case "unsupported":
		case "unknown_failure":
			return category
		default:
			return undefined
	}
}

function normalizeProbeType(value: unknown): LiveProbeResultPoint["type"] {
	const type = normalizeOptionalString(value)
	switch (type) {
		case "tcping":
		case "icmp_ping":
		case "http_get":
			return type
		default:
			return undefined
	}
}
