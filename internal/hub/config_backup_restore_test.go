//go:build testing

package hub

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

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
