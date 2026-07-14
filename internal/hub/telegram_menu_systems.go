package hub

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/pocketbase/pocketbase/core"
)

type telegramSystemResourceUsage struct {
	CPU    float64 `json:"cpu"`
	Memory float64 `json:"mp"`
	Disk   float64 `json:"dp"`
}

func (h *Hub) telegramSystemList() (string, *TelegramSendOptions, error) {
	systems, err := h.FindRecordsByFilter("systems", "id != ''", "name", -1, 0)
	if err != nil {
		return "", nil, err
	}
	if len(systems) == 0 {
		return "暂无节点。", nil, nil
	}
	lines := []string{"Beszel 节点列表"}
	buttonRows := make([][]map[string]string, 0)
	const maxTelegramSystemListItems = 25
	for i, system := range systems {
		if i >= maxTelegramSystemListItems {
			lines = append(lines, fmt.Sprintf("另有 %d 个节点，请在面板中查看。", len(systems)-maxTelegramSystemListItems))
			break
		}
		index := i + 1
		status := strings.TrimSpace(system.GetString("status"))
		if status == "" {
			status = "unknown"
		}
		usage, hasUsage := telegramSystemCurrentUsage(system)
		if hasUsage {
			lines = append(lines, fmt.Sprintf("%d. %s（%s） CPU %.1f%% · 内存 %.1f%% · 磁盘 %.1f%%", index, system.GetString("name"), status, usage.CPU, usage.Memory, usage.Disk))
		} else {
			lines = append(lines, fmt.Sprintf("%d. %s（%s）资源暂无数据", index, system.GetString("name"), status))
		}
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
		return "未找到对应节点。请使用 /systems 查看节点列表。", nil, nil
	}
	lines := []string{
		"Beszel 节点详情",
		"名称：" + system.GetString("name"),
		"状态：" + valueOrUnknown(system.GetString("status")),
	}
	usage, hasUsage := telegramSystemCurrentUsage(system)
	if hasUsage {
		lines = append(lines,
			fmt.Sprintf("CPU：%.1f%%", usage.CPU),
			fmt.Sprintf("内存：%.1f%%", usage.Memory),
			fmt.Sprintf("磁盘：%.1f%%", usage.Disk),
		)
	}
	if updated := system.GetDateTime("updated"); !updated.IsZero() {
		lines = append(lines, "最后上报："+updated.Time().Local().Format("2006-01-02 15:04:05"))
	}
	return strings.Join(lines, "\n"), &TelegramSendOptions{ReplyMarkup: telegramInlineKeyboard(
		[]map[string]string{telegramButton("返回节点列表", "systems")},
	)}, nil
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

func telegramSystemCurrentUsage(system *core.Record) (telegramSystemResourceUsage, bool) {
	var usage telegramSystemResourceUsage
	raw := system.GetString("info")
	if raw == "" {
		return usage, false
	}
	if err := json.Unmarshal([]byte(raw), &usage); err != nil {
		return usage, false
	}
	return usage, true
}

func valueOrUnknown(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return "unknown"
	}
	return value
}
