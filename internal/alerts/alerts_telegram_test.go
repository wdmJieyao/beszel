//go:build testing

package alerts_test

import (
	"testing"
	"time"

	"github.com/henrygd/beszel/internal/alerts"
	beszelTests "github.com/henrygd/beszel/internal/tests"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSendAlertInvokesTelegramSender(t *testing.T) {
	hub, user := beszelTests.GetHubWithUser(t)
	defer hub.Cleanup()

	sender := &fakeTelegramSender{}
	hub.GetAlertManager().SetTelegramSender(sender)

	err := hub.GetAlertManager().SendAlert(alerts.AlertMessageData{
		UserID:   user.Id,
		SystemID: "",
		Title:    "Test alert",
		Message:  "Body",
	})
	require.NoError(t, err)
	require.Len(t, sender.messages, 1)
	assert.Equal(t, "Test alert", sender.messages[0].Title)
}

func TestSendAlertPreservesStructuredTelegramContext(t *testing.T) {
	hub, user := beszelTests.GetHubWithUser(t)
	defer hub.Cleanup()

	sender := &fakeTelegramSender{}
	hub.GetAlertManager().SetTelegramSender(sender)
	eventTime := time.Date(2026, time.July, 10, 9, 30, 0, 0, time.UTC)

	err := hub.GetAlertManager().SendAlert(alerts.AlertMessageData{
		UserID:     user.Id,
		SystemID:   "system-1",
		SystemName: "node-1",
		AlertClass: "cpu",
		Severity:   alerts.AlertSeverityCritical,
		State:      alerts.AlertStateTriggered,
		EventTime:  eventTime,
		Title:      "CPU threshold exceeded",
		Message:    "CPU averaged 95%",
	})
	require.NoError(t, err)
	require.Len(t, sender.messages, 1)
	assert.Equal(t, "node-1", sender.messages[0].SystemName)
	assert.Equal(t, "cpu", sender.messages[0].AlertClass)
	assert.Equal(t, alerts.AlertSeverityCritical, sender.messages[0].Severity)
	assert.Equal(t, alerts.AlertStateTriggered, sender.messages[0].State)
	assert.Equal(t, eventTime, sender.messages[0].EventTime)
}

type fakeTelegramSender struct {
	messages []alerts.AlertMessageData
}

func (f *fakeTelegramSender) SendTelegramAlert(data alerts.AlertMessageData) error {
	f.messages = append(f.messages, data)
	return nil
}
