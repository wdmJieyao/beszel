package hub

import (
	"strings"

	"github.com/henrygd/beszel/internal/alerts"
)

func telegramDestinationCanUseAdminMenu(destination telegramDestinationRecord) bool {
	return destination.Enabled && destination.Role == TelegramRoleAdmin
}

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
	sanitized.Message = strings.ReplaceAll(sanitized.Message, data.Link, "")
	return sanitized
}
