package hub

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/pocketbase/pocketbase/core"
)

func (h *Hub) telegramSystemList() (string, *TelegramSendOptions, error) {
	systems, err := h.FindRecordsByFilter("systems", "id != ''", "name", -1, 0)
	if err != nil {
		return "", nil, err
	}
	if len(systems) == 0 {
		return "暂无节点。", telegramHelpKeyboard(), nil
	}
	lines := []string{"Beszel 节点列表"}
	buttonRows := make([][]map[string]string, 0)
	for i, system := range systems {
		index := i + 1
		status := strings.TrimSpace(system.GetString("status"))
		if status == "" {
			status = "unknown"
		}
		lines = append(lines, fmt.Sprintf("%d. %s（%s）", index, system.GetString("name"), status))
		buttonRows = append(buttonRows, []map[string]string{telegramButton(system.GetString("name"), "system", telegramSystemCommandArg(system.Id, index))})
	}
	return strings.Join(lines, "\n"), &TelegramSendOptions{ReplyMarkup: telegramInlineKeyboard(buttonRows...)}, nil
}

func (h *Hub) telegramSystemDetail(identifier string) (string, *TelegramSendOptions, error) {
	system, err := h.findTelegramSystem(identifier)
	if err != nil {
		return "", nil, err
	}
	if system == nil {
		return "未找到对应节点。请使用 /systems 查看节点列表。", telegramHelpKeyboard(), nil
	}
	lines := []string{
		"Beszel 节点详情",
		"名称：" + system.GetString("name"),
		"状态：" + valueOrUnknown(system.GetString("status")),
	}
	if system.GetString("info") != "" {
		lines = append(lines, "信息已在面板中更新。")
	}
	return strings.Join(lines, "\n"), telegramHelpKeyboard(), nil
}

func (h *Hub) findTelegramSystem(identifier string) (*core.Record, error) {
	identifier = strings.TrimSpace(identifier)
	if identifier == "" {
		return nil, nil
	}
	if index, err := strconv.Atoi(identifier); err == nil && index > 0 {
		systems, err := h.FindRecordsByFilter("systems", "id != ''", "name", -1, 0)
		if err != nil {
			return nil, err
		}
		if index <= len(systems) {
			return systems[index-1], nil
		}
		return nil, nil
	}
	system, err := h.FindRecordById("systems", identifier)
	if err != nil {
		return nil, nil
	}
	return system, nil
}

func valueOrUnknown(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return "unknown"
	}
	return value
}
