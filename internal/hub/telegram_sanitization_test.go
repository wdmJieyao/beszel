//go:build testing

package hub

import (
	"testing"

	"github.com/henrygd/beszel/internal/alerts"
	"github.com/stretchr/testify/assert"
)

func TestTelegramReadOnlyAlertMessageOmitsAdminLink(t *testing.T) {
	message := telegramReadOnlyAlertMessage(sanitizeTelegramReadOnlyAlert(alerts.AlertMessageData{
		Title:   "Connection to node-1 is down",
		Message: "Connection to node-1 is down https://panel.example/system/node-1",
		Link:    "https://panel.example/system/node-1",
	}))

	assert.Contains(t, message, "Beszel 告警摘要")
	assert.NotContains(t, message, "https://panel.example")
}
