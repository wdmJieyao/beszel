import type { AdminPublicSystem, NetworkProbe } from "@/types"

export function toPublicSystemPayload(data: Partial<AdminPublicSystem>) {
	return {
		publicEnabled: data.publicEnabled,
		publicName: data.publicName,
		showCpu: data.showCpu,
		showMemory: data.showMemory,
		showDisk: data.showDisk,
		publicProbeIds: data.publicProbeIds,
	}
}

export function availablePublicProbeIds(systemId: string, probes: Pick<NetworkProbe, "id" | "scope" | "systems">[]) {
	return probes.filter((probe) => probe.scope === "global" || probe.systems.includes(systemId)).map((probe) => probe.id)
}

export function normalizePublicProbeSelection(probeIds: string[]) {
	if (probeIds.length === 0) {
		return []
	}
	return [...new Set(probeIds.filter(Boolean))]
}

export function togglePublicProbeSelection(selectedProbeIds: string[], probeId: string, checked: boolean) {
	const normalized = normalizePublicProbeSelection(selectedProbeIds)
	if (!checked) {
		return normalized.filter((selectedProbeId) => selectedProbeId !== probeId)
	}
	if (normalized.includes(probeId)) {
		return normalized
	}
	return [...normalized, probeId]
}
