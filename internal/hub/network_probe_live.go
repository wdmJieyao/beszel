package hub

import (
	"context"
	"net/http"
	"slices"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/pocketbase/pocketbase/core"
)

const (
	networkProbeLiveRange          = "1m"
	networkProbeLiveCadenceSeconds = 1
	networkProbeLiveSessionTTL     = 15 * time.Second
)

type NetworkProbeLiveSessionInput struct {
	Range string `json:"range"`
}

type NetworkProbeLiveSessionResponse struct {
	SessionID      string `json:"sessionId"`
	SystemID       string `json:"systemId"`
	Range          string `json:"range"`
	CadenceSeconds int    `json:"cadenceSeconds"`
	ExpiresAt      string `json:"expiresAt"`
}

type networkProbeLiveSession struct {
	SessionID string
	SystemID  string
	UserID    string
	StartedAt time.Time
	LastSeen  time.Time
	ExpiresAt time.Time
}

type networkProbeLiveManager struct {
	mu       sync.Mutex
	sessions map[string]networkProbeLiveSession
	inFlight map[string]time.Time
}

func newNetworkProbeLiveManager() *networkProbeLiveManager {
	return &networkProbeLiveManager{
		sessions: make(map[string]networkProbeLiveSession),
		inFlight: make(map[string]time.Time),
	}
}

func (m *networkProbeLiveManager) create(systemID string, userID string, now time.Time) networkProbeLiveSession {
	m.mu.Lock()
	defer m.mu.Unlock()
	session := networkProbeLiveSession{
		SessionID: uuid.NewString(),
		SystemID:  systemID,
		UserID:    userID,
		StartedAt: now,
		LastSeen:  now,
		ExpiresAt: now.Add(networkProbeLiveSessionTTL),
	}
	m.sessions[session.SessionID] = session
	return session
}

func (m *networkProbeLiveManager) renew(systemID string, sessionID string, userID string, now time.Time) (networkProbeLiveSession, bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	session, ok := m.sessions[sessionID]
	if !ok || session.SystemID != systemID || session.UserID != userID || !session.ExpiresAt.After(now) {
		delete(m.sessions, sessionID)
		return networkProbeLiveSession{}, false
	}
	session.LastSeen = now
	session.ExpiresAt = now.Add(networkProbeLiveSessionTTL)
	m.sessions[sessionID] = session
	return session, true
}

func (m *networkProbeLiveManager) end(systemID string, sessionID string, userID string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	session, ok := m.sessions[sessionID]
	if ok && session.SystemID == systemID && session.UserID == userID {
		delete(m.sessions, sessionID)
	}
}

func (m *networkProbeLiveManager) activeSystems(now time.Time) []string {
	m.mu.Lock()
	defer m.mu.Unlock()
	active := make(map[string]struct{})
	for id, session := range m.sessions {
		if !session.ExpiresAt.After(now) {
			delete(m.sessions, id)
			continue
		}
		active[session.SystemID] = struct{}{}
	}
	systems := make([]string, 0, len(active))
	for systemID := range active {
		systems = append(systems, systemID)
	}
	slices.Sort(systems)
	return systems
}

func (m *networkProbeLiveManager) activeSessionCount(systemID string, now time.Time) int {
	m.mu.Lock()
	defer m.mu.Unlock()
	count := 0
	for id, session := range m.sessions {
		if !session.ExpiresAt.After(now) {
			delete(m.sessions, id)
			continue
		}
		if session.SystemID == systemID {
			count++
		}
	}
	return count
}

func (m *networkProbeLiveManager) beginAssignment(assignmentID string, now time.Time) bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	if started, ok := m.inFlight[assignmentID]; ok && now.Sub(started) < networkProbeLiveSessionTTL {
		return false
	}
	m.inFlight[assignmentID] = now
	return true
}

func (m *networkProbeLiveManager) endAssignment(assignmentID string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.inFlight, assignmentID)
}

func (h *Hub) liveProbeManager() *networkProbeLiveManager {
	if h.networkProbeLive == nil {
		h.networkProbeLive = newNetworkProbeLiveManager()
	}
	return h.networkProbeLive
}

func (h *Hub) createNetworkProbeLiveSession(e *core.RequestEvent) error {
	systemID := e.Request.PathValue("systemId")
	if err := h.ensureSystemVisibleToAuth(e.Auth, systemID); err != nil {
		return e.NotFoundError("System not found.", err)
	}
	var input NetworkProbeLiveSessionInput
	if err := e.BindBody(&input); err != nil {
		return e.BadRequestError("Invalid live latency session request.", err)
	}
	if input.Range != networkProbeLiveRange {
		return e.BadRequestError("Invalid live latency session request.", map[string]string{
			"range": "Only 1m is supported for live latency sessions.",
		})
	}
	session := h.liveProbeManager().create(systemID, e.Auth.Id, time.Now().UTC())
	return e.JSON(http.StatusCreated, networkProbeLiveSessionResponse(session))
}

func (h *Hub) renewNetworkProbeLiveSession(e *core.RequestEvent) error {
	systemID := e.Request.PathValue("systemId")
	sessionID := e.Request.PathValue("sessionId")
	if err := h.ensureSystemVisibleToAuth(e.Auth, systemID); err != nil {
		return e.NotFoundError("System not found.", err)
	}
	var input NetworkProbeLiveSessionInput
	if err := e.BindBody(&input); err != nil {
		return e.BadRequestError("Invalid live latency session request.", err)
	}
	if input.Range != networkProbeLiveRange {
		return e.BadRequestError("Invalid live latency session request.", map[string]string{
			"range": "Only 1m is supported for live latency sessions.",
		})
	}
	session, ok := h.liveProbeManager().renew(systemID, sessionID, e.Auth.Id, time.Now().UTC())
	if !ok {
		return e.NotFoundError("Live latency session not found.", nil)
	}
	return e.JSON(http.StatusOK, networkProbeLiveSessionResponse(session))
}

func (h *Hub) endNetworkProbeLiveSession(e *core.RequestEvent) error {
	systemID := e.Request.PathValue("systemId")
	sessionID := e.Request.PathValue("sessionId")
	if err := h.ensureSystemVisibleToAuth(e.Auth, systemID); err != nil {
		return e.NotFoundError("System not found.", err)
	}
	h.liveProbeManager().end(systemID, sessionID, e.Auth.Id)
	return e.NoContent(http.StatusNoContent)
}

func (h *Hub) startNetworkProbeLiveScheduler() {
	defer func() {
		if r := recover(); r != nil {
			h.Logger().Warn("Live network probe scheduler stopped after panic", "err", r)
		}
	}()
	ticker := time.NewTicker(time.Duration(networkProbeLiveCadenceSeconds) * time.Second)
	defer ticker.Stop()
	for range ticker.C {
		safeRunActiveNetworkProbeLiveSessions(h, context.Background())
	}
}

func (h *Hub) runActiveNetworkProbeLiveSessions(ctx context.Context) {
	now := time.Now().UTC()
	for _, systemID := range h.liveProbeManager().activeSystems(now) {
		go func(systemID string) {
			if err := h.RunLiveNetworkProbes(ctx, systemID); err != nil {
				h.Logger().Warn("Error running live network probes", "system", systemID, "err", err)
			}
		}(systemID)
	}
}

func safeRunActiveNetworkProbeLiveSessions(h *Hub, ctx context.Context) {
	defer func() {
		if r := recover(); r != nil {
			h.Logger().Warn("Error running live network probes", "err", r)
		}
	}()
	h.runActiveNetworkProbeLiveSessions(ctx)
}

func networkProbeLiveSessionResponse(session networkProbeLiveSession) NetworkProbeLiveSessionResponse {
	return NetworkProbeLiveSessionResponse{
		SessionID:      session.SessionID,
		SystemID:       session.SystemID,
		Range:          networkProbeLiveRange,
		CadenceSeconds: networkProbeLiveCadenceSeconds,
		ExpiresAt:      session.ExpiresAt.UTC().Format(time.RFC3339),
	}
}
