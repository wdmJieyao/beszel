//go:build testing

package hub

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConfigBackupNetworkProbeAssignmentPreviewUsesBackupReferences(t *testing.T) {
	hub, admin := newTelegramHubWithAdmin(t)
	document := ConfigBackupDocument{
		Meta: ConfigBackupMeta{BackupVersion: ConfigBackupVersion, Mode: ConfigBackupMode},
		Systems: []ConfigBackupSystem{
			{StableID: "sysbackup000001", Name: "node-1", Host: "10.0.0.1", Users: []ConfigBackupUserRef{{Email: admin.GetString("email")}}},
		},
		NetworkProbes: ConfigBackupNetworkProbes{
			Probes: []ConfigBackupNetworkProbe{
				{StableID: "probebackup0001", Name: "Telecom", Type: "tcping", Target: "example.com:443", Enabled: true, Scope: NetworkProbeScopeGlobal},
			},
			Assignments: []ConfigBackupNetworkProbeAssignment{
				{ProbeStableID: "probebackup0001", SystemStableID: "sysbackup000001", Enabled: true},
			},
		},
	}
	content, err := marshalConfigBackupDocument(document)
	require.NoError(t, err)

	preview, err := hub.previewConfigBackup(document, content, "")
	require.NoError(t, err)
	assert.Zero(t, preview.Summary.Conflict)
	assert.GreaterOrEqual(t, preview.Summary.Create, 2)
}
