import { useEffect, useMemo, useState } from "react"
import { timeDay, timeHour, timeMinute } from "d3-time"
import type { NetworkProbe, NetworkProbeLiveSession, NetworkProbeResultPoint, PublicChartRange } from "../../../types"
import { groupNetworkProbeData } from "./network-probe-groups.ts"
import {
	liveProbeSessionHeartbeatMs,
	needsNewLiveProbeSession,
	shouldUseLiveProbeCadence,
} from "./network-probe-live-cadence.ts"
import { applyLiveProbeResultEvent, initialLiveProbeResults } from "./network-probe-live-session.ts"
import { filterAssignedProbes } from "./network-probe-filter.ts"

type LoadedProbeData = {
	probes: NetworkProbe[]
	results: Record<string, { series: NetworkProbeResultPoint[] }>
}

type HistoricalProbeResultsLoad = {
	results: Record<string, { series: NetworkProbeResultPoint[] }>
	hasSuccesses: boolean
	hasFailures: boolean
}

type HistoricalProbeLoadState = {
	results: Record<string, { series: NetworkProbeResultPoint[] }>
	error: boolean
}

type HistoricalProbeResultsFetcher = (
	probeId: string,
	params: { system?: string; range: PublicChartRange }
) => Promise<{ series: NetworkProbeResultPoint[] }>

export function useNetworkProbeData(systemId: string, range: PublicChartRange) {
	const [probes, setProbes] = useState<NetworkProbe[]>([])
	const [results, setResults] = useState<Record<string, { series: NetworkProbeResultPoint[] }>>({})
	const [resultsMode, setResultsMode] = useState<"live" | "historical">(() =>
		shouldUseLiveProbeCadence(range) ? "live" : "historical"
	)
	const [liveWaiting, setLiveWaiting] = useState(() => shouldUseLiveProbeCadence(range))
	const [loading, setLoading] = useState(false)
	const [error, setError] = useState(false)

	useEffect(() => {
		if (!systemId) return
		let cancelled = false
		let unsubscribe: (() => void) | undefined
		let heartbeatId: number | undefined
		let currentSession: NetworkProbeLiveSession | undefined
		const useLiveCadence = shouldUseLiveProbeCadence(range)
		setLoading(true)
		setError(false)
		setProbes([])
		setResultsMode(useLiveCadence ? "live" : "historical")
		setLiveWaiting(useLiveCadence)
		if (useLiveCadence) {
			setResults(initialLiveProbeResults())
		}

		const refresh = async () => {
			if (cancelled) return
			try {
				const assigned = await loadAssignedProbes(systemId)
				if (cancelled) return
				setProbes(assigned)
				if (assigned.length === 0) {
					setLiveWaiting(false)
				}
				if (useLiveCadence) {
					return
				}
				setLiveWaiting(false)
				const history = await loadHistoricalProbeResults(assigned, systemId, range)
				if (cancelled) return
				if (shouldErrorHistoricalProbeLoad(history, assigned.length)) {
					setResults({})
					setError(true)
					return
				}
				setResults((current) => applyHistoricalProbeLoad(current, history, range, assigned.length).results)
				setError(false)
			} catch {
				if (!cancelled) {
					setProbes([])
					setResults({})
					setLiveWaiting(false)
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

		if (useLiveCadence) {
			subscribeToProbeResults(systemId, (event) => {
				setLiveWaiting(false)
				setResults((current) => applyLiveProbeResultEvent(current, event))
			})
				.then((nextUnsub) => {
					if (cancelled) {
						nextUnsub()
						return
					}
					unsubscribe = nextUnsub
				})
				.catch(() => {
					if (!cancelled) {
						setLiveWaiting(false)
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
					setLiveWaiting(false)
					setError(true)
				}
			})
			heartbeatId = window.setInterval(() => {
				ensureSession().catch(() => {
					if (!cancelled) {
						setLiveWaiting(false)
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
			if (currentSession && useLiveCadence) {
				endNetworkProbeLiveSession(systemId, currentSession.sessionId).catch(() => undefined)
			}
			unsubscribe?.()
		}
	}, [range, systemId])

	const displayResults = useMemo(() => {
		if (shouldUseLiveProbeCadence(range) && resultsMode !== "live") {
			return initialLiveProbeResults()
		}
		return results
	}, [range, results, resultsMode])

	const groups = useMemo(
		() => groupNetworkProbeData(probes, displayResults, systemId),
		[probes, displayResults, systemId]
	)

	return { probes, results: displayResults, loading: loading || liveWaiting, error, groups }
}

function subscribeToProbeResults(
	systemId: string,
	onChange: (event: { action: string; record: Record<string, unknown> }) => void
) {
	return loadApiModule().then(({ pb }) =>
		pb.collection("network_probe_results").subscribe(
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
	)
}

export async function loadAssignedProbes(systemId: string) {
	const { getNetworkProbes } = await loadApiModule()
	const { probes } = await getNetworkProbes()
	return filterAssignedProbes(probes, systemId)
}

export async function loadHistoricalProbeResults(
	assigned: NetworkProbe[],
	systemId: string,
	range: PublicChartRange,
	fetchResults: HistoricalProbeResultsFetcher = loadNetworkProbeResults
): Promise<HistoricalProbeResultsLoad> {
	const settled = await Promise.allSettled(
		assigned.map(async (probe) => {
			const data = await fetchResults(probe.id, { system: systemId, range })
			return [probe.id, data] as const
		})
	)

	const entries: Array<readonly [string, { series: NetworkProbeResultPoint[] }]> = []
	let failureCount = 0

	for (const result of settled) {
		if (result.status === "fulfilled") {
			entries.push(result.value)
			continue
		}
		failureCount++
	}

	return {
		results: Object.fromEntries(entries),
		hasSuccesses: entries.length > 0,
		hasFailures: failureCount > 0,
	}
}

export function applyHistoricalProbeLoad(
	_current: Record<string, { series: NetworkProbeResultPoint[] }>,
	history: HistoricalProbeResultsLoad,
	range: PublicChartRange,
	assignedCount: number
): HistoricalProbeLoadState {
	if (shouldErrorHistoricalProbeLoad(history, assignedCount)) {
		return { results: {}, error: true }
	}
	if (assignedCount === 0) {
		return { results: {}, error: false }
	}
	return {
		results: mergeProbeResults({}, history.results, range),
		error: false,
	}
}

export function shouldErrorHistoricalProbeLoad(history: HistoricalProbeResultsLoad, assignedCount: number) {
	return assignedCount > 0 && !history.hasSuccesses && history.hasFailures
}

function loadApiModule() {
	return import("../../../lib/api.ts")
}

async function loadNetworkProbeResults(probeId: string, params: { system?: string; range: PublicChartRange }) {
	const { getNetworkProbeResults } = await loadApiModule()
	return getNetworkProbeResults(probeId, params)
}

async function createNetworkProbeLiveSession(systemId: string) {
	const { createNetworkProbeLiveSession: createSession } = await loadApiModule()
	return createSession(systemId)
}

async function renewNetworkProbeLiveSession(systemId: string, sessionId: string) {
	const { renewNetworkProbeLiveSession: renewSession } = await loadApiModule()
	return renewSession(systemId, sessionId)
}

async function endNetworkProbeLiveSession(systemId: string, sessionId: string) {
	const { endNetworkProbeLiveSession: endSession } = await loadApiModule()
	return endSession(systemId, sessionId)
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
	const endTime = new Date()
	switch (range) {
		case "1m":
			return timeMinute.offset(endTime, -1).getTime()
		case "30m":
			return timeMinute.offset(endTime, -30).getTime()
		case "1h":
			return timeHour.offset(endTime, -1).getTime()
		case "12h":
			return timeHour.offset(endTime, -12).getTime()
		case "24h":
			return timeHour.offset(endTime, -24).getTime()
		case "1w":
			return timeDay.offset(endTime, -7).getTime()
		case "30d":
			return timeDay.offset(endTime, -30).getTime()
		default:
			return timeMinute.offset(endTime, -30).getTime()
	}
}
