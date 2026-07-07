//go:build testing

package hub

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConfigBackupNotificationsRestoreTelegramDestination(t *testing.T) {
	hub, admin := newTelegramHubWithAdmin(t)
	token, err := encryptConfigBackupSecret("123456:abcde_token_valid", "backup-pass", "telegram.botToken")
	require.NoError(t, err)
	document := ConfigBackupDocument{
		Meta: ConfigBackupMeta{BackupVersion: ConfigBackupVersion, Mode: ConfigBackupMode},
		Notifications: ConfigBackupNotifications{
			Telegram: ConfigBackupTelegramNotifications{
				Settings: ConfigBackupTelegramSettings{
					Enabled:        true,
					PollingEnabled: true,
					BotToken:       token,
				},
				Destinations: []ConfigBackupTelegramDestination{
					{
						StableID:  "tgdestbackup001",
						UserEmail: admin.GetString("email"),
						Name:      "Ops",
						ChatID:    "12345",
						ChatType:  TelegramChatTypePrivate,
						Role:      TelegramRoleAdmin,
						Enabled:   true,
					},
				},
			},
		},
	}

	summary := ConfigBackupApplySummary{}
	emailToUserID, err := hub.userIDByEmailMap()
	require.NoError(t, err)
	require.NoError(t, hub.applyConfigBackupNotifications(document.Notifications, emailToUserID, "backup-pass", &summary))
	assert.Equal(t, 2, summary.Created+summary.Updated)

	destination, err := hub.FindRecordById(CollectionTelegramDestinations, "tgdestbackup001")
	require.NoError(t, err)
	assert.Equal(t, "Ops", destination.GetString("name"))
	assert.Equal(t, admin.Id, destination.GetString("user"))

	settings, err := hub.loadTelegramSettings()
	require.NoError(t, err)
	plaintext, err := hub.decryptTelegramToken(settings)
	require.NoError(t, err)
	assert.Equal(t, "123456:abcde_token_valid", plaintext)
}
