package hub

import (
	"fmt"
	"regexp"
	"slices"
	"strings"
	"time"
)

const (
	CollectionTelegramSettings     = "telegram_settings"
	CollectionTelegramDestinations = "telegram_destinations"

	TelegramRoleAdmin    = "admin"
	TelegramRoleReadOnly = "read_only"

	TelegramChatTypePrivate    = "private"
	TelegramChatTypeGroup      = "group"
	TelegramChatTypeSupergroup = "supergroup"
	TelegramChatTypeChannel    = "channel"
	TelegramChatTypeUnknown    = "unknown"
)

var (
	telegramTokenPattern = regexp.MustCompile(`^\d+:[A-Za-z0-9_-]{10,}$`)
	telegramChatIDRegexp = regexp.MustCompile(`^-?\d+$`)
	validTelegramRoles   = []string{TelegramRoleAdmin, TelegramRoleReadOnly}
	validTelegramTypes   = []string{
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
	OK          bool   `json:"ok"`
	BotUsername string `json:"botUsername,omitempty"`
	Error       string `json:"error,omitempty"`
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

type TelegramDestinationResponse struct {
	ID              string   `json:"id"`
	UserID          string   `json:"userId,omitempty"`
	Name            string   `json:"name"`
	ChatID          string   `json:"chatId"`
	ChatType        string   `json:"chatType"`
	Role            string   `json:"role"`
	Enabled         bool     `json:"enabled"`
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
