import { t } from "@lingui/core/macro"
import { Trans } from "@lingui/react/macro"
import { useStore } from "@nanostores/react"
import { LoaderCircleIcon, PlusIcon, SaveIcon, SendIcon, Trash2Icon } from "lucide-react"
import { useEffect, useMemo, useState } from "react"
import { Button } from "@/components/ui/button"
import { Card } from "@/components/ui/card"
import { Checkbox } from "@/components/ui/checkbox"
import { Input } from "@/components/ui/input"
import { InputTags } from "@/components/ui/input-tags"
import { Label } from "@/components/ui/label"
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select"
import { Separator } from "@/components/ui/separator"
import { Switch } from "@/components/ui/switch"
import { toast } from "@/components/ui/use-toast"
import {
	deleteTelegramDestination,
	getTelegramDestinations,
	saveTelegramDestination,
	testTelegramDestination,
	testTelegramSettings,
} from "@/lib/api"
import { $systems } from "@/lib/stores"
import type { TelegramDestination, TelegramDestinationInput, TelegramSettingsInput } from "@/types"
import {
	buildTelegramDestinationPayload,
	buildTelegramSettingsPayload,
	defaultTelegramDestination,
} from "./telegram-utils"

export function TelegramDestinations({
	settings,
	onSettingsChange,
	onSettingsSave,
}: {
	settings: TelegramSettingsInput & { botUsername?: string; hasToken?: boolean; lastError?: string }
	onSettingsChange: (settings: TelegramSettingsInput) => void
	onSettingsSave: (settings: TelegramSettingsInput) => Promise<void>
}) {
	const systems = useStore($systems)
	const [destinations, setDestinations] = useState<TelegramDestination[]>([])
	const [draft, setDraft] = useState<TelegramDestinationInput>(defaultTelegramDestination())
	const [isSavingSettings, setIsSavingSettings] = useState(false)
	const [isTestingSettings, setIsTestingSettings] = useState(false)
	const [busyDestinationId, setBusyDestinationId] = useState("")

	useEffect(() => {
		getTelegramDestinations()
			.then((response) => setDestinations(response.destinations))
			.catch((error) => {
				toast({
					title: t`Failed to load Telegram destinations`,
					description: error.message,
					variant: "destructive",
				})
			})
	}, [])

	const hasSettingsReady = settings.enabled && (!!settings.botToken || settings.hasToken)
	const selectedSystems = useMemo(() => new Set(draft.nodeScope), [draft.nodeScope])

	async function saveSettings() {
		setIsSavingSettings(true)
		try {
			await onSettingsSave(buildTelegramSettingsPayload(settings))
		} catch (error: any) {
			toast({
				title: t`Failed to save Telegram settings`,
				description: error.message,
				variant: "destructive",
			})
		} finally {
			setIsSavingSettings(false)
		}
	}

	async function runSettingsTest() {
		setIsTestingSettings(true)
		try {
			const result = await testTelegramSettings(buildTelegramSettingsPayload(settings))
			if (!result.ok) {
				throw new Error(result.error || t`Telegram test failed`)
			}
			toast({
				title: t`Telegram bot verified`,
				description: result.botUsername || t`Telegram connectivity is working.`,
			})
		} catch (error: any) {
			toast({
				title: t`Failed to verify Telegram bot`,
				description: error.message,
				variant: "destructive",
			})
		} finally {
			setIsTestingSettings(false)
		}
	}

	async function saveDestination() {
		try {
			const payload = buildTelegramDestinationPayload(draft)
			const saved = await saveTelegramDestination(payload)
			setDestinations((previous) =>
				draft.id ? previous.map((item) => (item.id === saved.id ? saved : item)) : [...previous, saved]
			)
			setDraft(defaultTelegramDestination())
			toast({
				title: t`Telegram destination saved`,
				description: saved.name,
			})
		} catch (error: any) {
			toast({
				title: t`Failed to save Telegram destination`,
				description: error.message,
				variant: "destructive",
			})
		}
	}

	async function removeDestination(destination: TelegramDestination) {
		setBusyDestinationId(destination.id)
		try {
			await deleteTelegramDestination(destination.id)
			setDestinations((previous) => previous.filter((item) => item.id !== destination.id))
			if (draft.id === destination.id) {
				setDraft(defaultTelegramDestination())
			}
		} catch (error: any) {
			toast({
				title: t`Failed to delete Telegram destination`,
				description: error.message,
				variant: "destructive",
			})
		} finally {
			setBusyDestinationId("")
		}
	}

	async function sendDestinationTest(destination: TelegramDestination) {
		setBusyDestinationId(destination.id)
		try {
			const result = await testTelegramDestination(destination.id)
			if (!result.ok) {
				throw new Error(result.error || t`Telegram test failed`)
			}
			toast({
				title: t`Telegram test message sent`,
				description: destination.name,
			})
		} catch (error: any) {
			toast({
				title: t`Failed to send Telegram test message`,
				description: error.message,
				variant: "destructive",
			})
		} finally {
			setBusyDestinationId("")
		}
	}

	return (
		<div className="space-y-4">
			<div className="rounded-md border p-4 space-y-4">
				<div className="flex flex-wrap items-start justify-between gap-4">
					<div>
						<h3 className="text-lg font-medium">
							<Trans>Telegram bot</Trans>
						</h3>
						<p className="text-sm text-muted-foreground">
							<Trans>管理员可以在这里配置 Bot 凭据、验证连通性，并维护允许接收消息的会话。</Trans>
						</p>
					</div>
					<div className="flex items-center gap-2">
						<Button variant="outline" onClick={runSettingsTest} disabled={isTestingSettings}>
							{isTestingSettings ? (
								<LoaderCircleIcon className="size-4 animate-spin" />
							) : (
								<SendIcon className="size-4" />
							)}
							<span className="ms-1">
								<Trans>Test Bot</Trans>
							</span>
						</Button>
						<Button onClick={saveSettings} disabled={isSavingSettings}>
							{isSavingSettings ? (
								<LoaderCircleIcon className="size-4 animate-spin" />
							) : (
								<SaveIcon className="size-4" />
							)}
							<span className="ms-1">
								<Trans>Save Telegram</Trans>
							</span>
						</Button>
					</div>
				</div>
				<div className="grid gap-4 md:grid-cols-2">
					<div className="space-y-2">
						<Label htmlFor="telegram-token">
							<Trans>Bot Token</Trans>
						</Label>
						<Input
							id="telegram-token"
							type="password"
							placeholder="123456:AA..."
							value={settings.botToken ?? ""}
							onChange={(event) => onSettingsChange({ ...settings, botToken: event.target.value })}
						/>
						<p className="text-[0.8rem] text-muted-foreground">
							{settings.hasToken ? (
								<Trans>已保存旧 token；留空即可保留。</Trans>
							) : (
								<Trans>首次启用时需要填写有效 token。</Trans>
							)}
						</p>
					</div>
					<div className="space-y-3 rounded-md bg-muted/40 p-3">
						<div className="flex items-center justify-between gap-3">
							<div>
								<div className="font-medium">
									<Trans>Enable Telegram</Trans>
								</div>
								<div className="text-sm text-muted-foreground">
									<Trans>启用 Telegram 告警投递。</Trans>
								</div>
							</div>
							<Switch
								checked={settings.enabled}
								onCheckedChange={(checked) => onSettingsChange({ ...settings, enabled: checked })}
							/>
						</div>
						<div className="flex items-center justify-between gap-3">
							<div>
								<div className="font-medium">
									<Trans>Bot polling</Trans>
								</div>
								<div className="text-sm text-muted-foreground">
									<Trans>后续可用于管理 Telegram 菜单命令。</Trans>
								</div>
							</div>
							<Switch
								checked={settings.pollingEnabled}
								onCheckedChange={(checked) => onSettingsChange({ ...settings, pollingEnabled: checked })}
							/>
						</div>
						{settings.botUsername && <div className="text-sm text-muted-foreground">@{settings.botUsername}</div>}
						<div className="rounded-md border bg-background p-3 text-sm text-muted-foreground">
							<div className="mb-1 font-medium text-foreground">
								<Trans>Bot menu</Trans>
							</div>
							<div>
								<Trans>
									启用轮询后，管理员目的地可使用
									/status、/alerts、/systems、/system、/mute、/unmute；只读目的地只能接收授权范围内的通知。
								</Trans>
							</div>
						</div>
						{settings.lastError && <div className="text-sm text-destructive">{settings.lastError}</div>}
					</div>
				</div>
			</div>

			<div className="rounded-md border p-4 space-y-4">
				<div className="flex flex-wrap items-start justify-between gap-4">
					<div>
						<h3 className="text-lg font-medium">
							<Trans>Telegram destinations</Trans>
						</h3>
						<p className="text-sm text-muted-foreground">
							<Trans>维护允许接收告警的私聊、群组、频道或超群，并为只读渠道限制节点和告警范围。</Trans>
						</p>
					</div>
					<Button variant="outline" onClick={() => setDraft(defaultTelegramDestination())}>
						<PlusIcon className="size-4" />
						<span className="ms-1">
							<Trans>New Destination</Trans>
						</span>
					</Button>
				</div>
				<div className="grid gap-3 md:grid-cols-2">
					<div className="space-y-2">
						<Label htmlFor="telegram-destination-name">
							<Trans>Name</Trans>
						</Label>
						<Input
							id="telegram-destination-name"
							value={draft.name}
							onChange={(event) => setDraft((previous) => ({ ...previous, name: event.target.value }))}
						/>
					</div>
					<div className="space-y-2">
						<Label htmlFor="telegram-chat-id">
							<Trans>Chat ID</Trans>
						</Label>
						<Input
							id="telegram-chat-id"
							value={draft.chatId}
							onChange={(event) => setDraft((previous) => ({ ...previous, chatId: event.target.value }))}
						/>
					</div>
					<div className="space-y-2">
						<Label>
							<Trans>Chat Type</Trans>
						</Label>
						<Select
							value={draft.chatType}
							onValueChange={(value: TelegramDestination["chatType"]) =>
								setDraft((previous) => ({ ...previous, chatType: value }))
							}
						>
							<SelectTrigger>
								<SelectValue />
							</SelectTrigger>
							<SelectContent>
								<SelectItem value="private">Private</SelectItem>
								<SelectItem value="group">Group</SelectItem>
								<SelectItem value="supergroup">Supergroup</SelectItem>
								<SelectItem value="channel">Channel</SelectItem>
								<SelectItem value="unknown">Unknown</SelectItem>
							</SelectContent>
						</Select>
					</div>
					<div className="space-y-2">
						<Label>
							<Trans>Role</Trans>
						</Label>
						<Select
							value={draft.role}
							onValueChange={(value: TelegramDestination["role"]) =>
								setDraft((previous) => ({ ...previous, role: value }))
							}
						>
							<SelectTrigger>
								<SelectValue />
							</SelectTrigger>
							<SelectContent>
								<SelectItem value="admin">Admin</SelectItem>
								<SelectItem value="read_only">Read-only</SelectItem>
							</SelectContent>
						</Select>
					</div>
				</div>
				<div className="grid gap-4 lg:grid-cols-[1.1fr_1fr]">
					<div className="space-y-2">
						<Label>
							<Trans>Alert Scope Tags</Trans>
						</Label>
						<InputTags
							value={draft.alertLevelScope}
							onChange={(value) => setDraft((previous) => ({ ...previous, alertLevelScope: value }))}
							placeholder={t`status, cpu, memory ...`}
							className="w-full"
						/>
						<p className="text-[0.8rem] text-muted-foreground">
							<Trans>只读渠道可按告警类别限制投递；留空表示不过滤。</Trans>
						</p>
					</div>
					<div className="space-y-2">
						<Label>
							<Trans>Node Scope</Trans>
						</Label>
						<div className="max-h-40 space-y-2 overflow-auto rounded-md border p-3">
							{systems.map((system) => (
								<div key={system.id} className="flex items-center gap-2 text-sm">
									<Checkbox
										id={`tg-node-${system.id}`}
										checked={selectedSystems.has(system.id)}
										onCheckedChange={(checked) =>
											setDraft((previous) => ({
												...previous,
												nodeScope: checked
													? [...previous.nodeScope, system.id]
													: previous.nodeScope.filter((item) => item !== system.id),
											}))
										}
									/>
									<Label htmlFor={`tg-node-${system.id}`}>{system.name}</Label>
								</div>
							))}
							{systems.length === 0 && (
								<div className="text-sm text-muted-foreground">
									<Trans>No systems available.</Trans>
								</div>
							)}
						</div>
					</div>
				</div>
				<div className="flex flex-wrap items-center justify-between gap-3">
					<div className="flex items-center gap-2 text-sm">
						<Switch
							id="telegram-destination-enabled"
							checked={draft.enabled}
							onCheckedChange={(checked) => setDraft((previous) => ({ ...previous, enabled: checked }))}
						/>
						<Label htmlFor="telegram-destination-enabled">
							<Trans>Enabled</Trans>
						</Label>
					</div>
					<Button onClick={saveDestination} disabled={!hasSettingsReady}>
						<SaveIcon className="size-4" />
						<span className="ms-1">
							<Trans>Save Destination</Trans>
						</span>
					</Button>
				</div>
				<Separator />
				<div className="space-y-2">
					{destinations.map((destination) => (
						<Card key={destination.id} className="bg-table-header p-3">
							<div className="flex flex-wrap items-start justify-between gap-3">
								<button
									type="button"
									className="min-w-0 flex-1 text-start"
									onClick={() =>
										setDraft({
											id: destination.id,
											userId: destination.userId,
											name: destination.name,
											chatId: destination.chatId,
											chatType: destination.chatType,
											role: destination.role,
											enabled: destination.enabled,
											nodeScope: destination.nodeScope,
											alertLevelScope: destination.alertLevelScope,
											muteUntil: destination.muteUntil,
										})
									}
								>
									<div className="truncate font-medium">{destination.name}</div>
									<div className="truncate text-sm text-muted-foreground">
										{destination.chatId} · {destination.chatType} · {destination.role}
									</div>
									{destination.lastError && (
										<div className="mt-1 text-sm text-destructive">{destination.lastError}</div>
									)}
								</button>
								<div className="flex items-center gap-2">
									<Button
										variant="outline"
										size="sm"
										onClick={() => sendDestinationTest(destination)}
										disabled={busyDestinationId === destination.id || !hasSettingsReady}
									>
										{busyDestinationId === destination.id ? (
											<LoaderCircleIcon className="size-4 animate-spin" />
										) : (
											<SendIcon className="size-4" />
										)}
										<span className="ms-1">
											<Trans>Test</Trans>
										</span>
									</Button>
									<Button
										variant="outline"
										size="icon"
										onClick={() => removeDestination(destination)}
										disabled={busyDestinationId === destination.id}
									>
										<Trash2Icon className="size-4" />
									</Button>
								</div>
							</div>
						</Card>
					))}
					{destinations.length === 0 && (
						<div className="rounded-md border border-dashed p-6 text-center text-sm text-muted-foreground">
							<Trans>No Telegram destinations yet.</Trans>
						</div>
					)}
				</div>
			</div>
		</div>
	)
}
