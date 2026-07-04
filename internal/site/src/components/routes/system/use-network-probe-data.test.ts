import { describe, it } from "node:test"
import assert from "node:assert/strict"
import { groupNetworkProbeData, type ProbeDefinition, type ProbeResultsByProbeId } from "./network-probe-groups.ts"
import { filterAssignedProbes } from "./network-probe-filter.ts"
import { applyLiveProbeResultEvent, initialLiveProbeResults } from "./network-probe-live-session.ts"
import type { NetworkProbe } from "@/types"

const systemId = "sys-1"

const probes: ProbeDefinition[] = [
	{
		id: "probe-ct",
		name: "广东电信",
		type: "tcping",
		target: "ct.example.com",
		systems: [systemId],
	},
	{
		id: "probe-cu",
		name: "广东联通",
		type: "tcping",
		target: "cu.example.com",
		systems: [systemId],
	},
	{
		id: "probe-cm",
		name: "广东移动",
		type: "icmp_ping",
		target: "cm.example.com",
		systems: [systemId],
	},
]

describe("groupNetworkProbeData", () => {
	it("preserves configured latency series with zero live points", () => {
		const groups = groupNetworkProbeData(probes, {}, systemId)
		assert.equal(groups.length, 1)
		assert.equal(groups[0].label, "线路检测")
		assert.deepEqual(
			groups[0].series.map((series) => series.label),
			["广东电信", "广东联通", "广东移动"]
		)
		assert.deepEqual(
			groups[0].series.map((series) => series.points.length),
			[0, 0, 0]
		)
	})

	it("keeps missing-line legend entries when realtime arrivals are staggered", () => {
		const results: ProbeResultsByProbeId = {
			"probe-ct": {
				series: [
					{
						systemId,
						created: "2026-07-01T12:01:00.000Z",
						success: true,
						latencyMs: 8,
						type: "tcping",
					},
				],
			},
			"probe-cu": {
				series: [
					{
						systemId,
						created: "2026-07-01T12:01:05.000Z",
						success: true,
						latencyMs: 15,
						type: "tcping",
					},
				],
			},
		}
		const groups = groupNetworkProbeData(probes, results, systemId)
		assert.deepEqual(
			groups[0].series.map((series) => [series.label, series.points.length]),
			[
				["广东电信", 1],
				["广东联通", 1],
				["广东移动", 0],
			]
		)
	})
})

describe("useNetworkProbeData live result semantics", () => {
	it("starts the 1m live window empty", () => {
		assert.deepEqual(initialLiveProbeResults(), {})
	})

	it("appends realtime events for the active system without backfilling history", () => {
		const next = applyLiveProbeResultEvent(initialLiveProbeResults(), {
			action: "create",
			record: {
				probe: "probe-ct",
				system: systemId,
				created: "2026-07-01T12:01:10.000Z",
				success: true,
				latency_ms: 8,
				type: "tcping",
				target: "ct.example.com:443",
			},
		})
		assert.deepEqual(Object.keys(next), ["probe-ct"])
		assert.equal(next["probe-ct"].series.length, 1)
		assert.equal(next["probe-ct"].series[0].latencyMs, 8)
	})
})

describe("filterAssignedProbes", () => {
	it("includes global probes for the active system without assignment rows", () => {
		const activeSystem = "sys-new"
		const assigned = filterAssignedProbes(
			[
				networkProbe({ id: "global", scope: "global", systems: [] }),
				networkProbe({ id: "fixed-other", scope: "fixed", systems: ["sys-other"] }),
				networkProbe({ id: "fixed-active", scope: "fixed", systems: [activeSystem] }),
			],
			activeSystem
		)

		assert.deepEqual(
			assigned.map((probe) => probe.id),
			["global", "fixed-active"]
		)
	})
})

function networkProbe(overrides: Partial<NetworkProbe>): NetworkProbe {
	return {
		id: "probe",
		name: "线路",
		type: "tcping",
		target: "example.com:443",
		intervalSeconds: 20,
		timeoutSeconds: 5,
		enabled: true,
		publicVisible: true,
		scope: "fixed",
		systems: [],
		...overrides,
	}
}
