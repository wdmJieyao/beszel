import assert from "node:assert/strict"
import { describe, it } from "node:test"
import type { AdminPublicSystem, NetworkProbe } from "@/types"
import {
	availablePublicProbeIds,
	normalizePublicProbeSelection,
	toPublicSystemPayload,
	togglePublicProbeSelection,
} from "./public-status-utils.ts"

describe("toPublicSystemPayload", () => {
	it("includes public probe selections when saving public systems", () => {
		const payload = toPublicSystemPayload(publicSystem({ publicProbeIds: ["probe-a", "probe-b"] }))

		assert.deepEqual(payload.publicProbeIds, ["probe-a", "probe-b"])
		assert.equal(payload.publicEnabled, true)
	})
})

describe("public probe availability", () => {
	it("defaults newly public systems to an empty selected probe list", () => {
		const payload = toPublicSystemPayload(publicSystem({ publicProbeIds: [] }))

		assert.deepEqual(payload.publicProbeIds, [])
	})

	it("returns only probes that cover the target system", () => {
		const probeIds = availablePublicProbeIds("sys-1", [
			networkProbe({ id: "global-1", scope: "global", systems: [] }),
			networkProbe({ id: "fixed-1", scope: "fixed", systems: ["sys-1"] }),
			networkProbe({ id: "fixed-2", scope: "fixed", systems: ["sys-2"] }),
		])

		assert.deepEqual(probeIds, ["global-1", "fixed-1"])
	})

	it("supports zero, one, and many selected probes plus select-all behavior", () => {
		assert.deepEqual(normalizePublicProbeSelection([]), [])
		assert.deepEqual(normalizePublicProbeSelection(["probe-a"]), ["probe-a"])
		assert.deepEqual(normalizePublicProbeSelection(["probe-a", "probe-b", "probe-a"]), ["probe-a", "probe-b"])

		let selected = togglePublicProbeSelection([], "probe-a", true)
		assert.deepEqual(selected, ["probe-a"])

		selected = togglePublicProbeSelection(selected, "probe-b", true)
		assert.deepEqual(selected, ["probe-a", "probe-b"])

		selected = togglePublicProbeSelection(selected, "probe-a", false)
		assert.deepEqual(selected, ["probe-b"])

		const all = availablePublicProbeIds("sys-1", [
			networkProbe({ id: "probe-a", scope: "global", systems: [] }),
			networkProbe({ id: "probe-b", scope: "fixed", systems: ["sys-1"] }),
		])
		assert.deepEqual(all, ["probe-a", "probe-b"])
	})
})

function publicSystem(overrides: Partial<AdminPublicSystem>): AdminPublicSystem {
	return {
		id: "sys-1",
		name: "节点一",
		status: "up",
		publicEnabled: true,
		publicName: "公开节点一",
		showCpu: true,
		showMemory: true,
		showDisk: true,
		publicProbeIds: [],
		...overrides,
	}
}

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
