import { CircleAlertIcon, LoaderCircleIcon } from "lucide-react"
import { useEffect, useMemo, useState } from "react"
import { CartesianGrid, Line, LineChart, Tooltip, XAxis, YAxis } from "recharts"
import { ChartContainer, ChartTooltipContent } from "@/components/ui/chart"
import { getPublicChartTimeData } from "@/lib/utils"
import type { NetworkProbeChartSeries, NetworkProbeSeriesPoint, PublicChartRange } from "@/types"
import { prepareNetworkProbeChartData } from "./network-probe-chart-data"
import { getNetworkProbeChartEmptyState } from "./network-probe-chart-state.ts"

const NETWORK_PROBE_CHART_VERSION = "2026-07-02-cache-refresh"

export function NetworkProbeChart({
	series,
	range = "30m",
	heightClassName = "h-28",
	emptyLabel = "暂无检测结果",
	loading = false,
	error = false,
}: {
	series: NetworkProbeChartSeries[]
	range?: PublicChartRange
	heightClassName?: string
	emptyLabel?: string
	loading?: boolean
	error?: boolean
}) {
	const [hiddenSeries, setHiddenSeries] = useState<Set<string>>(() => new Set())
	const [nowMs, setNowMs] = useState(() => Date.now())
	const emptyState = getNetworkProbeChartEmptyState({ range, emptyLabel, loading, error })
	const { chartRows, renderedSeries, hasLatency, hasAnyPoints, latestPoint, timeData } = useMemo(
		() => prepareNetworkProbeChartData(series, range, nowMs),
		[range, series, nowMs]
	)
	const visibleSeries = useMemo(
		() => renderedSeries.filter((item) => !hiddenSeries.has(item.id)),
		[hiddenSeries, renderedSeries]
	)
	useEffect(() => {
		setNowMs(Date.now())
		if (range !== "1m") {
			return
		}
		const interval = window.setInterval(() => setNowMs(Date.now()), 1000)
		return () => window.clearInterval(interval)
	}, [range])
	useEffect(() => {
		setHiddenSeries((current) => {
			const ids = new Set(series.map((item) => item.id))
			let changed = false
			const next = new Set<string>()
			for (const id of current) {
				if (ids.has(id)) {
					next.add(id)
				} else {
					changed = true
				}
			}
			return changed ? next : current
		})
	}, [series])

	if (!series.length) {
		return (
			<div
				className={`${heightClassName} flex items-center justify-center rounded-md bg-muted/40 px-3 text-sm text-muted-foreground`}
			>
				{emptyState.kind === "loading" ? (
					<span className="inline-flex items-center gap-2">
						<LoaderCircleIcon className="size-4 animate-spin" />
						{emptyState.message}
					</span>
				) : (
					emptyState.message
				)}
			</div>
		)
	}

	const legend = (
		<div className="flex flex-wrap gap-2 text-xs">
			{renderedSeries.map((item) => {
				const hidden = hiddenSeries.has(item.id)
				return (
					<button
						key={item.id}
						type="button"
						className={`inline-flex items-center gap-1.5 rounded-md border px-2 py-1 transition-colors ${
							hidden ? "text-muted-foreground opacity-60" : "bg-muted/40 text-foreground"
						}`}
						onClick={() =>
							setHiddenSeries((current) => {
								const next = new Set(current)
								if (next.has(item.id)) {
									next.delete(item.id)
								} else {
									next.add(item.id)
								}
								return next
							})
						}
						title={hidden ? "点击显示" : "点击隐藏"}
					>
						<span className="h-2 w-2 rounded-[2px]" style={{ backgroundColor: item.color }} />
						<span>{item.label}</span>
					</button>
				)
			})}
		</div>
	)

	if (!hasAnyPoints) {
		return (
			<div className="space-y-2">
				{legend}
				<div
					className={`${heightClassName} flex items-center justify-center rounded-md bg-muted/40 px-3 text-sm text-muted-foreground`}
				>
					{emptyState.kind === "loading" ? (
						<span className="inline-flex items-center gap-2">
							<LoaderCircleIcon className="size-4 animate-spin" />
							{emptyState.message}
						</span>
					) : (
						emptyState.message
					)}
				</div>
			</div>
		)
	}

	if (!hasLatency) {
		return (
			<div className="space-y-2">
				{legend}
				<div
					className={`${heightClassName} flex flex-col items-center justify-center gap-1 rounded-md bg-muted/40 px-3 text-center text-sm text-muted-foreground`}
				>
					<span>暂无成功延迟数据</span>
					{latestPoint && !latestPoint.success && (
						<span className="inline-flex items-center gap-1 text-xs text-destructive">
							<CircleAlertIcon className="size-3.5" />
							{formatFailureLabel(latestPoint.failureCategory) || latestPoint.error || "检测失败"}
						</span>
					)}
				</div>
			</div>
		)
	}

	return (
		<div className="space-y-2" data-network-probe-chart-version={NETWORK_PROBE_CHART_VERSION}>
			{legend}
			{visibleSeries.length === 0 ? (
				<div
					className={`${heightClassName} flex items-center justify-center rounded-md bg-muted/40 text-sm text-muted-foreground`}
				>
					已隐藏所有线路
				</div>
			) : (
				<ChartContainer className={`${heightClassName} w-full`}>
					<LineChart accessibilityLayer data={chartRows} margin={{ top: 4, right: 12, bottom: 18, left: 4 }}>
						<CartesianGrid vertical={false} strokeDasharray="2 6" />
						<YAxis
							direction="ltr"
							orientation="left"
							width={64}
							domain={[0, "auto"]}
							tickFormatter={(value) => (hasLatency ? `${Math.round(Number(value))}ms` : "")}
							tickLine={false}
							axisLine={false}
							tickMargin={4}
							fontSize={11}
						/>
						<XAxis
							dataKey="created"
							type="number"
							scale="time"
							domain={timeData.domain}
							ticks={timeData.ticks}
							allowDataOverflow
							tickFormatter={(value) => getPublicChartTimeData(range).format(new Date(Number(value)).toISOString())}
							tickLine={false}
							axisLine={false}
							tickMargin={6}
							fontSize={11}
							interval="preserveStartEnd"
						/>
						<Tooltip
							content={
								<ChartTooltipContent
									contentFormatter={(item, key) => {
										if (typeof item.value === "number") {
											return `${item.value.toFixed(0)} ms`
										}
										const payload = item.payload as Record<string, any>
										const error = payload[`${key}:error`]
										const httpStatus = payload[`${key}:httpStatus`]
										const packetLossPercent = payload[`${key}:packetLossPercent`]
										if (httpStatus) return `HTTP ${httpStatus}`
										if (typeof packetLossPercent === "number") return `丢包 ${packetLossPercent.toFixed(0)}%`
										return error || "检测失败"
									}}
								/>
							}
						/>
						{visibleSeries.flatMap((item) =>
							item.segments.map((segment, index) => {
								const segmentPointCount = segment.filter((row) => typeof row[item.slug] === "number").length
								return (
									<Line
										key={`${item.id}:${index}`}
										dataKey={item.slug}
										data={segment}
										name={item.label}
										type="monotoneX"
										connectNulls
										dot={segmentPointCount < 2 ? { r: 3, fill: item.color, strokeWidth: 0 } : false}
										strokeWidth={1.5}
										stroke={item.color}
										isAnimationActive={false}
										strokeLinejoin="round"
										strokeLinecap="round"
									/>
								)
							})
						)}
					</LineChart>
				</ChartContainer>
			)}
		</div>
	)
}

export function NetworkProbeStatus({ latest }: { latest?: NetworkProbeSeriesPoint }) {
	if (!latest) {
		return <span className="text-muted-foreground">暂无结果</span>
	}
	if (latest.success) {
		return <span>{typeof latest.latencyMs === "number" ? `${latest.latencyMs.toFixed(0)} ms` : "可达"}</span>
	}
	return (
		<span className="inline-flex items-center gap-1 text-destructive">
			<CircleAlertIcon className="size-3.5" />
			{formatFailureLabel(latest.failureCategory) || latest.error || "检测失败"}
		</span>
	)
}

export function formatFailureLabel(category?: string) {
	switch (category) {
		case "invalid_target":
			return "目标格式无效"
		case "dns_failure":
			return "DNS 解析失败"
		case "timeout":
			return "检测超时"
		case "connection_refused":
			return "连接被拒绝"
		case "target_unreachable":
			return "目标不可达"
		case "execution_node_unavailable":
			return "执行节点不可用"
		case "unsupported":
			return "不支持"
		case "unknown_failure":
			return "未知失败"
		default:
			return undefined
	}
}
