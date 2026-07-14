//go:build testing

package hub

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/tools/router"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConfigBackupPreviewWarnsAndSkipsUnknownAndNewerSections(t *testing.T) {
	hub, _ := newTelegramHubWithAdmin(t)
	content := strings.TrimSpace(`
meta:
  backupVersion: "1"
  sourceVersion: test
  createdAt: "2026-07-10T00:00:00Z"
  mode: merge
  sections:
    - systems
    - futureSection
  sectionVersions:
    systems: "999"
    futureSection: "1"
systems:
  - stableId: sysbackup000001
    name: must-not-restore
    host: 10.0.0.1
futureSection:
  enabled: true
`) + "\n"

	document, err := parseConfigBackupDocument(content)
	require.NoError(t, err)
	preview, err := hub.previewConfigBackup(document, content, "")
	require.NoError(t, err)
	assert.Equal(t, 2, preview.Summary.Skip)
	assert.Contains(t, strings.Join(preview.Warnings, "\n"), "futureSection")
	assert.Contains(t, strings.Join(preview.Warnings, "\n"), "systems")

	applied, warnings, err := hub.applyConfigBackup(document, "")
	require.NoError(t, err)
	assert.Zero(t, applied.Created)
	assert.Zero(t, applied.Updated)
	assert.NotEmpty(t, warnings)
	_, err = hub.FindRecordById("systems", "sysbackup000001")
	require.Error(t, err)
}

func TestConfigBackupPreviewAndMergeRestorePreservesTargetOnlyRecords(t *testing.T) {
	hub, admin := newTelegramHubWithAdmin(t)
	targetOnly := mustCreateTelegramRecord(t, hub, "systems", map[string]any{
		"name":   "target-only",
		"host":   "127.0.0.1",
		"port":   "45876",
		"status": "up",
		"users":  []string{admin.Id},
	})
	secret, err := encryptConfigBackupSecret("restored-token", "backup-pass", "system.token")
	require.NoError(t, err)
	document := ConfigBackupDocument{
		Meta: ConfigBackupMeta{
			BackupVersion: ConfigBackupVersion,
			Mode:          ConfigBackupMode,
			Sections:      []string{ConfigBackupSectionSystems, ConfigBackupSectionNetworkProbes, ConfigBackupSectionPublicStatus},
		},
		Systems: []ConfigBackupSystem{
			{
				StableID: "sysbackup000001",
				Name:     "restored-node",
				Host:     "10.0.0.2",
				Port:     "45876",
				Status:   "up",
				Users:    []ConfigBackupUserRef{{Email: admin.GetString("email")}},
				Token:    secret,
			},
		},
		NetworkProbes: ConfigBackupNetworkProbes{
			Probes: []ConfigBackupNetworkProbe{
				{
					StableID:        "probebackup0001",
					Name:            "Guangdong Telecom",
					Type:            "tcping",
					Target:          "example.com:443",
					IntervalSeconds: 20,
					TimeoutSeconds:  3,
					Enabled:         true,
					Scope:           NetworkProbeScopeGlobal,
				},
			},
			Assignments: []ConfigBackupNetworkProbeAssignment{
				{ProbeStableID: "probebackup0001", SystemStableID: "sysbackup000001", Enabled: true},
			},
		},
		PublicStatus: ConfigBackupPublicStatus{
			Systems: []ConfigBackupPublicSystem{
				{
					SystemStableID:       "sysbackup000001",
					PublicEnabled:        true,
					PublicName:           "Public restored node",
					ShowCPU:              true,
					ShowMemory:           true,
					ShowDisk:             true,
					PublicProbeStableIDs: []string{"probebackup0001"},
				},
			},
		},
	}
	content, err := marshalConfigBackupDocument(document)
	require.NoError(t, err)

	preview, err := hub.previewConfigBackup(document, content, "backup-pass")
	require.NoError(t, err)
	assert.Zero(t, preview.Summary.Conflict)
	assert.Zero(t, preview.Summary.Error)
	assert.GreaterOrEqual(t, preview.Summary.Create, 2)
	assert.GreaterOrEqual(t, preview.Summary.Preserve, 1)

	applied, _, err := hub.applyConfigBackup(document, "backup-pass")
	require.NoError(t, err)
	assert.GreaterOrEqual(t, applied.Created, 3)

	_, err = hub.FindRecordById("systems", targetOnly.Id)
	require.NoError(t, err)
	restored, err := hub.FindRecordById("systems", "sysbackup000001")
	require.NoError(t, err)
	assert.Equal(t, "restored-node", restored.GetString("name"))

	fingerprint, err := hub.fingerprintBySystem("sysbackup000001")
	require.NoError(t, err)
	assert.Equal(t, "restored-token", fingerprint.GetString("token"))

	visibility, publicRecord := hub.findPublicVisibility("sysbackup000001")
	require.NotNil(t, publicRecord)
	assert.True(t, visibility.PublicEnabled)
	assert.Contains(t, visibility.PublicProbeIDs, "probebackup0001")
}

func TestConfigBackupPreviewBlocksMissingCredential(t *testing.T) {
	hub, _ := newTelegramHubWithAdmin(t)
	secret, err := encryptConfigBackupSecret("token", "backup-pass", "system.token")
	require.NoError(t, err)
	document := ConfigBackupDocument{
		Meta: ConfigBackupMeta{BackupVersion: ConfigBackupVersion, Mode: ConfigBackupMode},
		Systems: []ConfigBackupSystem{
			{StableID: "sysbackup000001", Name: "node", Host: "10.0.0.1", Token: secret},
		},
	}
	content, err := marshalConfigBackupDocument(document)
	require.NoError(t, err)

	preview, err := hub.previewConfigBackup(document, content, "")
	require.NoError(t, err)
	assert.True(t, preview.RequiresCredential)
	assert.Equal(t, 1, preview.Summary.Error)
}

func TestConfigBackupRestorePreservesRedactedAndOmittedSystemToken(t *testing.T) {
	hub, admin := newTelegramHubWithAdmin(t)
	system := mustCreateTelegramRecord(t, hub, "systems", map[string]any{
		"name": "existing", "host": "10.0.0.1", "port": "45876", "status": "up", "users": []string{admin.Id},
	})
	mustCreateTelegramRecord(t, hub, "fingerprints", map[string]any{
		"system": system.Id, "fingerprint": "existingfp", "token": "keep-system-token",
	})

	for name, secret := range map[string]*ConfigBackupSecret{
		"redacted": {Redacted: true},
		"omitted":  nil,
	} {
		t.Run(name, func(t *testing.T) {
			document := ConfigBackupDocument{
				Meta: ConfigBackupMeta{BackupVersion: ConfigBackupVersion, Mode: ConfigBackupMode, Sections: []string{ConfigBackupSectionSystems}},
				Systems: []ConfigBackupSystem{{
					StableID: system.Id, Name: "updated", Host: "10.0.0.2", Port: "45876", Status: "up",
					Users: []ConfigBackupUserRef{{Email: admin.GetString("email")}}, Token: secret,
				}},
			}
			_, _, err := hub.applyConfigBackup(document, "")
			require.NoError(t, err)
			fingerprint, err := hub.fingerprintBySystem(system.Id)
			require.NoError(t, err)
			assert.Equal(t, "keep-system-token", fingerprint.GetString("token"))
		})
	}
}

func TestConfigBackupSectionRestoreRollsBackFailedSection(t *testing.T) {
	hub, admin := newTelegramHubWithAdmin(t)
	document := ConfigBackupDocument{
		Meta: ConfigBackupMeta{BackupVersion: ConfigBackupVersion, Mode: ConfigBackupMode, Sections: []string{ConfigBackupSectionSystems}},
		Systems: []ConfigBackupSystem{
			{StableID: "sysbackup000001", Name: "created-before-failure", Host: "10.0.0.1", Users: []ConfigBackupUserRef{{Email: admin.GetString("email")}}},
			{StableID: "sysbackup000002", Name: "fails", Host: "10.0.0.2", Users: []ConfigBackupUserRef{{Email: "missing@example.com"}}},
		},
	}

	_, _, err := hub.applyConfigBackup(document, "")
	require.Error(t, err)
	_, findErr := hub.FindRecordById("systems", "sysbackup000001")
	require.Error(t, findErr)
}

func TestConfigBackupPreviewEnumeratesTargetOnlyRecordsOnceWithEmptyImports(t *testing.T) {
	hub, admin := newTelegramHubWithAdmin(t)
	system := mustCreateTelegramRecord(t, hub, "systems", map[string]any{"name": "target system", "host": "127.0.0.1", "port": "45876", "status": "up", "users": []string{admin.Id}})
	mustCreateTelegramRecord(t, hub, "alerts", map[string]any{"name": "Status", "user": admin.Id, "system": system.Id})
	mustCreateTelegramRecord(t, hub, "quiet_hours", map[string]any{"user": admin.Id, "system": system.Id, "type": "one-time", "start": time.Now().UTC().Add(-time.Hour), "end": time.Now().UTC().Add(time.Hour)})
	mustCreateTelegramRecord(t, hub, "user_settings", map[string]any{"user": admin.Id, "settings": map[string]any{"emails": []string{"ops@example.com"}}})
	_, err := hub.saveTelegramSettings(TelegramSettingsInput{Enabled: true, PollingEnabled: true, BotToken: "123456:preview_token"}, telegramSettingsRecord{})
	require.NoError(t, err)
	mustCreateTelegramRecord(t, hub, CollectionTelegramDestinations, map[string]any{"name": "target destination", "chat_id": "123", "chat_type": TelegramChatTypePrivate, "role": TelegramRoleReadOnly, "enabled": true})
	mustCreateTelegramRecord(t, hub, CollectionPublicSystemVisibility, map[string]any{"system": system.Id, "public_enabled": true})
	probe := mustCreateTelegramRecord(t, hub, CollectionNetworkProbes, map[string]any{"name": "target probe", "type": "tcping", "target": "example.com:443", "interval_seconds": 20, "timeout_seconds": 3, "enabled": true, "scope": NetworkProbeScopeGlobal})
	mustCreateTelegramRecord(t, hub, CollectionNetworkProbeAssignments, map[string]any{"probe": probe.Id, "system": system.Id, "enabled": true})
	document := ConfigBackupDocument{Meta: ConfigBackupMeta{BackupVersion: ConfigBackupVersion, Mode: ConfigBackupMode, Sections: []string{ConfigBackupSectionAlerts, ConfigBackupSectionNotifications, ConfigBackupSectionPublicStatus, ConfigBackupSectionNetworkProbes}}}
	preview, err := hub.previewConfigBackup(document, "empty imports", "")
	require.NoError(t, err)
	var preserveIDs []string
	for _, item := range preview.Items {
		if item.Action == configBackupActionPreserve {
			preserveIDs = append(preserveIDs, item.StableID)
		}
	}
	assert.Len(t, preserveIDs, 8)
	assert.Len(t, uniqueStrings(preserveIDs), len(preserveIDs))
}

func TestConfigBackupPreviewIgnoresSecretsInSkippedSections(t *testing.T) {
	hub, _ := newTelegramHubWithAdmin(t)
	secret, err := encryptConfigBackupSecret("future-token", "backup-pass", "system.token")
	require.NoError(t, err)
	document := ConfigBackupDocument{Meta: ConfigBackupMeta{BackupVersion: ConfigBackupVersion, Mode: ConfigBackupMode, Sections: []string{ConfigBackupSectionSystems}}, SkippedSections: map[string]string{ConfigBackupSectionSystems: "future section version"}, Systems: []ConfigBackupSystem{{StableID: "future", Token: secret}}}
	preview, err := hub.previewConfigBackup(document, "skipped encrypted section", "")
	require.NoError(t, err)
	assert.False(t, preview.RequiresCredential)
	assert.Zero(t, preview.Summary.Error)
}

func TestRestoreConfigBackupReturnsCommittedProgressWhenLaterSectionFails(t *testing.T) {
	hub, admin := newTelegramHubWithAdmin(t)
	document := ConfigBackupDocument{
		Meta:    ConfigBackupMeta{BackupVersion: ConfigBackupVersion, Mode: ConfigBackupMode, Sections: []string{ConfigBackupSectionSystems, ConfigBackupSectionAlerts}},
		Systems: []ConfigBackupSystem{{StableID: "progresssys0001", Name: "committed system", Host: "10.0.0.1", Users: []ConfigBackupUserRef{{Email: admin.GetString("email")}}}},
		Alerts:  ConfigBackupAlerts{Definitions: []ConfigBackupAlertDefinition{{StableID: "progressalert01", SystemStableID: "progresssys0001", UserEmail: admin.GetString("email"), Name: "invalid-alert-name"}}},
	}
	content, err := marshalConfigBackupDocument(document)
	require.NoError(t, err)
	body, err := json.Marshal(ConfigBackupRestoreRequest{Content: content, PreviewID: configBackupPreviewID(content, ""), Mode: ConfigBackupMode})
	require.NoError(t, err)
	request := httptest.NewRequest(http.MethodPost, "/api/beszel/config/restore", bytes.NewReader(body))
	request.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()
	require.NoError(t, hub.restoreConfigBackup(&core.RequestEvent{App: hub, Auth: admin, Event: router.Event{Request: request, Response: recorder}}))

	assert.Equal(t, http.StatusUnprocessableEntity, recorder.Code)
	var response ConfigBackupRestoreFailureResponse
	require.NoError(t, json.Unmarshal(recorder.Body.Bytes(), &response))
	assert.Equal(t, []string{ConfigBackupSectionSystems}, response.CompletedSections)
	assert.Equal(t, ConfigBackupSectionAlerts, response.FailedSection)
	assert.Equal(t, 1, response.Applied.Created)
	assert.Contains(t, response.Error, "rolled back")
	_, err = hub.FindRecordById("systems", "progresssys0001")
	require.NoError(t, err)
	_, err = hub.FindRecordById("alerts", "progressalert01")
	require.Error(t, err)
}

func uniqueStrings(values []string) map[string]struct{} {
	unique := make(map[string]struct{}, len(values))
	for _, value := range values {
		unique[value] = struct{}{}
	}
	return unique
}
