//go:build testing

package hub

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/tools/router"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateUpdateDeleteTelegramDestination(t *testing.T) {
	hub, admin := newTelegramHubWithAdmin(t)
	system := mustCreateTelegramRecord(t, hub, "systems", map[string]any{
		"name":   "node-1",
		"host":   "127.0.0.1",
		"status": "up",
		"users":  []string{admin.Id},
	})

	createBody := bytes.NewBufferString(`{"name":"Ops","chatId":"-10012345","chatType":"channel","role":"read_only","enabled":true,"nodeScope":["` + system.Id + `"],"alertLevelScope":["status"]}`)
	createReq := httptest.NewRequest(http.MethodPost, "/api/beszel/telegram/destinations", createBody)
	createReq.Header.Set("Content-Type", "application/json")
	createRecorder := httptest.NewRecorder()
	require.NoError(t, hub.createTelegramDestination(&core.RequestEvent{
		App:  hub,
		Auth: admin,
		Event: router.Event{
			Request:  createReq,
			Response: createRecorder,
		},
	}))

	var created TelegramDestinationResponse
	require.NoError(t, json.Unmarshal(createRecorder.Body.Bytes(), &created))
	assert.Equal(t, "Ops", created.Name)
	assert.Equal(t, []string{system.Id}, created.NodeScope)

	updateBody := bytes.NewBufferString(`{"name":"Ops Updated","chatId":"-10012345","chatType":"channel","role":"admin","enabled":false,"nodeScope":[],"alertLevelScope":[]}`)
	updateReq := httptest.NewRequest(http.MethodPatch, "/api/beszel/telegram/destinations/"+created.ID, updateBody)
	updateReq.Header.Set("Content-Type", "application/json")
	updateReq.SetPathValue("destinationId", created.ID)
	updateRecorder := httptest.NewRecorder()
	require.NoError(t, hub.updateTelegramDestination(&core.RequestEvent{
		App:  hub,
		Auth: admin,
		Event: router.Event{
			Request:  updateReq,
			Response: updateRecorder,
		},
	}))

	var updated TelegramDestinationResponse
	require.NoError(t, json.Unmarshal(updateRecorder.Body.Bytes(), &updated))
	assert.Equal(t, "Ops Updated", updated.Name)
	assert.Equal(t, TelegramRoleAdmin, updated.Role)
	assert.False(t, updated.Enabled)

	listReq := httptest.NewRequest(http.MethodGet, "/api/beszel/telegram/destinations", nil)
	listRecorder := httptest.NewRecorder()
	require.NoError(t, hub.listTelegramDestinationsHandler(&core.RequestEvent{
		App:  hub,
		Auth: admin,
		Event: router.Event{
			Request:  listReq,
			Response: listRecorder,
		},
	}))
	assert.Contains(t, listRecorder.Body.String(), "Ops Updated")

	deleteReq := httptest.NewRequest(http.MethodDelete, "/api/beszel/telegram/destinations/"+created.ID, nil)
	deleteReq.SetPathValue("destinationId", created.ID)
	deleteRecorder := httptest.NewRecorder()
	require.NoError(t, hub.deleteTelegramDestination(&core.RequestEvent{
		App:  hub,
		Auth: admin,
		Event: router.Event{
			Request:  deleteReq,
			Response: deleteRecorder,
		},
	}))

	destinations, err := hub.listTelegramDestinations()
	require.NoError(t, err)
	assert.Empty(t, destinations)
}

func TestCreateTelegramDestinationValidatesInput(t *testing.T) {
	hub, admin := newTelegramHubWithAdmin(t)

	req := httptest.NewRequest(http.MethodPost, "/api/beszel/telegram/destinations", bytes.NewBufferString(`{"name":"","chatId":"bad","role":"oops"}`))
	req.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()

	err := hub.createTelegramDestination(&core.RequestEvent{
		App:  hub,
		Auth: admin,
		Event: router.Event{
			Request:  req,
			Response: recorder,
		},
	})
	require.Error(t, err)
	apiErr := router.ToApiError(err)
	assert.Equal(t, http.StatusBadRequest, apiErr.Status)
}

func TestTelegramDestinationValidatesSupportedAlertScopes(t *testing.T) {
	valid := TelegramDestinationInput{
		Name:            "Ops",
		ChatID:          "12345",
		Role:            TelegramRoleReadOnly,
		AlertLevelScope: append([]string(nil), TelegramAlertLevelScopes...),
	}
	require.NoError(t, validateTelegramDestinationInput(valid))

	invalid := valid
	invalid.AlertLevelScope = []string{"status", "arbitrary-secret-category"}
	err := validateTelegramDestinationInput(invalid)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "alertLevelScope")
}

func TestTelegramDestinationPatchPreservesOmittedFieldsAndCanClearMute(t *testing.T) {
	hub, admin := newTelegramHubWithAdmin(t)
	muteUntil := time.Now().UTC().Add(time.Hour)
	destination, err := hub.upsertTelegramDestination(nil, TelegramDestinationInput{
		Name: "Ops", ChatID: "42021", ChatType: TelegramChatTypePrivate,
		Role: TelegramRoleAdmin, Enabled: boolPtr(true), MuteUntil: &muteUntil,
	})
	require.NoError(t, err)

	patchReq := httptest.NewRequest(http.MethodPatch, "/api/beszel/telegram/destinations/"+destination.Id, bytes.NewBufferString(`{"enabled":false}`))
	patchReq.Header.Set("Content-Type", "application/json")
	patchReq.SetPathValue("destinationId", destination.Id)
	patchRecorder := httptest.NewRecorder()
	require.NoError(t, hub.updateTelegramDestination(&core.RequestEvent{
		App: hub, Auth: admin, Event: router.Event{Request: patchReq, Response: patchRecorder},
	}))
	var patched TelegramDestinationResponse
	require.NoError(t, json.Unmarshal(patchRecorder.Body.Bytes(), &patched))
	assert.Equal(t, "Ops", patched.Name)
	assert.Equal(t, "42021", patched.ChatID)
	assert.Equal(t, TelegramRoleAdmin, patched.Role)
	assert.False(t, patched.Enabled)
	assert.NotEmpty(t, patched.MuteUntil)

	clearReq := httptest.NewRequest(http.MethodPatch, "/api/beszel/telegram/destinations/"+destination.Id, bytes.NewBufferString(`{"muteUntil":null}`))
	clearReq.Header.Set("Content-Type", "application/json")
	clearReq.SetPathValue("destinationId", destination.Id)
	clearRecorder := httptest.NewRecorder()
	require.NoError(t, hub.updateTelegramDestination(&core.RequestEvent{
		App: hub, Auth: admin, Event: router.Event{Request: clearReq, Response: clearRecorder},
	}))
	patched = TelegramDestinationResponse{}
	require.NoError(t, json.Unmarshal(clearRecorder.Body.Bytes(), &patched))
	assert.Empty(t, patched.MuteUntil)
}

func TestTelegramDestinationPatchRejectsNullForNonNullableFields(t *testing.T) {
	hub, admin := newTelegramHubWithAdmin(t)
	destination := mustCreateTelegramRecord(t, hub, CollectionTelegramDestinations, map[string]any{
		"name": "Ops", "chat_id": "42022", "chat_type": TelegramChatTypePrivate, "role": TelegramRoleAdmin, "enabled": true,
	})
	for _, field := range []string{"name", "chatId", "chatType", "role", "enabled", "nodeScope", "alertLevelScope"} {
		t.Run(field, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPatch, "/api/beszel/telegram/destinations/"+destination.Id, bytes.NewBufferString(`{"`+field+`":null}`))
			req.Header.Set("Content-Type", "application/json")
			req.SetPathValue("destinationId", destination.Id)
			err := hub.updateTelegramDestination(&core.RequestEvent{App: hub, Auth: admin, Event: router.Event{Request: req, Response: httptest.NewRecorder()}})
			require.Error(t, err)
			assert.Equal(t, http.StatusBadRequest, router.ToApiError(err).Status)
		})
	}
}

func TestConcurrentTelegramDestinationTestEditAndDeleteDoesNotResurrectChannel(t *testing.T) {
	hub, admin := newTelegramHubWithAdmin(t)
	_, err := hub.saveTelegramSettings(TelegramSettingsInput{
		Enabled: true, BotToken: "123456:concurrent_token_value",
	}, telegramSettingsRecord{})
	require.NoError(t, err)
	destination := mustCreateTelegramRecord(t, hub, CollectionTelegramDestinations, map[string]any{
		"name": "Concurrent", "chat_id": "42031", "chat_type": TelegramChatTypePrivate, "role": TelegramRoleAdmin, "enabled": true,
	})
	fake := &fakeTelegramTransport{sendStarted: make(chan struct{}), sendRelease: make(chan struct{})}
	hub.telegramTransport = fake

	testDone := make(chan error, 1)
	go func() {
		req := httptest.NewRequest(http.MethodPost, "/api/beszel/telegram/destinations/"+destination.Id+"/test", nil)
		req.SetPathValue("destinationId", destination.Id)
		testDone <- hub.testTelegramDestination(&core.RequestEvent{
			App: hub, Auth: admin, Event: router.Event{Request: req, Response: httptest.NewRecorder()},
		})
	}()
	<-fake.sendStarted

	patchReq := httptest.NewRequest(http.MethodPatch, "/api/beszel/telegram/destinations/"+destination.Id, bytes.NewBufferString(`{"name":"Edited while testing"}`))
	patchReq.Header.Set("Content-Type", "application/json")
	patchReq.SetPathValue("destinationId", destination.Id)
	require.NoError(t, hub.updateTelegramDestination(&core.RequestEvent{
		App: hub, Auth: admin, Event: router.Event{Request: patchReq, Response: httptest.NewRecorder()},
	}))

	deleteReq := httptest.NewRequest(http.MethodDelete, "/api/beszel/telegram/destinations/"+destination.Id, nil)
	deleteReq.SetPathValue("destinationId", destination.Id)
	require.NoError(t, hub.deleteTelegramDestination(&core.RequestEvent{
		App: hub, Auth: admin, Event: router.Event{Request: deleteReq, Response: httptest.NewRecorder()},
	}))
	close(fake.sendRelease)
	require.NoError(t, <-testDone)

	_, err = hub.findTelegramDestinationByID(destination.Id)
	require.Error(t, err)
	assert.Len(t, fake.sendCalls, 1)
}

func TestDeleteTelegramDestinationOnlyDeletesTargetAndPolicies(t *testing.T) {
	hub, admin := newTelegramHubWithAdmin(t)
	_, err := hub.saveTelegramSettings(TelegramSettingsInput{Enabled: true, BotToken: "123456:saved_token_value"}, telegramSettingsRecord{})
	require.NoError(t, err)
	target := mustCreateTelegramRecord(t, hub, CollectionTelegramDestinations, map[string]any{
		"name": "Delete me", "chat_id": "30001", "chat_type": TelegramChatTypePrivate, "role": TelegramRoleAdmin, "enabled": true,
	})
	sibling := mustCreateTelegramRecord(t, hub, CollectionTelegramDestinations, map[string]any{
		"name": "Keep me", "chat_id": "30002", "chat_type": TelegramChatTypePrivate, "role": TelegramRoleAdmin, "enabled": true,
	})
	require.NoError(t, hub.ensureTelegramDestinationDefaultPolicy(target))
	require.NoError(t, hub.ensureTelegramDestinationDefaultPolicy(sibling))

	deleteEvent := func() *core.RequestEvent {
		req := httptest.NewRequest(http.MethodDelete, "/api/beszel/telegram/destinations/"+target.Id, nil)
		req.SetPathValue("destinationId", target.Id)
		return &core.RequestEvent{App: hub, Auth: admin, Event: router.Event{Request: req, Response: httptest.NewRecorder()}}
	}
	require.NoError(t, hub.deleteTelegramDestination(deleteEvent()))
	_, err = hub.findTelegramDestinationByID(target.Id)
	require.Error(t, err)
	policies, err := hub.listTelegramNotificationPolicies(target.Id)
	require.NoError(t, err)
	assert.Empty(t, policies)
	_, err = hub.findTelegramDestinationByID(sibling.Id)
	require.NoError(t, err)
	siblingPolicies, err := hub.listTelegramNotificationPolicies(sibling.Id)
	require.NoError(t, err)
	assert.Len(t, siblingPolicies, 1)
	settings, err := hub.loadTelegramSettings()
	require.NoError(t, err)
	assert.True(t, settings.Enabled)

	err = hub.deleteTelegramDestination(deleteEvent())
	require.Error(t, err)
	assert.Equal(t, http.StatusNotFound, router.ToApiError(err).Status)
}
