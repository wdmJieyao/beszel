package hub

import (
	"fmt"
	"strconv"
	"strings"
)

type telegramCommand struct {
	ChatID          string
	Name            string
	Args            []string
	CallbackQueryID string
}

func parseTelegramCommand(update TelegramUpdate) (telegramCommand, bool) {
	if update.CallbackQuery != nil {
		chatID := ""
		if update.CallbackQuery.Message != nil {
			chatID = telegramChatIDString(update.CallbackQuery.Message.Chat.ID)
		}
		data := strings.TrimSpace(update.CallbackQuery.Data)
		if data == "" || chatID == "" {
			return telegramCommand{}, false
		}
		parts := strings.Split(data, ":")
		return telegramCommand{
			ChatID:          chatID,
			Name:            normalizeTelegramCommandName(parts[0]),
			Args:            parts[1:],
			CallbackQueryID: update.CallbackQuery.ID,
		}, true
	}
	if update.Message == nil {
		return telegramCommand{}, false
	}
	text := strings.TrimSpace(update.Message.Text)
	if text == "" {
		return telegramCommand{}, false
	}
	fields := strings.Fields(text)
	if len(fields) == 0 {
		return telegramCommand{}, false
	}
	return telegramCommand{
		ChatID: telegramChatIDString(update.Message.Chat.ID),
		Name:   normalizeTelegramCommandName(fields[0]),
		Args:   fields[1:],
	}, true
}

func normalizeTelegramCommandName(name string) string {
	name = strings.ToLower(strings.TrimSpace(name))
	name = strings.TrimPrefix(name, "/")
	if at := strings.Index(name, "@"); at >= 0 {
		name = name[:at]
	}
	switch name {
	case "start":
		return "help"
	case "system":
		return "system"
	case "systems":
		return "systems"
	case "alerts":
		return "alerts"
	case "mute":
		return "mute"
	case "unmute":
		return "unmute"
	case "status":
		return "status"
	case "help":
		return "help"
	default:
		return "help"
	}
}

func telegramChatIDString(id int64) string {
	return strconv.FormatInt(id, 10)
}

func telegramCommandArg(args []string, index int) string {
	if index < 0 || index >= len(args) {
		return ""
	}
	return strings.TrimSpace(args[index])
}

func telegramCallbackData(name string, args ...string) string {
	values := append([]string{strings.TrimSpace(name)}, args...)
	return strings.Join(values, ":")
}

func telegramInlineKeyboard(rows ...[]map[string]string) map[string]any {
	return map[string]any{"inline_keyboard": rows}
}

func telegramButton(text string, action string, args ...string) map[string]string {
	return map[string]string{
		"text":          text,
		"callback_data": telegramCallbackData(action, args...),
	}
}

func telegramSystemCommandArg(systemID string, index int) string {
	if systemID != "" {
		return systemID
	}
	return fmt.Sprintf("%d", index)
}
