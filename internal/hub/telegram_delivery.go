package hub

import (
	"context"
	"fmt"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/henrygd/beszel/internal/alerts"
)

type telegramRetryPolicy struct {
	MaxAttempts    int
	InitialBackoff time.Duration
}

type telegramDeliveryState struct {
	serial    sync.Mutex
	pendingMu sync.Mutex
	pending   int
	lastSent  time.Time
	interval  time.Duration
}

const telegramDeliveryQueueLimit = 32
const telegramDeliveryInterval = 35 * time.Millisecond

func telegramDeliveryStateFor(h *Hub) *telegramDeliveryState {
	if h.telegramDeliveryState == nil {
		h.telegramDeliveryState = &telegramDeliveryState{}
	}
	return h.telegramDeliveryState
}

func (state *telegramDeliveryState) enter() error {
	state.pendingMu.Lock()
	defer state.pendingMu.Unlock()
	if state.pending >= telegramDeliveryQueueLimit {
		return fmt.Errorf("telegram delivery queue is full")
	}
	state.pending++
	return nil
}

func (state *telegramDeliveryState) leave() {
	state.pendingMu.Lock()
	state.pending--
	state.pendingMu.Unlock()
}

func (state *telegramDeliveryState) waitForRateLimit(ctx context.Context) error {
	interval := state.interval
	if interval <= 0 {
		interval = telegramDeliveryInterval
	}
	wait := time.Until(state.lastSent.Add(interval))
	if wait > 0 {
		timer := time.NewTimer(wait)
		select {
		case <-ctx.Done():
			timer.Stop()
			return ctx.Err()
		case <-timer.C:
		}
	}
	state.lastSent = time.Now()
	return nil
}

func sendTelegramMessageWithRetry(ctx context.Context, transport TelegramTransport, token, chatID, message string, options *TelegramSendOptions, policy telegramRetryPolicy) error {
	if policy.MaxAttempts < 1 {
		policy.MaxAttempts = 1
	}
	backoff := policy.InitialBackoff
	var err error
	for attempt := 0; attempt < policy.MaxAttempts; attempt++ {
		err = transport.SendMessage(ctx, token, chatID, truncateTelegramMessage(message), options)
		if err == nil || !isRetryableTelegramError(err) {
			return err
		}
		if attempt+1 >= policy.MaxAttempts || backoff <= 0 {
			continue
		}
		timer := time.NewTimer(backoff)
		select {
		case <-ctx.Done():
			timer.Stop()
			return ctx.Err()
		case <-timer.C:
		}
		backoff *= 2
	}
	return err
}

func isRetryableTelegramError(err error) bool {
	if err == nil {
		return false
	}
	message := strings.ToLower(err.Error())
	return strings.Contains(message, "429") || strings.Contains(message, "too many requests") ||
		strings.Contains(message, "timeout") || strings.Contains(message, "temporary") ||
		strings.Contains(message, "502") || strings.Contains(message, "503") || strings.Contains(message, "504")
}

func (h *Hub) SendTelegramAlert(data alerts.AlertMessageData) error {
	settings, err := h.loadTelegramSettings()
	if err != nil || !settings.Enabled {
		return err
	}
	token, err := h.decryptTelegramToken(settings)
	if err != nil || token == "" {
		return err
	}
	destinations, err := h.listTelegramDestinations()
	if err != nil {
		return err
	}
	policiesByDestination, err := h.listAllTelegramNotificationPolicies()
	if err != nil {
		return err
	}
	state := telegramDeliveryStateFor(h)
	if err := state.enter(); err != nil {
		return err
	}
	defer state.leave()
	state.serial.Lock()
	defer state.serial.Unlock()
	var firstErr error
	for _, destination := range destinations {
		if !telegramDestinationCanReceiveAlert(destination, data) || !telegramPoliciesMatchAlert(policiesByDestination[destination.ID], data) {
			continue
		}
		message := telegramMessageForDestination(destination, data)
		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		if err := state.waitForRateLimit(ctx); err != nil {
			cancel()
			return err
		}
		err := sendTelegramMessageWithRetry(ctx, h.telegramTransport, token, destination.ChatID, message, nil, telegramRetryPolicy{MaxAttempts: 3, InitialBackoff: 200 * time.Millisecond})
		cancel()
		if err != nil {
			_ = h.setTelegramDestinationDeliveryState(destination.ID, time.Time{}, err.Error())
			if firstErr == nil {
				firstErr = err
			}
			continue
		}
		_ = h.setTelegramDestinationDeliveryState(destination.ID, time.Now().UTC(), "")
	}
	return firstErr
}

func (h *Hub) sendTelegramTestMessage(destination telegramDestinationRecord) error {
	settings, err := h.loadTelegramSettings()
	if err != nil {
		return err
	}
	token, err := h.decryptTelegramToken(settings)
	if err != nil {
		return err
	}
	if token == "" {
		return fmt.Errorf("telegram bot token is not configured")
	}
	message := fmt.Sprintf("Beszel Telegram test\n目的地: %s\n时间: %s", destination.Name, time.Now().UTC().Format(time.RFC3339))
	return h.telegramTransport.SendMessage(context.Background(), token, destination.ChatID, message, nil)
}

func telegramDestinationCanReceiveAlert(destination telegramDestinationRecord, data alerts.AlertMessageData) bool {
	if !destination.Enabled {
		return false
	}
	if destination.UserID != "" && destination.UserID != data.UserID {
		return false
	}
	if destination.MuteUntil != nil && destination.MuteUntil.After(time.Now().UTC()) {
		return false
	}
	return true
}

func telegramPoliciesMatchAlert(policies []telegramNotificationPolicyRecord, data alerts.AlertMessageData) bool {
	alertClass := telegramAlertClass(data)
	for _, policy := range policies {
		if !policy.Enabled {
			continue
		}
		if policy.NodeScopeMode == TelegramNodeScopeSelected && !slices.Contains(policy.NodeScope, data.SystemID) {
			continue
		}
		if len(policy.AlertLevelScope) != 0 && !slices.Contains(policy.AlertLevelScope, alertClass) {
			continue
		}
		return true
	}
	return false
}

func telegramMessageForDestination(destination telegramDestinationRecord, data alerts.AlertMessageData) string {
	if destination.Role == TelegramRoleReadOnly {
		return truncateTelegramMessage(telegramReadOnlyAlertMessage(sanitizeTelegramReadOnlyAlert(data)))
	}
	lines := telegramAlertContextLines(data)
	message := strings.TrimSpace(data.Message)
	if data.Link != "" {
		if message != "" {
			message += "\n\n"
		}
		message += data.Link
	}
	if message == "" {
		message = data.Title
	}
	if len(lines) > 0 {
		message = strings.Join(lines, "\n") + "\n\n" + message
	}
	return truncateTelegramMessage(message)
}

func telegramAlertContextLines(data alerts.AlertMessageData) []string {
	lines := make([]string, 0, 5)
	if value := strings.TrimSpace(data.SystemName); value != "" {
		lines = append(lines, "节点："+value)
	}
	if value := strings.TrimSpace(data.AlertClass); value != "" {
		lines = append(lines, "类型："+value)
	}
	if value := strings.TrimSpace(data.Severity); value != "" {
		lines = append(lines, "严重级别："+value)
	}
	if value := strings.TrimSpace(data.State); value != "" {
		lines = append(lines, "状态："+value)
	}
	if !data.EventTime.IsZero() {
		lines = append(lines, "时间："+data.EventTime.UTC().Format(time.RFC3339))
	}
	return lines
}

func telegramReadOnlyAlertMessage(data alerts.AlertMessageData) string {
	lines := []string{
		"Beszel 告警摘要",
		strings.TrimSpace(data.Title),
	}
	message := strings.TrimSpace(data.Message)
	if message != "" && message != strings.TrimSpace(data.Title) {
		lines = append(lines, message)
	}
	lines = append(lines, telegramAlertContextLines(data)...)
	return truncateTelegramMessage(strings.Join(lines, "\n"))
}

func telegramAlertClass(data alerts.AlertMessageData) string {
	if class := strings.ToLower(strings.TrimSpace(data.AlertClass)); class != "" {
		return class
	}
	text := strings.ToLower(strings.TrimSpace(data.Title + " " + data.Message))
	switch {
	case strings.Contains(text, "connection to"), strings.Contains(text, " status "):
		return "status"
	case strings.Contains(text, "cpu"):
		return "cpu"
	case strings.Contains(text, "memory"):
		return "memory"
	case strings.Contains(text, "disk"):
		return "disk"
	case strings.Contains(text, "temperature"):
		return "temperature"
	case strings.Contains(text, "bandwidth"):
		return "bandwidth"
	case strings.Contains(text, "gpu"):
		return "gpu"
	case strings.Contains(text, "loadavg1"), strings.Contains(text, "load avg 1"):
		return "loadavg1"
	case strings.Contains(text, "loadavg5"), strings.Contains(text, "load avg 5"):
		return "loadavg5"
	case strings.Contains(text, "loadavg15"), strings.Contains(text, "load avg 15"):
		return "loadavg15"
	case strings.Contains(text, "battery"):
		return "battery"
	default:
		return "status"
	}
}
