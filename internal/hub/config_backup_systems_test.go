//go:build testing

package hub

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConfigBackupSystemsExportUserRefsAndRedactedToken(t *testing.T) {
	hub, admin := newTelegramHubWithAdmin(t)
	system := mustCreateTelegramRecord(t, hub, "systems", map[string]any{
		"name":   "node-1",
		"host":   "10.0.0.1",
		"port":   "45876",
		"status": "up",
		"users":  []string{admin.Id},
	})
	mustCreateTelegramRecord(t, hub, "fingerprints", map[string]any{
		"system":      system.Id,
		"fingerprint": "fingerprintabc",
		"token":       "secret-token",
	})
	_, userIDToEmail, err := hub.configBackupUsers()
	require.NoError(t, err)

	items, err := hub.configBackupSystems(userIDToEmail, configBackupExportOptions{IncludeSecrets: false})
	require.NoError(t, err)
	require.Len(t, items, 1)
	assert.Equal(t, admin.GetString("email"), items[0].Users[0].Email)
	require.NotNil(t, items[0].Token)
	assert.True(t, items[0].Token.Redacted)
	assert.Empty(t, items[0].Token.Encrypted)
}
