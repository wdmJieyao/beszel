import { t } from "@lingui/core/macro"
import { Trans } from "@lingui/react/macro"
import { redirectPage } from "@nanostores/router"
import clsx from "clsx"
import { DownloadIcon, FileSlidersIcon, LoaderCircleIcon, RotateCwIcon, SearchCheckIcon } from "lucide-react"
import { useState } from "react"
import { $router } from "@/components/router"
import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert"
import { Button } from "@/components/ui/button"
import { Checkbox } from "@/components/ui/checkbox"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import { Separator } from "@/components/ui/separator"
import { Switch } from "@/components/ui/switch"
import { Textarea } from "@/components/ui/textarea"
import { toast } from "@/components/ui/use-toast"
import { exportConfigBackup, isAdmin, pb, restoreConfigBackup, validateConfigBackup } from "@/lib/api"
import type { ConfigBackupPreviewResponse, ConfigBackupSection } from "@/types"
import { ConfigBackupPreview } from "./config-backup-preview"
import {
	buildConfigBackupExportPayload,
	configBackupSections,
	previewHasBlockingIssues,
	toggleConfigBackupSection,
} from "./config-backup-utils"

const sectionLabels: Record<ConfigBackupSection, string> = {
	systems: "节点与指纹",
	alerts: "告警与静默时间",
	notifications: "通知与 Telegram",
	publicStatus: "公共看板",
	networkProbes: "线路检测",
}

export default function ConfigYaml() {
	const [backupContent, setBackupContent] = useState("")
	const [legacyConfigContent, setLegacyConfigContent] = useState("")
	const [credential, setCredential] = useState("")
	const [includeSecrets, setIncludeSecrets] = useState(true)
	const [sections, setSections] = useState<ConfigBackupSection[]>(configBackupSections)
	const [preview, setPreview] = useState<ConfigBackupPreviewResponse>()
	const [isExporting, setIsExporting] = useState(false)
	const [isValidating, setIsValidating] = useState(false)
	const [isRestoring, setIsRestoring] = useState(false)
	const [isLegacyLoading, setIsLegacyLoading] = useState(false)

	if (!isAdmin()) {
		redirectPage($router, "settings", { name: "general" })
	}

	async function runExport() {
		setIsExporting(true)
		try {
			const payload = buildConfigBackupExportPayload({
				includeSecrets,
				encryptionCredential: credential,
				sections,
			})
			const response = await exportConfigBackup(payload)
			setBackupContent(response.content)
			setPreview(undefined)
			toast({
				title: t`Configuration backup exported`,
				description: response.filename,
			})
		} catch (error: any) {
			toast({
				title: t`Failed to export configuration backup`,
				description: error.message,
				variant: "destructive",
			})
		} finally {
			setIsExporting(false)
		}
	}

	async function runValidation() {
		setIsValidating(true)
		try {
			const result = await validateConfigBackup({
				content: backupContent,
				decryptionCredential: credential,
			})
			setPreview(result)
			toast({
				title: t`Backup preview ready`,
				description: t`Review the merge result before restoring.`,
			})
		} catch (error: any) {
			const response = error?.data
			if (response?.previewId) {
				setPreview(response)
				return
			}
			toast({
				title: t`Failed to validate configuration backup`,
				description: error.message,
				variant: "destructive",
			})
		} finally {
			setIsValidating(false)
		}
	}

	async function runRestore() {
		if (!preview || previewHasBlockingIssues(preview)) {
			return
		}
		setIsRestoring(true)
		try {
			const response = await restoreConfigBackup({
				content: backupContent,
				decryptionCredential: credential,
				previewId: preview.previewId,
				mode: "merge",
			})
			toast({
				title: t`Configuration backup restored`,
				description: t`${response.applied.created} created, ${response.applied.updated} updated.`,
			})
		} catch (error: any) {
			toast({
				title: t`Failed to restore configuration backup`,
				description: error.message,
				variant: "destructive",
			})
		} finally {
			setIsRestoring(false)
		}
	}

	async function fetchLegacyConfig() {
		try {
			setIsLegacyLoading(true)
			const { config } = await pb.send<{ config: string }>("/api/beszel/config-yaml", {})
			setLegacyConfigContent(config)
		} catch (error: any) {
			toast({
				title: t`Error`,
				description: error.message,
				variant: "destructive",
			})
		} finally {
			setIsLegacyLoading(false)
		}
	}

	return (
		<div>
			<div>
				<h3 className="text-xl font-medium mb-2">
					<Trans>Configuration Backup</Trans>
				</h3>
				<p className="text-sm text-muted-foreground leading-relaxed">
					<Trans>导出完整面板配置，或通过预览后的合并恢复导入配置。</Trans>
				</p>
			</div>
			<Separator className="my-4" />
			<div className="space-y-5">
				<div className="rounded-md border p-4 space-y-4">
					<div className="grid gap-4 lg:grid-cols-[1fr_260px]">
						<div className="space-y-3">
							<div className="flex items-center justify-between gap-3 rounded-md bg-muted/40 p-3">
								<div>
									<div className="font-medium">
										<Trans>Encrypt sensitive values</Trans>
									</div>
									<div className="text-sm text-muted-foreground">
										<Trans>包含 token、Webhook URL、Telegram Bot Token 等敏感值时必须加密。</Trans>
									</div>
								</div>
								<Switch checked={includeSecrets} onCheckedChange={setIncludeSecrets} />
							</div>
							<div className="grid gap-2">
								<Label htmlFor="backup-credential">
									<Trans>Encryption credential</Trans>
								</Label>
								<Input
									id="backup-credential"
									type="password"
									value={credential}
									onChange={(event) => setCredential(event.target.value)}
								/>
							</div>
						</div>
						<div className="space-y-2 rounded-md bg-muted/40 p-3">
							<div className="text-sm font-medium">
								<Trans>Sections</Trans>
							</div>
							{configBackupSections.map((section) => (
								<div key={section} className="flex items-center gap-2 text-sm">
									<Checkbox
										id={`backup-section-${section}`}
										checked={sections.includes(section)}
										onCheckedChange={(checked) =>
											setSections((current) => toggleConfigBackupSection(current, section, checked === true))
										}
									/>
									<Label htmlFor={`backup-section-${section}`}>{sectionLabels[section]}</Label>
								</div>
							))}
						</div>
					</div>
					<div className="flex flex-wrap gap-2">
						<Button onClick={runExport} disabled={isExporting}>
							{isExporting ? <LoaderCircleIcon className="size-4 animate-spin" /> : <DownloadIcon className="size-4" />}
							<span className="ms-1">
								<Trans>Export Backup</Trans>
							</span>
						</Button>
						<Button variant="outline" onClick={runValidation} disabled={isValidating || backupContent.trim() === ""}>
							{isValidating ? (
								<LoaderCircleIcon className="size-4 animate-spin" />
							) : (
								<SearchCheckIcon className="size-4" />
							)}
							<span className="ms-1">
								<Trans>Validate Preview</Trans>
							</span>
						</Button>
						<Button variant="outline" onClick={runRestore} disabled={isRestoring || previewHasBlockingIssues(preview)}>
							{isRestoring ? <LoaderCircleIcon className="size-4 animate-spin" /> : <RotateCwIcon className="size-4" />}
							<span className="ms-1">
								<Trans>Restore Merge</Trans>
							</span>
						</Button>
					</div>
					<Textarea
						dir="ltr"
						value={backupContent}
						onChange={(event) => {
							setBackupContent(event.target.value)
							setPreview(undefined)
						}}
						spellCheck="false"
						rows={Math.min(26, Math.max(10, backupContent.split("\n").length))}
						className="font-mono whitespace-pre"
						placeholder="meta:"
					/>
					{preview && <ConfigBackupPreview preview={preview} />}
				</div>

				<div className="rounded-md border p-4">
					<div className="mb-4">
						<h3 className="mb-1 text-lg font-medium">
							<Trans>Legacy config.yml</Trans>
						</h3>
						<p className="text-sm text-muted-foreground leading-relaxed">
							<Trans>旧版 config.yml 只包含系统配置，并保留重启同步时删除缺失系统的原有行为。</Trans>
						</p>
						<Alert className="my-4 border-destructive text-destructive w-auto table md:pe-6">
							<AlertTitle>
								<Trans>Caution - potential data loss</Trans>
							</AlertTitle>
							<AlertDescription>
								<Trans>
									Existing systems not defined in <code>config.yml</code> will be deleted. Please make regular backups.
								</Trans>
							</AlertDescription>
						</Alert>
					</div>
					{legacyConfigContent && (
						<Textarea
							dir="ltr"
							value={legacyConfigContent}
							readOnly
							spellCheck="false"
							rows={Math.min(18, legacyConfigContent.split("\n").length)}
							className="font-mono whitespace-pre"
						/>
					)}
					<Button
						type="button"
						variant="outline"
						className="mt-4 flex items-center gap-1"
						onClick={fetchLegacyConfig}
						disabled={isLegacyLoading}
					>
						<FileSlidersIcon className={clsx("h-4 w-4 me-0.5", isLegacyLoading && "hidden")} />
						{isLegacyLoading && <LoaderCircleIcon className="h-4 w-4 me-0.5 animate-spin" />}
						<Trans>Export legacy config.yml</Trans>
					</Button>
				</div>
			</div>
		</div>
	)
}
