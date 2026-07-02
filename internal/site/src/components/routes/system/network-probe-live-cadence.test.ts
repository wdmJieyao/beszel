import assert from "node:assert/strict"
import { describe, it } from "node:test"
import {
	getNetworkProbeRenderingMode,
	isLiveProbeSessionExpired,
	isHistoricalNetworkProbeRange,
	liveProbeSessionBody,
	liveProbeSessionHeartbeatMs,
	needsNewLiveProbeSession,
	networkProbeChartRanges,
	shouldUseLiveProbeCadence,
} from "./network-probe-live-cadence.ts"

describe("network probe live cadence", () => {
	it("uses cadence only for 1m range", () => {
		assert.equal(shouldUseLiveProbeCadence("1m"), true)
		assert.equal(shouldUseLiveProbeCadence("30m"), false)
		assert.equal(shouldUseLiveProbeCadence("1h"), false)
		assert.equal(shouldUseLiveProbeCadence("12h"), false)
		assert.equal(shouldUseLiveProbeCadence("24h"), false)
		assert.equal(shouldUseLiveProbeCadence("1w"), false)
		assert.equal(shouldUseLiveProbeCadence("30d"), false)
	})

	it("maps ranges to live or historical rendering modes", () => {
		assert.deepEqual([...networkProbeChartRanges], ["1m", "30m", "1h", "12h", "24h", "1w", "30d"])
		assert.equal(getNetworkProbeRenderingMode("1m"), "live-realtime")
		for (const range of ["30m", "1h", "12h", "24h", "1w", "30d"] as const) {
			assert.equal(getNetworkProbeRenderingMode(range), "historical-range")
			assert.equal(isHistoricalNetworkProbeRange(range), true)
		}
	})

	it("provides a stable 1m live session body and heartbeat interval", () => {
		assert.deepEqual(liveProbeSessionBody(), { range: "1m" })
		assert.equal(liveProbeSessionHeartbeatMs, 5_000)
	})

	it("detects when a new live session is needed", () => {
		const now = Date.parse("2026-07-02T08:00:00.000Z")
		assert.equal(needsNewLiveProbeSession(undefined, "sys-1", "1m", now), true)
		assert.equal(
			needsNewLiveProbeSession(
				{
					sessionId: "session-1",
					systemId: "sys-1",
					range: "1m",
					cadenceSeconds: 1,
					expiresAt: "2026-07-02T08:00:10.000Z",
				},
				"sys-1",
				"1m",
				now
			),
			false
		)
		assert.equal(
			needsNewLiveProbeSession(
				{
					sessionId: "session-1",
					systemId: "sys-1",
					range: "1m",
					cadenceSeconds: 1,
					expiresAt: "2026-07-02T07:59:59.000Z",
				},
				"sys-1",
				"1m",
				now
			),
			true
		)
	})

	it("treats invalid or past expiry as expired", () => {
		const now = Date.parse("2026-07-02T08:00:00.000Z")
		assert.equal(isLiveProbeSessionExpired("invalid-date", now), true)
		assert.equal(isLiveProbeSessionExpired("2026-07-02T08:00:01.000Z", now), false)
		assert.equal(isLiveProbeSessionExpired("2026-07-02T07:59:59.000Z", now), true)
	})
})
