import { useEffect, useMemo, useState } from "react"
import {
	createNetworkProbeLiveSession,
	endNetworkProbeLiveSession,
	getNetworkProbeResults,
	getNetworkProbes,
	pb,
	renewNetworkProbeLiveSession,
} from "@/lib/api"
import { getPublicChartTimeData } from "@/lib/utils"
import type { NetworkProbe, NetworkProbeLiveSession, NetworkProbeResultPoint, PublicChartRange } from "@/types"
import { groupNetworkProbeData } from "./network-probe-groups"
import {
	liveProbeSessionHeartbeatMs,
	needsNewLiveProbeSession,
	shouldUseLiveProbeCadence,
} from "./network-probe-live-cadence"
import { applyLiveProbeResultEvent, initialLiveProbeResults } from "./network-probe-live-session"

type LoadedProbeData = {
	probes: NetworkProbe[]
	results: Record<string, { series: NetworkProbeResultPoint[] }>
}

export function useNetworkProbeData(systemId: string, range: PublicChartRange) {
	const [probes, setProbes] = useState<NetworkProbe[]>([])
	const [results, setResults] = useState<Record<string, { series: NetworkProbeResultPoint[] }>>({})
	const [loading, setLoading] = useState(false)
	const [error, setError] = useState(false)

	useEffect(() => {
		if (!systemId) return
		let cancelled = false
		let unsubscribe: (() => void) | undefined
		let heartbeatId: number | undefined
		let currentSession: NetworkProbeLiveSession | undefined
		setLoading(true)
		setError(false)

		const refresh = async () => {
			if (cancelled) return
			try {
				const assigned = await loadAssignedProbes(systemId)
				if (cancelled) return
				setProbes(assigned)
				if (shouldUseLiveProbeCadence(range)) {
					setResults(initialLiveProbeResults())
					return
				}
				const entries = await Promise.all(
					assigned.map(async (probe) => {
						const data = await getNetworkProbeResults(probe.id, { system: systemId, range })
						return [probe.id, data] as const
					})
				)
				if (cancelled) return
				setResults((current) => mergeProbeResults(current, Object.fromEntries(entries), range))
			} catch {
				if (!cancelled) {
					setProbes([])
					setResults({})
					setError(true)
				}
			}
		}

		refresh()
			.catch(() => undefined)
			.finally(() => {
				if (!cancelled) {
					setLoading(false)
				}
			})

		if (shouldUseLiveProbeCadence(range)) {
			subscribeToProbeResults(systemId, (event) => {
				setResults((current) => applyLiveProbeResultEvent(current, event))
			})
				.then((nextUnsub) => {
					unsubscribe = nextUnsub
				})
				.catch(() => {
					if (!cancelled) {
						setError(true)
					}
				})

			const ensureSession = async () => {
				try {
					if (needsNewLiveProbeSession(currentSession, systemId, range)) {
						const nextSession = await createNetworkProbeLiveSession(systemId)
						if (cancelled) {
							await endNetworkProbeLiveSession(systemId, nextSession.sessionId).catch(() => undefined)
							return
						}
						currentSession = nextSession
						return
					}
					const nextSession = await renewNetworkProbeLiveSession(systemId, currentSession.sessionId)
					if (!cancelled) {
						currentSession = nextSession
					}
				} catch {
					const nextSession = await createNetworkProbeLiveSession(systemId)
					if (cancelled) {
						await endNetworkProbeLiveSession(systemId, nextSession.sessionId).catch(() => undefined)
						return
					}
					currentSession = nextSession
				}
			}

			ensureSession().catch(() => {
				if (!cancelled) {
					setError(true)
				}
			})
			heartbeatId = window.setInterval(() => {
				ensureSession().catch(() => {
					if (!cancelled) {
						setError(true)
					}
				})
			}, liveProbeSessionHeartbeatMs)
		}

		return () => {
			cancelled = true
			if (heartbeatId) {
				window.clearInterval(heartbeatId)
			}
			if (currentSession && shouldUseLiveProbeCadence(range)) {
				endNetworkProbeLiveSession(systemId, currentSession.sessionId).catch(() => undefined)
			}
			unsubscribe?.()
		}
	}, [range, systemId])

	const groups = useMemo(() => groupNetworkProbeData(probes, results, systemId), [probes, results, systemId])

	return { probes, results, loading, error, groups }
}

function subscribeToProbeResults(
	systemId: string,
	onChange: (event: { action: string; record: Record<string, unknown> }) => void
) {
	return pb.collection("network_probe_results").subscribe(
		"*",
		(event) => {
			const record = event.record as Record<string, unknown> | undefined
			if (typeof record?.system === "string" && record.system === systemId) {
				onChange({ action: event.action, record })
			}
		},
		{
			fields:
				"id,probe,system,created,success,latency_ms,packet_loss_percent,http_status,error,failure_category,type,target",
		}
	)
}

async function loadAssignedProbes(systemId: string) {
	const { probes } = await getNetworkProbes()
	return probes.filter((probe) => probe.systems.includes(systemId))
}

function mergeProbeResults(
	current: Record<string, { series: NetworkProbeResultPoint[] }>,
	next: Record<string, { series: NetworkProbeResultPoint[] }>,
	range: PublicChartRange
) {
	const merged = { ...current }
	for (const [probeId, value] of Object.entries(next)) {
		const currentSeries = merged[probeId]?.series ?? []
		merged[probeId] = {
			series: mergeProbeSeries(currentSeries, value.series, range),
		}
	}
	return merged
}

function mergeProbeSeries(
	current: NetworkProbeResultPoint[],
	next: NetworkProbeResultPoint[],
	range: PublicChartRange
) {
	const cutoff = getWindowCutoffMs(range)
	const series = [...current, ...next].filter((point) => new Date(point.created).getTime() >= cutoff)
	const byCreated = new Map<string, NetworkProbeResultPoint>()
	for (const point of series) {
		byCreated.set(point.created, point)
	}
	return [...byCreated.values()].sort((a, b) => new Date(a.created).getTime() - new Date(b.created).getTime())
}

function getWindowCutoffMs(range: PublicChartRange) {
	return getPublicChartTimeData(range).getOffset(new Date()).getTime()
}
