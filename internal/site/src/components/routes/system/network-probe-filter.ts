import type { NetworkProbe } from "@/types"

export function filterAssignedProbes(probes: NetworkProbe[], systemId: string) {
	return probes.filter((probe) => probe.scope === "global" || probe.systems.includes(systemId))
}
