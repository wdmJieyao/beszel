//go:build testing

package hub

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"slices"
	"sync"
	"testing"

	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/tools/router"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTelegramNotificationPolicyNestedCRUDAndOwnership(t *testing.T) {
	hub, admin := newTelegramHubWithAdmin(t)
	system := mustCreateTelegramTestSystems(t, hub, admin.Id, 1)[0]
	destination := mustCreateTelegramRecord(t, hub, CollectionTelegramDestinations, map[string]any{
		"name": "Ops", "chat_id": "41001", "chat_type": TelegramChatTypePrivate, "role": TelegramRoleAdmin, "enabled": true,
	})
	other := mustCreateTelegramRecord(t, hub, CollectionTelegramDestinations, map[string]any{
		"name": "Other", "chat_id": "41002", "chat_type": TelegramChatTypePrivate, "role": TelegramRoleAdmin, "enabled": true,
	})

	create := telegramPolicyRequestEvent(hub, admin, http.MethodPost, destination.Id, "", `{"name":"CPU","enabled":true,"nodeScopeMode":"selected","nodeScope":["`+system.Id+`"],"alertLevelScope":["cpu"]}`)
	require.NoError(t, hub.createTelegramNotificationPolicy(create))
	assert.Equal(t, http.StatusCreated, create.Response.(*httptest.ResponseRecorder).Code)
	var created TelegramNotificationPolicyResponse
	require.NoError(t, json.Unmarshal(create.Response.(*httptest.ResponseRecorder).Body.Bytes(), &created))
	assert.Equal(t, destination.Id, created.DestinationID)

	list := telegramPolicyRequestEvent(hub, admin, http.MethodGet, destination.Id, "", "")
	require.NoError(t, hub.listTelegramNotificationPoliciesHandler(list))
	assert.Contains(t, list.Response.(*httptest.ResponseRecorder).Body.String(), `"name":"CPU"`)

	update := telegramPolicyRequestEvent(hub, admin, http.MethodPatch, destination.Id, created.ID, `{"name":"CPU critical","enabled":true,"nodeScopeMode":"all","nodeScope":[],"alertLevelScope":["cpu"]}`)
	require.NoError(t, hub.updateTelegramNotificationPolicy(update))
	assert.Contains(t, update.Response.(*httptest.ResponseRecorder).Body.String(), "CPU critical")

	crossChannel := telegramPolicyRequestEvent(hub, admin, http.MethodPatch, other.Id, created.ID, `{"name":"stolen","enabled":true,"nodeScopeMode":"all","nodeScope":[],"alertLevelScope":[]}`)
	err := hub.updateTelegramNotificationPolicy(crossChannel)
	require.Error(t, err)
	assert.Equal(t, http.StatusNotFound, router.ToApiError(err).Status)

	remove := telegramPolicyRequestEvent(hub, admin, http.MethodDelete, destination.Id, created.ID, "")
	require.NoError(t, hub.deleteTelegramNotificationPolicy(remove))
	policies, err := hub.listTelegramNotificationPolicies(destination.Id)
	require.NoError(t, err)
	assert.Empty(t, policies)
}

func TestTelegramNotificationPolicyDuplicateNameReturnsConflict(t *testing.T) {
	hub, admin := newTelegramHubWithAdmin(t)
	destination := mustCreateTelegramRecord(t, hub, CollectionTelegramDestinations, map[string]any{
		"name": "Ops", "chat_id": "41011", "chat_type": TelegramChatTypePrivate, "role": TelegramRoleAdmin, "enabled": true,
	})
	first := telegramPolicyRequestEvent(hub, admin, http.MethodPost, destination.Id, "", `{"name":"CPU","enabled":true,"nodeScopeMode":"all","nodeScope":[],"alertLevelScope":["cpu"]}`)
	require.NoError(t, hub.createTelegramNotificationPolicy(first))
	var created TelegramNotificationPolicyResponse
	require.NoError(t, json.Unmarshal(first.Response.(*httptest.ResponseRecorder).Body.Bytes(), &created))

	duplicate := telegramPolicyRequestEvent(hub, admin, http.MethodPost, destination.Id, "", `{"name":"CPU","enabled":true,"nodeScopeMode":"all","nodeScope":[],"alertLevelScope":[]}`)
	require.NoError(t, hub.createTelegramNotificationPolicy(duplicate))
	assert.Equal(t, http.StatusConflict, duplicate.Response.(*httptest.ResponseRecorder).Code)

	second := telegramPolicyRequestEvent(hub, admin, http.MethodPost, destination.Id, "", `{"name":"Memory","enabled":true,"nodeScopeMode":"all","nodeScope":[],"alertLevelScope":["memory"]}`)
	require.NoError(t, hub.createTelegramNotificationPolicy(second))
	var secondPolicy TelegramNotificationPolicyResponse
	require.NoError(t, json.Unmarshal(second.Response.(*httptest.ResponseRecorder).Body.Bytes(), &secondPolicy))
	rename := telegramPolicyRequestEvent(hub, admin, http.MethodPatch, destination.Id, secondPolicy.ID, `{"name":"CPU"}`)
	require.NoError(t, hub.updateTelegramNotificationPolicy(rename))
	assert.Equal(t, http.StatusConflict, rename.Response.(*httptest.ResponseRecorder).Code)

	unchanged, err := hub.findTelegramNotificationPolicy(destination.Id, created.ID)
	require.NoError(t, err)
	assert.Equal(t, "CPU", unchanged.GetString("name"))
}

func TestTelegramNotificationPolicyPatchPreservesOmittedFields(t *testing.T) {
	hub, admin := newTelegramHubWithAdmin(t)
	system := mustCreateTelegramTestSystems(t, hub, admin.Id, 1)[0]
	destination := mustCreateTelegramRecord(t, hub, CollectionTelegramDestinations, map[string]any{
		"name": "Ops", "chat_id": "41012", "chat_type": TelegramChatTypePrivate, "role": TelegramRoleAdmin, "enabled": true,
	})
	created, err := hub.upsertTelegramNotificationPolicy(nil, destination.Id, TelegramNotificationPolicyInput{
		Name: "CPU", Enabled: boolPtr(true), NodeScopeMode: TelegramNodeScopeSelected,
		NodeScope: []string{system.Id}, AlertLevelScope: []string{"cpu"},
	})
	require.NoError(t, err)

	patch := telegramPolicyRequestEvent(hub, admin, http.MethodPatch, destination.Id, created.Id, `{"enabled":false}`)
	require.NoError(t, hub.updateTelegramNotificationPolicy(patch))
	assert.Equal(t, http.StatusOK, patch.Response.(*httptest.ResponseRecorder).Code)
	var response TelegramNotificationPolicyResponse
	require.NoError(t, json.Unmarshal(patch.Response.(*httptest.ResponseRecorder).Body.Bytes(), &response))
	assert.Equal(t, "CPU", response.Name)
	assert.False(t, response.Enabled)
	assert.Equal(t, TelegramNodeScopeSelected, response.NodeScopeMode)
	assert.Equal(t, []string{system.Id}, response.NodeScope)
	assert.Equal(t, []string{"cpu"}, response.AlertLevelScope)
}

func TestTelegramNotificationPolicyPatchRejectsNullFields(t *testing.T) {
	hub, admin := newTelegramHubWithAdmin(t)
	destination := mustCreateTelegramRecord(t, hub, CollectionTelegramDestinations, map[string]any{
		"name": "Ops", "chat_id": "41014", "chat_type": TelegramChatTypePrivate, "role": TelegramRoleAdmin, "enabled": true,
	})
	policy, err := hub.upsertTelegramNotificationPolicy(nil, destination.Id, TelegramNotificationPolicyInput{
		Name: "Policy", Enabled: boolPtr(true), NodeScopeMode: TelegramNodeScopeAll,
	})
	require.NoError(t, err)
	for _, field := range []string{"name", "enabled", "nodeScopeMode", "nodeScope", "alertLevelScope"} {
		t.Run(field, func(t *testing.T) {
			event := telegramPolicyRequestEvent(hub, admin, http.MethodPatch, destination.Id, policy.Id, `{"`+field+`":null}`)
			err := hub.updateTelegramNotificationPolicy(event)
			require.Error(t, err)
			assert.Equal(t, http.StatusBadRequest, router.ToApiError(err).Status)
		})
	}
}

func TestTelegramPolicyUniqueIndexValidationIsClassifiedAsConflict(t *testing.T) {
	hub, _ := newTelegramHubWithAdmin(t)
	destination := mustCreateTelegramRecord(t, hub, CollectionTelegramDestinations, map[string]any{
		"name": "Ops", "chat_id": "41015", "chat_type": TelegramChatTypePrivate, "role": TelegramRoleAdmin, "enabled": true,
	})
	_, err := hub.upsertTelegramNotificationPolicy(nil, destination.Id, TelegramNotificationPolicyInput{
		Name: "Duplicate", Enabled: boolPtr(true), NodeScopeMode: TelegramNodeScopeAll,
	})
	require.NoError(t, err)
	collection, err := hub.FindCachedCollectionByNameOrId(CollectionTelegramNotificationPolicies)
	require.NoError(t, err)
	duplicate := core.NewRecord(collection)
	duplicate.Set("destination", destination.Id)
	duplicate.Set("name", "Duplicate")
	duplicate.Set("enabled", true)
	duplicate.Set("node_scope_mode", TelegramNodeScopeAll)
	err = hub.Save(duplicate)
	require.Error(t, err)
	assert.True(t, isTelegramPolicyNameConflictError(err))
}

func TestConcurrentTelegramPolicyDuplicateCreateAndRenameReturnConflict(t *testing.T) {
	t.Run("create", func(t *testing.T) {
		hub, admin := newTelegramHubWithAdmin(t)
		destination := mustCreateTelegramRecord(t, hub, CollectionTelegramDestinations, map[string]any{
			"name": "Ops", "chat_id": "41016", "chat_type": TelegramChatTypePrivate, "role": TelegramRoleAdmin, "enabled": true,
		})
		statuses := runConcurrentTelegramPolicyRequests(2, func() *core.RequestEvent {
			return telegramPolicyRequestEvent(hub, admin, http.MethodPost, destination.Id, "", `{"name":"Same","enabled":true,"nodeScopeMode":"all","nodeScope":[],"alertLevelScope":[]}`)
		}, hub.createTelegramNotificationPolicy)
		assert.Equal(t, []int{http.StatusCreated, http.StatusConflict}, statuses)
	})

	t.Run("rename", func(t *testing.T) {
		hub, admin := newTelegramHubWithAdmin(t)
		destination := mustCreateTelegramRecord(t, hub, CollectionTelegramDestinations, map[string]any{
			"name": "Ops", "chat_id": "41017", "chat_type": TelegramChatTypePrivate, "role": TelegramRoleAdmin, "enabled": true,
		})
		policies := make([]*core.Record, 2)
		for index, name := range []string{"First", "Second"} {
			var err error
			policies[index], err = hub.upsertTelegramNotificationPolicy(nil, destination.Id, TelegramNotificationPolicyInput{
				Name: name, Enabled: boolPtr(true), NodeScopeMode: TelegramNodeScopeAll,
			})
			require.NoError(t, err)
		}
		requestIndex := 0
		var indexMu sync.Mutex
		statuses := runConcurrentTelegramPolicyRequests(2, func() *core.RequestEvent {
			indexMu.Lock()
			policy := policies[requestIndex]
			requestIndex++
			indexMu.Unlock()
			return telegramPolicyRequestEvent(hub, admin, http.MethodPatch, destination.Id, policy.Id, `{"name":"Renamed"}`)
		}, hub.updateTelegramNotificationPolicy)
		assert.Equal(t, []int{http.StatusOK, http.StatusConflict}, statuses)
	})
}

func runConcurrentTelegramPolicyRequests(count int, event func() *core.RequestEvent, handler func(*core.RequestEvent) error) []int {
	start := make(chan struct{})
	statuses := make(chan int, count)
	var wait sync.WaitGroup
	wait.Add(count)
	for range count {
		e := event()
		go func() {
			defer wait.Done()
			<-start
			if err := handler(e); err != nil {
				statuses <- router.ToApiError(err).Status
				return
			}
			statuses <- e.Response.(*httptest.ResponseRecorder).Code
		}()
	}
	close(start)
	wait.Wait()
	close(statuses)
	result := make([]int, 0, count)
	for status := range statuses {
		result = append(result, status)
	}
	slices.Sort(result)
	return result
}

func TestConcurrentTelegramPolicyUpdateAndSiblingDeleteRemainConsistent(t *testing.T) {
	hub, admin := newTelegramHubWithAdmin(t)
	destination := mustCreateTelegramRecord(t, hub, CollectionTelegramDestinations, map[string]any{
		"name": "Ops", "chat_id": "41013", "chat_type": TelegramChatTypePrivate, "role": TelegramRoleAdmin, "enabled": true,
	})
	first, err := hub.upsertTelegramNotificationPolicy(nil, destination.Id, TelegramNotificationPolicyInput{
		Name: "Delete", Enabled: boolPtr(true), NodeScopeMode: TelegramNodeScopeAll,
	})
	require.NoError(t, err)
	second, err := hub.upsertTelegramNotificationPolicy(nil, destination.Id, TelegramNotificationPolicyInput{
		Name: "Update", Enabled: boolPtr(true), NodeScopeMode: TelegramNodeScopeAll,
	})
	require.NoError(t, err)

	start := make(chan struct{})
	errors := make(chan error, 2)
	var wait sync.WaitGroup
	wait.Add(2)
	go func() {
		defer wait.Done()
		<-start
		event := telegramPolicyRequestEvent(hub, admin, http.MethodDelete, destination.Id, first.Id, "")
		errors <- hub.deleteTelegramNotificationPolicy(event)
	}()
	go func() {
		defer wait.Done()
		<-start
		event := telegramPolicyRequestEvent(hub, admin, http.MethodPatch, destination.Id, second.Id, `{"enabled":false}`)
		errors <- hub.updateTelegramNotificationPolicy(event)
	}()
	close(start)
	wait.Wait()
	close(errors)
	for operationErr := range errors {
		require.NoError(t, operationErr)
	}

	_, err = hub.findTelegramNotificationPolicy(destination.Id, first.Id)
	require.Error(t, err)
	updated, err := hub.findTelegramNotificationPolicy(destination.Id, second.Id)
	require.NoError(t, err)
	assert.False(t, updated.GetBool("enabled"))
	assert.Equal(t, "Update", updated.GetString("name"))
}

func TestCreateTelegramDestinationDuplicateReturnsConflictMetadata(t *testing.T) {
	hub, admin := newTelegramHubWithAdmin(t)
	existing := mustCreateTelegramRecord(t, hub, CollectionTelegramDestinations, map[string]any{
		"name": "Existing", "chat_id": "42001", "chat_type": TelegramChatTypePrivate, "role": TelegramRoleAdmin, "enabled": true,
	})
	req := httptest.NewRequest(http.MethodPost, "/api/beszel/telegram/destinations", bytes.NewBufferString(`{"name":"Duplicate","chatId":"42001","chatType":"private","role":"admin","enabled":true,"nodeScope":[],"alertLevelScope":[]}`))
	req.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()
	require.NoError(t, hub.createTelegramDestination(&core.RequestEvent{App: hub, Auth: admin, Event: router.Event{Request: req, Response: recorder}}))
	assert.Equal(t, http.StatusConflict, recorder.Code)
	assert.Contains(t, recorder.Body.String(), existing.Id)
}

func TestCreateTelegramDestinationRollsBackWhenDefaultPolicyCannotBeSaved(t *testing.T) {
	hub, admin := newTelegramHubWithAdmin(t)
	collection, err := hub.FindCollectionByNameOrId(CollectionTelegramNotificationPolicies)
	require.NoError(t, err)
	require.NoError(t, hub.Delete(collection))

	req := httptest.NewRequest(http.MethodPost, "/api/beszel/telegram/destinations", bytes.NewBufferString(`{"name":"Rollback","chatId":"42003","chatType":"private","role":"admin","enabled":true,"nodeScope":[],"alertLevelScope":[]}`))
	req.Header.Set("Content-Type", "application/json")
	err = hub.createTelegramDestination(&core.RequestEvent{App: hub, Auth: admin, Event: router.Event{Request: req, Response: httptest.NewRecorder()}})
	require.Error(t, err)
	_, findErr := hub.findTelegramDestinationByChatID("42003")
	require.Error(t, findErr)
}

func telegramPolicyRequestEvent(hub *Hub, admin *core.Record, method, destinationID, policyID, body string) *core.RequestEvent {
	request := httptest.NewRequest(method, "/api/beszel/telegram/destinations/"+destinationID+"/policies/"+policyID, bytes.NewBufferString(body))
	request.Header.Set("Content-Type", "application/json")
	request.SetPathValue("destinationId", destinationID)
	request.SetPathValue("policyId", policyID)
	return &core.RequestEvent{App: hub, Auth: admin, Event: router.Event{Request: request, Response: httptest.NewRecorder()}}
}
