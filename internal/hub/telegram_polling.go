package hub

import (
	"context"
	"time"
)

const telegramPollingTimeoutSeconds = 25

func (h *Hub) startTelegramPolling(ctx context.Context) {
	defer func() {
		if r := recover(); r != nil {
			h.Logger().Warn("Telegram polling stopped after panic", "err", r)
		}
	}()
	for {
		if err := h.pollTelegramOnce(ctx); err != nil && ctx.Err() == nil {
			h.Logger().Warn("Telegram polling error", "err", err)
			_ = h.setTelegramSettingsRuntimeState("", -1, err.Error())
			select {
			case <-ctx.Done():
				return
			case <-time.After(10 * time.Second):
			}
			continue
		}
		select {
		case <-ctx.Done():
			return
		default:
		}
	}
}

func (h *Hub) pollTelegramOnce(ctx context.Context) error {
	settings, err := h.loadTelegramSettings()
	if err != nil {
		return err
	}
	if !settings.Enabled || !settings.PollingEnabled {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(15 * time.Second):
			return nil
		}
	}
	token, err := h.decryptTelegramToken(settings)
	if err != nil || token == "" {
		return err
	}
	updates, err := h.telegramTransport.GetUpdates(ctx, token, settings.LastPollOffset+1, telegramPollingTimeoutSeconds)
	if err != nil {
		return err
	}
	nextOffset := settings.LastPollOffset
	for _, update := range updates {
		if update.UpdateID > nextOffset {
			nextOffset = update.UpdateID
		}
		command, ok := parseTelegramCommand(update)
		if !ok {
			continue
		}
		if err := h.handleTelegramCommand(ctx, token, command); err != nil {
			h.Logger().Warn("Telegram menu command failed", "err", err)
		}
	}
	if len(updates) > 0 {
		return h.setTelegramSettingsRuntimeState("", nextOffset, "")
	}
	return nil
}
