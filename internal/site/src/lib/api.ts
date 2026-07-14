import { t } from "@lingui/core/macro"
import PocketBase from "pocketbase"
import { basePath } from "@/components/router"
import { toast } from "@/components/ui/use-toast"
import type {
	AdminPublicSystem,
	ChartTimes,
	ConfigBackupExportRequest,
	ConfigBackupExportResponse,
	ConfigBackupPreviewResponse,
	ConfigBackupRestoreRequest,
	ConfigBackupRestoreResponse,
	ConfigBackupValidationRequest,
	NetworkProbe,
	NetworkProbeLiveSession,
	NetworkProbeInput,
	NetworkProbeResultsResponse,
	PublicChartRange,
	PublicStatusResponse,
	TelegramDestination,
	TelegramDestinationInput,
	TelegramNotificationPolicy,
	TelegramNotificationPolicyInput,
	TelegramPolicyListResponse,
	TelegramSettings,
	TelegramSettingsInput,
	TelegramTestResponse,
	UserSettings,
} from "@/types"
import { $alerts, $allSystemsById, $allSystemsByName, $userSettings } from "./stores"
import { toPublicSystemPayload } from "@/components/routes/settings/public-status-utils"
import { chartTimeData } from "./utils"

/** PocketBase JS Client */
export const pb = new PocketBase(basePath)

export const isAdmin = () => pb.authStore.record?.role === "admin"
export const isReadOnlyUser = () => pb.authStore.record?.role === "readonly"

export const verifyAuth = () => {
	pb.collection("users")
		.authRefresh()
		.catch(() => {
			logOut()
			toast({
				title: t`Failed to authenticate`,
				description: t`Please log in again`,
				variant: "destructive",
			})
		})
}

/** Logs the user out by clearing the auth store and unsubscribing from realtime updates. */
export function logOut() {
	$allSystemsByName.set({})
	$allSystemsById.set({})
	$alerts.set({})
	$userSettings.set({} as UserSettings)
	sessionStorage.setItem("lo", "t") // prevent auto login on logout
	pb.authStore.clear()
	pb.realtime.unsubscribe()
}

/** Fetch or create user settings in database */
export async function updateUserSettings() {
	try {
		const req = await pb.collection("user_settings").getFirstListItem("", { fields: "settings" })
		$userSettings.set(req.settings)
		return
	} catch (e) {
		console.error("get settings", e)
	}
	// create user settings if error fetching existing
	try {
		const createdSettings = await pb.collection("user_settings").create({ user: pb.authStore.record?.id })
		$userSettings.set(createdSettings.settings)
	} catch (e) {
		console.error("create settings", e)
	}
}

export function getPbTimestamp(timeString: ChartTimes, d?: Date) {
	d ||= chartTimeData[timeString].getOffset(new Date())
	const year = d.getUTCFullYear()
	const month = String(d.getUTCMonth() + 1).padStart(2, "0")
	const day = String(d.getUTCDate()).padStart(2, "0")
	const hours = String(d.getUTCHours()).padStart(2, "0")
	const minutes = String(d.getUTCMinutes()).padStart(2, "0")
	const seconds = String(d.getUTCSeconds()).padStart(2, "0")

	return `${year}-${month}-${day} ${hours}:${minutes}:${seconds}`
}

export function getPublicStatus(range?: PublicChartRange) {
	const query = range ? { range } : undefined
	return pb.send<PublicStatusResponse>("/api/beszel/public/status", { query })
}

export function getPublicSystems() {
	return pb.send<{ systems: AdminPublicSystem[] }>("/api/beszel/public/systems", {})
}

export function updatePublicSystem(systemId: string, data: Partial<AdminPublicSystem>) {
	return pb.send<AdminPublicSystem>(`/api/beszel/public/systems/${systemId}`, {
		method: "PATCH",
		body: toPublicSystemPayload(data),
	})
}

export function getNetworkProbes() {
	return pb.send<{ probes: NetworkProbe[] }>("/api/beszel/network-probes", {})
}

export function saveNetworkProbe(probe: Partial<NetworkProbeInput>) {
	const method = probe.id ? "PATCH" : "POST"
	const path = probe.id ? `/api/beszel/network-probes/${probe.id}` : "/api/beszel/network-probes"
	return pb.send<NetworkProbe>(path, { method, body: toNetworkProbePayload(probe) })
}

export function deleteNetworkProbe(probeId: string) {
	return pb.send(`/api/beszel/network-probes/${probeId}`, { method: "DELETE" })
}

export function getNetworkProbeResults(
	probeId: string,
	params: { system?: string; range: ChartTimes | PublicChartRange }
) {
	return pb.send<NetworkProbeResultsResponse>(`/api/beszel/network-probes/${probeId}/results`, {
		query: params,
	})
}

export function createNetworkProbeLiveSession(systemId: string) {
	return pb.send<NetworkProbeLiveSession>(`/api/beszel/systems/${systemId}/network-probe-live-sessions`, {
		method: "POST",
		body: { range: "1m" },
	})
}

export function renewNetworkProbeLiveSession(systemId: string, sessionId: string) {
	return pb.send<NetworkProbeLiveSession>(`/api/beszel/systems/${systemId}/network-probe-live-sessions/${sessionId}`, {
		method: "PATCH",
		body: { range: "1m" },
	})
}

export function endNetworkProbeLiveSession(systemId: string, sessionId: string) {
	return pb.send(`/api/beszel/systems/${systemId}/network-probe-live-sessions/${sessionId}`, {
		method: "DELETE",
	})
}

export function getTelegramSettings() {
	return pb.send<TelegramSettings>("/api/beszel/telegram/settings", {})
}

export function saveTelegramSettings(settings: TelegramSettingsInput) {
	return pb.send<TelegramSettings>("/api/beszel/telegram/settings", {
		method: "PUT",
		body: settings,
	})
}

export function testTelegramSettings(settings?: { botToken?: string }) {
	return pb.send<TelegramTestResponse>("/api/beszel/telegram/settings/test", {
		method: "POST",
		body: settings ?? {},
	})
}

export function getTelegramDestinations() {
	return pb.send<{ destinations: TelegramDestination[] }>("/api/beszel/telegram/destinations", {})
}

export function saveTelegramDestination(destination: Partial<TelegramDestinationInput>) {
	const method = destination.id ? "PATCH" : "POST"
	const path = destination.id
		? `/api/beszel/telegram/destinations/${destination.id}`
		: "/api/beszel/telegram/destinations"
	return pb.send<TelegramDestination>(path, {
		method,
		body: destination,
	})
}

export function deleteTelegramDestination(destinationId: string) {
	return pb.send(`/api/beszel/telegram/destinations/${destinationId}`, {
		method: "DELETE",
	})
}

export function testTelegramDestination(destinationId: string) {
	return pb.send<TelegramTestResponse & { sentAt?: string }>(
		`/api/beszel/telegram/destinations/${destinationId}/test`,
		{
			method: "POST",
		}
	)
}

export function getTelegramNotificationPolicies(destinationId: string) {
	return pb.send<TelegramPolicyListResponse>(`/api/beszel/telegram/destinations/${destinationId}/policies`, {})
}

export function saveTelegramNotificationPolicy(
	destinationId: string,
	policy: TelegramNotificationPolicyInput & { id?: string }
) {
	const path = policy.id
		? `/api/beszel/telegram/destinations/${destinationId}/policies/${policy.id}`
		: `/api/beszel/telegram/destinations/${destinationId}/policies`
	return pb.send<TelegramNotificationPolicy>(path, {
		method: policy.id ? "PATCH" : "POST",
		body: policy,
	})
}

export function deleteTelegramNotificationPolicy(destinationId: string, policyId: string) {
	return pb.send(`/api/beszel/telegram/destinations/${destinationId}/policies/${policyId}`, { method: "DELETE" })
}

export function exportConfigBackup(request: ConfigBackupExportRequest) {
	return pb.send<ConfigBackupExportResponse>("/api/beszel/config-backups/exports", {
		method: "POST",
		body: request,
	})
}

export function validateConfigBackup(request: ConfigBackupValidationRequest) {
	return pb.send<ConfigBackupPreviewResponse>("/api/beszel/config-backups/validations", {
		method: "POST",
		body: request,
	})
}

export function restoreConfigBackup(request: ConfigBackupRestoreRequest) {
	return pb.send<ConfigBackupRestoreResponse>("/api/beszel/config-backups/restores", {
		method: "POST",
		body: request,
	})
}

function toNetworkProbePayload(probe: Partial<NetworkProbeInput>) {
	return {
		name: probe.name,
		type: probe.type,
		target: probe.target,
		intervalSeconds: probe.intervalSeconds,
		timeoutSeconds: probe.timeoutSeconds,
		enabled: probe.enabled,
		scope: probe.scope,
		systems: probe.systems,
	}
}
