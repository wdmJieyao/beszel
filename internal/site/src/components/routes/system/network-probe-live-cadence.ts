import type { NetworkProbeLiveSession, PublicChartRange } from "@/types"

export const liveProbeSessionHeartbeatMs = 5_000
export const networkProbeChartRanges = ["1m", "30m", "1h", "12h", "24h", "1w", "30d"] as const

export type NetworkProbeRenderingMode = "live-realtime" | "historical-range"

export function shouldUseLiveProbeCadence(range: PublicChartRange) {
	return range === "1m"
}

export function getNetworkProbeRenderingMode(range: PublicChartRange): NetworkProbeRenderingMode {
	return shouldUseLiveProbeCadence(range) ? "live-realtime" : "historical-range"
}

export function isHistoricalNetworkProbeRange(range: PublicChartRange) {
	return getNetworkProbeRenderingMode(range) === "historical-range"
}

export function liveProbeSessionBody() {
	return { range: "1m" as const }
}

export function isLiveProbeSessionExpired(expiresAt: string, nowMs = Date.now()) {
	const expiresMs = new Date(expiresAt).getTime()
	return !Number.isFinite(expiresMs) || expiresMs <= nowMs
}

export function needsNewLiveProbeSession(
	current: NetworkProbeLiveSession | undefined,
	systemId: string,
	range: PublicChartRange,
	nowMs = Date.now()
) {
	if (!shouldUseLiveProbeCadence(range)) {
		return false
	}
	if (!current) {
		return true
	}
	if (current.systemId !== systemId || current.range !== "1m") {
		return true
	}
	return isLiveProbeSessionExpired(current.expiresAt, nowMs)
}
