import assert from "node:assert/strict"
import { describe, it } from "node:test"
import { applyHistoricalProbeLoad, loadHistoricalProbeResults } from "./use-network-probe-data.ts"
import type { NetworkProbe, NetworkProbeResultsResponse } from "@/types"

describe("loadHistoricalProbeResults", () => {
	it("keeps successful probe history when another probe request fails", async () => {
		const assigned = [networkProbe("probe-ct"), networkProbe("probe-cu")]
		const result = await loadHistoricalProbeResults(assigned, "sys-1", "30m", (probeId) => {
			if (probeId === "probe-cu") {
				throw new Error("probe unavailable")
			}
			return {
				probeId,
				series: [{ systemId: "sys-1", created: "2026-07-01T12:00:00.000Z", success: true, latencyMs: 12 }],
			} satisfies NetworkProbeResultsResponse
		})

		assert.equal(result.hasSuccesses, true)
		assert.equal(result.hasFailures, true)
		assert.deepEqual(Object.keys(result.results), ["probe-ct"])
		assert.equal(result.results["probe-ct"].series[0].latencyMs, 12)
	})

	it("marks the historical load as failed only when every probe request fails", async () => {
		const assigned = [networkProbe("probe-ct"), networkProbe("probe-cu")]
		const result = await loadHistoricalProbeResults(assigned, "sys-1", "30m", () => {
			throw new Error("probe unavailable")
		})

		assert.equal(result.hasSuccesses, false)
		assert.equal(result.hasFailures, true)
		assert.deepEqual(result.results, {})
	})
})

describe("applyHistoricalProbeLoad", () => {
	it("keeps successful lines visible when some historical probe requests fail", async () => {
		const assigned = [networkProbe("probe-ct"), networkProbe("probe-cu")]
		const created = new Date().toISOString()
		const history = await loadHistoricalProbeResults(assigned, "sys-1", "30m", (probeId) => {
			if (probeId === "probe-cu") {
				throw new Error("probe unavailable")
			}
			return {
				probeId,
				series: [{ systemId: "sys-1", created, success: true, latencyMs: 12 }],
			} satisfies NetworkProbeResultsResponse
		})

		const state = applyHistoricalProbeLoad({}, history, "30m", assigned.length)

		assert.equal(state.error, false)
		assert.deepEqual(Object.keys(state.results), ["probe-ct"])
		assert.equal(state.results["probe-ct"].series[0].latencyMs, 12)
	})

	it("does not treat an empty assignment set as a failed historical load", () => {
		const state = applyHistoricalProbeLoad({}, { results: {}, hasSuccesses: false, hasFailures: false }, "30m", 0)

		assert.equal(state.error, false)
		assert.deepEqual(state.results, {})
	})

	it("drops stale lines that failed in the current historical refresh", async () => {
		const assigned = [networkProbe("probe-ct"), networkProbe("probe-cu")]
		const created = new Date().toISOString()
		const history = await loadHistoricalProbeResults(assigned, "sys-1", "24h", (probeId) => {
			if (probeId === "probe-cu") {
				throw new Error("probe unavailable")
			}
			return {
				probeId,
				series: [{ systemId: "sys-1", created, success: true, latencyMs: 18 }],
			} satisfies NetworkProbeResultsResponse
		})

		const current = {
			"probe-ct": {
				series: [{ systemId: "sys-1", created, success: true, latencyMs: 11 }],
			},
			"probe-cu": {
				series: [{ systemId: "sys-1", created, success: true, latencyMs: 22 }],
			},
		}

		const state = applyHistoricalProbeLoad(current, history, "24h", assigned.length)

		assert.equal(state.error, false)
		assert.deepEqual(Object.keys(state.results), ["probe-ct"])
		assert.equal(state.results["probe-ct"].series[0].latencyMs, 18)
	})
})

function networkProbe(id: string): NetworkProbe {
	return {
		id,
		name: id,
		type: "tcping",
		target: `${id}.example.com:443`,
		intervalSeconds: 20,
		timeoutSeconds: 5,
		enabled: true,
		publicVisible: true,
		scope: "fixed",
		systems: ["sys-1"],
	}
}
