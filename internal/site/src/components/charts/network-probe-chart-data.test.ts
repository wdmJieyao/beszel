import { describe, it } from "node:test"
import assert from "node:assert/strict"
import { getNetworkChartTimeData, prepareNetworkProbeChartData } from "./network-probe-chart-data.ts"

const now = Date.parse("2026-07-01T12:01:00.000Z")

describe("network probe chart data", () => {
	it("uses a rolling one-minute domain with CPU-like second-level ticks", () => {
		const timeData = getNetworkChartTimeData("1m", now)
		assert.equal(timeData.domain[0], Date.parse("2026-07-01T12:00:00.000Z"))
		assert.equal(timeData.domain[1], now)
		assert.ok(timeData.ticks.length >= 2)
		assert.ok(timeData.ticks.every((tick) => tick >= timeData.domain[0] && tick <= timeData.domain[1]))
	})

	it("keeps a single fresh point as point state without fabricating a long segment", () => {
		const prepared = prepareNetworkProbeChartData(
			[
				{
					id: "probe-a:sys-1",
					label: "广东电信",
					probeId: "probe-a",
					systemId: "sys-1",
					type: "tcping",
					targetLabel: "",
					points: [{ created: "2026-07-01T12:00:55.000Z", success: true, latencyMs: 8 }],
				},
			],
			"1m",
			now
		)
		assert.equal(prepared.renderedSeries[0].pointCount, 1)
		assert.equal(prepared.renderedSeries[0].rows.length, 1)
		assert.equal(prepared.renderedSeries[0].rows[0].created, Date.parse("2026-07-01T12:00:55.000Z"))
	})

	it("keeps a two-point live session as a short current-session segment", () => {
		const prepared = prepareNetworkProbeChartData(
			[
				{
					id: "probe-a:sys-1",
					label: "广东电信",
					probeId: "probe-a",
					systemId: "sys-1",
					type: "tcping",
					targetLabel: "",
					points: [
						{ created: "2026-07-01T12:00:50.000Z", success: true, latencyMs: 8 },
						{ created: "2026-07-01T12:00:58.000Z", success: true, latencyMs: 9 },
					],
				},
			],
			"1m",
			now
		)
		const rows = prepared.renderedSeries[0].rows
		assert.equal(prepared.renderedSeries[0].pointCount, 2)
		assert.equal(rows.length, 2)
		assert.equal(rows[1].created - rows[0].created, 8_000)
	})

	it("keeps failed live samples as gaps instead of fake successful latency points", () => {
		const prepared = prepareNetworkProbeChartData(
			[
				{
					id: "probe-a:sys-1",
					label: "广东电信",
					probeId: "probe-a",
					systemId: "sys-1",
					type: "tcping",
					targetLabel: "",
					points: [
						{ created: "2026-07-01T12:00:54.000Z", success: true, latencyMs: 8 },
						{
							created: "2026-07-01T12:00:56.000Z",
							success: false,
							failureCategory: "timeout",
							error: "timeout",
						},
					],
				},
			],
			"1m",
			now
		)
		const [successRow, failedRow] = prepared.renderedSeries[0].rows
		assert.equal(successRow["series-0"], 8)
		assert.equal(failedRow["series-0"], null)
		assert.equal(failedRow["series-0:success"], false)
		assert.equal(failedRow["series-0:error"], "timeout")
	})

	it("keeps short live failures in one render segment while leaving the failed value null", () => {
		const prepared = prepareNetworkProbeChartData(
			[
				{
					id: "probe-a:sys-1",
					label: "广东电信",
					probeId: "probe-a",
					systemId: "sys-1",
					type: "tcping",
					targetLabel: "",
					points: [
						{ created: "2026-07-01T12:00:54.000Z", success: true, latencyMs: 8 },
						{
							created: "2026-07-01T12:00:56.000Z",
							success: false,
							failureCategory: "timeout",
							error: "timeout",
						},
						{ created: "2026-07-01T12:00:58.000Z", success: true, latencyMs: 9 },
					],
				},
			],
			"1m",
			now
		)
		assert.equal(prepared.renderedSeries[0].segments.length, 1)
		assert.equal(prepared.renderedSeries[0].segments[0][1]["series-0"], null)
	})

	it("keeps consecutive short live failures in one render segment", () => {
		const prepared = prepareNetworkProbeChartData(
			[
				{
					id: "probe-a:sys-1",
					label: "广东移动",
					probeId: "probe-a",
					systemId: "sys-1",
					type: "tcping",
					targetLabel: "",
					points: [
						{ created: "2026-07-01T12:00:50.000Z", success: true, latencyMs: 45 },
						{
							created: "2026-07-01T12:00:52.000Z",
							success: false,
							failureCategory: "timeout",
							error: "timeout",
						},
						{
							created: "2026-07-01T12:00:54.000Z",
							success: false,
							failureCategory: "timeout",
							error: "timeout",
						},
						{ created: "2026-07-01T12:00:56.000Z", success: true, latencyMs: 47 },
					],
				},
			],
			"1m",
			now
		)
		assert.equal(prepared.renderedSeries[0].segments.length, 1)
		assert.equal(prepared.renderedSeries[0].segments[0].length, 4)
	})

	it("breaks one-minute render segments across long gaps", () => {
		const prepared = prepareNetworkProbeChartData(
			[
				{
					id: "probe-a:sys-1",
					label: "广东电信",
					probeId: "probe-a",
					systemId: "sys-1",
					type: "tcping",
					targetLabel: "",
					points: [
						{ created: "2026-07-01T12:00:20.000Z", success: true, latencyMs: 8 },
						{
							created: "2026-07-01T12:00:30.000Z",
							success: false,
							failureCategory: "timeout",
							error: "timeout",
						},
						{ created: "2026-07-01T12:00:44.000Z", success: true, latencyMs: 9 },
					],
				},
			],
			"1m",
			now
		)
		assert.equal(prepared.renderedSeries[0].segments.length, 2)
	})

	it("buckets high-cadence historical samples for 30m readability", () => {
		const points = Array.from({ length: 120 }, (_, index) => ({
			created: new Date(Date.parse("2026-07-01T12:00:00.000Z") + index * 1000).toISOString(),
			success: true,
			latencyMs: 10 + index,
		}))
		const prepared = prepareNetworkProbeChartData(
			[
				{
					id: "probe-a:sys-1",
					label: "广东电信",
					probeId: "probe-a",
					systemId: "sys-1",
					type: "tcping",
					targetLabel: "",
					points,
				},
			],
			"30m",
			Date.parse("2026-07-01T12:02:00.000Z")
		)

		assert.ok(prepared.renderedSeries[0].rows.length < points.length)
		assert.ok(prepared.renderedSeries[0].rows.length <= 10)
		assert.equal(prepared.renderedSeries[0].segments.length, 1)
		assert.equal(prepared.chartRows.length, prepared.renderedSeries[0].rows.length)
	})

	it("keeps failed historical buckets as null latency values", () => {
		const prepared = prepareNetworkProbeChartData(
			[
				{
					id: "probe-a:sys-1",
					label: "广东移动",
					probeId: "probe-a",
					systemId: "sys-1",
					type: "tcping",
					targetLabel: "",
					points: [
						{
							created: "2026-07-01T12:00:05.000Z",
							success: false,
							failureCategory: "timeout",
							error: "timeout",
						},
						{
							created: "2026-07-01T12:00:10.000Z",
							success: false,
							failureCategory: "timeout",
							error: "timeout",
						},
					],
				},
			],
			"30m",
			Date.parse("2026-07-01T12:01:00.000Z")
		)

		assert.equal(prepared.hasLatency, false)
		assert.equal(prepared.renderedSeries[0].rows.length, 1)
		assert.equal(prepared.renderedSeries[0].rows[0]["series-0"], null)
		assert.equal(prepared.renderedSeries[0].rows[0]["series-0:success"], false)
		assert.equal(prepared.renderedSeries[0].rows[0]["series-0:error"], "timeout")
	})
})
