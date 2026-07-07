package hub

import (
	"fmt"
	"time"

	"github.com/pocketbase/pocketbase/tools/types"
)

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
	return fmt.Sprintf("已暂停 %s 的 Telegram 通知，恢复时间：%s。", destination.Name, until.Format(time.RFC3339)), telegramHelpKeyboard(), nil
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
	return fmt.Sprintf("已恢复 %s 的 Telegram 通知。", destination.Name), telegramHelpKeyboard(), nil
}
