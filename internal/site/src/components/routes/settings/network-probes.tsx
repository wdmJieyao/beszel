import { Trans } from "@lingui/react/macro"
import { ChevronDownIcon, PlusIcon, TrashIcon } from "lucide-react"
import { useEffect, useState } from "react"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select"
import { Switch } from "@/components/ui/switch"
import { deleteNetworkProbe, getNetworkProbes, saveNetworkProbe } from "@/lib/api"
import { $systems } from "@/lib/stores"
import { useStore } from "@nanostores/react"
import type { NetworkProbe, NetworkProbeType } from "@/types"
import { buildNetworkProbePayload, probeScopeLabel } from "./network-probes-utils"

const emptyProbe: Partial<NetworkProbe> = {
	name: "",
	type: "tcping",
	target: "",
	intervalSeconds: 10,
	timeoutSeconds: 5,
	enabled: true,
	publicVisible: true,
	scope: "global",
	systems: [],
}

export default function NetworkProbesSettings() {
	const systems = useStore($systems)
	const [probes, setProbes] = useState<NetworkProbe[]>([])
	const [draft, setDraft] = useState<Partial<NetworkProbe>>(emptyProbe)
	const [advancedOpen, setAdvancedOpen] = useState(false)

	useEffect(() => {
		getNetworkProbes().then((data) => setProbes(data.probes))
	}, [])

	async function submit() {
		const saved = await saveNetworkProbe(buildNetworkProbePayload(draft))
		setProbes((prev) => (draft.id ? prev.map((probe) => (probe.id === saved.id ? saved : probe)) : [...prev, saved]))
		setDraft(emptyProbe)
		setAdvancedOpen(false)
	}

	async function remove(probe: NetworkProbe) {
		await deleteNetworkProbe(probe.id)
		setProbes((prev) => prev.filter((item) => item.id !== probe.id))
	}

	return (
		<div className="space-y-5">
			<div>
				<h3 className="text-lg font-medium">
					<Trans>线路检测</Trans>
				</h3>
				<p className="text-sm text-muted-foreground">
					<Trans>配置线路、观测点和检测目标，系统会定期记录延迟走势。</Trans>
				</p>
			</div>
			<div className="space-y-3 rounded-md border p-3">
				<div className="grid gap-2 md:grid-cols-[minmax(0,1fr)_150px_minmax(0,1fr)_auto]">
					<Input
						placeholder="线路名称"
						value={draft.name ?? ""}
						onChange={(e) => setDraft((prev) => ({ ...prev, name: e.target.value }))}
					/>
					<Select
						value={draft.type}
						onValueChange={(value: NetworkProbeType) => setDraft((prev) => ({ ...prev, type: value }))}
					>
						<SelectTrigger>
							<SelectValue />
						</SelectTrigger>
						<SelectContent>
							<SelectItem value="tcping">TCPing</SelectItem>
							<SelectItem value="icmp_ping">ICMP Ping</SelectItem>
							<SelectItem value="http_get">HTTP GET</SelectItem>
						</SelectContent>
					</Select>
					<Input
						placeholder="检测目标，例如 example.com:443"
						value={draft.target ?? ""}
						onChange={(e) => setDraft((prev) => ({ ...prev, target: e.target.value }))}
					/>
					<Button onClick={submit} disabled={!draft.name || !draft.target || systems.length === 0}>
						<PlusIcon className="me-2 size-4" />
						<Trans>保存线路</Trans>
					</Button>
				</div>
				<div className="flex flex-wrap items-center justify-between gap-2 text-sm text-muted-foreground">
					<span>
						<Trans>默认由所有可用节点定期检测。</Trans>
					</span>
					<Button variant="ghost" size="sm" onClick={() => setAdvancedOpen((open) => !open)}>
						<ChevronDownIcon className="me-1 size-4" />
						<Trans>高级设置</Trans>
					</Button>
				</div>
				{advancedOpen && (
					<div className="grid gap-2 rounded-md bg-muted/40 p-3 md:grid-cols-3">
						<Input
							type="number"
							min={10}
							value={draft.intervalSeconds ?? 10}
							onChange={(e) => setDraft((prev) => ({ ...prev, intervalSeconds: Number(e.target.value) }))}
							aria-label="检测间隔"
						/>
						<Input
							type="number"
							min={1}
							value={draft.timeoutSeconds ?? 5}
							onChange={(e) => setDraft((prev) => ({ ...prev, timeoutSeconds: Number(e.target.value) }))}
							aria-label="超时时间"
						/>
						<Select
							value={draft.scope === "fixed" ? (draft.systems?.[0] ?? "__all") : "__all"}
							onValueChange={(value) =>
								setDraft((prev) => ({
									...prev,
									scope: value === "__all" ? "global" : "fixed",
									systems: value === "__all" ? [] : [value],
								}))
							}
						>
							<SelectTrigger>
								<SelectValue placeholder="执行节点" />
							</SelectTrigger>
							<SelectContent>
								<SelectItem value="__all">全部可用节点</SelectItem>
								{systems.map((system) => (
									<SelectItem key={system.id} value={system.id}>
										{system.name}
									</SelectItem>
								))}
							</SelectContent>
						</Select>
					</div>
				)}
			</div>
			<div className="divide-y rounded-md border">
				{probes.map((probe) => (
					<div key={probe.id} className="flex items-center justify-between gap-3 p-3">
						<button type="button" className="min-w-0 text-start" onClick={() => setDraft(probe)}>
							<div className="truncate font-medium">{probe.name}</div>
							<div className="truncate text-sm text-muted-foreground">
								{probeTypeLabel(probe.type)} · {probe.target} · {probeScopeLabel(probe, systems)}
							</div>
						</button>
						<div className="flex items-center gap-3">
							<span className="text-sm text-muted-foreground">
								{probe.publicVisible ? <Trans>公开展示</Trans> : <Trans>仅后台</Trans>}
							</span>
							<Switch
								checked={probe.publicVisible}
								onCheckedChange={async (checked) => {
									const saved = await saveNetworkProbe({ ...probe, publicVisible: checked })
									setProbes((prev) => prev.map((item) => (item.id === saved.id ? saved : item)))
								}}
							/>
							<Button variant="ghost" size="icon" onClick={() => remove(probe)}>
								<TrashIcon className="size-4" />
							</Button>
						</div>
					</div>
				))}
				{probes.length === 0 && (
					<div className="p-6 text-center text-sm text-muted-foreground">
						<Trans>暂无线路检测配置。</Trans>
					</div>
				)}
			</div>
		</div>
	)
}

function probeTypeLabel(type: NetworkProbeType) {
	switch (type) {
		case "tcping":
			return "TCPing"
		case "icmp_ping":
			return "ICMP Ping"
		case "http_get":
			return "HTTP GET"
		default:
			return type
	}
}
