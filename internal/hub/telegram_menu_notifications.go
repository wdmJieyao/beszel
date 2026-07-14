package hub

import (
	"fmt"
	"time"

	"github.com/pocketbase/pocketbase/tools/types"
)

func (h *Hub) telegramNotificationSettingsStatus(destination telegramDestinationRecord) (string, *TelegramSendOptions, error) {
	settings, err := h.loadTelegramSettings()
	if err != nil {
		return "", nil, err
	}
	status := "已关闭"
	if settings.Enabled {
		status = "已启用"
	}
	polling := "已关闭"
	if settings.PollingEnabled {
		polling = "已启用"
	}
	mute := "未静音"
	if destination.MuteUntil != nil && destination.MuteUntil.After(time.Now().UTC()) {
		mute = "静音至 " + destination.MuteUntil.UTC().Format(time.RFC3339)
	}
	return fmt.Sprintf("Telegram 通知设置\n机器人：%s\n菜单轮询：%s\n当前目的地：%s\n通知状态：%s", status, polling, destination.Name, mute), nil, nil
}

func (h *Hub) telegramBindingStatus(destination telegramDestinationRecord) (string, *TelegramSendOptions, error) {
	return fmt.Sprintf("Telegram 绑定已验证\n名称：%s\n角色：%s\n会话 ID：%s", destination.Name, destination.Role, destination.ChatID), nil, nil
}

func (h *Hub) telegramMuteDestination(destination telegramDestinationRecord) (string, *TelegramSendOptions, error) {
	record, err := h.findTelegramDestinationByID(destination.ID)
	if err != nil {
		return "", nil, err
	}
	until := time.Now().UTC().Add(time.Hour)
	dt, _ := types.ParseDateTime(until)
	record.Set("mute_until", dt)
	if err := h.Save(record); err != nil {
		return "", nil, err
	}
	return fmt.Sprintf("已暂停 %s 的 Telegram 通知，恢复时间：%s。", destination.Name, until.Format(time.RFC3339)), nil, nil
}

func (h *Hub) telegramUnmuteDestination(destination telegramDestinationRecord) (string, *TelegramSendOptions, error) {
	record, err := h.findTelegramDestinationByID(destination.ID)
	if err != nil {
		return "", nil, err
	}
	record.Set("mute_until", nil)
	if err := h.Save(record); err != nil {
		return "", nil, err
	}
	return fmt.Sprintf("已恢复 %s 的 Telegram 通知。", destination.Name), nil, nil
}
