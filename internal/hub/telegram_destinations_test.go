//go:build testing

package hub

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

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
