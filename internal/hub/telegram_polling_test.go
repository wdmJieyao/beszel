//go:build testing

package hub

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseTelegramCommandFromMessageAndCallback(t *testing.T) {
	messageCommand, ok := parseTelegramCommand(telegramUpdate{
		Message: &telegramMessage{
			Text: "/system 2",
			Chat: telegramChat{ID: 12345},
		},
	})
	require.True(t, ok)
	assert.Equal(t, "12345", messageCommand.ChatID)
	assert.Equal(t, "system", messageCommand.Name)
	assert.Equal(t, []string{"2"}, messageCommand.Args)

	callbackCommand, ok := parseTelegramCommand(telegramUpdate{
		CallbackQuery: &telegramCallbackQuery{
			ID:   "callback-1",
			Data: "status",
			Message: &telegramMessage{
				Chat: telegramChat{ID: -10012345},
			},
		},
	})
	require.True(t, ok)
	assert.Equal(t, "-10012345", callbackCommand.ChatID)
	assert.Equal(t, "status", callbackCommand.Name)
	assert.Equal(t, "callback-1", callbackCommand.CallbackQueryID)
}

func TestTelegramPollingProcessesUpdatesAndPersistsOffset(t *testing.T) {
	hub, _ := newTelegramHubWithAdmin(t)
	fake := &fakeTelegramTransport{
		getUpdates: []telegramUpdate{
			{
				UpdateID: 41,
				Message: &telegramMessage{
					Text: "/status",
					Chat: telegramChat{ID: 12345},
				},
			},
		},
	}
	hub.telegramTransport = fake
	_, err := hub.saveTelegramSettings(TelegramSettingsInput{
		Enabled:        true,
		PollingEnabled: true,
		BotToken:       "123456:abcde_token_valid",
	}, telegramSettingsRecord{})
	require.NoError(t, err)
	_, err = hub.upsertTelegramDestination(nil, TelegramDestinationInput{
		Name:     "Admin",
		ChatID:   "12345",
		ChatType: TelegramChatTypePrivate,
		Role:     TelegramRoleAdmin,
		Enabled:  boolPtr(true),
	})
	require.NoError(t, err)

	require.NoError(t, hub.pollTelegramOnce(context.Background()))
	require.Len(t, fake.sendCalls, 1)
	assert.Contains(t, fake.sendCalls[0].Text, "Beszel 状态总览")

	settings, err := hub.loadTelegramSettings()
	require.NoError(t, err)
	assert.Equal(t, int64(41), settings.LastPollOffset)
}
