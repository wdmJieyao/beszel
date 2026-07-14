package hub

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/pocketbase/pocketbase/core"
)

func (h *Hub) listTelegramNotificationPoliciesHandler(e *core.RequestEvent) error {
	destinationID := e.Request.PathValue("destinationId")
	if _, err := h.findTelegramDestinationByID(destinationID); err != nil {
		return e.NotFoundError("destination not found", err)
	}
	policies, err := h.listTelegramNotificationPolicies(destinationID)
	if err != nil {
		return err
	}
	response := TelegramPolicyListResponse{Policies: make([]TelegramNotificationPolicyResponse, 0, len(policies))}
	for _, policy := range policies {
		response.Policies = append(response.Policies, telegramNotificationPolicyResponse(policy))
	}
	return e.JSON(http.StatusOK, response)
}

func (h *Hub) createTelegramNotificationPolicy(e *core.RequestEvent) error {
	destinationID := e.Request.PathValue("destinationId")
	if _, err := h.findTelegramDestinationByID(destinationID); err != nil {
		return e.NotFoundError("destination not found", err)
	}
	input := TelegramNotificationPolicyInput{}
	if err := e.BindBody(&input); err != nil {
		return e.BadRequestError("invalid request body", err)
	}
	record, err := h.upsertTelegramNotificationPolicy(nil, destinationID, input)
	if err != nil {
		if isTelegramPolicyNameConflictError(err) {
			return e.JSON(http.StatusConflict, map[string]any{"status": http.StatusConflict, "message": err.Error()})
		}
		return e.BadRequestError(err.Error(), nil)
	}
	return e.JSON(http.StatusCreated, telegramNotificationPolicyResponse(telegramNotificationPolicyFromRecord(record)))
}

func (h *Hub) updateTelegramNotificationPolicy(e *core.RequestEvent) error {
	destinationID := e.Request.PathValue("destinationId")
	record, err := h.findTelegramNotificationPolicy(destinationID, e.Request.PathValue("policyId"))
	if err != nil {
		return e.NotFoundError("policy not found", err)
	}
	patch := TelegramNotificationPolicyPatchInput{}
	if err := json.NewDecoder(e.Request.Body).Decode(&patch); err != nil {
		return e.BadRequestError("invalid request body", err)
	}
	input, err := mergeTelegramNotificationPolicyPatch(telegramNotificationPolicyFromRecord(record), patch)
	if err != nil {
		return e.BadRequestError(err.Error(), nil)
	}
	record, err = h.upsertTelegramNotificationPolicy(record, destinationID, input)
	if err != nil {
		if isTelegramPolicyNameConflictError(err) {
			return e.JSON(http.StatusConflict, map[string]any{"status": http.StatusConflict, "message": err.Error()})
		}
		return e.BadRequestError(err.Error(), nil)
	}
	return e.JSON(http.StatusOK, telegramNotificationPolicyResponse(telegramNotificationPolicyFromRecord(record)))
}

func mergeTelegramNotificationPolicyPatch(current telegramNotificationPolicyRecord, patch TelegramNotificationPolicyPatchInput) (TelegramNotificationPolicyInput, error) {
	input := TelegramNotificationPolicyInput{
		Name: current.Name, Enabled: telegramBoolPointer(current.Enabled), NodeScopeMode: current.NodeScopeMode,
		NodeScope: append([]string(nil), current.NodeScope...), AlertLevelScope: append([]string(nil), current.AlertLevelScope...),
	}
	if err := assignTelegramPatchField(patch.Name, "name", &input.Name); err != nil {
		return TelegramNotificationPolicyInput{}, err
	}
	if patch.Enabled.Set {
		if patch.Enabled.Null {
			return TelegramNotificationPolicyInput{}, fmt.Errorf("enabled cannot be null")
		}
		input.Enabled = telegramBoolPointer(patch.Enabled.Value)
	}
	if err := assignTelegramPatchField(patch.NodeScopeMode, "nodeScopeMode", &input.NodeScopeMode); err != nil {
		return TelegramNotificationPolicyInput{}, err
	}
	if err := assignTelegramPatchField(patch.NodeScope, "nodeScope", &input.NodeScope); err != nil {
		return TelegramNotificationPolicyInput{}, err
	}
	if err := assignTelegramPatchField(patch.AlertLevelScope, "alertLevelScope", &input.AlertLevelScope); err != nil {
		return TelegramNotificationPolicyInput{}, err
	}
	return input, nil
}

func telegramBoolPointer(value bool) *bool {
	return &value
}

func (h *Hub) deleteTelegramNotificationPolicy(e *core.RequestEvent) error {
	record, err := h.findTelegramNotificationPolicy(e.Request.PathValue("destinationId"), e.Request.PathValue("policyId"))
	if err != nil {
		return e.NotFoundError("policy not found", err)
	}
	if err := h.Delete(record); err != nil {
		return err
	}
	return e.NoContent(http.StatusNoContent)
}

func telegramNotificationPolicyResponse(policy telegramNotificationPolicyRecord) TelegramNotificationPolicyResponse {
	return TelegramNotificationPolicyResponse(policy)
}
