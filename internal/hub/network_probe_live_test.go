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
	"github.com/pocketbase/pocketbase/tests"
	"github.com/pocketbase/pocketbase/tools/router"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNetworkProbeLiveManagerCreateRenewEnd(t *testing.T) {
	manager := newNetworkProbeLiveManager()
	now := time.Date(2026, 7, 2, 8, 0, 0, 0, time.UTC)

	session := manager.create("sys-1", "user-1", now)
	require.Equal(t, "sys-1", session.SystemID)
	require.Equal(t, "user-1", session.UserID)
	require.True(t, session.ExpiresAt.After(now))

	renewed, ok := manager.renew("sys-1", session.SessionID, "user-1", now.Add(5*time.Second))
	require.True(t, ok)
	assert.True(t, renewed.ExpiresAt.After(session.ExpiresAt))

	manager.end("sys-1", session.SessionID, "user-1")
	_, ok = manager.renew("sys-1", session.SessionID, "user-1", now.Add(6*time.Second))
	assert.False(t, ok)
}

func TestNetworkProbeLiveManagerCoalescesSystemsAndExpiresSessions(t *testing.T) {
	manager := newNetworkProbeLiveManager()
	now := time.Date(2026, 7, 2, 8, 0, 0, 0, time.UTC)

	first := manager.create("sys-1", "user-1", now)
	second := manager.create("sys-1", "user-2", now.Add(time.Second))
	manager.create("sys-2", "user-3", now.Add(2*time.Second))

	active := manager.activeSystems(now.Add(3 * time.Second))
	assert.Equal(t, []string{"sys-1", "sys-2"}, active)
	assert.Equal(t, 2, manager.activeSessionCount("sys-1", now.Add(3*time.Second)))

	_, ok := manager.renew("sys-1", first.SessionID, "user-1", now.Add(networkProbeLiveSessionTTL+time.Second))
	assert.False(t, ok)
	assert.Equal(t, 0, manager.activeSessionCount("sys-1", now.Add(networkProbeLiveSessionTTL+2*time.Second)))

	_ = second
}

func TestNetworkProbeLiveManagerPreventsOverlappingAssignments(t *testing.T) {
	manager := newNetworkProbeLiveManager()
	now := time.Date(2026, 7, 2, 8, 0, 0, 0, time.UTC)

	assert.True(t, manager.beginAssignment("assign-1", now))
	assert.False(t, manager.beginAssignment("assign-1", now.Add(time.Second)))
	manager.endAssignment("assign-1")
	assert.True(t, manager.beginAssignment("assign-1", now.Add(2*time.Second)))
}

func TestNetworkProbeLiveCadenceIsOneSecond(t *testing.T) {
	assert.Equal(t, 1, networkProbeLiveCadenceSeconds)
}

func TestCreateAndRenewNetworkProbeLiveSessionHandlers(t *testing.T) {
	hub, err := tests.NewTestApp(t.TempDir())
	require.NoError(t, err)
	t.Cleanup(hub.Cleanup)

	app := NewHub(hub)
	user := createNetworkProbeTestRecord(t, app, "users", map[string]any{
		"email":    "live-user@example.com",
		"password": "password123",
	})
	system := createNetworkProbeTestRecord(t, app, "systems", map[string]any{
		"name":   "node",
		"host":   "127.0.0.1",
		"status": "up",
		"users":  []string{user.Id},
	})

	body, err := json.Marshal(NetworkProbeLiveSessionInput{Range: networkProbeLiveRange})
	require.NoError(t, err)
	req := httptest.NewRequest(http.MethodPost, "/api/beszel/systems/"+system.Id+"/network-probe-live-sessions", bytes.NewReader(body))
	req.SetPathValue("systemId", system.Id)
	req.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()

	require.NoError(t, app.createNetworkProbeLiveSession(&core.RequestEvent{
		App:  app,
		Auth: user,
		Event: router.Event{
			Request:  req,
			Response: recorder,
		},
	}))
	require.Equal(t, http.StatusCreated, recorder.Code)
	var created NetworkProbeLiveSessionResponse
	require.NoError(t, json.Unmarshal(recorder.Body.Bytes(), &created))
	assert.Equal(t, system.Id, created.SystemID)
	assert.Equal(t, networkProbeLiveCadenceSeconds, created.CadenceSeconds)

	renewReq := httptest.NewRequest(http.MethodPatch, "/api/beszel/systems/"+system.Id+"/network-probe-live-sessions/"+created.SessionID, bytes.NewReader(body))
	renewReq.SetPathValue("systemId", system.Id)
	renewReq.SetPathValue("sessionId", created.SessionID)
	renewReq.Header.Set("Content-Type", "application/json")
	renewRecorder := httptest.NewRecorder()

	require.NoError(t, app.renewNetworkProbeLiveSession(&core.RequestEvent{
		App:  app,
		Auth: user,
		Event: router.Event{
			Request:  renewReq,
			Response: renewRecorder,
		},
	}))
	require.Equal(t, http.StatusOK, renewRecorder.Code)
	var renewed NetworkProbeLiveSessionResponse
	require.NoError(t, json.Unmarshal(renewRecorder.Body.Bytes(), &renewed))
	assert.Equal(t, created.SessionID, renewed.SessionID)
}

func TestNetworkProbeLiveSessionHandlersRejectInvalidRangeAndSupportDelete(t *testing.T) {
	hub, err := tests.NewTestApp(t.TempDir())
	require.NoError(t, err)
	t.Cleanup(hub.Cleanup)

	app := NewHub(hub)
	user := createNetworkProbeTestRecord(t, app, "users", map[string]any{
		"email":    "live-user@example.com",
		"password": "password123",
	})
	system := createNetworkProbeTestRecord(t, app, "systems", map[string]any{
		"name":   "node",
		"host":   "127.0.0.1",
		"status": "up",
		"users":  []string{user.Id},
	})

	invalidBody, err := json.Marshal(NetworkProbeLiveSessionInput{Range: "30m"})
	require.NoError(t, err)
	invalidReq := httptest.NewRequest(http.MethodPost, "/api/beszel/systems/"+system.Id+"/network-probe-live-sessions", bytes.NewReader(invalidBody))
	invalidReq.SetPathValue("systemId", system.Id)
	invalidReq.Header.Set("Content-Type", "application/json")
	invalidRecorder := httptest.NewRecorder()
	err = app.createNetworkProbeLiveSession(&core.RequestEvent{
		App:  app,
		Auth: user,
		Event: router.Event{
			Request:  invalidReq,
			Response: invalidRecorder,
		},
	})
	require.Error(t, err)
	apiErr := router.ToApiError(err)
	assert.Equal(t, http.StatusBadRequest, apiErr.Status)

	validBody, err := json.Marshal(NetworkProbeLiveSessionInput{Range: networkProbeLiveRange})
	require.NoError(t, err)
	createReq := httptest.NewRequest(http.MethodPost, "/api/beszel/systems/"+system.Id+"/network-probe-live-sessions", bytes.NewReader(validBody))
	createReq.SetPathValue("systemId", system.Id)
	createReq.Header.Set("Content-Type", "application/json")
	createRecorder := httptest.NewRecorder()
	require.NoError(t, app.createNetworkProbeLiveSession(&core.RequestEvent{
		App:  app,
		Auth: user,
		Event: router.Event{
			Request:  createReq,
			Response: createRecorder,
		},
	}))
	var created NetworkProbeLiveSessionResponse
	require.NoError(t, json.Unmarshal(createRecorder.Body.Bytes(), &created))

	deleteReq := httptest.NewRequest(http.MethodDelete, "/api/beszel/systems/"+system.Id+"/network-probe-live-sessions/"+created.SessionID, nil)
	deleteReq.SetPathValue("systemId", system.Id)
	deleteReq.SetPathValue("sessionId", created.SessionID)
	deleteRecorder := httptest.NewRecorder()
	require.NoError(t, app.endNetworkProbeLiveSession(&core.RequestEvent{
		App:  app,
		Auth: user,
		Event: router.Event{
			Request:  deleteReq,
			Response: deleteRecorder,
		},
	}))
	assert.Equal(t, http.StatusNoContent, deleteRecorder.Code)
	assert.Equal(t, 0, app.liveProbeManager().activeSessionCount(system.Id, time.Now().UTC()))
}
