import { describe, it } from "node:test"
import assert from "node:assert/strict"
import {
	applyLiveProbeResultEvent,
	initialLiveProbeResults,
	shouldUseLiveProbeSession,
	type ProbeResultsByProbeId,
	type NetworkProbeRealtimeEvent,
} from "./network-probe-live-session.ts"

const systemId = "sys-1"

const history: ProbeResultsByProbeId = {
	"probe-a": {
		series: [
			{
				systemId,
				created: "2026-07-01T12:00:00.000Z",
				success: true,
				latencyMs: 8,
				type: "tcping",
				target: "a.example.com",
			},
		],
	},
}

function realtimeEvent(
	probeId: string,
	created: string,
	overrides: Partial<Record<string, unknown>> = {}
): NetworkProbeRealtimeEvent {
	return {
		action: "create",
		record: {
			id: `${probeId}-${created}`,
			probe: probeId,
			system: systemId,
			created,
			success: true,
			latency_ms: 12,
			packet_loss_percent: undefined,
			http_status: undefined,
			error: "",
			failure_category: "",
			type: "tcping",
			target: `${probeId}.example.com`,
			...overrides,
		},
	}
}

describe("network probe live session", () => {
	it("uses live session semantics only for the 1m range", () => {
		assert.equal(shouldUseLiveProbeSession("1m"), true)
		assert.equal(shouldUseLiveProbeSession("30m"), false)
		assert.equal(shouldUseLiveProbeSession("1h"), false)
	})

	it("starts empty and does not retain historical probe results", () => {
		const live = initialLiveProbeResults()
		assert.deepEqual(live, {})
		assert.notDeepEqual(live, history)
	})

	it("appends realtime events received by the active browser session", () => {
		const next = applyLiveProbeResultEvent(
			initialLiveProbeResults(),
			realtimeEvent("probe-a", "2026-07-01T12:01:00.000Z")
		)
		assert.equal(next["probe-a"].series.length, 1)
		assert.equal(next["probe-a"].series[0].latencyMs, 12)
	})

	it("replaces duplicate realtime points for the same probe and created time", () => {
		const first = applyLiveProbeResultEvent(
			initialLiveProbeResults(),
			realtimeEvent("probe-a", "2026-07-01T12:01:00.000Z", { latency_ms: 12 })
		)
		const second = applyLiveProbeResultEvent(
			first,
			realtimeEvent("probe-a", "2026-07-01T12:01:00.000Z", { latency_ms: 18 })
		)
		assert.equal(second["probe-a"].series.length, 1)
		assert.equal(second["probe-a"].series[0].latencyMs, 18)
	})

	it("preserves failed realtime points without drawing stale successful history", () => {
		const next = applyLiveProbeResultEvent(
			initialLiveProbeResults(),
			realtimeEvent("probe-b", "2026-07-01T12:01:10.000Z", {
				success: false,
				latency_ms: undefined,
				error: "timeout",
				failure_category: "timeout",
			})
		)
		assert.equal(next["probe-b"].series[0].success, false)
		assert.equal(next["probe-b"].series[0].latencyMs, undefined)
		assert.equal(next["probe-b"].series[0].failureCategory, "timeout")
	})

	it("clears previous live points by creating a new empty live session", () => {
		const firstSession = applyLiveProbeResultEvent(
			initialLiveProbeResults(),
			realtimeEvent("probe-a", "2026-07-01T12:01:00.000Z")
		)
		assert.equal(firstSession["probe-a"].series.length, 1)
		assert.deepEqual(initialLiveProbeResults(), {})
	})
})
