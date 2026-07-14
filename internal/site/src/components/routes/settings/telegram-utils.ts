import * as v from "valibot"
import type {
	TelegramDestination,
	TelegramDestinationInput,
	TelegramNotificationPolicy,
	TelegramNotificationPolicyInput,
	TelegramSettingsInput,
	TelegramTestResponse,
} from "@/types"

export const TELEGRAM_ALERT_SCOPE_OPTIONS = [
	{ value: "status" },
	{ value: "cpu" },
	{ value: "memory" },
	{ value: "disk" },
	{ value: "temperature" },
	{ value: "bandwidth" },
	{ value: "gpu" },
	{ value: "loadavg1" },
	{ value: "loadavg5" },
	{ value: "loadavg15" },
	{ value: "battery" },
	{ value: "smart" },
] as const

const TELEGRAM_ALERT_SCOPE_VALUES = TELEGRAM_ALERT_SCOPE_OPTIONS.map((option) => option.value) as [
	(typeof TELEGRAM_ALERT_SCOPE_OPTIONS)[number]["value"],
	...(typeof TELEGRAM_ALERT_SCOPE_OPTIONS)[number]["value"][],
]

export const TelegramSettingsSchema = v.object({
	enabled: v.boolean(),
	pollingEnabled: v.boolean(),
	botToken: v.optional(v.string()),
})

export const TelegramDestinationSchema = v.object({
	id: v.optional(v.string()),
	userId: v.optional(v.string()),
	name: v.pipe(v.string(), v.minLength(1)),
	chatId: v.pipe(v.string(), v.regex(/^-?\d+$/)),
	chatType: v.picklist(["private", "group", "supergroup", "channel", "unknown"]),
	role: v.picklist(["admin", "read_only"]),
	enabled: v.boolean(),
	nodeScope: v.array(v.string()),
	alertLevelScope: v.array(v.picklist(TELEGRAM_ALERT_SCOPE_VALUES)),
	muteUntil: v.optional(v.string()),
})

export const TelegramPolicySchema = v.object({
	name: v.pipe(v.string(), v.minLength(1)),
	enabled: v.boolean(),
	nodeScopeMode: v.picklist(["all", "selected"]),
	nodeScope: v.array(v.string()),
	alertLevelScope: v.array(v.picklist(TELEGRAM_ALERT_SCOPE_VALUES)),
})

export function buildTelegramSettingsPayload(input: TelegramSettingsInput) {
	return v.parse(TelegramSettingsSchema, {
		enabled: input.enabled,
		pollingEnabled: input.pollingEnabled,
		botToken: input.botToken?.trim() || undefined,
	})
}

export function buildTelegramBotTestPayload(input: TelegramSettingsInput) {
	const botToken = input.botToken?.trim()
	return botToken ? { botToken } : {}
}

export function formatTelegramBotTestResult(result: TelegramTestResponse) {
	const credentials = result.stages?.credentials
	const commandMenu = result.stages?.commandMenu
	return {
		credentials: {
			ok: credentials?.ok === true,
			error: sanitizeTelegramTroubleshootingError(credentials?.error || result.error),
		},
		commandMenu: {
			ok: commandMenu?.ok === true,
			error: sanitizeTelegramTroubleshootingError(commandMenu?.error || result.error),
		},
	}
}

export function maskTelegramChatID(chatID: string) {
	const prefixLength = chatID.startsWith("-100") ? 4 : Math.min(2, Math.max(0, chatID.length - 4))
	if (chatID.length <= prefixLength + 4) return chatID
	return `${chatID.slice(0, prefixLength)}••••${chatID.slice(-4)}`
}

export function removeTelegramDestinationByID(destinations: TelegramDestination[], destinationID: string) {
	return destinations.filter((destination) => destination.id !== destinationID)
}

export interface TelegramPolicyEditorState {
	destination?: TelegramDestination
	policies: TelegramNotificationPolicy[]
	draft: TelegramNotificationPolicyInput & { id?: string }
	search: string
}

export function clearTelegramPolicyEditorForDeletedDestination(
	state: TelegramPolicyEditorState,
	destinationID: string
): TelegramPolicyEditorState {
	if (state.destination?.id !== destinationID) return state
	return {
		destination: undefined,
		policies: [],
		draft: defaultTelegramPolicy(),
		search: "",
	}
}

export function getExistingTelegramDestinationID(error: unknown) {
	if (!error || typeof error !== "object") return ""
	const value = error as {
		data?: { data?: { existingDestinationId?: unknown }; existingDestinationId?: unknown }
		response?: { data?: { existingDestinationId?: unknown } }
	}
	const candidate =
		value.data?.data?.existingDestinationId ??
		value.data?.existingDestinationId ??
		value.response?.data?.existingDestinationId
	return typeof candidate === "string" ? candidate : ""
}

export function buildTelegramDestinationPayload(input: Partial<TelegramDestinationInput>) {
	return v.parse(TelegramDestinationSchema, {
		id: input.id,
		userId: input.userId,
		name: input.name?.trim() ?? "",
		chatId: input.chatId?.trim() ?? "",
		chatType: input.chatType ?? "private",
		role: input.role ?? "admin",
		enabled: input.enabled ?? true,
		nodeScope: input.nodeScope ?? [],
		alertLevelScope: input.alertLevelScope ?? [],
		muteUntil: input.muteUntil || undefined,
	})
}

export function buildTelegramChannelPayload(input: Partial<TelegramDestinationInput>) {
	const destination = buildTelegramDestinationPayload(input)
	return {
		id: destination.id,
		userId: destination.userId,
		name: destination.name,
		chatId: destination.chatId,
		chatType: destination.chatType,
		role: destination.role,
		enabled: destination.enabled,
		muteUntil: destination.muteUntil,
	}
}

export function buildTelegramPolicyPayload(input: TelegramNotificationPolicyInput) {
	const payload = v.parse(TelegramPolicySchema, {
		...input,
		name: input.name.trim(),
		nodeScope: input.nodeScopeMode === "all" ? [] : [...new Set(input.nodeScope)],
		alertLevelScope: [...new Set(input.alertLevelScope)],
	})
	if (payload.nodeScopeMode === "selected" && payload.nodeScope.length === 0) {
		throw new Error("selected_node_required")
	}
	return payload
}

export function defaultTelegramPolicy(): TelegramNotificationPolicyInput {
	return {
		name: "",
		enabled: true,
		nodeScopeMode: "all",
		nodeScope: [],
		alertLevelScope: [],
	}
}

type TelegramSystemOption = { id: string; name: string }

export function searchTelegramSystems<T extends TelegramSystemOption>(systems: T[], query: string): T[] {
	const normalized = query.trim().toLocaleLowerCase()
	if (!normalized) return systems
	return systems.filter(
		(system) =>
			system.name.toLocaleLowerCase().includes(normalized) || system.id.toLocaleLowerCase().includes(normalized)
	)
}

export function selectAllTelegramSystems(selected: string[], visibleSystems: TelegramSystemOption[]) {
	return [...new Set([...selected, ...visibleSystems.map((system) => system.id)])]
}

export function clearTelegramSystems() {
	return [] as string[]
}

export type TelegramHealthStatus = "disabled" | "pending" | "healthy" | "muted" | "error"

export interface TelegramDeliveryHealth {
	status: TelegramHealthStatus
	error: string
	lastTestAt: string
	lastDeliveryAt: string
	muteUntil: string
}

export function sanitizeTelegramTroubleshootingError(message?: string) {
	return (message ?? "")
		.replace(/bot\d+:[A-Za-z0-9_-]+/gi, "bot[redacted]")
		.replace(/\b\d{6,}:[A-Za-z0-9_-]{6,}\b/g, "[redacted]")
		.replace(/\s+/g, " ")
		.trim()
}

function formatTelegramTimestamp(timestamp: string | undefined, locale?: string) {
	if (!timestamp) return ""
	const date = new Date(timestamp)
	if (Number.isNaN(date.getTime())) return ""
	return new Intl.DateTimeFormat(locale, { dateStyle: "short", timeStyle: "medium" }).format(date)
}

export function getTelegramBotHealth(settings: {
	enabled: boolean
	hasToken?: boolean
	botToken?: string
	botUsername?: string
	lastError?: string
}): Pick<TelegramDeliveryHealth, "status" | "error"> {
	const error = sanitizeTelegramTroubleshootingError(settings.lastError)
	if (!settings.enabled) return { status: "disabled", error }
	if (error) return { status: "error", error }
	if (!(settings.hasToken || settings.botToken?.trim())) {
		return { status: "pending", error: "" }
	}
	return { status: "healthy", error: "" }
}

export function getTelegramDestinationHealth(
	destination: TelegramDestination,
	now = new Date(),
	locale?: string
): TelegramDeliveryHealth {
	const error = sanitizeTelegramTroubleshootingError(destination.lastError)
	const muteUntilDate = destination.muteUntil ? new Date(destination.muteUntil) : undefined
	const isMuted = muteUntilDate && !Number.isNaN(muteUntilDate.getTime()) && muteUntilDate > now
	let status: TelegramHealthStatus = "pending"
	if (!destination.enabled) {
		status = "disabled"
	} else if (isMuted) {
		status = "muted"
	} else if (error) {
		status = "error"
	} else if (destination.lastTestAt || destination.lastDeliveryAt) {
		status = "healthy"
	}
	return {
		status,
		error,
		lastTestAt: formatTelegramTimestamp(destination.lastTestAt, locale),
		lastDeliveryAt: formatTelegramTimestamp(destination.lastDeliveryAt, locale),
		muteUntil: formatTelegramTimestamp(destination.muteUntil, locale),
	}
}

export function defaultTelegramSettings(): TelegramSettingsInput {
	return {
		enabled: false,
		pollingEnabled: false,
		botToken: "",
	}
}

export function defaultTelegramDestination(): TelegramDestinationInput {
	return {
		name: "",
		chatId: "",
		chatType: "private",
		role: "admin",
		enabled: true,
		nodeScope: [],
		alertLevelScope: [],
	}
}
