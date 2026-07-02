import { Trans } from "@lingui/react/macro"
import { ActivityIcon, HardDriveIcon, LogInIcon, MemoryStickIcon, MicrochipIcon } from "lucide-react"
import { Component, useEffect, useMemo, useState } from "react"
import { CartesianGrid, Line, LineChart, XAxis, YAxis } from "recharts"
import { NetworkProbeChart, NetworkProbeStatus } from "@/components/charts/network-probe-chart"
import { Link } from "@/components/router"
import { Badge } from "@/components/ui/badge"
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import { ChartContainer, ChartTooltip, ChartTooltipContent } from "@/components/ui/chart"
import { Dialog, DialogContent, DialogHeader, DialogTitle } from "@/components/ui/dialog"
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select"
import { getPublicStatus } from "@/lib/api"
import { formatShortDate, getPublicChartTimeData, publicChartRangeOptions } from "@/lib/utils"
import type {
	NetworkProbeChartSeries,
	PublicChartRange,
	PublicMetricPoint,
	PublicProbeSummary,
	PublicStatusResponse,
	PublicStatusSystem,
} from "@/types"

const DEFAULT_PUBLIC_RANGE: PublicChartRange = "30m"

export default function PublicStatus() {
	const [data, setData] = useState<PublicStatusResponse>()
	const [refreshError, setRefreshError] = useState(false)
	const [range, setRange] = useState<PublicChartRange>(DEFAULT_PUBLIC_RANGE)
	const now = useNow()

	useEffect(() => {
		document.title = "公共看板 / Beszel"
		let cancelled = false
		let timer: ReturnType<typeof setTimeout> | undefined

		const load = async () => {
			try {
				const next = await getPublicStatus(range)
				if (!cancelled) {
					setData(next)
					setRefreshError(false)
				}
			} catch (_error) {
				if (!cancelled) {
					setRefreshError(true)
				}
			} finally {
				if (!cancelled) {
					timer = setTimeout(load, 20_000)
				}
			}
		}

		load()
		return () => {
			cancelled = true
			if (timer) clearTimeout(timer)
		}
	}, [range])

	if (refreshError && !data) {
		return (
			<PublicShell>
				<Card>
					<CardContent className="py-10 text-center text-destructive">
						<Trans>公共看板加载失败。</Trans>
					</CardContent>
				</Card>
			</PublicShell>
		)
	}
	if (!data) {
		return (
			<PublicShell>
				<Card>
					<CardContent className="py-10 text-center text-muted-foreground">
						<Trans>正在加载...</Trans>
					</CardContent>
				</Card>
			</PublicShell>
		)
	}
	return (
		<PublicShell>
			{refreshError && (
				<Card>
					<CardContent className="py-3 text-sm text-muted-foreground">
						<Trans>最新刷新失败，正在保留当前数据。</Trans>
					</CardContent>
				</Card>
			)}
			{data.systems.length === 0 ? (
				<Card>
					<CardContent className="py-10 text-center text-muted-foreground">
						<Trans>当前没有公开展示的节点。</Trans>
					</CardContent>
				</Card>
			) : (
				<div className="grid gap-3 xl:grid-cols-2">
					{data.systems.map((system) => (
						<SystemStatusCard key={system.id} system={system} now={now} range={range} onRangeChange={setRange} />
					))}
				</div>
			)}
		</PublicShell>
	)
}

function PublicShell({ children }: { children: React.ReactNode }) {
	return (
		<main className="mx-auto flex w-full max-w-7xl flex-col gap-4 px-4 py-6">
			<header className="flex flex-wrap items-center justify-between gap-3">
				<div>
					<h1 className="text-2xl font-semibold tracking-normal">
						<Trans>公共看板</Trans>
					</h1>
					<p className="text-sm text-muted-foreground">
						<Trans>公开节点状态与线路检测走势。</Trans>
					</p>
				</div>
				<div className="flex flex-wrap items-center gap-3">
					<Link
						href="/settings/general"
						className="inline-flex items-center gap-1.5 rounded-md border px-3 py-1.5 text-sm hover:bg-muted"
					>
						<LogInIcon className="size-4" />
						<Trans>管理登录</Trans>
					</Link>
				</div>
			</header>
			{children}
		</main>
	)
}

function SystemStatusCard({
	system,
	now,
	range,
	onRangeChange,
}: {
	system: PublicStatusSystem
	now: number
	range: PublicChartRange
	onRangeChange: (range: PublicChartRange) => void
}) {
	const freshness = useMemo(() => {
		if (!system.freshness) {
			return undefined
		}
		return relativeTime(system.freshness, now)
	}, [system.freshness, now])
	const unavailable = system.metrics.unavailable ?? []
	const latencySeries = useMemo(() => publicLatencySeries(system), [system])
	const summaryProbes = system.probes.filter((probe) => probe.type === "http_get")
	const metricSeries = useMemo(
		() => publicMetricSeries(system.history ?? [], system, range),
		[range, system.history, system.metrics, system.freshness]
	)
	const [metricDialogOpen, setMetricDialogOpen] = useState(false)
	return (
		<Card>
			<CardHeader className="flex flex-row items-center justify-between gap-3 space-y-0 pb-3">
				<CardTitle className="min-w-0 truncate text-base">{system.name}</CardTitle>
				<Badge variant={system.status === "up" ? "default" : "secondary"}>{statusLabel(system.status)}</Badge>
			</CardHeader>
			<CardContent className="space-y-4">
				<div className="grid grid-cols-3 gap-2 text-sm">
					<Metric
						icon={<MicrochipIcon className="size-3.5" />}
						label="CPU"
						value={system.metrics.cpuPercent}
						unavailable={unavailable}
						metric="cpu"
						onClick={() => setMetricDialogOpen(true)}
					/>
					<Metric
						icon={<MemoryStickIcon className="size-3.5" />}
						label="内存"
						value={system.metrics.memoryPercent}
						unavailable={unavailable}
						metric="memory"
						onClick={() => setMetricDialogOpen(true)}
					/>
					<Metric
						icon={<HardDriveIcon className="size-3.5" />}
						label="磁盘"
						value={system.metrics.diskPercent}
						unavailable={unavailable}
						metric="disk"
						onClick={() => setMetricDialogOpen(true)}
					/>
				</div>
				<Dialog open={metricDialogOpen} onOpenChange={setMetricDialogOpen}>
					<DialogContent className="max-h-[90dvh] w-[calc(100vw-24px)] max-w-5xl overflow-y-auto">
						<DialogHeader>
							<DialogTitle>{system.name} 资源趋势</DialogTitle>
						</DialogHeader>
						<div className="flex justify-end">
							<PublicChartRangeSelect value={range} onChange={onRangeChange} className="w-32" />
						</div>
						<div className="grid gap-3">
							<PublicMetricTrendChart
								data={metricSeries}
								dataKey="cpuPercent"
								label="CPU 使用率"
								color="var(--chart-1)"
								range={range}
								heightClassName="h-40"
							/>
							<PublicMetricTrendChart
								data={metricSeries}
								dataKey="memoryPercent"
								label="内存使用率"
								color="var(--chart-2)"
								range={range}
								heightClassName="h-40"
							/>
							<PublicMetricTrendChart
								data={metricSeries}
								dataKey="diskPercent"
								label="磁盘使用率"
								color="var(--chart-3)"
								range={range}
								heightClassName="h-40"
							/>
						</div>
					</DialogContent>
				</Dialog>
				{latencySeries.length > 0 && (
					<div className="space-y-3">
						<div className="rounded-md border p-3">
							<div className="mb-2 flex items-center justify-between gap-2 text-sm">
								<div className="min-w-0">
									<div className="truncate font-medium">线路检测</div>
									<div className="truncate text-xs text-muted-foreground">{latencySeries.length} 条线路</div>
								</div>
								<PublicChartRangeSelect value={range} onChange={onRangeChange} className="w-32" />
							</div>
							<ChartErrorBoundary>
								<NetworkProbeChart series={latencySeries} range={range} heightClassName="h-40" />
							</ChartErrorBoundary>
						</div>
					</div>
				)}
				{summaryProbes.length > 0 && (
					<div className="grid gap-2">
						{summaryProbes.map((probe) => (
							<div key={probe.id} className="rounded-md border bg-muted/30 px-3 py-2 text-sm">
								<div className="flex items-center justify-between gap-2">
									<div className="min-w-0">
										<div className="truncate font-medium">{probe.name}</div>
									</div>
									<NetworkProbeStatus latest={probe.latest} />
								</div>
							</div>
						))}
					</div>
				)}
				{freshness ? (
					<div className="inline-flex items-center gap-1.5 text-xs text-muted-foreground">
						<ActivityIcon className="size-3.5" />
						<Trans>节点最后上报</Trans>: {freshness}
					</div>
				) : unavailable.includes("freshness") ? (
					<div className="inline-flex items-center gap-1.5 text-xs text-muted-foreground">
						<ActivityIcon className="size-3.5" />
						<Trans>节点最后上报</Trans>: <span>不可用</span>
					</div>
				) : null}
			</CardContent>
		</Card>
	)
}

function publicMetricSeries(history: PublicMetricPoint[], system: PublicStatusSystem, range: PublicChartRange) {
	const cutoff = getPublicChartTimeData(range).getOffset(new Date()).getTime()
	const visible = history.filter((point) => {
		if (!point.created) return false
		const timestamp = new Date(point.created).getTime()
		return Number.isFinite(timestamp) && timestamp >= cutoff
	})
	const latest = {
		created: system.freshness ?? "",
		cpuPercent: system.metrics.cpuPercent,
		memoryPercent: system.metrics.memoryPercent,
		diskPercent: system.metrics.diskPercent,
	}
	if (!latest.created) {
		return visible
	}
	if (visible.length === 0) {
		return [latest]
	}
	const next = [...visible]
	const latestTimestamp = new Date(latest.created).getTime()
	const lastTimestamp = new Date(next[next.length - 1].created).getTime()
	if (Number.isFinite(latestTimestamp) && Number.isFinite(lastTimestamp) && latestTimestamp <= lastTimestamp) {
		next[next.length - 1] = { ...next[next.length - 1], ...latest }
		return next
	}
	next.push(latest)
	return next
}

function PublicMetricTrendChart({
	data,
	dataKey,
	label,
	color,
	range,
	heightClassName,
}: {
	data: PublicMetricPoint[]
	dataKey: "cpuPercent" | "memoryPercent" | "diskPercent"
	label: string
	color: string
	range: PublicChartRange
	heightClassName?: string
}) {
	const values = data
		.map((point) => point[dataKey])
		.filter((value): value is number => typeof value === "number" && Number.isFinite(value))
	const domain = metricPercentDomain(values)
	const latest = values.at(-1)
	if (values.length === 0) {
		return (
			<div className="flex h-36 flex-col justify-center rounded-md border bg-muted/20 px-3 text-sm text-muted-foreground">
				<span className="font-medium text-foreground">{label}</span>
				<span>暂无数据</span>
			</div>
		)
	}
	return (
		<div className="space-y-2 rounded-md border bg-muted/20 p-3">
			<div className="flex items-center justify-between gap-2 text-sm">
				<div className="min-w-0">
					<div className="truncate font-medium">{label}</div>
				</div>
				<span className="shrink-0 text-xs text-muted-foreground">
					{typeof latest === "number" ? `${latest.toFixed(1)}%` : "-"}
				</span>
			</div>
			<ChartContainer className={`${heightClassName ?? "h-36"} w-full`}>
				<LineChart data={data} margin={{ top: 4, right: 8, bottom: 18, left: 0 }}>
					<CartesianGrid vertical={false} strokeDasharray="2 6" />
					<YAxis
						width={38}
						domain={domain}
						tickFormatter={(value) => `${Number(value).toFixed(0)}%`}
						tickLine={false}
						axisLine={false}
						tickMargin={4}
						fontSize={11}
					/>
					<XAxis
						dataKey="created"
						type="category"
						ticks={metricAxisTicks(data, range)}
						tickFormatter={(value) => getPublicChartTimeData(range).format(String(value))}
						tickLine={false}
						axisLine={false}
						tickMargin={6}
						fontSize={11}
						interval="preserveStartEnd"
					/>
					<ChartTooltip
						content={
							<ChartTooltipContent
								labelFormatter={(_, payload) => formatShortDate(payload[0]?.payload.created ?? "")}
								contentFormatter={(item) => (typeof item.value === "number" ? `${item.value.toFixed(1)}%` : "-")}
							/>
						}
					/>
					<Line
						type="monotone"
						dataKey={dataKey}
						name={label}
						stroke={color}
						strokeWidth={1.75}
						dot={false}
						isAnimationActive={false}
						connectNulls
					/>
				</LineChart>
			</ChartContainer>
		</div>
	)
}

function PublicChartRangeSelect({
	value,
	onChange,
	className,
}: {
	value: PublicChartRange
	onChange: (range: PublicChartRange) => void
	className?: string
}) {
	return (
		<Select value={value} onValueChange={(next) => onChange(next as PublicChartRange)}>
			<SelectTrigger className={className}>
				<SelectValue />
			</SelectTrigger>
			<SelectContent>
				{publicChartRangeOptions.map((range) => (
					<SelectItem key={range} value={range}>
						{getPublicChartTimeData(range).label()}
					</SelectItem>
				))}
			</SelectContent>
		</Select>
	)
}

function getPublicTimeData(range: PublicChartRange) {
	const now = new Date()
	const rangeData = getPublicChartTimeData(range)
	return {
		domain: [rangeData.getOffset(now).getTime(), now.getTime()],
		ticks: rangeData.ticks ?? 6,
	}
}

function metricAxisTicks(data: PublicMetricPoint[], range: PublicChartRange) {
	const maxTicks = getPublicTimeData(range).ticks
	if (data.length <= maxTicks) {
		return data.map((point) => point.created)
	}
	const step = Math.max(1, Math.ceil((data.length - 1) / (maxTicks - 1)))
	const ticks = data.filter((_, index) => index % step === 0).map((point) => point.created)
	const last = data[data.length - 1]?.created
	if (last && ticks.at(-1) !== last) {
		ticks.push(last)
	}
	return ticks
}

function metricPercentDomain(values: number[]): [number, number] {
	if (values.length === 0) {
		return [0, 100]
	}
	const min = Math.min(...values)
	const max = Math.max(...values)
	const spread = max - min
	const padding = Math.max(spread * 0.25, 1)
	const lower = Math.max(0, Math.floor(min - padding))
	const upper = Math.min(100, Math.ceil(max + padding))
	if (upper - lower < 2) {
		return [Math.max(0, lower - 1), Math.min(100, upper + 1)]
	}
	return [lower, upper]
}

function statusLabel(status: PublicStatusSystem["status"]) {
	switch (status) {
		case "up":
			return "在线"
		case "down":
			return "离线"
		case "paused":
			return "暂停"
		case "pending":
			return "等待中"
		case "stale":
			return "数据过期"
		default:
			return status
	}
}

function publicLatencySeries(system: PublicStatusSystem): NetworkProbeChartSeries[] {
	return system.probes
		.filter((probe) => probe.type === "tcping" || probe.type === "icmp_ping")
		.map((probe) => publicProbeToSeries(system.id, probe))
}

function publicProbeToSeries(systemId: string, probe: PublicProbeSummary): NetworkProbeChartSeries {
	return {
		id: `${probe.id}:${systemId}`,
		label: probe.name,
		probeId: probe.id,
		systemId,
		type: probe.type,
		targetLabel: "",
		points: probe.series,
	}
}

class ChartErrorBoundary extends Component<{ children: React.ReactNode }, { hasError: boolean }> {
	state = { hasError: false }

	static getDerivedStateFromError() {
		return { hasError: true }
	}

	componentDidCatch(error: unknown) {
		console.error("Public network probe chart failed", error)
	}

	render() {
		if (this.state.hasError) {
			return (
				<div className="flex h-28 items-center justify-center rounded-md bg-muted/40 text-sm text-muted-foreground">
					线路检测图表暂不可用
				</div>
			)
		}
		return this.props.children
	}
}

function Metric({
	icon,
	label,
	value,
	unavailable,
	metric,
	onClick,
}: {
	icon: React.ReactNode
	label: string
	value?: number
	unavailable: string[]
	metric: "cpu" | "memory" | "disk"
	onClick?: () => void
}) {
	const isUnavailable = unavailable.includes(metric)
	return (
		<button
			type="button"
			className="rounded-md bg-muted px-3 py-2 text-left transition-colors hover:bg-muted/80 focus-visible:outline-hidden focus-visible:ring-2 focus-visible:ring-ring"
			onClick={onClick}
		>
			<div className="flex items-center gap-1 text-xs text-muted-foreground">
				{icon}
				{label}
			</div>
			<div className="font-medium">
				{typeof value === "number" ? `${value.toFixed(1)}%` : isUnavailable ? "不可用" : "-"}
			</div>
		</button>
	)
}

function useNow() {
	const [now, setNow] = useState(Date.now())
	useEffect(() => {
		const timer = setInterval(() => setNow(Date.now()), 1_000)
		return () => clearInterval(timer)
	}, [])
	return now
}

function relativeTime(value: string, now: number) {
	const timestamp = new Date(value).getTime()
	if (!Number.isFinite(timestamp)) {
		return value
	}
	const seconds = Math.max(0, Math.floor((now - timestamp) / 1000))
	if (seconds < 60) return `${seconds} 秒前`
	const minutes = Math.floor(seconds / 60)
	if (minutes < 60) return `${minutes} 分钟前`
	const hours = Math.floor(minutes / 60)
	if (hours < 24) return `${hours} 小时前`
	return new Date(timestamp).toLocaleString()
}
