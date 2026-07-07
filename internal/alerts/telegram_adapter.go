package alerts

type TelegramAlertSender interface {
	SendTelegramAlert(AlertMessageData) error
}

func (am *AlertManager) SetTelegramSender(sender TelegramAlertSender) {
	am.telegramSender = sender
}
