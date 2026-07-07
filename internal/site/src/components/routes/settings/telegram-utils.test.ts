import assert from "node:assert/strict"
import { describe, it } from "node:test"
import {
	buildTelegramDestinationPayload,
	buildTelegramSettingsPayload,
	defaultTelegramDestination,
} from "./telegram-utils.ts"

describe("buildTelegramSettingsPayload", () => {
	it("trims optional bot token and preserves toggles", () => {
		const payload = buildTelegramSettingsPayload({
			enabled: true,
			pollingEnabled: true,
			botToken: " 123456:abcde_token_valid ",
		})

		assert.equal(payload.enabled, true)
		assert.equal(payload.pollingEnabled, true)
		assert.equal(payload.botToken, "123456:abcde_token_valid")
	})
})

describe("buildTelegramDestinationPayload", () => {
	it("normalizes telegram destination payloads", () => {
		const payload = buildTelegramDestinationPayload({
			...defaultTelegramDestination(),
			name: " 运维群 ",
			chatId: " -10012345 ",
			alertLevelScope: ["status"],
		})

		assert.equal(payload.name, "运维群")
		assert.equal(payload.chatId, "-10012345")
		assert.deepEqual(payload.alertLevelScope, ["status"])
		assert.equal(payload.role, "admin")
	})
})
