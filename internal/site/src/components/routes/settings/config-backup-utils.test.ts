import assert from "node:assert/strict"
import { describe, it } from "node:test"
import {
	buildConfigBackupExportPayload,
	configBackupSections,
	previewHasBlockingIssues,
	toggleConfigBackupSection,
} from "./config-backup-utils.ts"

describe("buildConfigBackupExportPayload", () => {
	it("requires an encryption credential when secrets are included", () => {
		assert.throws(() => buildConfigBackupExportPayload({ includeSecrets: true }), /加密凭据/)
	})

	it("defaults to all backup sections", () => {
		const payload = buildConfigBackupExportPayload({
			includeSecrets: false,
			encryptionCredential: "",
		})

		assert.deepEqual(payload.sections, configBackupSections)
	})
})

describe("previewHasBlockingIssues", () => {
	it("blocks restore when conflicts or errors are present", () => {
		assert.equal(previewHasBlockingIssues(undefined), true)
		assert.equal(
			previewHasBlockingIssues({
				previewId: "sha256:test",
				mode: "merge",
				backupMeta: {
					backupVersion: "1",
					sourceVersion: "test",
					createdAt: "2026-07-06T00:00:00Z",
					mode: "merge",
					sections: [],
				},
				summary: { create: 0, update: 0, preserve: 0, skip: 0, conflict: 1, error: 0 },
				items: [],
				warnings: [],
				requiresCredential: false,
			}),
			true
		)
	})
})

describe("toggleConfigBackupSection", () => {
	it("adds and removes sections idempotently", () => {
		assert.deepEqual(toggleConfigBackupSection(["systems"], "alerts", true), ["systems", "alerts"])
		assert.deepEqual(toggleConfigBackupSection(["systems"], "systems", true), ["systems"])
		assert.deepEqual(toggleConfigBackupSection(["systems", "alerts"], "alerts", false), ["systems"])
	})
})
