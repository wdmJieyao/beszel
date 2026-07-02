import assert from "node:assert/strict"
import { describe, it } from "node:test"
import { appendData } from "./append-data.ts"

describe("appendData", () => {
	it("does not insert realtime chart gaps for normal 1m metric cadence jitter", () => {
		const points = appendData([{ created: 1_000, stats: { c: 1 } }], [{ created: 3_100, stats: { c: 2 } }], 2_000, 60)

		assert.equal(points.length, 2)
		assert.equal(points[0].created, 1_000)
		assert.equal(points[1].created, 3_100)
	})

	it("keeps explicit gaps when metric cadence is clearly missed", () => {
		const points = appendData([{ created: 1_000, stats: { c: 1 } }], [{ created: 5_000, stats: { c: 2 } }], 2_000, 60)

		assert.equal(points.length, 3)
		assert.equal(points[1].created, null)
	})
})
