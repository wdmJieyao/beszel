import { memo, useState } from "react"
import { Trans } from "@lingui/react/macro"
import { compareSemVer, parseSemVer } from "@/lib/utils"
import type { GPUData } from "@/types"
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs"
import InfoBar from "./system/info-bar"
import { useSystemData } from "./system/use-system-data"
import { useNetworkProbeData } from "./system/use-network-probe-data"
import { CpuChart, ContainerCpuChart } from "./system/charts/cpu-charts"
import { MemoryChart, ContainerMemoryChart, SwapChart } from "./system/charts/memory-charts"
import { RootDiskCharts, ExtraFsCharts } from "./system/charts/disk-charts"
import { BandwidthChart, ContainerNetworkChart } from "./system/charts/network-charts"
import { TemperatureChart, BatteryChart } from "./system/charts/sensor-charts"
import { GpuPowerChart, GpuDetailCharts } from "./system/charts/gpu-charts"
import { LazyContainersTable, LazySmartTable, LazySystemdTable } from "./system/lazy-tables"
import { LoadAverageChart } from "./system/charts/load-average-chart"
import { ContainerIcon, CpuIcon, HardDriveIcon, TerminalSquareIcon } from "lucide-react"
import { GpuIcon } from "../ui/icons"
import SystemdTable from "../systemd-table/systemd-table"
import ContainersTable from "../containers-table/containers-table"
import { NetworkProbeChart, NetworkProbeStatus } from "../charts/network-probe-chart"

const SEMVER_0_14_0 = parseSemVer("0.14.0")
const SEMVER_0_15_0 = parseSemVer("0.15.0")

export default memo(function SystemDetail({ id }: { id: string }) {
	const systemData = useSystemData(id)
	const probeData = useNetworkProbeData(id, systemData.chartData.chartTime)

	const {
		system,
		systemStats,
		containerData,
		chartData,
		containerChartConfigs,
		details,
		grid,
		setGrid,
		displayMode,
		setDisplayMode,
		activeTab,
		setActiveTab,
		mountedTabs,
		tabsRef,
		maxValues,
		isLongerChart,
		showMax,
		dataEmpty,
		isPodman,
		lastGpus,
		hasGpuData,
		hasGpuEnginesData,
		hasGpuPowerData,
	} = systemData

	// extra margin to add to bottom of page, specifically for temperature chart,
	// where the tooltip can go past the bottom of the page if lots of sensors
	const [pageBottomExtraMargin, setPageBottomExtraMargin] = useState(0)

	if (!system.id) {
		return null
	}

	const hasContainers = containerData.length > 0
	const maybeHasSmartData = compareSemVer(chartData.agentVersion, SEMVER_0_15_0) >= 0
	const hasContainersTable = hasContainers && compareSemVer(chartData.agentVersion, SEMVER_0_14_0) >= 0
	const hasSystemd = !!system.info?.sv
	const hasGpu = hasGpuData || hasGpuPowerData
	const latencyProbeGroups = probeData.groups.filter((group) => group.type === "tcping" || group.type === "icmp_ping")
	const summaryProbeGroups = probeData.groups.filter((group) => group.type === "http_get")
	const latencySeries = latencyProbeGroups.flatMap((group) => group.series)

	// keep tabsRef in sync for keyboard navigation
	const tabs = ["core", "disk"]
	if (hasGpu) tabs.push("gpu")
	if (hasContainers) tabs.push("containers")
	if (hasSystemd) tabs.push("services")
	tabsRef.current = tabs

	// shared chart props
	const coreProps = { chartData, grid, dataEmpty, showMax, isLongerChart, maxValues }

	function defaultLayout() {
		return (
			<>
				{/* main charts */}
				<div className="grid xl:grid-cols-2 gap-4">
					<CpuChart {...coreProps} />

					{hasContainers && (
						<ContainerCpuChart
							chartData={chartData}
							grid={grid}
							dataEmpty={dataEmpty}
							isPodman={isPodman}
							cpuConfig={containerChartConfigs.cpu}
						/>
					)}

					<MemoryChart {...coreProps} />

					{hasContainers && (
						<ContainerMemoryChart
							chartData={chartData}
							grid={grid}
							dataEmpty={dataEmpty}
							isPodman={isPodman}
							memoryConfig={containerChartConfigs.memory}
						/>
					)}

					<RootDiskCharts systemData={systemData} />

					<BandwidthChart {...coreProps} systemStats={systemStats} />

					{hasContainers && (
						<ContainerNetworkChart
							chartData={chartData}
							grid={grid}
							dataEmpty={dataEmpty}
							isPodman={isPodman}
							networkConfig={containerChartConfigs.network}
						/>
					)}

					<SwapChart chartData={chartData} grid={grid} dataEmpty={dataEmpty} systemStats={systemStats} />

					<LoadAverageChart chartData={chartData} grid={grid} dataEmpty={dataEmpty} />

					<TemperatureChart {...coreProps} />

					<BatteryChart {...coreProps} />

					{hasGpuPowerData && <GpuPowerChart chartData={chartData} grid={grid} dataEmpty={dataEmpty} />}
				</div>

				{hasGpuData && lastGpus && (
					<GpuDetailCharts
						chartData={chartData}
						grid={grid}
						dataEmpty={dataEmpty}
						lastGpus={lastGpus as Record<string, GPUData>}
						hasGpuEnginesData={hasGpuEnginesData}
					/>
				)}

				<ExtraFsCharts systemData={systemData} />

				{maybeHasSmartData && <LazySmartTable systemId={system.id} />}

				{hasContainersTable && <LazyContainersTable systemId={system.id} />}

				{hasSystemd && <LazySystemdTable systemId={system.id} />}
			</>
		)
	}

	function tabbedLayout() {
		return (
			<Tabs value={activeTab} onValueChange={setActiveTab} className="contents">
				<TabsList className="h-11 p-1.5 w-full shadow-xs overflow-auto justify-start">
					<TabsTrigger value="core" className="w-full flex items-center gap-1.5">
						<CpuIcon className="size-3.5" />
						<Trans context="Core system metrics">Core</Trans>
					</TabsTrigger>
					<TabsTrigger value="disk" className="w-full flex items-center gap-1.5">
						<HardDriveIcon className="size-3.5" />
						<Trans>Disk</Trans>
					</TabsTrigger>
					{hasGpu && (
						<TabsTrigger value="gpu" className="w-full flex items-center gap-2">
							<GpuIcon className="size-3.5" />
							<Trans>GPU</Trans>
						</TabsTrigger>
					)}
					{hasContainers && (
						<TabsTrigger value="containers" className="w-full flex items-center gap-2">
							<ContainerIcon className="size-3.5" />
							<Trans>Containers</Trans>
						</TabsTrigger>
					)}
					{hasSystemd && (
						<TabsTrigger value="services" className="w-full flex items-center gap-2">
							<TerminalSquareIcon className="size-3.5" />
							<Trans>Services</Trans>
						</TabsTrigger>
					)}
				</TabsList>

				<TabsContent value="core" forceMount className={activeTab === "core" ? "contents" : "hidden"}>
					<div className="grid xl:grid-cols-2 gap-4">
						<CpuChart {...coreProps} />
						<MemoryChart {...coreProps} />
						<LoadAverageChart chartData={chartData} grid={grid} dataEmpty={dataEmpty} />
						<BandwidthChart {...coreProps} systemStats={systemStats} />
						<TemperatureChart {...coreProps} setPageBottomExtraMargin={setPageBottomExtraMargin} />
						<BatteryChart {...coreProps} />
						<SwapChart chartData={chartData} grid={grid} dataEmpty={dataEmpty} systemStats={systemStats} />
						{pageBottomExtraMargin > 0 && <div style={{ marginBottom: pageBottomExtraMargin }}></div>}
					</div>
				</TabsContent>

				<TabsContent value="disk" forceMount className={activeTab === "disk" ? "contents" : "hidden"}>
					{mountedTabs.has("disk") && (
						<>
							<div className="grid xl:grid-cols-2 gap-4">
								<RootDiskCharts systemData={systemData} />
							</div>
							<ExtraFsCharts systemData={systemData} />
							{maybeHasSmartData && <LazySmartTable systemId={system.id} />}
						</>
					)}
				</TabsContent>

				{hasGpu && (
					<TabsContent value="gpu" forceMount className={activeTab === "gpu" ? "contents" : "hidden"}>
						<div className="grid xl:grid-cols-2 gap-4">
							{hasGpuPowerData && <GpuPowerChart chartData={chartData} grid={grid} dataEmpty={dataEmpty} />}
						</div>
						{hasGpuData && lastGpus && (
							<GpuDetailCharts
								chartData={chartData}
								grid={grid}
								dataEmpty={dataEmpty}
								lastGpus={lastGpus as Record<string, GPUData>}
								hasGpuEnginesData={hasGpuEnginesData}
							/>
						)}
					</TabsContent>
				)}

				{hasContainers && (
					<TabsContent value="containers" forceMount className={activeTab === "containers" ? "contents" : "hidden"}>
						{mountedTabs.has("containers") && (
							<>
								<div className="grid xl:grid-cols-2 gap-4">
									<ContainerCpuChart
										chartData={chartData}
										grid={grid}
										dataEmpty={dataEmpty}
										isPodman={isPodman}
										cpuConfig={containerChartConfigs.cpu}
									/>
									<ContainerMemoryChart
										chartData={chartData}
										grid={grid}
										dataEmpty={dataEmpty}
										isPodman={isPodman}
										memoryConfig={containerChartConfigs.memory}
									/>
									<ContainerNetworkChart
										chartData={chartData}
										grid={grid}
										dataEmpty={dataEmpty}
										isPodman={isPodman}
										networkConfig={containerChartConfigs.network}
									/>
								</div>
								{hasContainersTable && <ContainersTable systemId={system.id} />}
							</>
						)}
					</TabsContent>
				)}

				{hasSystemd && (
					<TabsContent value="services" forceMount className={activeTab === "services" ? "contents" : "hidden"}>
						{mountedTabs.has("services") && <SystemdTable systemId={system.id} />}
					</TabsContent>
				)}
			</Tabs>
		)
	}

	return (
		<div className="grid gap-4 mb-14 overflow-x-clip">
			{/* system info */}
			<InfoBar
				system={system}
				chartData={chartData}
				grid={grid}
				setGrid={setGrid}
				displayMode={displayMode}
				setDisplayMode={setDisplayMode}
				details={details}
			/>

			{latencySeries.length > 0 && (
				<div className="space-y-3 rounded-md border bg-card p-4">
					<div className="flex items-center justify-between gap-2">
						<div>
							<div className="font-medium">线路检测</div>
							<div className="text-sm text-muted-foreground">点击图例显示或隐藏对应线路。</div>
						</div>
						<div className="text-sm text-muted-foreground">{latencySeries.length} 条线路</div>
					</div>
					<NetworkProbeChart series={latencySeries} range={chartData.chartTime} />
				</div>
			)}

			{summaryProbeGroups.length > 0 && (
				<div className="grid gap-3 md:grid-cols-2">
					{summaryProbeGroups.map((group) => (
						<div key={group.id} className="rounded-md border bg-card p-4">
							<div className="flex items-center justify-between gap-2">
								<div className="min-w-0">
									<div className="truncate font-medium">{group.label}</div>
								</div>
								<NetworkProbeStatus latest={group.latest} />
							</div>
						</div>
					))}
				</div>
			)}

			{displayMode === "tabs" ? tabbedLayout() : defaultLayout()}
		</div>
	)
})
