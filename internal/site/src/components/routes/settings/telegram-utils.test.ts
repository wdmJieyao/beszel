import assert from "node:assert/strict"
import { readFileSync } from "node:fs"
import { describe, it } from "node:test"
import {
	TELEGRAM_ALERT_SCOPE_OPTIONS,
	buildTelegramBotTestPayload,
	buildTelegramDestinationPayload,
	buildTelegramPolicyPayload,
	buildTelegramSettingsPayload,
	clearTelegramSystems,
	clearTelegramPolicyEditorForDeletedDestination,
	defaultTelegramDestination,
	defaultTelegramPolicy,
	formatTelegramBotTestResult,
	getTelegramBotHealth,
	getTelegramDestinationHealth,
	getExistingTelegramDestinationID,
	maskTelegramChatID,
	removeTelegramDestinationByID,
	searchTelegramSystems,
	selectAllTelegramSystems,
} from "./telegram-utils.ts"

interface TelegramPolicyFixture {
	id: string
	name: string
	enabled: boolean
	nodeScopeMode: "all" | "selected"
	nodeScope: string[]
	alertLevelScope: string[]
}

function makeTelegramPolicyFixture(overrides: Partial<TelegramPolicyFixture> = {}): TelegramPolicyFixture {
	return {
		id: "policy-1",
		name: "默认规则",
		enabled: true,
		nodeScopeMode: "all",
		nodeScope: [],
		alertLevelScope: [],
		...overrides,
	}
}

function makeTelegramSystemFixtures(count: number) {
	return Array.from({ length: count }, (_, index) => ({
		id: `system-${index + 1}`,
		name: `node-${String(index + 1).padStart(3, "0")}`,
	}))
}

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

	it("accepts only supported alert scope values", () => {
		const alertLevelScope = TELEGRAM_ALERT_SCOPE_OPTIONS.map((option) => option.value)
		const payload = buildTelegramDestinationPayload({
			...defaultTelegramDestination(),
			name: "运维群",
			chatId: "-10012345",
			alertLevelScope,
		})

		assert.deepEqual(payload.alertLevelScope, alertLevelScope)
		assert.throws(() =>
			buildTelegramDestinationPayload({
				...defaultTelegramDestination(),
				name: "运维群",
				chatId: "-10012345",
				alertLevelScope: ["custom-alert"],
			})
		)
	})
})

describe("getTelegramBotHealth", () => {
	it("describes disabled, incomplete, ready, and error states", () => {
		assert.equal(getTelegramBotHealth({ enabled: false, hasToken: false }).status, "disabled")
		assert.equal(getTelegramBotHealth({ enabled: true, hasToken: false }).status, "pending")
		assert.equal(getTelegramBotHealth({ enabled: true, hasToken: true, botUsername: "beszel_bot" }).status, "healthy")
		const failed = getTelegramBotHealth({
			enabled: true,
			hasToken: true,
			lastError: "request https://api.telegram.org/bot123456789:secret_token/getMe failed",
		})
		assert.equal(failed.status, "error")
		assert.equal(failed.error, "request https://api.telegram.org/bot[redacted]/getMe failed")
	})
})

describe("getTelegramDestinationHealth", () => {
	it("formats delivery timestamps, mute state, and sanitized errors", () => {
		const health = getTelegramDestinationHealth(
			{
				...defaultTelegramDestination(),
				id: "destination-1",
				lastTestAt: "2026-07-10T01:02:03Z",
				lastDeliveryAt: "2026-07-10T02:03:04Z",
				muteUntil: "2026-07-10T04:00:00Z",
				lastError: "send failed\n123456789:secret_token chat not found",
			},
			new Date("2026-07-10T03:00:00Z"),
			"zh-CN"
		)

		assert.equal(health.status, "muted")
		assert.match(health.lastTestAt, /2026\/0?7\/10/)
		assert.match(health.lastDeliveryAt, /2026\/0?7\/10/)
		assert.match(health.muteUntil, /2026\/0?7\/10/)
		assert.equal(health.error, "send failed [redacted] chat not found")
	})
})

describe("Telegram Bot verification and deletion helpers", () => {
	it("tests an entered token without sending unrelated settings", () => {
		assert.deepEqual(
			buildTelegramBotTestPayload({ enabled: true, pollingEnabled: true, botToken: " 123456:token_value " }),
			{
				botToken: "123456:token_value",
			}
		)
		assert.deepEqual(buildTelegramBotTestPayload({ enabled: true, pollingEnabled: true, botToken: " " }), {})
	})

	it("formats credential and menu stages independently", () => {
		const result = formatTelegramBotTestResult({
			ok: false,
			botUsername: "beszel_bot",
			stages: {
				credentials: { ok: true, error: "" },
				commandMenu: { ok: false, error: "menu denied" },
			},
		})
		assert.deepEqual(result.credentials, { ok: true, error: "" })
		assert.deepEqual(result.commandMenu, { ok: false, error: "menu denied" })
	})

	it("masks chat IDs and removes only the confirmed destination", () => {
		assert.equal(maskTelegramChatID("-100123456789"), "-100••••6789")
		const destinations = [
			{ ...defaultTelegramDestination(), id: "one", lastError: "" },
			{ ...defaultTelegramDestination(), id: "two", lastError: "" },
		]
		assert.deepEqual(
			removeTelegramDestinationByID(destinations, "one").map((item) => item.id),
			["two"]
		)
	})

	it("clears policy editor state only when its channel is deleted", () => {
		const activeDestination = { ...defaultTelegramDestination(), id: "one", lastError: "" }
		const policy = makeTelegramPolicyFixture()
		const cleared = clearTelegramPolicyEditorForDeletedDestination(
			{
				destination: activeDestination,
				policies: [policy],
				draft: policy,
				search: "node",
			},
			"one"
		)
		assert.equal(cleared.destination, undefined)
		assert.deepEqual(cleared.policies, [])
		assert.deepEqual(cleared.draft, defaultTelegramPolicy())
		assert.equal(cleared.search, "")

		const retained = clearTelegramPolicyEditorForDeletedDestination(
			{ destination: activeDestination, policies: [policy], draft: policy, search: "node" },
			"two"
		)
		assert.equal(retained.destination?.id, "one")
		assert.equal(retained.policies.length, 1)
		assert.equal(retained.search, "node")
	})
})

describe("Telegram notification policy helpers", () => {
	it("normalizes all and selected policy payloads without ambiguous empty selection", () => {
		assert.deepEqual(
			buildTelegramPolicyPayload({
				name: " All nodes ",
				enabled: true,
				nodeScopeMode: "all",
				nodeScope: ["ignored"],
				alertLevelScope: ["status"],
			}),
			{
				name: "All nodes",
				enabled: true,
				nodeScopeMode: "all",
				nodeScope: [],
				alertLevelScope: ["status"],
			}
		)
		assert.throws(() =>
			buildTelegramPolicyPayload({
				name: "Selected",
				enabled: true,
				nodeScopeMode: "selected",
				nodeScope: [],
				alertLevelScope: [],
			})
		)
	})

	it("searches and bulk-selects 500 nodes", () => {
		const systems = makeTelegramSystemFixtures(500)
		assert.equal(searchTelegramSystems(systems, "node-49").length, 10)
		const selected = selectAllTelegramSystems(["system-1"], searchTelegramSystems(systems, "node-49"))
		assert.equal(selected.length, 11)
		assert.deepEqual(clearTelegramSystems(), [])
	})

	it("extracts the existing channel from a duplicate Chat ID conflict", () => {
		assert.equal(
			getExistingTelegramDestinationID({ data: { data: { existingDestinationId: "channel-1" } } }),
			"channel-1"
		)
		assert.equal(getExistingTelegramDestinationID(new Error("other")), "")
	})
})

describe("Telegram localized copy boundary", () => {
	it("keeps user-visible copy out of pure helpers so Lingui can extract it from the component", () => {
		const source = readFileSync(new URL("./telegram-utils.ts", import.meta.url), "utf8")
		assert.doesNotMatch(source, /[\u3400-\u9fff]/)
		assert.deepEqual(
			TELEGRAM_ALERT_SCOPE_OPTIONS.filter((item) => item.value.startsWith("loadavg")).map((item) => item.value),
			["loadavg1", "loadavg5", "loadavg15"]
		)
	})

	it("keeps role, chat capability, and load-average copy translated in Simplified Chinese", () => {
		const component = readFileSync(new URL("./telegram-destinations.tsx", import.meta.url), "utf8")
		const catalog = readFileSync(new URL("../../../locales/zh-CN/zh-CN.po", import.meta.url), "utf8")
		const expected = new Map([
			[
				"Admin channels receive full alert details within policy scope. Only private admins can use the management menu.",
				"管理员渠道接收策略范围内的完整告警详情；仅私聊管理员可使用管理菜单。",
			],
			[
				"Read-only channels receive sanitized monitoring messages within policy scope and cannot run management commands.",
				"只读渠道接收策略范围内的脱敏监控消息，不能执行管理命令。",
			],
			[
				"Groups, supergroups, and channels can receive full notifications but cannot use the management menu.",
				"群组、超群和频道可以接收完整通知，但不能使用管理菜单。",
			],
			["System load (1-minute average)", "系统负载（1 分钟均值）"],
			["System load (5-minute average)", "系统负载（5 分钟均值）"],
			["System load (15-minute average)", "系统负载（15 分钟均值）"],
		])
		for (const [message, translation] of expected) {
			assert.ok(component.includes(message), `missing Lingui source message: ${message}`)
			const block = catalog.split("\n\n").find((entry) => entry.includes(`msgid "${message}"`))
			assert.ok(block, `missing zh-CN catalog entry: ${message}`)
			assert.ok(block.includes(`msgstr "${translation}"`), `incorrect zh-CN translation: ${message}`)
		}
		assert.match(component, /policyDraft\.nodeScopeMode/)
		assert.match(component, /policyDraft\.alertLevelScope/)
	})
})
