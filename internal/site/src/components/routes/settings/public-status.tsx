import { Trans } from "@lingui/react/macro"
import { useEffect, useState } from "react"
import { Input } from "@/components/ui/input"
import { Switch } from "@/components/ui/switch"
import { getPublicSystems, updatePublicSystem } from "@/lib/api"
import type { AdminPublicSystem } from "@/types"

export default function PublicStatusSettings() {
	const [systems, setSystems] = useState<AdminPublicSystem[]>([])

	useEffect(() => {
		getPublicSystems().then((data) => setSystems(data.systems))
	}, [])

	async function save(system: AdminPublicSystem, patch: Partial<AdminPublicSystem>) {
		const updated = await updatePublicSystem(system.id, { ...system, ...patch })
		setSystems((prev) => prev.map((item) => (item.id === system.id ? { ...item, ...updated } : item)))
	}

	const publicCount = systems.filter((system) => system.publicEnabled).length

	return (
		<div className="space-y-4">
			<div className="space-y-1">
				<h3 className="text-lg font-medium">
					<Trans>公共看板</Trans>
				</h3>
				<p className="text-sm text-muted-foreground">
					<Trans>选择匿名首页展示哪些节点，并控制公开指标范围。</Trans>
				</p>
				<div className="text-sm text-muted-foreground">
					<Trans>当前公开节点</Trans>: {publicCount} / {systems.length}
				</div>
			</div>
			<div className="divide-y rounded-md border">
				{systems.map((system) => (
					<div key={system.id} className="grid gap-3 p-3 md:grid-cols-[minmax(0,1fr)_auto] md:items-start">
						<div className="min-w-0 space-y-3">
							<div>
								<div className="truncate font-medium">{system.publicName || system.name}</div>
								<div className="text-sm text-muted-foreground">
									{system.name} · {statusLabel(system.status)}
								</div>
							</div>
							<div className="grid gap-2 sm:grid-cols-[minmax(0,1fr)_auto] sm:items-center">
								<Input
									value={system.publicName}
									placeholder={system.name}
									disabled={!system.publicEnabled}
									onBlur={(event) => {
										if (event.currentTarget.value !== system.publicName) {
											save(system, { publicName: event.currentTarget.value })
										}
									}}
									onChange={(event) => {
										const publicName = event.currentTarget.value
										setSystems((prev) => prev.map((item) => (item.id === system.id ? { ...item, publicName } : item)))
									}}
								/>
								<div className="flex flex-wrap gap-3 text-sm">
									<FieldToggle
										label="CPU"
										checked={system.showCpu}
										disabled={!system.publicEnabled}
										onChange={(showCpu) => save(system, { showCpu })}
									/>
									<FieldToggle
										label="内存"
										checked={system.showMemory}
										disabled={!system.publicEnabled}
										onChange={(showMemory) => save(system, { showMemory })}
									/>
									<FieldToggle
										label="磁盘"
										checked={system.showDisk}
										disabled={!system.publicEnabled}
										onChange={(showDisk) => save(system, { showDisk })}
									/>
								</div>
							</div>
						</div>
						<div className="flex items-center gap-2">
							<span className="text-sm text-muted-foreground">
								{system.publicEnabled ? <Trans>已展示</Trans> : <Trans>隐藏</Trans>}
							</span>
							<Switch
								checked={system.publicEnabled}
								onCheckedChange={(publicEnabled) => save(system, { publicEnabled })}
							/>
						</div>
					</div>
				))}
				{systems.length === 0 && (
					<div className="p-6 text-center text-sm text-muted-foreground">
						<Trans>暂无可配置节点。</Trans>
					</div>
				)}
			</div>
		</div>
	)
}

function FieldToggle({
	label,
	checked,
	disabled,
	onChange,
}: {
	label: string
	checked: boolean
	disabled?: boolean
	onChange: (checked: boolean) => void
}) {
	return (
		<span className="inline-flex items-center gap-1.5">
			<Switch checked={checked} disabled={disabled} onCheckedChange={onChange} />
			<span>{label}</span>
		</span>
	)
}

function statusLabel(status: AdminPublicSystem["status"]) {
	switch (status) {
		case "up":
			return "在线"
		case "down":
			return "离线"
		case "paused":
			return "暂停"
		case "pending":
			return "等待中"
		default:
			return status
	}
}
