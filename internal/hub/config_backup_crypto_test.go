//go:build testing

package hub

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConfigBackupSecretEncryptionRoundTrip(t *testing.T) {
	secret, err := encryptConfigBackupSecret("secret-token", "passphrase", "system.token")
	require.NoError(t, err)
	require.NotNil(t, secret)
	assert.NotContains(t, secret.Encrypted, "secret-token")

	plaintext, err := decryptConfigBackupSecret(secret, "passphrase")
	require.NoError(t, err)
	assert.Equal(t, "secret-token", plaintext)

	_, err = decryptConfigBackupSecret(secret, "wrong-passphrase")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "decryption failed")
}

func TestConfigBackupExportOmitsPlaintextSecrets(t *testing.T) {
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
		"token":       "system-secret-token",
	})
	_, err := hub.saveTelegramSettings(TelegramSettingsInput{
		Enabled:        true,
		PollingEnabled: true,
		BotToken:       "123456:abcde_token_valid",
	}, telegramSettingsRecord{})
	require.NoError(t, err)

	document, _, err := hub.buildConfigBackupDocument(configBackupExportOptions{
		IncludeSecrets: true,
		Credential:     "backup-pass",
		Sections:       []string{ConfigBackupSectionSystems, ConfigBackupSectionNotifications},
	})
	require.NoError(t, err)
	content, err := marshalConfigBackupDocument(document)
	require.NoError(t, err)

	assert.NotContains(t, content, "system-secret-token")
	assert.NotContains(t, content, "123456:abcde_token_valid")
	assert.Contains(t, content, "encrypted:")
	assert.True(t, strings.Contains(content, "system.token") || strings.Contains(content, "telegram.botToken"))
}
