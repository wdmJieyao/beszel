package hub

import (
	"context"
	"fmt"
	"strings"
)

func (h *Hub) handleTelegramCommand(ctx context.Context, token string, command telegramCommand) error {
	destination, authorized, err := h.authorizeTelegramCommand(command.ChatID)
	if err != nil {
		return err
	}
	if !authorized {
		return h.telegramTransport.SendMessage(ctx, token, command.ChatID, "此 Telegram 会话未在 Beszel 面板中授权。", nil)
	}
	if !telegramDestinationCanUseAdminMenu(destination) && command.Name != "help" {
		return h.telegramTransport.SendMessage(ctx, token, command.ChatID, telegramReadOnlyMenuMessage(destination), nil)
	}

	var message string
	var options *TelegramSendOptions
	switch command.Name {
	case "status":
		message, options, err = h.telegramStatusOverview()
	case "alerts":
		message, options, err = h.telegramAlertSummary()
	case "systems":
		message, options, err = h.telegramSystemList()
	case "system":
		message, options, err = h.telegramSystemDetail(telegramCommandArg(command.Args, 0))
	case "mute":
		message, options, err = h.telegramMuteDestination(destination)
	case "unmute":
		message, options, err = h.telegramUnmuteDestination(destination)
	default:
		message, options = telegramHelpMessage(destination), telegramHelpKeyboard()
	}
	if err != nil {
		message = "Telegram 菜单处理失败：" + sanitizeTelegramError(err.Error()).Error()
		options = nil
	}
	if command.CallbackQueryID != "" {
		_ = h.telegramTransport.AnswerCallbackQuery(ctx, token, command.CallbackQueryID, "已处理")
	}
	return h.telegramTransport.SendMessage(ctx, token, command.ChatID, message, options)
}

func (h *Hub) authorizeTelegramCommand(chatID string) (telegramDestinationRecord, bool, error) {
	record, err := h.findTelegramDestinationByChatID(chatID)
	if err != nil {
		return telegramDestinationRecord{}, false, nil
	}
	destination := telegramDestinationFromRecord(record)
	if !destination.Enabled {
		return destination, false, nil
	}
	return destination, true, nil
}

func telegramHelpMessage(destination telegramDestinationRecord) string {
	if destination.Role == TelegramRoleReadOnly {
		return telegramReadOnlyMenuMessage(destination)
	}
	return strings.Join([]string{
		"Beszel Telegram 管理菜单",
		"/status - 查看面板状态总览",
		"/alerts - 查看告警概要",
		"/systems - 查看节点列表",
		"/system <序号或ID> - 查看节点详情",
		"/mute - 暂停当前 Telegram 目的地通知 1 小时",
		"/unmute - 恢复当前 Telegram 目的地通知",
	}, "\n")
}

func telegramHelpKeyboard() *TelegramSendOptions {
	return &TelegramSendOptions{ReplyMarkup: telegramInlineKeyboard(
		[]map[string]string{
			telegramButton("状态", "status"),
			telegramButton("告警", "alerts"),
		},
		[]map[string]string{
			telegramButton("节点", "systems"),
			telegramButton("静音", "mute"),
			telegramButton("恢复", "unmute"),
		},
	)}
}

func (h *Hub) telegramStatusOverview() (string, *TelegramSendOptions, error) {
	systems, err := h.FindRecordsByFilter("systems", "id != ''", "name", -1, 0)
	if err != nil {
		return "", nil, err
	}
	up, down, paused := 0, 0, 0
	for _, system := range systems {
		switch strings.ToLower(system.GetString("status")) {
		case "up":
			up++
		case "paused":
			paused++
		default:
			down++
		}
	}
	alertsCount, _ := h.CountRecords("alerts")
	return fmt.Sprintf("Beszel 状态总览\n节点：%d 个（在线 %d，异常 %d，暂停 %d）\n告警规则：%d 条", len(systems), up, down, paused, alertsCount), telegramHelpKeyboard(), nil
}

func (h *Hub) telegramAlertSummary() (string, *TelegramSendOptions, error) {
	alerts, err := h.FindRecordsByFilter("alerts", "id != ''", "name", -1, 0)
	if err != nil {
		return "", nil, err
	}
	triggered := 0
	for _, alert := range alerts {
		if alert.GetBool("triggered") {
			triggered++
		}
	}
	return fmt.Sprintf("Beszel 告警概要\n告警规则：%d 条\n当前触发：%d 条", len(alerts), triggered), telegramHelpKeyboard(), nil
}
