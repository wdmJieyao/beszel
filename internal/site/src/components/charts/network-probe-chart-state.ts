import type { PublicChartRange } from "@/types"

export type NetworkProbeChartEmptyState =
	| { kind: "loading"; message: string }
	| { kind: "error"; message: string }
	| { kind: "empty"; message: string }

export function getNetworkProbeChartEmptyState({
	range,
	emptyLabel,
	loading,
	error,
}: {
	range: PublicChartRange
	emptyLabel: string
	loading: boolean
	error: boolean
}): NetworkProbeChartEmptyState {
	if (loading) {
		return { kind: "loading", message: range === "1m" ? "等待实时数据" : "加载中" }
	}
	if (error) {
		return { kind: "error", message: "检测结果加载失败" }
	}
	return { kind: "empty", message: emptyLabel }
}
