import * as v from "valibot"
import type { TelegramDestinationInput, TelegramSettingsInput } from "@/types"

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
	alertLevelScope: v.array(v.string()),
	muteUntil: v.optional(v.string()),
})

export function buildTelegramSettingsPayload(input: TelegramSettingsInput) {
	return v.parse(TelegramSettingsSchema, {
		enabled: input.enabled,
		pollingEnabled: input.pollingEnabled,
		botToken: input.botToken?.trim() || undefined,
	})
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
