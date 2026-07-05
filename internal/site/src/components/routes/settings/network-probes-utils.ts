import type { NetworkProbe } from "@/types"

type SystemSummary = {
	id: string
	name: string
}

export function buildNetworkProbePayload(draft: Partial<NetworkProbe>) {
	const scope = draft.systems?.length ? "fixed" : "global"
	return {
		id: draft.id,
		name: draft.name,
		type: draft.type,
		target: draft.target,
		intervalSeconds: draft.intervalSeconds,
		timeoutSeconds: draft.timeoutSeconds,
		enabled: draft.enabled,
		scope,
		systems: scope === "global" ? [] : draft.systems,
	}
}

export function probeScopeLabel(probe: Pick<NetworkProbe, "scope" | "systems">, systems: SystemSummary[]) {
	if (probe.scope === "global") {
		return "全部可用节点"
	}
	const names = probe.systems.map((systemId) => systems.find((system) => system.id === systemId)?.name).filter(Boolean)
	if (names.length === 0) {
		return "固定节点: 0 个"
	}
	return `固定节点: ${names.join(", ")}`
}
