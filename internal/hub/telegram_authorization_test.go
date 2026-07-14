//go:build testing

package hub

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTelegramAdminMenuRequiresPrivateChat(t *testing.T) {
	assert.True(t, telegramDestinationCanUseAdminMenu(telegramDestinationRecord{Enabled: true, Role: TelegramRoleAdmin, ChatType: TelegramChatTypePrivate}))
	assert.False(t, telegramDestinationCanUseAdminMenu(telegramDestinationRecord{Enabled: true, Role: TelegramRoleAdmin, ChatType: TelegramChatTypeGroup}))
	assert.False(t, telegramDestinationCanUseAdminMenu(telegramDestinationRecord{Enabled: true, Role: TelegramRoleAdmin, ChatType: TelegramChatTypeSupergroup}))
}

func TestTelegramAuthorizationUsesEnabledAllowlist(t *testing.T) {
	hub, _ := newTelegramHubWithAdmin(t)
	_, err := hub.upsertTelegramDestination(nil, TelegramDestinationInput{
		Name:     "Disabled admin",
		ChatID:   "12345",
		ChatType: TelegramChatTypePrivate,
		Role:     TelegramRoleAdmin,
		Enabled:  boolPtr(false),
	})
	require.NoError(t, err)

	_, authorized, err := hub.authorizeTelegramCommand("12345")
	require.NoError(t, err)
	assert.False(t, authorized)

	_, authorized, err = hub.authorizeTelegramCommand("99999")
	require.NoError(t, err)
	assert.False(t, authorized)
}
