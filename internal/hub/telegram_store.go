package hub

import (
	"crypto/sha256"
	"errors"
	"fmt"
	"strings"
	"time"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/tools/security"
	"github.com/pocketbase/pocketbase/tools/types"
)

func isTelegramPolicyNameConflictError(err error) bool {
	if errors.Is(err, errTelegramPolicyNameConflict) {
		return true
	}
	var validationErrors validation.Errors
	if !errors.As(err, &validationErrors) {
		return false
	}
	for _, field := range []string{"destination", "name"} {
		var validationError validation.Error
		if errors.As(validationErrors[field], &validationError) && validationError.Code() == "validation_not_unique" {
			return true
		}
	}
	return false
}

func telegramCipherKey(value string) string {
	sum := sha256.Sum256([]byte(value))
	return string(sum[:])
}

func (h *Hub) getTelegramSettingsRecord() (*core.Record, error) {
	records, err := h.FindRecordsByFilter(CollectionTelegramSettings, "id != ''", "", 1, 0)
	if err != nil {
		return nil, err
	}
	if len(records) == 0 {
		return nil, nil
	}
	return records[0], nil
}

func (h *Hub) loadTelegramSettings() (telegramSettingsRecord, error) {
	record, err := h.getTelegramSettingsRecord()
	if err != nil || record == nil {
		return telegramSettingsRecord{}, err
	}
	return telegramSettingsFromRecord(record), nil
}

func telegramSettingsFromRecord(record *core.Record) telegramSettingsRecord {
	if record == nil {
		return telegramSettingsRecord{}
	}
	return telegramSettingsRecord{
		Enabled:           record.GetBool("enabled"),
		PollingEnabled:    record.GetBool("polling_enabled"),
		BotTokenEncrypted: record.GetString("bot_token_encrypted"),
		BotUsername:       record.GetString("bot_username"),
		LastPollOffset:    int64(record.GetInt("last_poll_offset")),
		LastError:         record.GetString("last_error"),
		Updated:           record.GetString("updated"),
	}
}

func (h *Hub) saveTelegramSettings(input TelegramSettingsInput, current telegramSettingsRecord) (telegramSettingsRecord, error) {
	record, err := h.getTelegramSettingsRecord()
	if err != nil {
		return telegramSettingsRecord{}, err
	}
	if record == nil {
		collection, err := h.FindCachedCollectionByNameOrId(CollectionTelegramSettings)
		if err != nil {
			return telegramSettingsRecord{}, err
		}
		record = core.NewRecord(collection)
	}
	record.Set("enabled", input.Enabled)
	record.Set("polling_enabled", input.PollingEnabled)
	if token := strings.TrimSpace(input.BotToken); token != "" {
		encrypted, err := security.Encrypt([]byte(token), telegramCipherKey(h.EncryptionEnv()))
		if err != nil {
			return telegramSettingsRecord{}, err
		}
		record.Set("bot_token_encrypted", encrypted)
	}
	if err := h.Save(record); err != nil {
		return telegramSettingsRecord{}, err
	}
	return telegramSettingsFromRecord(record), nil
}

func (h *Hub) decryptTelegramToken(settings telegramSettingsRecord) (string, error) {
	if strings.TrimSpace(settings.BotTokenEncrypted) == "" {
		return "", nil
	}
	raw, err := security.Decrypt(settings.BotTokenEncrypted, telegramCipherKey(h.EncryptionEnv()))
	if err != nil {
		return "", err
	}
	return string(raw), nil
}

func (h *Hub) listTelegramDestinations() ([]telegramDestinationRecord, error) {
	records, err := h.FindRecordsByFilter(CollectionTelegramDestinations, "id != ''", "name", -1, 0)
	if err != nil {
		return nil, err
	}
	destinations := make([]telegramDestinationRecord, 0, len(records))
	for _, record := range records {
		destinations = append(destinations, telegramDestinationFromRecord(record))
	}
	return destinations, nil
}

func (h *Hub) findTelegramDestinationByID(id string) (*core.Record, error) {
	return h.FindRecordById(CollectionTelegramDestinations, id)
}

func (h *Hub) findTelegramDestinationByChatID(chatID string) (*core.Record, error) {
	return h.FindFirstRecordByFilter(CollectionTelegramDestinations, "chat_id = {:chat_id}", dbx.Params{"chat_id": chatID})
}

func (h *Hub) listTelegramNotificationPolicies(destinationID string) ([]telegramNotificationPolicyRecord, error) {
	records, err := h.FindRecordsByFilter(CollectionTelegramNotificationPolicies, "destination = {:destination}", "name,id", -1, 0, dbx.Params{"destination": destinationID})
	if err != nil {
		return nil, err
	}
	policies := make([]telegramNotificationPolicyRecord, 0, len(records))
	for _, record := range records {
		policies = append(policies, telegramNotificationPolicyFromRecord(record))
	}
	return policies, nil
}

func (h *Hub) listAllTelegramNotificationPolicies() (map[string][]telegramNotificationPolicyRecord, error) {
	records, err := h.FindRecordsByFilter(CollectionTelegramNotificationPolicies, "id != ''", "destination,name,id", -1, 0)
	if err != nil {
		return nil, err
	}
	grouped := make(map[string][]telegramNotificationPolicyRecord)
	for _, record := range records {
		policy := telegramNotificationPolicyFromRecord(record)
		grouped[policy.DestinationID] = append(grouped[policy.DestinationID], policy)
	}
	return grouped, nil
}

func (h *Hub) findTelegramNotificationPolicy(destinationID, policyID string) (*core.Record, error) {
	return h.FindFirstRecordByFilter(CollectionTelegramNotificationPolicies, "id = {:id} && destination = {:destination}", dbx.Params{
		"id": policyID, "destination": destinationID,
	})
}

func telegramNotificationPolicyFromRecord(record *core.Record) telegramNotificationPolicyRecord {
	policy := telegramNotificationPolicyRecord{NodeScope: []string{}, AlertLevelScope: []string{}}
	if record == nil {
		return policy
	}
	policy.ID = record.Id
	policy.DestinationID = record.GetString("destination")
	policy.Name = record.GetString("name")
	policy.Enabled = record.GetBool("enabled")
	policy.NodeScopeMode = record.GetString("node_scope_mode")
	_ = record.UnmarshalJSONField("node_scope", &policy.NodeScope)
	_ = record.UnmarshalJSONField("alert_level_scope", &policy.AlertLevelScope)
	return policy
}

func (h *Hub) validateTelegramNotificationPolicyReferences(destinationID string, input TelegramNotificationPolicyInput) error {
	if _, err := h.findTelegramDestinationByID(destinationID); err != nil {
		return fmt.Errorf("destination is invalid")
	}
	for _, systemID := range normalizeTelegramStringSlice(input.NodeScope) {
		if _, err := h.FindRecordById("systems", systemID); err != nil {
			return fmt.Errorf("nodeScope contains an unknown system")
		}
	}
	return nil
}

func (h *Hub) upsertTelegramNotificationPolicy(record *core.Record, destinationID string, input TelegramNotificationPolicyInput) (*core.Record, error) {
	if err := validateTelegramNotificationPolicyInput(input); err != nil {
		return nil, err
	}
	if err := h.validateTelegramNotificationPolicyReferences(destinationID, input); err != nil {
		return nil, err
	}
	if record != nil && record.GetString("destination") != "" && record.GetString("destination") != destinationID {
		return nil, fmt.Errorf("policy does not belong to destination")
	}
	existing, err := h.FindRecordsByFilter(CollectionTelegramNotificationPolicies, "destination = {:destination} && name = {:name}", "", 1, 0, dbx.Params{
		"destination": destinationID, "name": strings.TrimSpace(input.Name),
	})
	if err != nil {
		return nil, err
	}
	if len(existing) != 0 && (record == nil || existing[0].Id != record.Id) {
		return nil, errTelegramPolicyNameConflict
	}
	if record == nil {
		collection, err := h.FindCachedCollectionByNameOrId(CollectionTelegramNotificationPolicies)
		if err != nil {
			return nil, err
		}
		record = core.NewRecord(collection)
	}
	record.Set("destination", destinationID)
	record.Set("name", strings.TrimSpace(input.Name))
	record.Set("enabled", boolValue(input.Enabled, true))
	record.Set("node_scope_mode", strings.ToLower(strings.TrimSpace(input.NodeScopeMode)))
	record.Set("node_scope", normalizeTelegramStringSlice(input.NodeScope))
	record.Set("alert_level_scope", normalizeTelegramAlertScopes(input.AlertLevelScope))
	if err := h.Save(record); err != nil {
		return nil, err
	}
	return record, nil
}

func (h *Hub) ensureTelegramDestinationDefaultPolicy(destination *core.Record) error {
	if destination == nil {
		return fmt.Errorf("destination is required")
	}
	existing, err := h.FindRecordsByFilter(CollectionTelegramNotificationPolicies, "destination = {:destination} && name = {:name}", "", 1, 0, dbx.Params{
		"destination": destination.Id, "name": TelegramDefaultPolicyName,
	})
	if err != nil || len(existing) != 0 {
		return err
	}
	legacy := telegramDestinationFromRecord(destination)
	mode := TelegramNodeScopeAll
	if len(legacy.NodeScope) != 0 {
		mode = TelegramNodeScopeSelected
	}
	_, err = h.upsertTelegramNotificationPolicy(nil, destination.Id, TelegramNotificationPolicyInput{
		Name: TelegramDefaultPolicyName, Enabled: &legacy.Enabled, NodeScopeMode: mode,
		NodeScope: legacy.NodeScope, AlertLevelScope: legacy.AlertLevelScope,
	})
	return err
}

func telegramDestinationFromRecord(record *core.Record) telegramDestinationRecord {
	if record == nil {
		return telegramDestinationRecord{}
	}
	destination := telegramDestinationRecord{
		ID:              record.Id,
		UserID:          record.GetString("user"),
		Name:            record.GetString("name"),
		ChatID:          record.GetString("chat_id"),
		ChatType:        record.GetString("chat_type"),
		Role:            record.GetString("role"),
		Enabled:         record.GetBool("enabled"),
		LastError:       record.GetString("last_error"),
		NodeScope:       []string{},
		AlertLevelScope: []string{},
	}
	_ = record.UnmarshalJSONField("node_scope", &destination.NodeScope)
	_ = record.UnmarshalJSONField("alert_level_scope", &destination.AlertLevelScope)

	if value := record.GetDateTime("mute_until"); !value.IsZero() {
		parsed := value.Time()
		destination.MuteUntil = &parsed
	}
	if value := record.GetDateTime("last_test_at"); !value.IsZero() {
		parsed := value.Time()
		destination.LastTestAt = &parsed
	}
	if value := record.GetDateTime("last_delivery_at"); !value.IsZero() {
		parsed := value.Time()
		destination.LastDeliveryAt = &parsed
	}
	return destination
}

func (h *Hub) validateTelegramDestinationReferences(input TelegramDestinationInput) error {
	if input.UserID != "" {
		if _, err := h.FindRecordById("users", input.UserID); err != nil {
			return fmt.Errorf("userId is invalid")
		}
	}
	for _, systemID := range normalizeTelegramStringSlice(input.NodeScope) {
		if _, err := h.FindRecordById("systems", systemID); err != nil {
			return fmt.Errorf("nodeScope contains an unknown system")
		}
	}
	return nil
}

func (h *Hub) upsertTelegramDestination(record *core.Record, input TelegramDestinationInput) (*core.Record, error) {
	if record == nil {
		collection, err := h.FindCachedCollectionByNameOrId(CollectionTelegramDestinations)
		if err != nil {
			return nil, err
		}
		record = core.NewRecord(collection)
	}
	setTelegramDestinationRecord(record, input)
	if err := h.Save(record); err != nil {
		return nil, err
	}
	if input.NodeScope != nil || input.AlertLevelScope != nil {
		if err := h.upsertTelegramDefaultPolicyFromLegacy(record, input); err != nil {
			return nil, err
		}
	}
	return record, nil
}

func setTelegramDestinationRecord(record *core.Record, input TelegramDestinationInput) {
	record.Set("user", strings.TrimSpace(input.UserID))
	record.Set("name", strings.TrimSpace(input.Name))
	record.Set("chat_id", strings.TrimSpace(input.ChatID))
	record.Set("chat_type", normalizeTelegramChatType(input.ChatType))
	record.Set("role", normalizeTelegramRole(input.Role))
	record.Set("enabled", boolValue(input.Enabled, true))
	record.Set("node_scope", normalizeTelegramStringSlice(input.NodeScope))
	record.Set("alert_level_scope", normalizeTelegramAlertScopes(input.AlertLevelScope))
	if input.MuteUntil != nil {
		dt, _ := types.ParseDateTime(input.MuteUntil.UTC())
		record.Set("mute_until", dt)
	} else {
		record.Set("mute_until", nil)
	}
}

func (h *Hub) setTelegramDestinationLastTest(destinationID string, when time.Time, lastError string) error {
	record, err := h.findTelegramDestinationByID(destinationID)
	if err != nil {
		return err
	}
	record.Set("last_error", strings.TrimSpace(lastError))
	if when.IsZero() {
		record.Set("last_test_at", nil)
	} else {
		dt, _ := types.ParseDateTime(when.UTC())
		record.Set("last_test_at", dt)
	}
	return h.Save(record)
}

func (h *Hub) setTelegramDestinationDeliveryState(destinationID string, when time.Time, lastError string) error {
	record, err := h.findTelegramDestinationByID(destinationID)
	if err != nil {
		return err
	}
	record.Set("last_error", strings.TrimSpace(lastError))
	if when.IsZero() {
		record.Set("last_delivery_at", nil)
	} else {
		dt, _ := types.ParseDateTime(when.UTC())
		record.Set("last_delivery_at", dt)
	}
	return h.Save(record)
}

func (h *Hub) setTelegramSettingsRuntimeState(username string, pollOffset int64, lastError string) error {
	record, err := h.getTelegramSettingsRecord()
	if err != nil || record == nil {
		return err
	}
	if username != "" {
		record.Set("bot_username", username)
	}
	if pollOffset >= 0 {
		record.Set("last_poll_offset", pollOffset)
	}
	record.Set("last_error", strings.TrimSpace(lastError))
	return h.Save(record)
}
