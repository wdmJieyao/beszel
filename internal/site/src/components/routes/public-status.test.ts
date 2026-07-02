import assert from "node:assert/strict"
import { readFileSync } from "node:fs"
import { describe, it } from "node:test"

describe("public status route", () => {
	it("does not use live probe session APIs", () => {
		const source = readFileSync(new URL("./public-status.tsx", import.meta.url), "utf8")
		assert.equal(source.includes("createNetworkProbeLiveSession"), false)
		assert.equal(source.includes("renewNetworkProbeLiveSession"), false)
		assert.equal(source.includes("endNetworkProbeLiveSession"), false)
		assert.equal(source.includes("network-probe-live-cadence"), false)
		assert.equal(source.includes("network-probe-live-session"), false)
	})
})
