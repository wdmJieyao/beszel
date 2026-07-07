//go:build testing

package hub

import (
	"strings"
	"testing"
	"time"

	"github.com/henrygd/beszel/internal/alerts"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTelegramDestinationTestSendUsesFakeTransport(t *testing.T) {
	hub, admin := newTelegramHubWithAdmin(t)
	fake := &fakeTelegramTransport{}
	hub.telegramTransport = fake
	_, err := hub.saveTelegramSettings(TelegramSettingsInput{
		Enabled:        true,
		PollingEnabled: false,
		BotToken:       "123456:abcde_token_valid",
	}, telegramSettingsRecord{})
	require.NoError(t, err)

	record, err := hub.upsertTelegramDestination(nil, TelegramDestinationInput{
		Name:     "Ops",
		ChatID:   "-10012345",
		ChatType: TelegramChatTypeChannel,
		Role:     TelegramRoleAdmin,
	})
	require.NoError(t, err)

	err = hub.sendTelegramTestMessage(telegramDestinationFromRecord(record))
	require.NoError(t, err)
	require.Len(t, fake.sendCalls, 1)
	assert.Equal(t, "-10012345", fake.sendCalls[0].ChatID)
	assert.Contains(t, fake.sendCalls[0].Text, "Beszel Telegram test")
	_ = admin
}

func TestSendTelegramAlertHonorsRoleAndScopes(t *testing.T) {
	hub, admin := newTelegramHubWithAdmin(t)
	fake := &fakeTelegramTransport{}
	hub.telegramTransport = fake
	_, err := hub.saveTelegramSettings(TelegramSettingsInput{
		Enabled:        true,
		PollingEnabled: false,
		BotToken:       "123456:abcde_token_valid",
	}, telegramSettingsRecord{})
	require.NoError(t, err)
	system := mustCreateTelegramRecord(t, hub, "systems", map[string]any{
		"name":   "node-1",
		"host":   "127.0.0.1",
		"status": "up",
		"users":  []string{admin.Id},
	})

	_, err = hub.upsertTelegramDestination(nil, TelegramDestinationInput{
		Name:            "Admin",
		ChatID:          "12345",
		ChatType:        TelegramChatTypePrivate,
		Role:            TelegramRoleAdmin,
		Enabled:         boolPtr(true),
		UserID:          admin.Id,
		NodeScope:       []string{system.Id},
		AlertLevelScope: []string{"status"},
	})
	require.NoError(t, err)
	_, err = hub.upsertTelegramDestination(nil, TelegramDestinationInput{
		Name:            "Readonly",
		ChatID:          "-10012345",
		ChatType:        TelegramChatTypeChannel,
		Role:            TelegramRoleReadOnly,
		Enabled:         boolPtr(true),
		NodeScope:       []string{system.Id},
		AlertLevelScope: []string{"status"},
		MuteUntil:       nil,
	})
	require.NoError(t, err)
	_, err = hub.upsertTelegramDestination(nil, TelegramDestinationInput{
		Name:            "Muted",
		ChatID:          "-10054321",
		ChatType:        TelegramChatTypeChannel,
		Role:            TelegramRoleReadOnly,
		Enabled:         boolPtr(true),
		NodeScope:       []string{system.Id},
		AlertLevelScope: []string{"status"},
		MuteUntil:       timePtr(time.Now().UTC().Add(5 * time.Minute)),
	})
	require.NoError(t, err)

	err = hub.SendTelegramAlert(alerts.AlertMessageData{
		UserID:   admin.Id,
		SystemID: system.Id,
		Title:    "Connection to node-1 is down",
		Message:  "Connection to node-1 is down",
		Link:     "https://example.com/system/node-1",
	})
	require.NoError(t, err)
	require.Len(t, fake.sendCalls, 2)
	assert.Contains(t, fake.sendCalls[0].Text, "https://example.com/system/node-1")
	assert.True(t, strings.HasPrefix(fake.sendCalls[1].Text, "Beszel 告警摘要"))
	assert.NotContains(t, fake.sendCalls[1].Text, "https://example.com/system/node-1")
}

func boolPtr(v bool) *bool {
	return &v
}

func timePtr(v time.Time) *time.Time {
	return &v
}
