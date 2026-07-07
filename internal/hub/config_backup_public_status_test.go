//go:build testing

package hub

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConfigBackupPublicStatusPreviewUsesBackupSystemReferences(t *testing.T) {
	hub, admin := newTelegramHubWithAdmin(t)
	document := ConfigBackupDocument{
		Meta: ConfigBackupMeta{BackupVersion: ConfigBackupVersion, Mode: ConfigBackupMode},
		Systems: []ConfigBackupSystem{
			{StableID: "sysbackup000001", Name: "node-1", Host: "10.0.0.1", Users: []ConfigBackupUserRef{{Email: admin.GetString("email")}}},
		},
		PublicStatus: ConfigBackupPublicStatus{
			Systems: []ConfigBackupPublicSystem{
				{SystemStableID: "sysbackup000001", PublicEnabled: true, PublicName: "Public node", ShowCPU: true},
			},
		},
	}
	content, err := marshalConfigBackupDocument(document)
	require.NoError(t, err)

	preview, err := hub.previewConfigBackup(document, content, "")
	require.NoError(t, err)
	assert.Zero(t, preview.Summary.Conflict)
	assert.Contains(t, preview.Items, ConfigBackupPreviewItem{
		Section:     ConfigBackupSectionPublicStatus,
		StableID:    "sysbackup000001",
		DisplayName: "Public node",
		Action:      configBackupActionUpdate,
		Reason:      "public visibility will be merged",
	})
}
