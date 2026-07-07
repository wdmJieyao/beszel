package hub

import (
	"crypto/sha256"
	"fmt"
	"strings"
	"time"

	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/tools/security"
	"github.com/pocketbase/pocketbase/tools/types"
)

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
	record.Set("user", strings.TrimSpace(input.UserID))
	record.Set("name", strings.TrimSpace(input.Name))
	record.Set("chat_id", strings.TrimSpace(input.ChatID))
	record.Set("chat_type", normalizeTelegramChatType(input.ChatType))
	record.Set("role", normalizeTelegramRole(input.Role))
	record.Set("enabled", boolValue(input.Enabled, true))
	record.Set("node_scope", normalizeTelegramStringSlice(input.NodeScope))
	record.Set("alert_level_scope", normalizeTelegramStringSlice(input.AlertLevelScope))
	if input.MuteUntil != nil {
		dt, _ := types.ParseDateTime(input.MuteUntil.UTC())
		record.Set("mute_until", dt)
	} else {
		record.Set("mute_until", nil)
	}
	if err := h.Save(record); err != nil {
		return nil, err
	}
	return record, nil
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
