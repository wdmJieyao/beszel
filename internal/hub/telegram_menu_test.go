//go:build testing

package hub

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTelegramAdminMenuReturnsStatusAndSystemList(t *testing.T) {
	hub, admin := newTelegramHubWithAdmin(t)
	fake := &fakeTelegramTransport{}
	hub.telegramTransport = fake
	system := mustCreateTelegramRecord(t, hub, "systems", map[string]any{
		"name":   "node-1",
		"host":   "127.0.0.1",
		"status": "up",
		"users":  []string{admin.Id},
	})
	_, err := hub.upsertTelegramDestination(nil, TelegramDestinationInput{
		Name:     "Admin",
		ChatID:   "12345",
		ChatType: TelegramChatTypePrivate,
		Role:     TelegramRoleAdmin,
		Enabled:  boolPtr(true),
	})
	require.NoError(t, err)

	require.NoError(t, hub.handleTelegramCommand(context.Background(), "token", telegramCommand{ChatID: "12345", Name: "systems"}))
	require.NoError(t, hub.handleTelegramCommand(context.Background(), "token", telegramCommand{ChatID: "12345", Name: "system", Args: []string{system.Id}}))
	require.Len(t, fake.sendCalls, 2)
	assert.Contains(t, fake.sendCalls[0].Text, "Beszel 节点列表")
	assert.Contains(t, fake.sendCalls[0].Text, "node-1")
	assert.Contains(t, fake.sendCalls[1].Text, "Beszel 节点详情")
	assert.NotContains(t, fake.sendCalls[1].Text, "127.0.0.1")
}

func TestTelegramReadOnlyAndUnknownChatsCannotUsePrivilegedMenu(t *testing.T) {
	hub, _ := newTelegramHubWithAdmin(t)
	fake := &fakeTelegramTransport{}
	hub.telegramTransport = fake
	_, err := hub.upsertTelegramDestination(nil, TelegramDestinationInput{
		Name:     "Readonly",
		ChatID:   "-10012345",
		ChatType: TelegramChatTypeChannel,
		Role:     TelegramRoleReadOnly,
		Enabled:  boolPtr(true),
	})
	require.NoError(t, err)

	require.NoError(t, hub.handleTelegramCommand(context.Background(), "token", telegramCommand{ChatID: "-10012345", Name: "status"}))
	require.NoError(t, hub.handleTelegramCommand(context.Background(), "token", telegramCommand{ChatID: "99999", Name: "status"}))
	require.Len(t, fake.sendCalls, 2)
	assert.Contains(t, fake.sendCalls[0].Text, "只读通知渠道")
	assert.NotContains(t, fake.sendCalls[0].Text, "节点：")
	assert.Contains(t, fake.sendCalls[1].Text, "未在 Beszel 面板中授权")
	assert.NotContains(t, fake.sendCalls[1].Text, "节点：")
}
