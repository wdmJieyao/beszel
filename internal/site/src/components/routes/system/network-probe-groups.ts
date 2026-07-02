export type ProbeType = "tcping" | "icmp_ping" | "http_get"

export type ProbeSeriesPoint = {
	created: string
	success: boolean
	latencyMs?: number
	packetLossPercent?: number
	httpStatus?: number
	error?: string
	failureCategory?: string
}

export type ProbeResultPoint = ProbeSeriesPoint & {
	systemId: string
	type?: ProbeType
	target?: string
}

export type ProbeDefinition = {
	id: string
	name: string
	type: ProbeType
	target: string
	systems: string[]
}

export type ProbeResultsByProbeId = Record<string, { series: ProbeResultPoint[] }>

export type ProbeChartSeries = {
	id: string
	label: string
	probeId: string
	systemId: string
	type: ProbeType
	targetLabel: string
	points: ProbeSeriesPoint[]
}

export type ProbeChartGroup = {
	id: string
	label: string
	type: ProbeType
	targetLabel: string
	latest?: ProbeSeriesPoint
	series: ProbeChartSeries[]
}

const LATENCY_TYPES = new Set<ProbeType>(["tcping", "icmp_ping"])

export function groupNetworkProbeData(
	probes: ProbeDefinition[],
	results: ProbeResultsByProbeId,
	systemId: string
): ProbeChartGroup[] {
	const groups = new Map<string, ProbeChartGroup>()
	for (const probe of probes) {
		const seriesPoints = results[probe.id]?.series ?? []
		const targetLabel = probe.target || "unknown"
		const latest = latestPoint(seriesPoints)
		const series = createProbeSeries(probe, seriesPoints, systemId, targetLabel)
		const groupId = groupKey(probe, systemId)
		const existing = groups.get(groupId)
		if (existing) {
			existing.series.push(series)
			existing.latest = pickLatest(existing.latest, latest)
			if (!existing.targetLabel && targetLabel) {
				existing.targetLabel = targetLabel
			}
			continue
		}
		groups.set(groupId, {
			id: groupId,
			label: isLatencyProbe(probe) ? "线路检测" : probe.name,
			type: probe.type,
			targetLabel: isLatencyProbe(probe) ? "" : targetLabel,
			latest,
			series: [series],
		})
	}
	return [...groups.values()].map((group) => ({
		...group,
		series: group.series,
	}))
}

function createProbeSeries(
	probe: ProbeDefinition,
	points: ProbeResultPoint[],
	systemId: string,
	targetLabel: string
): ProbeChartSeries {
	const seriesPoints: ProbeSeriesPoint[] = points.map((point) => ({
		created: point.created,
		success: point.success,
		latencyMs: point.latencyMs,
		packetLossPercent: point.packetLossPercent,
		httpStatus: point.httpStatus,
		error: point.error,
		failureCategory: point.failureCategory,
	}))
	return {
		id: `${probe.id}:${systemId}`,
		label: probe.name,
		probeId: probe.id,
		systemId,
		type: probe.type,
		targetLabel,
		points: seriesPoints,
	}
}

function groupKey(probe: ProbeDefinition, systemId: string) {
	if (isLatencyProbe(probe)) {
		return `${systemId}:latency`
	}
	return `${systemId}:${probe.type}:${probe.id}`
}

function isLatencyProbe(probe: ProbeDefinition) {
	return LATENCY_TYPES.has(probe.type)
}

function latestPoint(points: ProbeSeriesPoint[]) {
	return points.length > 0 ? points[points.length - 1] : undefined
}

function pickLatest(current: ProbeSeriesPoint | undefined, candidate: ProbeSeriesPoint | undefined) {
	if (!current) return candidate
	if (!candidate) return current
	return new Date(candidate.created).getTime() >= new Date(current.created).getTime() ? candidate : current
}
