package hub

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/pocketbase/pocketbase/core"
)

func (h *Hub) listTelegramDestinationsHandler(e *core.RequestEvent) error {
	destinations, err := h.listTelegramDestinations()
	if err != nil {
		return err
	}
	policiesByDestination, err := h.listAllTelegramNotificationPolicies()
	if err != nil {
		return err
	}
	response := TelegramDestinationsResponse{Destinations: make([]TelegramDestinationResponse, 0, len(destinations))}
	for _, destination := range destinations {
		item := telegramDestinationResponse(destination)
		policies := policiesByDestination[destination.ID]
		item.PolicyCount = len(policies)
		for _, policy := range policies {
			if policy.Name == TelegramDefaultPolicyName {
				item.NodeScope = policy.NodeScope
				item.AlertLevelScope = policy.AlertLevelScope
				break
			}
		}
		response.Destinations = append(response.Destinations, item)
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
	if existing, err := h.findTelegramDestinationByChatID(strings.TrimSpace(input.ChatID)); err == nil && existing != nil {
		return e.JSON(http.StatusConflict, map[string]any{
			"status": http.StatusConflict, "message": "Telegram channel already exists",
			"data": map[string]string{"existingDestinationId": existing.Id},
		})
	}
	var record *core.Record
	err := h.RunInTransaction(func(txApp core.App) error {
		txHub := h.configBackupTransactionHub(txApp)
		var err error
		record, err = txHub.upsertTelegramDestination(nil, input)
		if err != nil {
			return err
		}
		if input.NodeScope != nil || input.AlertLevelScope != nil {
			return txHub.ensureTelegramDestinationDefaultPolicy(record)
		}
		return nil
	})
	if err != nil {
		if existing, findErr := h.findTelegramDestinationByChatID(strings.TrimSpace(input.ChatID)); findErr == nil && existing != nil {
			return e.JSON(http.StatusConflict, map[string]any{
				"status": http.StatusConflict, "message": "Telegram channel already exists",
				"data": map[string]string{"existingDestinationId": existing.Id},
			})
		}
		return err
	}
	return e.JSON(http.StatusCreated, h.telegramDestinationResponseWithPolicies(telegramDestinationFromRecord(record)))
}

func (h *Hub) updateTelegramDestination(e *core.RequestEvent) error {
	patch := TelegramDestinationPatchInput{}
	if err := json.NewDecoder(e.Request.Body).Decode(&patch); err != nil {
		return e.BadRequestError("invalid request body", err)
	}
	record, err := h.findTelegramDestinationByID(e.Request.PathValue("destinationId"))
	if err != nil {
		return e.NotFoundError("destination not found", err)
	}
	input, err := mergeTelegramDestinationPatch(telegramDestinationFromRecord(record), patch)
	if err != nil {
		return e.BadRequestError(err.Error(), nil)
	}
	if err := validateTelegramDestinationInput(input); err != nil {
		return e.BadRequestError(err.Error(), nil)
	}
	if err := h.validateTelegramDestinationReferences(input); err != nil {
		return e.BadRequestError(err.Error(), nil)
	}
	if existing, err := h.findTelegramDestinationByChatID(strings.TrimSpace(input.ChatID)); err == nil && existing != nil && existing.Id != record.Id {
		return e.JSON(http.StatusConflict, map[string]any{
			"status": http.StatusConflict, "message": "Telegram channel already exists",
			"data": map[string]string{"existingDestinationId": existing.Id},
		})
	}
	err = h.RunInTransaction(func(txApp core.App) error {
		txHub := h.configBackupTransactionHub(txApp)
		txRecord, err := txApp.FindRecordById(CollectionTelegramDestinations, record.Id)
		if err != nil {
			return err
		}
		setTelegramDestinationRecord(txRecord, input)
		if err := txHub.Save(txRecord); err != nil {
			return err
		}
		record = txRecord
		if patch.NodeScope.Set || patch.AlertLevelScope.Set {
			return txHub.upsertTelegramDefaultPolicyFromLegacy(record, input)
		}
		return nil
	})
	if err != nil {
		return err
	}
	return e.JSON(http.StatusOK, h.telegramDestinationResponseWithPolicies(telegramDestinationFromRecord(record)))
}

func mergeTelegramDestinationPatch(current telegramDestinationRecord, patch TelegramDestinationPatchInput) (TelegramDestinationInput, error) {
	input := TelegramDestinationInput{
		UserID: current.UserID, Name: current.Name, ChatID: current.ChatID, ChatType: current.ChatType,
		Role: current.Role, Enabled: telegramBoolPointer(current.Enabled), NodeScope: append([]string(nil), current.NodeScope...),
		AlertLevelScope: append([]string(nil), current.AlertLevelScope...), MuteUntil: current.MuteUntil,
	}
	if patch.UserID.Set {
		if patch.UserID.Null {
			input.UserID = ""
		} else {
			input.UserID = patch.UserID.Value
		}
	}
	if err := assignTelegramPatchField(patch.Name, "name", &input.Name); err != nil {
		return TelegramDestinationInput{}, err
	}
	if err := assignTelegramPatchField(patch.ChatID, "chatId", &input.ChatID); err != nil {
		return TelegramDestinationInput{}, err
	}
	if err := assignTelegramPatchField(patch.ChatType, "chatType", &input.ChatType); err != nil {
		return TelegramDestinationInput{}, err
	}
	if err := assignTelegramPatchField(patch.Role, "role", &input.Role); err != nil {
		return TelegramDestinationInput{}, err
	}
	if patch.Enabled.Set {
		if patch.Enabled.Null {
			return TelegramDestinationInput{}, fmt.Errorf("enabled cannot be null")
		}
		input.Enabled = telegramBoolPointer(patch.Enabled.Value)
	}
	if err := assignTelegramPatchField(patch.NodeScope, "nodeScope", &input.NodeScope); err != nil {
		return TelegramDestinationInput{}, err
	}
	if err := assignTelegramPatchField(patch.AlertLevelScope, "alertLevelScope", &input.AlertLevelScope); err != nil {
		return TelegramDestinationInput{}, err
	}
	if patch.MuteUntil.Set {
		input.MuteUntil = patch.MuteUntil.Value
	}
	return input, nil
}

func (h *Hub) upsertTelegramDefaultPolicyFromLegacy(destination *core.Record, input TelegramDestinationInput) error {
	policies, err := h.listTelegramNotificationPolicies(destination.Id)
	if err != nil {
		return err
	}
	var record *core.Record
	for _, policy := range policies {
		if policy.Name == TelegramDefaultPolicyName {
			record, err = h.findTelegramNotificationPolicy(destination.Id, policy.ID)
			if err != nil {
				return err
			}
			break
		}
	}
	mode := TelegramNodeScopeAll
	if len(normalizeTelegramStringSlice(input.NodeScope)) != 0 {
		mode = TelegramNodeScopeSelected
	}
	_, err = h.upsertTelegramNotificationPolicy(record, destination.Id, TelegramNotificationPolicyInput{
		Name: TelegramDefaultPolicyName, Enabled: input.Enabled, NodeScopeMode: mode,
		NodeScope: input.NodeScope, AlertLevelScope: input.AlertLevelScope,
	})
	return err
}

func (h *Hub) deleteTelegramDestination(e *core.RequestEvent) error {
	destinationID := e.Request.PathValue("destinationId")
	_, err := h.findTelegramDestinationByID(destinationID)
	if err != nil {
		return e.NotFoundError("destination not found", err)
	}
	if err := h.RunInTransaction(func(txApp core.App) error {
		record, err := txApp.FindRecordById(CollectionTelegramDestinations, destinationID)
		if err != nil {
			return err
		}
		return txApp.Delete(record)
	}); err != nil {
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

func (h *Hub) telegramDestinationResponseWithPolicies(destination telegramDestinationRecord) TelegramDestinationResponse {
	response := telegramDestinationResponse(destination)
	policies, err := h.listTelegramNotificationPolicies(destination.ID)
	if err != nil {
		return response
	}
	response.PolicyCount = len(policies)
	for _, policy := range policies {
		if policy.Name == TelegramDefaultPolicyName {
			response.NodeScope = policy.NodeScope
			response.AlertLevelScope = policy.AlertLevelScope
			break
		}
	}
	return response
}
