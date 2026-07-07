package hub

import (
	"context"
	"fmt"
	"slices"
	"strings"
	"time"

	"github.com/henrygd/beszel/internal/alerts"
)

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
	var firstErr error
	for _, destination := range destinations {
		if !telegramDestinationMatchesAlert(destination, data) {
			continue
		}
		message := telegramMessageForDestination(destination, data)
		err := h.telegramTransport.SendMessage(context.Background(), token, destination.ChatID, message, nil)
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

func telegramDestinationMatchesAlert(destination telegramDestinationRecord, data alerts.AlertMessageData) bool {
	if !destination.Enabled {
		return false
	}
	if destination.UserID != "" && destination.UserID != data.UserID {
		return false
	}
	if destination.MuteUntil != nil && destination.MuteUntil.After(time.Now().UTC()) {
		return false
	}
	if len(destination.NodeScope) > 0 && !slices.Contains(destination.NodeScope, data.SystemID) {
		return false
	}
	if destination.Role == TelegramRoleReadOnly && len(destination.AlertLevelScope) > 0 {
		alertClass := telegramAlertClass(data)
		if !slices.Contains(destination.AlertLevelScope, alertClass) {
			return false
		}
	}
	return true
}

func telegramMessageForDestination(destination telegramDestinationRecord, data alerts.AlertMessageData) string {
	if destination.Role == TelegramRoleReadOnly {
		return telegramReadOnlyAlertMessage(data)
	}
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
	return message
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
	return strings.Join(lines, "\n")
}

func telegramAlertClass(data alerts.AlertMessageData) string {
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
