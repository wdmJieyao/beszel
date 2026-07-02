import { timeDay, timeHour, timeMinute, timeTicks } from "d3-time"

export type NetworkProbeChartRange = "1m" | "30m" | "1h" | "12h" | "24h" | "1w" | "30d"

export type NetworkProbeSeriesPointInput = {
	created: string
	success: boolean
	latencyMs?: number
	packetLossPercent?: number
	httpStatus?: number
	error?: string
	failureCategory?: string
}

export type NetworkProbeSeriesInput = {
	id: string
	label: string
	probeId: string
	systemId: string
	type: string
	targetLabel: string
	points: NetworkProbeSeriesPointInput[]
}

export type NetworkProbeChartRow = Record<string, number | string | null | boolean> & {
	created: number
}

export type RenderedNetworkProbeSeries = NetworkProbeSeriesInput & {
	color: string
	slug: string
	rows: NetworkProbeChartRow[]
	segments: NetworkProbeChartRow[][]
	pointCount: number
}

export type NetworkProbePreparedChart = {
	chartRows: NetworkProbeChartRow[]
	renderedSeries: RenderedNetworkProbeSeries[]
	hasLatency: boolean
	hasAnyPoints: boolean
	latestPoint?: NetworkProbeSeriesPointInput
	timeData: { domain: [number, number]; ticks: number[] }
}

const COLORS = ["var(--chart-1)", "var(--chart-2)", "var(--chart-3)", "var(--chart-4)", "var(--chart-5)"]

export function prepareNetworkProbeChartData(
	series: NetworkProbeSeriesInput[],
	range: NetworkProbeChartRange,
	nowMs = Date.now()
): NetworkProbePreparedChart {
	const timeData = getNetworkChartTimeData(range, nowMs)
	const rowMap = new Map<number, NetworkProbeChartRow>()
	let hasLatency = false
	let hasAnyPoints = false
	let latestPoint: NetworkProbeSeriesPointInput | undefined
	const renderedSeries: RenderedNetworkProbeSeries[] = series.map((item, index) => {
		const slug = `series-${index}`
		const pointRows: NetworkProbeChartRow[] = []
		for (const point of item.points) {
			const timestamp = new Date(point.created).getTime()
			if (!Number.isFinite(timestamp)) continue
			if (timestamp < timeData.domain[0] || timestamp > timeData.domain[1]) continue
			hasAnyPoints = true
			if (!latestPoint || timestamp > new Date(latestPoint.created).getTime()) {
				latestPoint = point
			}
			const row: NetworkProbeChartRow = { created: timestamp }
			row[slug] = point.success ? (point.latencyMs ?? null) : null
			row[`${slug}:success`] = point.success
			row[`${slug}:error`] = point.error || point.failureCategory || null
			row[`${slug}:httpStatus`] = point.httpStatus ?? null
			row[`${slug}:packetLossPercent`] = point.packetLossPercent ?? null
			if (typeof point.latencyMs === "number") {
				hasLatency = true
			}
			pointRows.push(row)
		}
		const sortedRows = bucketHistoricalRows(
			pointRows.sort((a, b) => Number(a.created) - Number(b.created)),
			slug,
			range
		)
		for (const row of sortedRows) {
			const timestamp = Number(row.created)
			rowMap.set(timestamp, { ...rowMap.get(timestamp), ...row, created: timestamp })
		}
		return {
			...item,
			color: COLORS[index % COLORS.length],
			slug,
			rows: sortedRows,
			segments: buildRenderSegments(sortedRows, slug, range),
			pointCount: sortedRows.filter((row) => typeof row[slug] === "number").length,
		}
	})
	return {
		chartRows: [...rowMap.values()].sort((a, b) => Number(a.created) - Number(b.created)),
		renderedSeries,
		hasLatency,
		hasAnyPoints,
		latestPoint,
		timeData,
	}
}

function buildRenderSegments(
	rows: NetworkProbeChartRow[],
	slug: string,
	range: NetworkProbeChartRange
): NetworkProbeChartRow[][] {
	if (range !== "1m") {
		return rows.length ? [rows] : []
	}
	const maxLiveSuccessGapMs = 10_000
	const segments: NetworkProbeChartRow[][] = []
	let current: NetworkProbeChartRow[] = []
	let lastSuccessMs: number | undefined
	for (const row of rows) {
		const timestamp = Number(row.created)
		const hasLatency = typeof row[slug] === "number"
		if (hasLatency) {
			if (lastSuccessMs !== undefined && timestamp - lastSuccessMs > maxLiveSuccessGapMs && current.length) {
				segments.push(current)
				current = []
			}
			current.push(row)
			lastSuccessMs = timestamp
			continue
		}
		if (current.length) {
			current.push(row)
		}
	}
	if (current.length) {
		segments.push(current)
	}
	return segments.filter((segment) => segment.some((row) => typeof row[slug] === "number"))
}

function bucketHistoricalRows(
	rows: NetworkProbeChartRow[],
	slug: string,
	range: NetworkProbeChartRange
): NetworkProbeChartRow[] {
	if (range === "1m" || rows.length < 2) {
		return rows
	}
	const bucketMs = getHistoricalBucketMs(range)
	const buckets = new Map<number, NetworkProbeChartRow[]>()
	for (const row of rows) {
		const bucketStart = Math.floor(Number(row.created) / bucketMs) * bucketMs
		const bucket = buckets.get(bucketStart)
		if (bucket) {
			bucket.push(row)
		} else {
			buckets.set(bucketStart, [row])
		}
	}
	return [...buckets.entries()]
		.sort(([a], [b]) => a - b)
		.map(([bucketStart, bucketRows]) => summarizeHistoricalBucket(bucketStart, bucketMs, bucketRows, slug))
}

function summarizeHistoricalBucket(
	bucketStart: number,
	bucketMs: number,
	rows: NetworkProbeChartRow[],
	slug: string
): NetworkProbeChartRow {
	const successRows = rows.filter((row) => typeof row[slug] === "number")
	const latestRow = rows[rows.length - 1]
	const created = Math.min(bucketStart + Math.floor(bucketMs / 2), Number(latestRow.created))
	const row: NetworkProbeChartRow = { created }
	if (successRows.length > 0) {
		const average = successRows.reduce((sum, item) => sum + Number(item[slug]), 0) / Math.max(successRows.length, 1)
		row[slug] = Number(average.toFixed(2))
		row[`${slug}:success`] = true
		row[`${slug}:error`] = null
		row[`${slug}:httpStatus`] = latestDefined(successRows, `${slug}:httpStatus`)
		row[`${slug}:packetLossPercent`] = latestDefined(successRows, `${slug}:packetLossPercent`)
		return row
	}
	row[slug] = null
	row[`${slug}:success`] = false
	row[`${slug}:error`] = latestRow[`${slug}:error`] ?? null
	row[`${slug}:httpStatus`] = latestRow[`${slug}:httpStatus`] ?? null
	row[`${slug}:packetLossPercent`] = latestRow[`${slug}:packetLossPercent`] ?? null
	return row
}

function latestDefined(rows: NetworkProbeChartRow[], key: string) {
	for (let index = rows.length - 1; index >= 0; index--) {
		const value = rows[index][key]
		if (value !== null && value !== undefined) {
			return value
		}
	}
	return null
}

function getHistoricalBucketMs(range: NetworkProbeChartRange) {
	switch (range) {
		case "30m":
			return 15_000
		case "1h":
			return 30_000
		case "12h":
			return 6 * 60_000
		case "24h":
			return 12 * 60_000
		case "1w":
			return 90 * 60_000
		case "30d":
			return 6 * 60 * 60_000
		default:
			return 60_000
	}
}

export function getNetworkChartTimeData(
	range: NetworkProbeChartRange,
	nowMs = Date.now()
): { domain: [number, number]; ticks: number[] } {
	const now = new Date(nowMs)
	const rangeData = getRangeData(range)
	const start = rangeData.getOffset(now)
	return {
		domain: [start.getTime(), now.getTime()],
		ticks: timeTicks(start, now, rangeData.ticks).map((date) => date.getTime()),
	}
}

function getRangeData(range: NetworkProbeChartRange) {
	switch (range) {
		case "1m":
			return { ticks: 3, getOffset: (endTime: Date) => timeMinute.offset(endTime, -1) }
		case "30m":
			return { ticks: 6, getOffset: (endTime: Date) => timeMinute.offset(endTime, -30) }
		case "1h":
			return { ticks: 12, getOffset: (endTime: Date) => timeHour.offset(endTime, -1) }
		case "12h":
			return { ticks: 12, getOffset: (endTime: Date) => timeHour.offset(endTime, -12) }
		case "24h":
			return { ticks: 12, getOffset: (endTime: Date) => timeHour.offset(endTime, -24) }
		case "1w":
			return { ticks: 7, getOffset: (endTime: Date) => timeDay.offset(endTime, -7) }
		case "30d":
			return { ticks: 30, getOffset: (endTime: Date) => timeDay.offset(endTime, -30) }
		default:
			return { ticks: 6, getOffset: (endTime: Date) => timeMinute.offset(endTime, -30) }
	}
}
