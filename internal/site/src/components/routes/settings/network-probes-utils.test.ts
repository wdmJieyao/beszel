import assert from "node:assert/strict"
import { describe, it } from "node:test"
import type { NetworkProbe } from "@/types"
import { buildNetworkProbePayload, probeScopeLabel } from "./network-probes-utils.ts"

describe("buildNetworkProbePayload", () => {
	it("sends global scope with an empty fixed systems list for all-node probes", () => {
		const payload = buildNetworkProbePayload(networkProbe({ systems: [] }))

		assert.equal(payload.scope, "global")
		assert.deepEqual(payload.systems, [])
	})

	it("sends fixed scope with selected systems for fixed-node probes", () => {
		const payload = buildNetworkProbePayload(networkProbe({ systems: ["sys-1"] }))

		assert.equal(payload.scope, "fixed")
		assert.deepEqual(payload.systems, ["sys-1"])
	})
})

describe("probeScopeLabel", () => {
	it("labels global and fixed probes for administrators", () => {
		assert.equal(probeScopeLabel(networkProbe({ scope: "global", systems: [] }), []), "全部可用节点")
		assert.equal(
			probeScopeLabel(networkProbe({ scope: "fixed", systems: ["sys-1"] }), [{ id: "sys-1", name: "节点一" }]),
			"固定节点: 节点一"
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
		scope: "global",
		systems: [],
		...overrides,
	}
}
