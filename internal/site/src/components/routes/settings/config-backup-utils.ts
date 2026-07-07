import * as v from "valibot"
import type { ConfigBackupExportRequest, ConfigBackupPreviewResponse, ConfigBackupSection } from "@/types"

export const configBackupSections: ConfigBackupSection[] = [
	"systems",
	"alerts",
	"notifications",
	"publicStatus",
	"networkProbes",
]

export const ConfigBackupExportSchema = v.object({
	includeSecrets: v.boolean(),
	encryptionCredential: v.string(),
	sections: v.array(v.picklist(configBackupSections)),
})

export function buildConfigBackupExportPayload(input: Partial<ConfigBackupExportRequest>): ConfigBackupExportRequest {
	const payload = {
		includeSecrets: input.includeSecrets ?? true,
		encryptionCredential: input.encryptionCredential?.trim() ?? "",
		sections: input.sections?.length ? input.sections : configBackupSections,
	}
	if (payload.includeSecrets && payload.encryptionCredential === "") {
		throw new Error("导出敏感配置时需要填写加密凭据")
	}
	return v.parse(ConfigBackupExportSchema, payload)
}

export function previewHasBlockingIssues(preview?: ConfigBackupPreviewResponse) {
	if (!preview) {
		return true
	}
	return preview.summary.conflict > 0 || preview.summary.error > 0
}

export function configBackupActionLabel(action: ConfigBackupPreviewResponse["items"][number]["action"]) {
	switch (action) {
		case "create":
			return "创建"
		case "update":
			return "更新"
		case "preserve":
			return "保留"
		case "skip":
			return "跳过"
		case "conflict":
			return "冲突"
		case "error":
			return "错误"
		default:
			return action
	}
}

export function toggleConfigBackupSection(
	sections: ConfigBackupSection[],
	section: ConfigBackupSection,
	enabled: boolean
) {
	if (enabled) {
		return sections.includes(section) ? sections : [...sections, section]
	}
	return sections.filter((item) => item !== section)
}
