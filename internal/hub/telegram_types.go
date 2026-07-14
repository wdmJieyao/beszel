package hub

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"slices"
	"strings"
	"time"
)

const (
	CollectionTelegramSettings             = "telegram_settings"
	CollectionTelegramDestinations         = "telegram_destinations"
	CollectionTelegramNotificationPolicies = "telegram_notification_policies"

	TelegramDefaultPolicyName = "默认规则"
	TelegramNodeScopeAll      = "all"
	TelegramNodeScopeSelected = "selected"

	TelegramRoleAdmin    = "admin"
	TelegramRoleReadOnly = "read_only"

	TelegramChatTypePrivate    = "private"
	TelegramChatTypeGroup      = "group"
	TelegramChatTypeSupergroup = "supergroup"
	TelegramChatTypeChannel    = "channel"
	TelegramChatTypeUnknown    = "unknown"
)

var (
	errTelegramPolicyNameConflict = errors.New("policy name already exists for destination")
	telegramTokenPattern          = regexp.MustCompile(`^\d+:[A-Za-z0-9_-]{10,}$`)
	telegramChatIDRegexp          = regexp.MustCompile(`^-?\d+$`)
	validTelegramRoles            = []string{TelegramRoleAdmin, TelegramRoleReadOnly}
	TelegramAlertLevelScopes      = []string{
		"status", "cpu", "memory", "disk", "bandwidth", "temperature",
		"loadavg1", "loadavg5", "loadavg15", "gpu", "battery", "smart",
	}
	validTelegramTypes = []string{
		TelegramChatTypePrivate,
		TelegramChatTypeGroup,
		TelegramChatTypeSupergroup,
		TelegramChatTypeChannel,
		TelegramChatTypeUnknown,
	}
)

type TelegramSettingsInput struct {
	Enabled        bool   `json:"enabled"`
	PollingEnabled bool   `json:"pollingEnabled"`
	BotToken       string `json:"botToken"`
}

type TelegramSettingsTestInput struct {
	BotToken string `json:"botToken"`
}

type TelegramSettingsResponse struct {
	Enabled        bool   `json:"enabled"`
	PollingEnabled bool   `json:"pollingEnabled"`
	BotUsername    string `json:"botUsername"`
	HasToken       bool   `json:"hasToken"`
	LastError      string `json:"lastError"`
	Updated        string `json:"updated,omitempty"`
}

type TelegramTestResponse struct {
	OK          bool               `json:"ok"`
	BotUsername string             `json:"botUsername,omitempty"`
	Error       string             `json:"error,omitempty"`
	Stages      TelegramTestStages `json:"stages"`
}

type TelegramTestStages struct {
	Credentials TelegramTestStage `json:"credentials"`
	CommandMenu TelegramTestStage `json:"commandMenu"`
}

type TelegramTestStage struct {
	OK    bool   `json:"ok"`
	Error string `json:"error"`
}

type TelegramDestinationInput struct {
	UserID          string     `json:"userId"`
	Name            string     `json:"name"`
	ChatID          string     `json:"chatId"`
	ChatType        string     `json:"chatType"`
	Role            string     `json:"role"`
	Enabled         *bool      `json:"enabled,omitempty"`
	NodeScope       []string   `json:"nodeScope"`
	AlertLevelScope []string   `json:"alertLevelScope"`
	MuteUntil       *time.Time `json:"muteUntil,omitempty"`
}

type TelegramDestinationPatchInput struct {
	UserID          TelegramPatchField[string]   `json:"userId"`
	Name            TelegramPatchField[string]   `json:"name"`
	ChatID          TelegramPatchField[string]   `json:"chatId"`
	ChatType        TelegramPatchField[string]   `json:"chatType"`
	Role            TelegramPatchField[string]   `json:"role"`
	Enabled         TelegramPatchField[bool]     `json:"enabled"`
	NodeScope       TelegramPatchField[[]string] `json:"nodeScope"`
	AlertLevelScope TelegramPatchField[[]string] `json:"alertLevelScope"`
	MuteUntil       TelegramOptionalTime         `json:"muteUntil"`
}

type TelegramPatchField[T any] struct {
	Set   bool
	Null  bool
	Value T
}

func (field *TelegramPatchField[T]) UnmarshalJSON(data []byte) error {
	field.Set = true
	if bytes.Equal(bytes.TrimSpace(data), []byte("null")) {
		field.Null = true
		return nil
	}
	return json.Unmarshal(data, &field.Value)
}

func assignTelegramPatchField[T any](field TelegramPatchField[T], name string, target *T) error {
	if !field.Set {
		return nil
	}
	if field.Null {
		return fmt.Errorf("%s cannot be null", name)
	}
	*target = field.Value
	return nil
}

type TelegramOptionalTime struct {
	Set   bool
	Value *time.Time
}

func (value *TelegramOptionalTime) UnmarshalJSON(data []byte) error {
	value.Set = true
	if bytes.Equal(bytes.TrimSpace(data), []byte("null")) {
		value.Value = nil
		return nil
	}
	var timestamp time.Time
	if err := json.Unmarshal(data, &timestamp); err != nil {
		return fmt.Errorf("muteUntil must be an RFC3339 timestamp or null")
	}
	value.Value = &timestamp
	return nil
}

type TelegramDestinationResponse struct {
	ID              string   `json:"id"`
	UserID          string   `json:"userId,omitempty"`
	Name            string   `json:"name"`
	ChatID          string   `json:"chatId"`
	ChatType        string   `json:"chatType"`
	Role            string   `json:"role"`
	Enabled         bool     `json:"enabled"`
	PolicyCount     int      `json:"policyCount"`
	NodeScope       []string `json:"nodeScope"`
	AlertLevelScope []string `json:"alertLevelScope"`
	MuteUntil       string   `json:"muteUntil,omitempty"`
	LastTestAt      string   `json:"lastTestAt,omitempty"`
	LastDeliveryAt  string   `json:"lastDeliveryAt,omitempty"`
	LastError       string   `json:"lastError"`
}

type TelegramDestinationsResponse struct {
	Destinations []TelegramDestinationResponse `json:"destinations"`
}

type TelegramNotificationPolicyInput struct {
	Name            string   `json:"name"`
	Enabled         *bool    `json:"enabled,omitempty"`
	NodeScopeMode   string   `json:"nodeScopeMode"`
	NodeScope       []string `json:"nodeScope"`
	AlertLevelScope []string `json:"alertLevelScope"`
}

type TelegramNotificationPolicyPatchInput struct {
	Name            TelegramPatchField[string]   `json:"name"`
	Enabled         TelegramPatchField[bool]     `json:"enabled"`
	NodeScopeMode   TelegramPatchField[string]   `json:"nodeScopeMode"`
	NodeScope       TelegramPatchField[[]string] `json:"nodeScope"`
	AlertLevelScope TelegramPatchField[[]string] `json:"alertLevelScope"`
}

type TelegramNotificationPolicyResponse struct {
	ID              string   `json:"id"`
	DestinationID   string   `json:"destinationId"`
	Name            string   `json:"name"`
	Enabled         bool     `json:"enabled"`
	NodeScopeMode   string   `json:"nodeScopeMode"`
	NodeScope       []string `json:"nodeScope"`
	AlertLevelScope []string `json:"alertLevelScope"`
}

type TelegramPolicyListResponse struct {
	Policies []TelegramNotificationPolicyResponse `json:"policies"`
}

type telegramSettingsRecord struct {
	Enabled           bool
	PollingEnabled    bool
	BotTokenEncrypted string
	BotUsername       string
	LastPollOffset    int64
	LastError         string
	Updated           string
}

type telegramDestinationRecord struct {
	ID              string
	UserID          string
	Name            string
	ChatID          string
	ChatType        string
	Role            string
	Enabled         bool
	NodeScope       []string
	AlertLevelScope []string
	MuteUntil       *time.Time
	LastTestAt      *time.Time
	LastDeliveryAt  *time.Time
	LastError       string
}

type telegramNotificationPolicyRecord struct {
	ID              string
	DestinationID   string
	Name            string
	Enabled         bool
	NodeScopeMode   string
	NodeScope       []string
	AlertLevelScope []string
}

func validateTelegramSettingsInput(input TelegramSettingsInput) error {
	if strings.TrimSpace(input.BotToken) == "" {
		return nil
	}
	if !telegramTokenPattern.MatchString(strings.TrimSpace(input.BotToken)) {
		return fmt.Errorf("invalid telegram bot token format")
	}
	return nil
}

func validateTelegramDestinationInput(input TelegramDestinationInput) error {
	if strings.TrimSpace(input.Name) == "" {
		return fmt.Errorf("destination name is required")
	}
	if !telegramChatIDRegexp.MatchString(strings.TrimSpace(input.ChatID)) {
		return fmt.Errorf("chatId must be a valid telegram chat id")
	}
	if input.ChatType != "" && !slices.Contains(validTelegramTypes, strings.ToLower(strings.TrimSpace(input.ChatType))) {
		return fmt.Errorf("chatType is invalid")
	}
	if !slices.Contains(validTelegramRoles, strings.ToLower(strings.TrimSpace(input.Role))) {
		return fmt.Errorf("role is invalid")
	}
	for _, scope := range normalizeTelegramStringSlice(input.AlertLevelScope) {
		if !slices.Contains(TelegramAlertLevelScopes, strings.ToLower(scope)) {
			return fmt.Errorf("alertLevelScope contains an unsupported value")
		}
	}
	return nil
}

func validateTelegramNotificationPolicyInput(input TelegramNotificationPolicyInput) error {
	if strings.TrimSpace(input.Name) == "" {
		return fmt.Errorf("policy name is required")
	}
	mode := strings.ToLower(strings.TrimSpace(input.NodeScopeMode))
	if mode != TelegramNodeScopeAll && mode != TelegramNodeScopeSelected {
		return fmt.Errorf("nodeScopeMode must be all or selected")
	}
	nodes := normalizeTelegramStringSlice(input.NodeScope)
	if mode == TelegramNodeScopeAll && len(nodes) != 0 {
		return fmt.Errorf("nodeScope must be empty when nodeScopeMode is all")
	}
	if mode == TelegramNodeScopeSelected && len(nodes) == 0 {
		return fmt.Errorf("nodeScope must contain at least one system when nodeScopeMode is selected")
	}
	for _, scope := range normalizeTelegramAlertScopes(input.AlertLevelScope) {
		if !slices.Contains(TelegramAlertLevelScopes, scope) {
			return fmt.Errorf("alertLevelScope contains an unsupported value")
		}
	}
	return nil
}

func normalizeTelegramRole(role string) string {
	role = strings.ToLower(strings.TrimSpace(role))
	if role == TelegramRoleReadOnly {
		return TelegramRoleReadOnly
	}
	return TelegramRoleAdmin
}

func normalizeTelegramChatType(chatType string) string {
	chatType = strings.ToLower(strings.TrimSpace(chatType))
	if slices.Contains(validTelegramTypes, chatType) {
		return chatType
	}
	return TelegramChatTypeUnknown
}

func normalizeTelegramAlertScopes(values []string) []string {
	normalized := normalizeTelegramStringSlice(values)
	for index := range normalized {
		normalized[index] = strings.ToLower(normalized[index])
	}
	return normalized
}

func normalizeTelegramStringSlice(values []string) []string {
	if len(values) == 0 {
		return []string{}
	}
	normalized := make([]string, 0, len(values))
	seen := make(map[string]struct{}, len(values))
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" {
			continue
		}
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		normalized = append(normalized, value)
	}
	return normalized
}
