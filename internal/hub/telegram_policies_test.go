//go:build testing

package hub

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTelegramNotificationPolicyValidationAndStorage(t *testing.T) {
	hub, admin := newTelegramHubWithAdmin(t)
	systems := mustCreateTelegramTestSystems(t, hub, admin.Id, 2)
	destination := mustCreateTelegramRecord(t, hub, CollectionTelegramDestinations, map[string]any{
		"name": "Ops", "chat_id": "20001", "chat_type": TelegramChatTypePrivate,
		"role": TelegramRoleAdmin, "enabled": true,
	})

	created, err := hub.upsertTelegramNotificationPolicy(nil, destination.Id, TelegramNotificationPolicyInput{
		Name: "Primary", Enabled: boolPointer(true), NodeScopeMode: TelegramNodeScopeSelected,
		NodeScope: []string{systems[0].Id}, AlertLevelScope: []string{"status", "cpu"},
	})
	require.NoError(t, err)
	assert.Equal(t, destination.Id, created.GetString("destination"))

	_, err = hub.upsertTelegramNotificationPolicy(nil, destination.Id, TelegramNotificationPolicyInput{
		Name: "Primary", Enabled: boolPointer(true), NodeScopeMode: TelegramNodeScopeAll,
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "name")

	tests := []struct {
		name  string
		input TelegramNotificationPolicyInput
	}{
		{name: "unsupported mode", input: TelegramNotificationPolicyInput{Name: "Bad mode", NodeScopeMode: "some"}},
		{name: "all with nodes", input: TelegramNotificationPolicyInput{Name: "All with nodes", NodeScopeMode: TelegramNodeScopeAll, NodeScope: []string{systems[0].Id}}},
		{name: "selected empty", input: TelegramNotificationPolicyInput{Name: "Selected empty", NodeScopeMode: TelegramNodeScopeSelected}},
		{name: "unknown node", input: TelegramNotificationPolicyInput{Name: "Unknown node", NodeScopeMode: TelegramNodeScopeSelected, NodeScope: []string{"missing-node"}}},
		{name: "unknown category", input: TelegramNotificationPolicyInput{Name: "Unknown category", NodeScopeMode: TelegramNodeScopeAll, AlertLevelScope: []string{"secret"}}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := hub.upsertTelegramNotificationPolicy(nil, destination.Id, test.input)
			require.Error(t, err)
		})
	}

	_, err = hub.upsertTelegramNotificationPolicy(nil, "missing-destination", TelegramNotificationPolicyInput{
		Name: "Missing parent", NodeScopeMode: TelegramNodeScopeAll,
	})
	require.Error(t, err)

	require.NoError(t, hub.Delete(destination))
	policies, err := hub.listTelegramNotificationPolicies(destination.Id)
	require.NoError(t, err)
	assert.Empty(t, policies)
}

func boolPointer(value bool) *bool {
	return &value
}
