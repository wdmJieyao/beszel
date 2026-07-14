//go:build testing

package hub

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"
	"unicode/utf8"

	"github.com/henrygd/beszel/internal/alerts"
	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase/core"
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
		Message:  "Connection to node-1 is down at 127.0.0.1 https://internal.example/probe",
		Link:     "https://example.com/system/node-1",
	})
	require.NoError(t, err)
	require.Len(t, fake.sendCalls, 2)
	assert.Contains(t, fake.sendCalls[0].Text, "https://example.com/system/node-1")
	assert.True(t, strings.HasPrefix(fake.sendCalls[1].Text, "Beszel 告警摘要"))
	assert.NotContains(t, fake.sendCalls[1].Text, "https://example.com/system/node-1")
	assert.NotContains(t, fake.sendCalls[1].Text, "127.0.0.1")
}

func TestTelegramMessageLengthIsBounded(t *testing.T) {
	message := strings.Repeat("测", telegramMessageMaxRunes+100)
	bounded := truncateTelegramMessage(message)
	assert.LessOrEqual(t, utf8.RuneCountInString(bounded), telegramMessageMaxRunes)
	assert.True(t, strings.HasSuffix(bounded, telegramMessageTruncatedSuffix))
}

func TestTelegramAlertFormattingIncludesStructuredContext(t *testing.T) {
	eventTime := time.Date(2026, time.July, 10, 9, 30, 0, 0, time.UTC)
	message := telegramMessageForDestination(telegramDestinationRecord{Role: TelegramRoleAdmin}, alerts.AlertMessageData{
		SystemName: "node-1", AlertClass: "cpu", Severity: alerts.AlertSeverityCritical,
		State: alerts.AlertStateTriggered, EventTime: eventTime, Title: "CPU threshold exceeded", Message: "CPU averaged 95%",
	})
	assert.Contains(t, message, "节点：node-1")
	assert.Contains(t, message, "类型：cpu")
	assert.Contains(t, message, "严重级别：critical")
	assert.Contains(t, message, "状态：triggered")
	assert.Contains(t, message, eventTime.Format(time.RFC3339))
}

func TestTelegramSendRetriesRetryableFailures(t *testing.T) {
	fake := &sequenceTelegramTransport{errors: []error{errors.New("telegram 429 too many requests"), errors.New("temporary timeout"), nil}}
	err := sendTelegramMessageWithRetry(context.Background(), fake, "token", "12345", "message", nil, telegramRetryPolicy{MaxAttempts: 3, InitialBackoff: time.Millisecond})
	require.NoError(t, err)
	assert.Equal(t, 3, fake.calls)
}

func TestTelegramDeliveryQueueIsBounded(t *testing.T) {
	state := &telegramDeliveryState{pending: telegramDeliveryQueueLimit}
	err := state.enter()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "queue is full")
}

func TestTelegramDeliveryRateLimitWaitsBetweenSends(t *testing.T) {
	state := &telegramDeliveryState{interval: 10 * time.Millisecond}
	require.NoError(t, state.waitForRateLimit(context.Background()))
	started := time.Now()
	require.NoError(t, state.waitForRateLimit(context.Background()))
	assert.GreaterOrEqual(t, time.Since(started), 8*time.Millisecond)
}

func TestTelegramPoliciesUseORSemanticsAndDeliverOncePerChannel(t *testing.T) {
	hub, admin := newTelegramHubWithAdmin(t)
	fake := &fakeTelegramTransport{}
	hub.telegramTransport = fake
	_, err := hub.saveTelegramSettings(TelegramSettingsInput{Enabled: true, BotToken: "123456:abcde_token_valid"}, telegramSettingsRecord{})
	require.NoError(t, err)
	systems := mustCreateTelegramTestSystems(t, hub, admin.Id, 2)
	destination := mustCreateTelegramRecord(t, hub, CollectionTelegramDestinations, map[string]any{
		"name": "Ops", "chat_id": "43001", "chat_type": TelegramChatTypePrivate, "role": TelegramRoleAdmin, "enabled": true,
	})
	for _, name := range []string{"CPU all", "CPU node"} {
		mode := TelegramNodeScopeAll
		nodes := []string{}
		if name == "CPU node" {
			mode = TelegramNodeScopeSelected
			nodes = []string{systems[0].Id}
		}
		_, err = hub.upsertTelegramNotificationPolicy(nil, destination.Id, TelegramNotificationPolicyInput{
			Name: name, Enabled: boolPtr(true), NodeScopeMode: mode, NodeScope: nodes, AlertLevelScope: []string{"cpu"},
		})
		require.NoError(t, err)
	}

	require.NoError(t, hub.SendTelegramAlert(alerts.AlertMessageData{UserID: admin.Id, SystemID: systems[0].Id, AlertClass: "cpu", Title: "CPU high"}))
	require.Len(t, fake.sendCalls, 1)
	assert.Equal(t, "43001", fake.sendCalls[0].ChatID)

	for _, policy := range mustListTelegramPolicyRecords(t, hub, destination.Id) {
		policy.Set("enabled", false)
		require.NoError(t, hub.Save(policy))
	}
	require.NoError(t, hub.SendTelegramAlert(alerts.AlertMessageData{UserID: admin.Id, SystemID: systems[0].Id, AlertClass: "cpu", Title: "CPU high"}))
	assert.Len(t, fake.sendCalls, 1)
}

func TestTelegramChannelWithoutPoliciesDoesNotReceiveAlerts(t *testing.T) {
	hub, admin := newTelegramHubWithAdmin(t)
	fake := &fakeTelegramTransport{}
	hub.telegramTransport = fake
	_, err := hub.saveTelegramSettings(TelegramSettingsInput{Enabled: true, BotToken: "123456:abcde_token_valid"}, telegramSettingsRecord{})
	require.NoError(t, err)
	system := mustCreateTelegramTestSystems(t, hub, admin.Id, 1)[0]
	mustCreateTelegramRecord(t, hub, CollectionTelegramDestinations, map[string]any{
		"name": "No policy", "chat_id": "43002", "chat_type": TelegramChatTypePrivate, "role": TelegramRoleAdmin, "enabled": true,
	})
	require.NoError(t, hub.SendTelegramAlert(alerts.AlertMessageData{UserID: admin.Id, SystemID: system.Id, AlertClass: "status", Title: "Down"}))
	assert.Empty(t, fake.sendCalls)
}

func TestTelegramPolicyScopesApplyToAdminAndReadOnlyBeforeFormatting(t *testing.T) {
	hub, admin := newTelegramHubWithAdmin(t)
	fake := &fakeTelegramTransport{}
	hub.telegramTransport = fake
	_, err := hub.saveTelegramSettings(TelegramSettingsInput{Enabled: true, BotToken: "123456:abcde_token_valid"}, telegramSettingsRecord{})
	require.NoError(t, err)
	systems := mustCreateTelegramTestSystems(t, hub, admin.Id, 2)
	for index, role := range []string{TelegramRoleAdmin, TelegramRoleReadOnly} {
		destination := mustCreateTelegramRecord(t, hub, CollectionTelegramDestinations, map[string]any{
			"name": role, "chat_id": fmt.Sprintf("4500%d", index), "chat_type": TelegramChatTypePrivate,
			"role": role, "enabled": true,
		})
		_, err := hub.upsertTelegramNotificationPolicy(nil, destination.Id, TelegramNotificationPolicyInput{
			Name: "CPU node one", Enabled: boolPtr(true), NodeScopeMode: TelegramNodeScopeSelected,
			NodeScope: []string{systems[0].Id}, AlertLevelScope: []string{"cpu"},
		})
		require.NoError(t, err)
	}

	for _, alert := range []alerts.AlertMessageData{
		{UserID: admin.Id, SystemID: systems[0].Id, AlertClass: "status", Title: "Down"},
		{UserID: admin.Id, SystemID: systems[1].Id, AlertClass: "cpu", Title: "CPU", Link: "https://private.example"},
	} {
		require.NoError(t, hub.SendTelegramAlert(alert))
	}
	assert.Empty(t, fake.sendCalls)

	require.NoError(t, hub.SendTelegramAlert(alerts.AlertMessageData{
		UserID: admin.Id, SystemID: systems[0].Id, AlertClass: "cpu", Title: "CPU",
		Message: "host 10.0.0.1", Link: "https://private.example",
	}))
	require.Len(t, fake.sendCalls, 2)
	assert.Contains(t, fake.sendCalls[0].Text, "https://private.example")
	assert.NotContains(t, fake.sendCalls[1].Text, "https://private.example")
	assert.NotContains(t, fake.sendCalls[1].Text, "10.0.0.1")
}

func TestTelegramAllNodePolicyIncludesFutureSystems(t *testing.T) {
	policy := telegramNotificationPolicyRecord{Enabled: true, NodeScopeMode: TelegramNodeScopeAll, NodeScope: []string{}}
	assert.True(t, telegramPoliciesMatchAlert([]telegramNotificationPolicyRecord{policy}, alerts.AlertMessageData{SystemID: "future-system", AlertClass: "status"}))
}

func mustListTelegramPolicyRecords(t *testing.T, hub *Hub, destinationID string) []*core.Record {
	t.Helper()
	records, err := hub.FindRecordsByFilter(CollectionTelegramNotificationPolicies, "destination = {:destination}", "", -1, 0, dbx.Params{"destination": destinationID})
	require.NoError(t, err)
	return records
}

type sequenceTelegramTransport struct {
	fakeTelegramTransport
	errors []error
	calls  int
}

func (f *sequenceTelegramTransport) SendMessage(ctx context.Context, token, chatID, text string, options *TelegramSendOptions) error {
	f.calls++
	if f.calls <= len(f.errors) {
		return f.errors[f.calls-1]
	}
	return nil
}

func boolPtr(v bool) *bool {
	return &v
}

func timePtr(v time.Time) *time.Time {
	return &v
}
