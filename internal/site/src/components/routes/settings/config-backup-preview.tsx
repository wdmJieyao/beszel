import { Trans } from "@lingui/react/macro"
import { Badge } from "@/components/ui/badge"
import type { ConfigBackupPreviewResponse } from "@/types"
import { configBackupActionLabel } from "./config-backup-utils"

export function ConfigBackupPreview({ preview }: { preview: ConfigBackupPreviewResponse }) {
	return (
		<div className="space-y-3">
			<div className="grid grid-cols-2 gap-2 text-sm md:grid-cols-6">
				<PreviewCount label="创建" value={preview.summary.create} />
				<PreviewCount label="更新" value={preview.summary.update} />
				<PreviewCount label="保留" value={preview.summary.preserve} />
				<PreviewCount label="跳过" value={preview.summary.skip} />
				<PreviewCount label="冲突" value={preview.summary.conflict} />
				<PreviewCount label="错误" value={preview.summary.error} />
			</div>
			<div className="max-h-72 overflow-auto rounded-md border">
				{preview.items.map((item, index) => (
					<div
						key={`${item.section}-${item.stableId}-${index}`}
						className="flex items-start justify-between gap-3 border-b p-3 last:border-b-0"
					>
						<div className="min-w-0">
							<div className="truncate font-medium">{item.displayName || item.stableId || item.section}</div>
							<div className="text-sm text-muted-foreground">{item.reason}</div>
						</div>
						<Badge variant={item.action === "conflict" || item.action === "error" ? "destructive" : "outline"}>
							{configBackupActionLabel(item.action)}
						</Badge>
					</div>
				))}
				{preview.items.length === 0 && (
					<div className="p-6 text-center text-sm text-muted-foreground">
						<Trans>没有需要应用的配置项。</Trans>
					</div>
				)}
			</div>
		</div>
	)
}

function PreviewCount({ label, value }: { label: string; value: number }) {
	return (
		<div className="rounded-md border p-2">
			<div className="text-xs text-muted-foreground">{label}</div>
			<div className="text-lg font-medium">{value}</div>
		</div>
	)
}
