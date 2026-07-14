package hub

import (
	"regexp"
	"strings"

	"github.com/henrygd/beszel/internal/alerts"
)

func telegramDestinationCanUseAdminMenu(destination telegramDestinationRecord) bool {
	return destination.Enabled && destination.Role == TelegramRoleAdmin && destination.ChatType == TelegramChatTypePrivate
}

var (
	telegramURLPattern  = regexp.MustCompile(`https?://\S+`)
	telegramIPv4Pattern = regexp.MustCompile(`\b(?:\d{1,3}\.){3}\d{1,3}\b`)
)

func telegramReadOnlyMenuMessage(destination telegramDestinationRecord) string {
	name := strings.TrimSpace(destination.Name)
	if name == "" {
		name = "当前 Telegram 目的地"
	}
	return name + " 是只读通知渠道，只会接收已授权范围内的监控通知，不能使用管理菜单。"
}

func sanitizeTelegramReadOnlyAlert(data alerts.AlertMessageData) alerts.AlertMessageData {
	sanitized := data
	sanitized.Link = ""
	sanitized.LinkText = ""
	for _, text := range []*string{&sanitized.Title, &sanitized.Message} {
		*text = strings.ReplaceAll(*text, data.Link, "")
		*text = telegramURLPattern.ReplaceAllString(*text, "[链接已隐藏]")
		*text = telegramIPv4Pattern.ReplaceAllString(*text, "[地址已隐藏]")
		*text = strings.TrimSpace(*text)
	}
	return sanitized
}
