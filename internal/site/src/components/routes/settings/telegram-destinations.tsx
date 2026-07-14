import { t } from "@lingui/core/macro"
import { Trans } from "@lingui/react/macro"
import { useStore } from "@nanostores/react"
import { LoaderCircleIcon, PlusIcon, SaveIcon, SendIcon, Trash2Icon } from "lucide-react"
import { useEffect, useMemo, useState } from "react"
import { Badge } from "@/components/ui/badge"
import {
	AlertDialog,
	AlertDialogAction,
	AlertDialogCancel,
	AlertDialogContent,
	AlertDialogDescription,
	AlertDialogFooter,
	AlertDialogHeader,
	AlertDialogTitle,
} from "@/components/ui/alert-dialog"
import { Button } from "@/components/ui/button"
import { Card } from "@/components/ui/card"
import { Checkbox } from "@/components/ui/checkbox"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select"
import { Separator } from "@/components/ui/separator"
import { Switch } from "@/components/ui/switch"
import { toast } from "@/components/ui/use-toast"
import {
	deleteTelegramDestination,
	deleteTelegramNotificationPolicy,
	getTelegramDestinations,
	getTelegramNotificationPolicies,
	saveTelegramDestination,
	saveTelegramNotificationPolicy,
	testTelegramDestination,
	testTelegramSettings,
} from "@/lib/api"
import { $systems } from "@/lib/stores"
import type {
	TelegramAlertScope,
	TelegramDestination,
	TelegramDestinationInput,
	TelegramNotificationPolicy,
	TelegramNotificationPolicyInput,
	TelegramSettingsInput,
} from "@/types"
import {
	TELEGRAM_ALERT_SCOPE_OPTIONS,
	buildTelegramBotTestPayload,
	buildTelegramChannelPayload,
	buildTelegramPolicyPayload,
	buildTelegramSettingsPayload,
	clearTelegramPolicyEditorForDeletedDestination,
	defaultTelegramDestination,
	defaultTelegramPolicy,
	formatTelegramBotTestResult,
	getTelegramBotHealth,
	getTelegramDestinationHealth,
	getExistingTelegramDestinationID,
	maskTelegramChatID,
	removeTelegramDestinationByID,
	searchTelegramSystems,
	selectAllTelegramSystems,
} from "./telegram-utils"

function destinationHealthVariant(status: ReturnType<typeof getTelegramDestinationHealth>["status"]) {
	if (status === "error") return "danger" as const
	if (status === "healthy") return "success" as const
	if (status === "muted") return "warning" as const
	return "secondary" as const
}

function telegramDestinationHealthLabel(status: ReturnType<typeof getTelegramDestinationHealth>["status"]) {
	if (status === "disabled") return t`Disabled`
	if (status === "muted") return t`Muted`
	if (status === "error") return t`Delivery error`
	if (status === "healthy") return t`Delivery healthy`
	return t`Waiting for test`
}

function telegramBotHealthLabel(status: ReturnType<typeof getTelegramBotHealth>["status"], hasUsername: boolean) {
	if (status === "disabled") return t`Disabled`
	if (status === "error") return t`Connection error`
	if (status === "pending") return t`Waiting for configuration`
	return hasUsername ? t`Running normally` : t`Credentials configured`
}

function telegramRoleDescription(role: TelegramDestination["role"]) {
	return role === "admin"
		? t`Admin channels receive full alert details within policy scope. Only private admins can use the management menu.`
		: t`Read-only channels receive sanitized monitoring messages within policy scope and cannot run management commands.`
}

function telegramChatCapabilityText(role: TelegramDestination["role"], chatType: TelegramDestination["chatType"]) {
	if (role !== "admin") return t`The read-only role cannot use the management menu.`
	if (chatType !== "private") {
		return t`Groups, supergroups, and channels can receive full notifications but cannot use the management menu.`
	}
	return t`This private admin can use the Bot management menu.`
}

function telegramAlertScopeLabel(scope: TelegramAlertScope) {
	switch (scope) {
		case "status":
			return t`Node status`
		case "cpu":
			return t`CPU`
		case "memory":
			return t`Memory`
		case "disk":
			return t`Disk`
		case "temperature":
			return t`Temperature`
		case "bandwidth":
			return t`Bandwidth`
		case "gpu":
			return t`GPU`
		case "loadavg1":
			return t`System load (1-minute average)`
		case "loadavg5":
			return t`System load (5-minute average)`
		case "loadavg15":
			return t`System load (15-minute average)`
		case "battery":
			return t`Battery`
		case "smart":
			return t`S.M.A.R.T.`
	}
}

function telegramBotStageText(stage: { ok: boolean; error: string }, stageName: "credentials" | "commandMenu") {
	if (stageName === "credentials") {
		return stage.ok ? t`Bot credentials verified` : t`Bot credential verification failed: ${stage.error}`
	}
	return stage.ok ? t`Command menu initialized` : t`Command menu initialization failed: ${stage.error}`
}

function TelegramDestinationHealth({ destination }: { destination: TelegramDestination }) {
	const health = getTelegramDestinationHealth(destination)
	return (
		<div className="mt-2 space-y-1 text-xs text-muted-foreground">
			<div className="flex flex-wrap items-center gap-2">
				<Badge variant={destinationHealthVariant(health.status)}>{telegramDestinationHealthLabel(health.status)}</Badge>
				<span>
					<Trans>Last test: {health.lastTestAt || t`Not tested yet`}</Trans>
				</span>
				<span>
					<Trans>Last delivery: {health.lastDeliveryAt || t`No delivery yet`}</Trans>
				</span>
			</div>
			{health.status === "muted" && (
				<div>
					<Trans>Muted until: {health.muteUntil || t`No mute deadline`}</Trans>
				</div>
			)}
			{health.error && (
				<div className="text-destructive">
					<Trans>Error: {health.error}</Trans>
				</div>
			)}
		</div>
	)
}

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
	const [botTestStatus, setBotTestStatus] = useState<ReturnType<typeof formatTelegramBotTestResult>>()
	const [deleteCandidate, setDeleteCandidate] = useState<TelegramDestination>()
	const [policyDestination, setPolicyDestination] = useState<TelegramDestination>()
	const [policies, setPolicies] = useState<TelegramNotificationPolicy[]>([])
	const [policyDraft, setPolicyDraft] = useState<TelegramNotificationPolicyInput & { id?: string }>(
		defaultTelegramPolicy()
	)
	const [policySearch, setPolicySearch] = useState("")

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
	const botHealth = getTelegramBotHealth(settings)
	const selectedPolicySystems = useMemo(() => new Set(policyDraft.nodeScope), [policyDraft.nodeScope])
	const visiblePolicySystems = useMemo(() => searchTelegramSystems(systems, policySearch), [systems, policySearch])

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
			const result = await testTelegramSettings(buildTelegramBotTestPayload(settings))
			const staged = formatTelegramBotTestResult(result)
			setBotTestStatus(staged)
			toast({
				title: result.ok ? t`Telegram bot verified` : t`Telegram bot verification incomplete`,
				description: result.ok
					? result.botUsername || telegramBotStageText(staged.commandMenu, "commandMenu")
					: telegramBotStageText(staged.commandMenu, "commandMenu"),
				variant: result.ok ? "default" : "destructive",
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
			const payload = buildTelegramChannelPayload(draft)
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
			const existingID = getExistingTelegramDestinationID(error)
			const existing = destinations.find((destination) => destination.id === existingID)
			if (existing) {
				setDraft({
					...existing,
					nodeScope: existing.nodeScope ?? [],
					alertLevelScope: existing.alertLevelScope ?? [],
				})
			}
			toast({
				title: t`Failed to save Telegram destination`,
				description: existing ? t`This Chat ID already exists. The existing channel is now selected.` : error.message,
				variant: "destructive",
			})
		}
	}

	async function openPolicyManager(destination: TelegramDestination) {
		setPolicyDestination(destination)
		setPolicyDraft(defaultTelegramPolicy())
		setPolicySearch("")
		try {
			const response = await getTelegramNotificationPolicies(destination.id)
			setPolicies(response.policies)
		} catch (error: any) {
			toast({ title: t`Failed to load notification policies`, description: error.message, variant: "destructive" })
		}
	}

	async function savePolicy() {
		if (!policyDestination) return
		try {
			const saved = await saveTelegramNotificationPolicy(policyDestination.id, {
				...buildTelegramPolicyPayload(policyDraft),
				id: policyDraft.id,
			})
			setPolicies((previous) =>
				policyDraft.id ? previous.map((policy) => (policy.id === saved.id ? saved : policy)) : [...previous, saved]
			)
			if (!policyDraft.id) {
				setDestinations((previous) =>
					previous.map((destination) =>
						destination.id === policyDestination.id
							? { ...destination, policyCount: (destination.policyCount ?? 0) + 1 }
							: destination
					)
				)
			}
			setPolicyDraft(defaultTelegramPolicy())
			toast({ title: t`Notification policy saved`, description: saved.name })
		} catch (error: any) {
			toast({
				title: t`Failed to save notification policy`,
				description:
					error.message === "selected_node_required"
						? t`Selected-node mode requires at least one node.`
						: error.message,
				variant: "destructive",
			})
		}
	}

	async function removePolicy(policy: TelegramNotificationPolicy) {
		if (!policyDestination) return
		try {
			await deleteTelegramNotificationPolicy(policyDestination.id, policy.id)
			setPolicies((previous) => previous.filter((item) => item.id !== policy.id))
			setDestinations((previous) =>
				previous.map((destination) =>
					destination.id === policyDestination.id
						? { ...destination, policyCount: Math.max(0, (destination.policyCount ?? 1) - 1) }
						: destination
				)
			)
			if (policyDraft.id === policy.id) setPolicyDraft(defaultTelegramPolicy())
		} catch (error: any) {
			toast({ title: t`Failed to delete notification policy`, description: error.message, variant: "destructive" })
		}
	}

	async function removeDestination(destination: TelegramDestination) {
		setBusyDestinationId(destination.id)
		try {
			await deleteTelegramDestination(destination.id)
			setDestinations((previous) => removeTelegramDestinationByID(previous, destination.id))
			if (draft.id === destination.id) {
				setDraft(defaultTelegramDestination())
			}
			const policyEditor = clearTelegramPolicyEditorForDeletedDestination(
				{ destination: policyDestination, policies, draft: policyDraft, search: policySearch },
				destination.id
			)
			setPolicyDestination(policyEditor.destination)
			setPolicies(policyEditor.policies)
			setPolicyDraft(policyEditor.draft)
			setPolicySearch(policyEditor.search)
			setDeleteCandidate(undefined)
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
						<div className="flex flex-wrap items-center gap-2">
							<Badge
								variant={
									botHealth.status === "error" ? "danger" : botHealth.status === "healthy" ? "success" : "secondary"
								}
							>
								{telegramBotHealthLabel(botHealth.status, !!settings.botUsername)}
							</Badge>
							{settings.botUsername && <span className="text-sm text-muted-foreground">@{settings.botUsername}</span>}
						</div>
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
						{botHealth.error && <div className="text-sm text-destructive">{botHealth.error}</div>}
						{botTestStatus && (
							<div className="space-y-1 rounded-md border bg-background p-3 text-sm" aria-live="polite">
								<div>{telegramBotStageText(botTestStatus.credentials, "credentials")}</div>
								<div>{telegramBotStageText(botTestStatus.commandMenu, "commandMenu")}</div>
							</div>
						)}
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
							<Trans>维护允许接收告警的私聊、群组、频道或超群，并通过通知规则限制节点和告警范围。</Trans>
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
								<SelectItem value="admin">
									<Trans>Administrator</Trans>
								</SelectItem>
								<SelectItem value="read_only">
									<Trans>Read-only notifications</Trans>
								</SelectItem>
							</SelectContent>
						</Select>
					</div>
				</div>
				<div className="rounded-md border bg-muted/30 p-3 text-sm text-muted-foreground">
					<div>{telegramRoleDescription(draft.role)}</div>
					<div>{telegramChatCapabilityText(draft.role, draft.chatType)}</div>
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
									<TelegramDestinationHealth destination={destination} />
								</button>
								<div className="flex items-center gap-2">
									<Button variant="outline" size="sm" onClick={() => openPolicyManager(destination)}>
										<Trans>Policies: {destination.policyCount ?? 0}</Trans>
									</Button>
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
										size="sm"
										title={t`Delete Telegram destination`}
										onClick={() => setDeleteCandidate(destination)}
										disabled={busyDestinationId === destination.id}
									>
										<Trash2Icon className="size-4" />
										<span className="ms-1">
											<Trans>Delete</Trans>
										</span>
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
				{policyDestination && (
					<div className="space-y-4 border-t pt-4">
						<div className="flex flex-wrap items-center justify-between gap-3">
							<div>
								<h4 className="font-medium">
									<Trans>Notification policies for {policyDestination.name}</Trans>
								</h4>
								<p className="text-sm text-muted-foreground">
									<Trans>A matching enabled policy delivers once, even when policies overlap.</Trans>
								</p>
							</div>
							<Button variant="outline" size="sm" onClick={() => setPolicyDraft(defaultTelegramPolicy())}>
								<PlusIcon className="size-4" />
								<Trans>New policy</Trans>
							</Button>
						</div>
						<div className="grid gap-3 md:grid-cols-2">
							<div className="space-y-2">
								<Label htmlFor="telegram-policy-name">
									<Trans>Policy name</Trans>
								</Label>
								<Input
									id="telegram-policy-name"
									value={policyDraft.name}
									onChange={(event) => setPolicyDraft((previous) => ({ ...previous, name: event.target.value }))}
								/>
							</div>
							<div className="space-y-2">
								<Label>
									<Trans>Node scope</Trans>
								</Label>
								<Select
									value={policyDraft.nodeScopeMode}
									onValueChange={(value: "all" | "selected") =>
										setPolicyDraft((previous) => ({
											...previous,
											nodeScopeMode: value,
											nodeScope: value === "all" ? [] : previous.nodeScope,
										}))
									}
								>
									<SelectTrigger>
										<SelectValue />
									</SelectTrigger>
									<SelectContent>
										<SelectItem value="all">
											<Trans>All nodes (including future nodes)</Trans>
										</SelectItem>
										<SelectItem value="selected">
											<Trans>Selected nodes</Trans>
										</SelectItem>
									</SelectContent>
								</Select>
							</div>
						</div>
						{policyDraft.nodeScopeMode === "selected" && (
							<div className="space-y-2">
								<div className="flex flex-wrap items-center gap-2">
									<Input
										className="max-w-sm"
										placeholder={t`Search nodes`}
										value={policySearch}
										onChange={(event) => setPolicySearch(event.target.value)}
									/>
									<Button
										variant="outline"
										size="sm"
										onClick={() =>
											setPolicyDraft((previous) => ({
												...previous,
												nodeScope: selectAllTelegramSystems(previous.nodeScope, visiblePolicySystems),
											}))
										}
									>
										<Trans>Select all current results</Trans>
									</Button>
									<Button
										variant="outline"
										size="sm"
										onClick={() => setPolicyDraft((previous) => ({ ...previous, nodeScope: [] }))}
									>
										<Trans>Clear</Trans>
									</Button>
									<span className="text-sm text-muted-foreground">
										<Trans>{policyDraft.nodeScope.length} nodes selected</Trans>
									</span>
								</div>
								<div className="grid max-h-52 grid-cols-2 gap-2 overflow-auto rounded-md border p-3 md:grid-cols-3">
									{visiblePolicySystems.map((system) => (
										<div key={system.id} className="flex items-center gap-2 text-sm">
											<Checkbox
												id={`telegram-policy-node-${system.id}`}
												checked={selectedPolicySystems.has(system.id)}
												onCheckedChange={(checked) =>
													setPolicyDraft((previous) => ({
														...previous,
														nodeScope: checked
															? [...new Set([...previous.nodeScope, system.id])]
															: previous.nodeScope.filter((id) => id !== system.id),
													}))
												}
											/>
											<Label htmlFor={`telegram-policy-node-${system.id}`}>{system.name}</Label>
										</div>
									))}
								</div>
							</div>
						)}
						<div className="space-y-2">
							<Label>
								<Trans>Alert categories</Trans>
							</Label>
							<div className="grid max-h-52 grid-cols-2 gap-2 overflow-auto rounded-md border p-3 md:grid-cols-3">
								{TELEGRAM_ALERT_SCOPE_OPTIONS.map((option) => (
									<div key={option.value} className="flex items-center gap-2 text-sm">
										<Checkbox
											id={`telegram-policy-alert-${option.value}`}
											checked={policyDraft.alertLevelScope.includes(option.value)}
											onCheckedChange={(checked) =>
												setPolicyDraft((previous) => ({
													...previous,
													alertLevelScope: checked
														? [...previous.alertLevelScope, option.value]
														: previous.alertLevelScope.filter((scope: TelegramAlertScope) => scope !== option.value),
												}))
											}
										/>
										<Label htmlFor={`telegram-policy-alert-${option.value}`}>
											{telegramAlertScopeLabel(option.value)}
										</Label>
									</div>
								))}
							</div>
							<p className="text-xs text-muted-foreground">
								<Trans>
									Leave empty to include all alert categories. System load is the average number of runnable tasks over
									1, 5, and 15 minutes. It is not a percentage or alert duration; compare it with the node's logical CPU
									core count.
								</Trans>
							</p>
						</div>
						<div className="flex flex-wrap items-center justify-between gap-3">
							<div className="flex items-center gap-2">
								<Switch
									id="telegram-policy-enabled"
									checked={policyDraft.enabled}
									onCheckedChange={(enabled) => setPolicyDraft((previous) => ({ ...previous, enabled }))}
								/>
								<Label htmlFor="telegram-policy-enabled">
									<Trans>Enable policy</Trans>
								</Label>
							</div>
							<Button onClick={savePolicy}>
								<SaveIcon className="size-4" />
								<Trans>Save policy</Trans>
							</Button>
						</div>
						<div className="space-y-2">
							{policies.map((policy) => (
								<div
									key={policy.id}
									className="flex flex-wrap items-center justify-between gap-2 rounded-md border p-3"
								>
									<button
										type="button"
										className="min-w-0 flex-1 text-start"
										onClick={() =>
											setPolicyDraft({
												id: policy.id,
												name: policy.name,
												enabled: policy.enabled,
												nodeScopeMode: policy.nodeScopeMode,
												nodeScope: policy.nodeScope,
												alertLevelScope: policy.alertLevelScope,
											})
										}
									>
										<div className="font-medium">{policy.name}</div>
										<div className="text-xs text-muted-foreground">
											{policy.enabled ? t`Enabled` : t`Disabled`} ·{" "}
											{policy.nodeScopeMode === "all" ? t`All nodes` : t`${policy.nodeScope.length} selected nodes`}
										</div>
									</button>
									<Button variant="outline" size="sm" onClick={() => removePolicy(policy)}>
										<Trash2Icon className="size-4" />
										<Trans>Delete policy</Trans>
									</Button>
								</div>
							))}
							{policies.length === 0 && (
								<div className="text-sm text-muted-foreground">
									<Trans>This channel has no notification policies.</Trans>
								</div>
							)}
						</div>
					</div>
				)}
			</div>
			<AlertDialog open={!!deleteCandidate} onOpenChange={(open) => !open && setDeleteCandidate(undefined)}>
				<AlertDialogContent>
					<AlertDialogHeader>
						<AlertDialogTitle>
							<Trans>Delete Telegram notification channel?</Trans>
						</AlertDialogTitle>
						<AlertDialogDescription>
							<Trans>
								Delete {deleteCandidate?.name} (Chat ID {maskTelegramChatID(deleteCandidate?.chatId ?? "")}) and its{" "}
								{deleteCandidate?.policyCount ?? 0} notification policies. Bot settings and other channels are not
								affected.
							</Trans>
						</AlertDialogDescription>
					</AlertDialogHeader>
					<AlertDialogFooter>
						<AlertDialogCancel>
							<Trans>Cancel</Trans>
						</AlertDialogCancel>
						<AlertDialogAction
							className="bg-destructive text-destructive-foreground hover:bg-destructive/90"
							disabled={!deleteCandidate || busyDestinationId === deleteCandidate.id}
							onClick={() => deleteCandidate && removeDestination(deleteCandidate)}
						>
							<Trans>Confirm delete</Trans>
						</AlertDialogAction>
					</AlertDialogFooter>
				</AlertDialogContent>
			</AlertDialog>
		</div>
	)
}
