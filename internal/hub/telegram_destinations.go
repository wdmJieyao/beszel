package hub

import (
	"net/http"
	"time"

	"github.com/pocketbase/pocketbase/core"
)

func (h *Hub) listTelegramDestinationsHandler(e *core.RequestEvent) error {
	destinations, err := h.listTelegramDestinations()
	if err != nil {
		return err
	}
	response := TelegramDestinationsResponse{Destinations: make([]TelegramDestinationResponse, 0, len(destinations))}
	for _, destination := range destinations {
		response.Destinations = append(response.Destinations, telegramDestinationResponse(destination))
	}
	return e.JSON(http.StatusOK, response)
}

func (h *Hub) createTelegramDestination(e *core.RequestEvent) error {
	input := TelegramDestinationInput{}
	if err := e.BindBody(&input); err != nil {
		return e.BadRequestError("invalid request body", err)
	}
	if err := validateTelegramDestinationInput(input); err != nil {
		return e.BadRequestError(err.Error(), nil)
	}
	if err := h.validateTelegramDestinationReferences(input); err != nil {
		return e.BadRequestError(err.Error(), nil)
	}
	if existing, err := h.findTelegramDestinationByChatID(input.ChatID); err == nil && existing != nil {
		return e.BadRequestError("chatId already exists", nil)
	}
	record, err := h.upsertTelegramDestination(nil, input)
	if err != nil {
		return err
	}
	return e.JSON(http.StatusCreated, telegramDestinationResponse(telegramDestinationFromRecord(record)))
}

func (h *Hub) updateTelegramDestination(e *core.RequestEvent) error {
	input := TelegramDestinationInput{}
	if err := e.BindBody(&input); err != nil {
		return e.BadRequestError("invalid request body", err)
	}
	if err := validateTelegramDestinationInput(input); err != nil {
		return e.BadRequestError(err.Error(), nil)
	}
	if err := h.validateTelegramDestinationReferences(input); err != nil {
		return e.BadRequestError(err.Error(), nil)
	}
	record, err := h.findTelegramDestinationByID(e.Request.PathValue("destinationId"))
	if err != nil {
		return e.NotFoundError("destination not found", err)
	}
	if existing, err := h.findTelegramDestinationByChatID(input.ChatID); err == nil && existing != nil && existing.Id != record.Id {
		return e.BadRequestError("chatId already exists", nil)
	}
	record, err = h.upsertTelegramDestination(record, input)
	if err != nil {
		return err
	}
	return e.JSON(http.StatusOK, telegramDestinationResponse(telegramDestinationFromRecord(record)))
}

func (h *Hub) deleteTelegramDestination(e *core.RequestEvent) error {
	record, err := h.findTelegramDestinationByID(e.Request.PathValue("destinationId"))
	if err != nil {
		return e.NotFoundError("destination not found", err)
	}
	if err := h.Delete(record); err != nil {
		return err
	}
	return e.NoContent(http.StatusNoContent)
}

func (h *Hub) testTelegramDestination(e *core.RequestEvent) error {
	record, err := h.findTelegramDestinationByID(e.Request.PathValue("destinationId"))
	if err != nil {
		return e.NotFoundError("destination not found", err)
	}
	destination := telegramDestinationFromRecord(record)
	if err := h.sendTelegramTestMessage(destination); err != nil {
		_ = h.setTelegramDestinationLastTest(destination.ID, time.Time{}, err.Error())
		return e.JSON(http.StatusOK, map[string]any{"ok": false, "error": err.Error()})
	}
	now := time.Now().UTC()
	_ = h.setTelegramDestinationLastTest(destination.ID, now, "")
	return e.JSON(http.StatusOK, map[string]any{"ok": true, "sentAt": now.Format(time.RFC3339)})
}

func telegramDestinationResponse(destination telegramDestinationRecord) TelegramDestinationResponse {
	response := TelegramDestinationResponse{
		ID:              destination.ID,
		UserID:          destination.UserID,
		Name:            destination.Name,
		ChatID:          destination.ChatID,
		ChatType:        destination.ChatType,
		Role:            destination.Role,
		Enabled:         destination.Enabled,
		NodeScope:       destination.NodeScope,
		AlertLevelScope: destination.AlertLevelScope,
		LastError:       destination.LastError,
	}
	if destination.MuteUntil != nil {
		response.MuteUntil = destination.MuteUntil.UTC().Format(time.RFC3339)
	}
	if destination.LastTestAt != nil {
		response.LastTestAt = destination.LastTestAt.UTC().Format(time.RFC3339)
	}
	if destination.LastDeliveryAt != nil {
		response.LastDeliveryAt = destination.LastDeliveryAt.UTC().Format(time.RFC3339)
	}
	return response
}
