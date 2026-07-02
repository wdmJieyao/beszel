import { t } from "@lingui/core/macro"
import PocketBase from "pocketbase"
import { basePath } from "@/components/router"
import { toast } from "@/components/ui/use-toast"
import type {
	AdminPublicSystem,
	ChartTimes,
	NetworkProbe,
	NetworkProbeLiveSession,
	NetworkProbeInput,
	NetworkProbeResultsResponse,
	PublicChartRange,
	PublicStatusResponse,
	UserSettings,
} from "@/types"
import { $alerts, $allSystemsById, $allSystemsByName, $userSettings } from "./stores"
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

function toPublicSystemPayload(data: Partial<AdminPublicSystem>) {
	return {
		publicEnabled: data.publicEnabled,
		publicName: data.publicName,
		showCpu: data.showCpu,
		showMemory: data.showMemory,
		showDisk: data.showDisk,
	}
}

function toNetworkProbePayload(probe: Partial<NetworkProbeInput>) {
	return {
		name: probe.name,
		type: probe.type,
		target: probe.target,
		intervalSeconds: probe.intervalSeconds,
		timeoutSeconds: probe.timeoutSeconds,
		enabled: probe.enabled,
		publicVisible: probe.publicVisible,
		systems: probe.systems,
	}
}
