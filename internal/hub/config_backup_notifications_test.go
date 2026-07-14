//go:build testing

package hub

import (
	"os"
	"slices"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTelegramBackupCompatibilityFixturesPreviewAndApply(t *testing.T) {
	t.Run("version 2 preserves all channel and policy fields", func(t *testing.T) {
		hub, _ := newTelegramHubWithAdmin(t)
		_, err := hub.saveTelegramSettings(TelegramSettingsInput{
			Enabled: true, PollingEnabled: true, BotToken: "123456:fixture_target_token",
		}, telegramSettingsRecord{})
		require.NoError(t, err)
		targetOnly := mustCreateTelegramRecord(t, hub, CollectionTelegramDestinations, map[string]any{
			"name": "Target only", "chat_id": "10999", "chat_type": TelegramChatTypePrivate, "role": TelegramRoleAdmin, "enabled": true,
		})
		require.NoError(t, hub.ensureTelegramDestinationDefaultPolicy(targetOnly))

		content, err := os.ReadFile("testdata/telegram_backup/notifications-v2.yml")
		require.NoError(t, err)
		document, err := parseConfigBackupDocument(string(content))
		require.NoError(t, err)
		preview, err := hub.previewConfigBackup(document, string(content), "")
		require.NoError(t, err)
		assert.Zero(t, preview.Summary.Conflict)
		assert.Zero(t, preview.Summary.Error)
		_, _, err = hub.applyConfigBackup(document, "")
		require.NoError(t, err)
		_, _, err = hub.applyConfigBackup(document, "")
		require.NoError(t, err)

		destination, err := hub.FindRecordById(CollectionTelegramDestinations, "channel00000001")
		require.NoError(t, err)
		assert.Equal(t, "Operations", destination.GetString("name"))
		assert.Equal(t, "10002", destination.GetString("chat_id"))
		assert.Equal(t, TelegramChatTypeGroup, destination.GetString("chat_type"))
		assert.Equal(t, TelegramRoleReadOnly, destination.GetString("role"))
		assert.False(t, destination.GetBool("enabled"))
		assert.False(t, destination.GetDateTime("mute_until").IsZero())
		policies, err := hub.listTelegramNotificationPolicies(destination.Id)
		require.NoError(t, err)
		require.Len(t, policies, 1)
		assert.Equal(t, "All status alerts", policies[0].Name)
		assert.False(t, policies[0].Enabled)
		assert.Equal(t, TelegramNodeScopeAll, policies[0].NodeScopeMode)
		assert.Empty(t, policies[0].NodeScope)
		assert.Equal(t, []string{"status"}, policies[0].AlertLevelScope)
		_, err = hub.FindRecordById(CollectionTelegramDestinations, targetOnly.Id)
		require.NoError(t, err)
		targetPolicies, err := hub.listTelegramNotificationPolicies(targetOnly.Id)
		require.NoError(t, err)
		assert.Len(t, targetPolicies, 1)
		settings, err := hub.loadTelegramSettings()
		require.NoError(t, err)
		token, err := hub.decryptTelegramToken(settings)
		require.NoError(t, err)
		assert.Equal(t, "123456:fixture_target_token", token)
		assert.False(t, settings.Enabled)
		assert.False(t, settings.PollingEnabled)
	})

	t.Run("version 1 creates one idempotent default policy", func(t *testing.T) {
		hub, _ := newTelegramHubWithAdmin(t)
		content, err := os.ReadFile("testdata/telegram_backup/notifications-v1.yml")
		require.NoError(t, err)
		document, err := parseConfigBackupDocument(string(content))
		require.NoError(t, err)
		preview, err := hub.previewConfigBackup(document, string(content), "")
		require.NoError(t, err)
		assert.Zero(t, preview.Summary.Conflict)
		_, _, err = hub.applyConfigBackup(document, "")
		require.NoError(t, err)
		_, _, err = hub.applyConfigBackup(document, "")
		require.NoError(t, err)
		policies, err := hub.listTelegramNotificationPolicies("legacychannel01")
		require.NoError(t, err)
		require.Len(t, policies, 1)
		assert.Equal(t, TelegramDefaultPolicyName, policies[0].Name)
		assert.True(t, policies[0].Enabled)
		assert.Equal(t, TelegramNodeScopeAll, policies[0].NodeScopeMode)
		assert.Empty(t, policies[0].NodeScope)
		assert.Equal(t, []string{"status"}, policies[0].AlertLevelScope)
	})
}

func TestConfigBackupNotificationsPreviewReportsRoutingConflicts(t *testing.T) {
	hub, admin := newTelegramHubWithAdmin(t)
	system := mustCreateTelegramTestSystems(t, hub, admin.Id, 1)[0]
	owner := mustCreateTelegramRecord(t, hub, CollectionTelegramDestinations, map[string]any{
		"name": "Owner", "chat_id": "44901", "chat_type": TelegramChatTypePrivate, "role": TelegramRoleAdmin, "enabled": true,
	})
	other := mustCreateTelegramRecord(t, hub, CollectionTelegramDestinations, map[string]any{
		"name": "Other", "chat_id": "44902", "chat_type": TelegramChatTypePrivate, "role": TelegramRoleAdmin, "enabled": true,
	})
	existingPolicy, err := hub.upsertTelegramNotificationPolicy(nil, owner.Id, TelegramNotificationPolicyInput{
		Name: "Existing", Enabled: boolPtr(true), NodeScopeMode: TelegramNodeScopeAll,
		NodeScope: []string{}, AlertLevelScope: []string{"status"},
	})
	require.NoError(t, err)

	document := ConfigBackupDocument{
		Meta: ConfigBackupMeta{BackupVersion: ConfigBackupVersion, Mode: ConfigBackupMode, Sections: []string{ConfigBackupSectionNotifications}},
		Notifications: ConfigBackupNotifications{SectionVersion: ConfigBackupNotificationsVersion, Telegram: ConfigBackupTelegramNotifications{
			Destinations: []ConfigBackupTelegramDestination{
				{StableID: "newdestconflict", Name: "Collision", ChatID: owner.GetString("chat_id"), ChatType: TelegramChatTypePrivate, Role: TelegramRoleAdmin, Enabled: true},
				{StableID: other.Id, Name: "Other", ChatID: other.GetString("chat_id"), ChatType: TelegramChatTypePrivate, Role: TelegramRoleAdmin, Enabled: true},
			},
			Policies: []ConfigBackupTelegramPolicy{
				{StableID: existingPolicy.Id, DestinationStableID: other.Id, Name: "Moved", Enabled: true, NodeScopeMode: TelegramNodeScopeAll},
				{StableID: "unknownsystempol", DestinationStableID: other.Id, Name: "Unknown system", Enabled: true, NodeScopeMode: TelegramNodeScopeSelected, NodeScope: []string{"missing-system"}},
				{StableID: "duplicatepolicy1", DestinationStableID: other.Id, Name: "Duplicate", Enabled: true, NodeScopeMode: TelegramNodeScopeSelected, NodeScope: []string{system.Id}},
				{StableID: "duplicatepolicy2", DestinationStableID: other.Id, Name: "Duplicate", Enabled: true, NodeScopeMode: TelegramNodeScopeAll},
			},
		}},
	}
	preview, err := hub.previewConfigBackup(document, "routing conflicts", "")
	require.NoError(t, err)
	assert.GreaterOrEqual(t, preview.Summary.Conflict, 4)
	reasons := make([]string, 0, len(preview.Items))
	for _, item := range preview.Items {
		if item.Action == configBackupActionConflict {
			reasons = append(reasons, item.Reason)
		}
	}
	joined := strings.Join(reasons, "\n")
	assert.Contains(t, joined, "Chat ID")
	assert.Contains(t, joined, "unknown system")
	assert.Contains(t, joined, "different destination")
	assert.Contains(t, joined, "duplicate policy name")

	_, _, err = hub.applyConfigBackup(document, "")
	require.Error(t, err)
	ownerAfterFailure, findErr := hub.FindRecordById(CollectionTelegramDestinations, owner.Id)
	require.NoError(t, findErr)
	assert.Equal(t, "44901", ownerAfterFailure.GetString("chat_id"))
	_, findErr = hub.FindRecordById(CollectionTelegramDestinations, "newdestconflict")
	require.Error(t, findErr)
	existingAfterFailure, findErr := hub.FindRecordById(CollectionTelegramNotificationPolicies, existingPolicy.Id)
	require.NoError(t, findErr)
	assert.Equal(t, owner.Id, existingAfterFailure.GetString("destination"))
}

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

func TestConfigBackupNotificationsPreserveRedactedSecrets(t *testing.T) {
	hub, admin := newTelegramHubWithAdmin(t)
	_, err := hub.saveTelegramSettings(TelegramSettingsInput{
		Enabled: true, PollingEnabled: true, BotToken: "123456:keep_telegram_token",
	}, telegramSettingsRecord{})
	require.NoError(t, err)
	settingsRecord := mustCreateTelegramRecord(t, hub, "user_settings", map[string]any{
		"user": admin.Id,
		"settings": map[string]any{
			"emails": []string{"old@example.com"}, "webhooks": []string{"https://keep.example/webhook"},
		},
	})

	document := ConfigBackupDocument{
		Meta: ConfigBackupMeta{BackupVersion: ConfigBackupVersion, Mode: ConfigBackupMode, Sections: []string{ConfigBackupSectionNotifications}},
		Notifications: ConfigBackupNotifications{
			UserSettings: []ConfigBackupUserNotificationSettings{{
				UserEmail: admin.GetString("email"), Emails: []string{"new@example.com"},
				Webhooks: []ConfigBackupSecret{{Redacted: true}},
			}},
			Telegram: ConfigBackupTelegramNotifications{Settings: ConfigBackupTelegramSettings{
				Enabled: true, PollingEnabled: true, BotToken: &ConfigBackupSecret{Redacted: true},
			}},
		},
	}

	summary := ConfigBackupApplySummary{}
	emailToUserID, err := hub.userIDByEmailMap()
	require.NoError(t, err)
	require.NoError(t, hub.applyConfigBackupNotifications(document.Notifications, emailToUserID, "", &summary))

	settings, err := hub.loadTelegramSettings()
	require.NoError(t, err)
	token, err := hub.decryptTelegramToken(settings)
	require.NoError(t, err)
	assert.Equal(t, "123456:keep_telegram_token", token)

	refreshed, err := hub.FindRecordById("user_settings", settingsRecord.Id)
	require.NoError(t, err)
	var values map[string]any
	require.NoError(t, refreshed.UnmarshalJSONField("settings", &values))
	assert.Equal(t, []any{"https://keep.example/webhook"}, values["webhooks"])
}

func TestConfigBackupNotificationsOmittedTelegramSettingsPreserveTargetBot(t *testing.T) {
	hub, _ := newTelegramHubWithAdmin(t)
	_, err := hub.saveTelegramSettings(TelegramSettingsInput{
		Enabled: true, PollingEnabled: true, BotToken: "123456:keep_omitted_token",
	}, telegramSettingsRecord{})
	require.NoError(t, err)

	summary := ConfigBackupApplySummary{}
	require.NoError(t, hub.applyConfigBackupNotifications(ConfigBackupNotifications{}, map[string]string{}, "", &summary))

	settings, err := hub.loadTelegramSettings()
	require.NoError(t, err)
	token, err := hub.decryptTelegramToken(settings)
	require.NoError(t, err)
	assert.True(t, settings.Enabled)
	assert.True(t, settings.PollingEnabled)
	assert.Equal(t, "123456:keep_omitted_token", token)
	assert.Zero(t, summary.Updated)
}

func TestConfigBackupNotificationsMixedWebhooksRestoreAvailableItemsAndPreserveRedactedPositions(t *testing.T) {
	hub, admin := newTelegramHubWithAdmin(t)
	record := mustCreateTelegramRecord(t, hub, "user_settings", map[string]any{
		"user": admin.Id,
		"settings": map[string]any{
			"webhooks": []string{"https://keep.example/first", "https://replace.example/second"},
		},
	})
	restored, err := encryptConfigBackupSecret("https://new.example/second", "backup-pass", "notifications.webhook")
	require.NoError(t, err)
	document := ConfigBackupNotifications{UserSettings: []ConfigBackupUserNotificationSettings{{
		UserEmail: admin.GetString("email"),
		Webhooks: []ConfigBackupSecret{
			{Redacted: true, ContentType: "notifications.webhook"},
			*restored,
		},
	}}}
	summary := ConfigBackupApplySummary{}
	require.NoError(t, hub.applyConfigBackupNotifications(document, map[string]string{admin.GetString("email"): admin.Id}, "backup-pass", &summary))

	refreshed, err := hub.FindRecordById("user_settings", record.Id)
	require.NoError(t, err)
	var settings map[string]any
	require.NoError(t, refreshed.UnmarshalJSONField("settings", &settings))
	webhooks, ok := settings["webhooks"].([]any)
	require.True(t, ok)
	values := make([]string, 0, len(webhooks))
	for _, webhook := range webhooks {
		values = append(values, webhook.(string))
	}
	assert.True(t, slices.Equal([]string{"https://keep.example/first", "https://new.example/second"}, values))
}

func TestConfigBackupNotificationsV2RoundTripsPoliciesAndV1CreatesDefault(t *testing.T) {
	hub, admin := newTelegramHubWithAdmin(t)
	system := mustCreateTelegramTestSystems(t, hub, admin.Id, 1)[0]
	destination := mustCreateTelegramRecord(t, hub, CollectionTelegramDestinations, map[string]any{
		"name": "Ops", "chat_id": "44001", "chat_type": TelegramChatTypePrivate, "role": TelegramRoleAdmin, "enabled": true,
	})
	_, err := hub.upsertTelegramNotificationPolicy(nil, destination.Id, TelegramNotificationPolicyInput{
		Name: "CPU", Enabled: boolPtr(true), NodeScopeMode: TelegramNodeScopeSelected,
		NodeScope: []string{system.Id}, AlertLevelScope: []string{"cpu"},
	})
	require.NoError(t, err)

	exported, err := hub.configBackupNotifications(map[string]string{admin.Id: admin.GetString("email")}, configBackupExportOptions{})
	require.NoError(t, err)
	assert.Equal(t, ConfigBackupNotificationsVersion, exported.SectionVersion)
	require.Len(t, exported.Telegram.Policies, 1)
	assert.Equal(t, destination.Id, exported.Telegram.Policies[0].DestinationStableID)

	otherHub, otherAdmin := newTelegramHubWithAdmin(t)
	mustCreateTelegramRecord(t, otherHub, "systems", map[string]any{
		"id": system.Id, "name": "node", "host": "127.0.0.1", "status": "up", "users": []string{otherAdmin.Id},
	})
	summary := ConfigBackupApplySummary{}
	require.NoError(t, otherHub.applyConfigBackupNotifications(exported, map[string]string{admin.GetString("email"): otherAdmin.Id}, "", &summary))
	require.NoError(t, otherHub.applyConfigBackupNotifications(exported, map[string]string{admin.GetString("email"): otherAdmin.Id}, "", &summary))
	policies, err := otherHub.listTelegramNotificationPolicies(destination.Id)
	require.NoError(t, err)
	require.Len(t, policies, 1)
	assert.Equal(t, "CPU", policies[0].Name)

	legacy := ConfigBackupNotifications{SectionVersion: "1", Telegram: ConfigBackupTelegramNotifications{Destinations: []ConfigBackupTelegramDestination{{
		StableID: "legacydest00001", Name: "Legacy", ChatID: "44002", ChatType: TelegramChatTypePrivate,
		Role: TelegramRoleReadOnly, Enabled: true, NodeScope: []string{system.Id}, AlertLevelScope: []string{"status"},
	}}}}
	require.NoError(t, otherHub.applyConfigBackupNotifications(legacy, map[string]string{}, "", &summary))
	legacyPolicies, err := otherHub.listTelegramNotificationPolicies("legacydest00001")
	require.NoError(t, err)
	require.Len(t, legacyPolicies, 1)
	assert.Equal(t, TelegramNodeScopeSelected, legacyPolicies[0].NodeScopeMode)
	require.NoError(t, otherHub.applyConfigBackupNotifications(legacy, map[string]string{}, "", &summary))
	legacyPolicies, err = otherHub.listTelegramNotificationPolicies("legacydest00001")
	require.NoError(t, err)
	assert.Len(t, legacyPolicies, 1)
}
