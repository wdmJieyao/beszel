//go:build testing

package hub

import (
	"context"
	"fmt"
	"strings"
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
		"info":   map[string]any{"cpu": 12.34, "mp": 45.67, "dp": 78.9},
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
	require.NoError(t, hub.handleTelegramCommand(context.Background(), "token", telegramCommand{ChatID: "12345", Name: "status"}))
	require.Len(t, fake.sendCalls, 3)
	assert.Contains(t, fake.sendCalls[0].Text, "Beszel 节点列表")
	assert.Contains(t, fake.sendCalls[0].Text, "node-1")
	assert.Contains(t, fake.sendCalls[1].Text, "Beszel 节点详情")
	assert.NotContains(t, fake.sendCalls[1].Text, "127.0.0.1")
	assert.Contains(t, fake.sendCalls[1].Text, "CPU：12.3%")
	assert.Contains(t, fake.sendCalls[1].Text, "内存：45.7%")
	assert.Contains(t, fake.sendCalls[1].Text, "磁盘：78.9%")
	assert.Contains(t, fake.sendCalls[1].Text, "最后上报：")
	require.NotNil(t, fake.sendCalls[1].Options)
	assert.NotNil(t, fake.sendCalls[1].Options.ReplyMarkup)
	assert.Nil(t, fake.sendCalls[2].Options)
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

func TestTelegramAdminMenuSupportsSettingsBindingAndProblematicNodes(t *testing.T) {
	hub, admin := newTelegramHubWithAdmin(t)
	fake := &fakeTelegramTransport{}
	hub.telegramTransport = fake
	_, err := hub.saveTelegramSettings(TelegramSettingsInput{Enabled: true, PollingEnabled: true, BotToken: "123456:abcde_token_valid"}, telegramSettingsRecord{})
	require.NoError(t, err)
	_, err = hub.upsertTelegramDestination(nil, TelegramDestinationInput{Name: "Admin", ChatID: "12345", ChatType: TelegramChatTypePrivate, Role: TelegramRoleAdmin, Enabled: boolPtr(true)})
	require.NoError(t, err)
	for index := 0; index < 7; index++ {
		mustCreateTelegramRecord(t, hub, "systems", map[string]any{
			"name":   fmt.Sprintf("down-%d", index),
			"host":   fmt.Sprintf("127.0.0.%d", index+1),
			"status": "down",
			"users":  []string{admin.Id},
		})
	}

	require.NoError(t, hub.handleTelegramCommand(context.Background(), "token", telegramCommand{ChatID: "12345", Name: "settings"}))
	require.NoError(t, hub.handleTelegramCommand(context.Background(), "token", telegramCommand{ChatID: "12345", Name: "binding"}))
	require.NoError(t, hub.handleTelegramCommand(context.Background(), "token", telegramCommand{ChatID: "12345", Name: "status"}))
	require.Len(t, fake.sendCalls, 3)
	assert.Contains(t, fake.sendCalls[0].Text, "Telegram 通知设置")
	assert.Contains(t, fake.sendCalls[1].Text, "绑定已验证")
	assert.Contains(t, fake.sendCalls[2].Text, "最近异常节点")
	assert.LessOrEqual(t, strings.Count(fake.sendCalls[2].Text, "（down）"), 5)
}
