import assert from "node:assert/strict"
import { describe, it } from "node:test"
import { getNetworkProbeChartEmptyState } from "./network-probe-chart-state.ts"

describe("getNetworkProbeChartEmptyState", () => {
	it("prefers loading state for the live one-minute range", () => {
		assert.deepEqual(
			getNetworkProbeChartEmptyState({
				range: "1m",
				emptyLabel: "暂无检测结果",
				loading: true,
				error: false,
			}),
			{ kind: "loading", message: "等待实时数据" }
		)
	})

	it("returns an error state even when there are no visible series yet", () => {
		assert.deepEqual(
			getNetworkProbeChartEmptyState({
				range: "30m",
				emptyLabel: "暂无检测结果",
				loading: false,
				error: true,
			}),
			{ kind: "error", message: "检测结果加载失败" }
		)
	})

	it("falls back to the empty label when loading and error are both false", () => {
		assert.deepEqual(
			getNetworkProbeChartEmptyState({
				range: "30m",
				emptyLabel: "暂无检测结果",
				loading: false,
				error: false,
			}),
			{ kind: "empty", message: "暂无检测结果" }
		)
	})
})
