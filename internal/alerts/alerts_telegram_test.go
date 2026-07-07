//go:build testing

package alerts_test

import (
	"testing"

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

type fakeTelegramSender struct {
	messages []alerts.AlertMessageData
}

func (f *fakeTelegramSender) SendTelegramAlert(data alerts.AlertMessageData) error {
	f.messages = append(f.messages, data)
	return nil
}
