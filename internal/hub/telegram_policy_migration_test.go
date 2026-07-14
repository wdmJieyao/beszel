//go:build testing

package hub

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTelegramPolicyCollectionAndLegacyBackfill(t *testing.T) {
	hub, admin := newTelegramHubWithAdmin(t)
	_, err := hub.FindCollectionByNameOrId(CollectionTelegramNotificationPolicies)
	require.NoError(t, err)

	systems := mustCreateTelegramTestSystems(t, hub, admin.Id, 2)
	tests := []struct {
		name          string
		chatID        string
		nodeScope     []string
		expectedMode  string
		expectedNodes []string
	}{
		{name: "empty legacy scope maps to all", chatID: "10001", expectedMode: TelegramNodeScopeAll, expectedNodes: []string{}},
		{name: "legacy nodes map to selected", chatID: "10002", nodeScope: telegramRecordIDs(systems), expectedMode: TelegramNodeScopeSelected, expectedNodes: telegramRecordIDs(systems)},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			destination := mustCreateTelegramRecord(t, hub, CollectionTelegramDestinations, map[string]any{
				"name":              test.name,
				"chat_id":           test.chatID,
				"chat_type":         TelegramChatTypePrivate,
				"role":              TelegramRoleReadOnly,
				"enabled":           true,
				"node_scope":        test.nodeScope,
				"alert_level_scope": []string{"status", "cpu"},
			})

			require.NoError(t, hub.ensureTelegramDestinationDefaultPolicy(destination))
			require.NoError(t, hub.ensureTelegramDestinationDefaultPolicy(destination))

			policies, err := hub.listTelegramNotificationPolicies(destination.Id)
			require.NoError(t, err)
			require.Len(t, policies, 1)
			assert.Equal(t, TelegramDefaultPolicyName, policies[0].Name)
			assert.Equal(t, test.expectedMode, policies[0].NodeScopeMode)
			assert.Equal(t, test.expectedNodes, policies[0].NodeScope)
			assert.Equal(t, []string{"status", "cpu"}, policies[0].AlertLevelScope)
			assert.True(t, policies[0].Enabled)
		})
	}
}
