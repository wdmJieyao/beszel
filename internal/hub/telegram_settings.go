package hub

import (
	"context"
	"net/http"
	"strings"

	"github.com/pocketbase/pocketbase/core"
)

func (h *Hub) getTelegramSettings(e *core.RequestEvent) error {
	settings, err := h.loadTelegramSettings()
	if err != nil {
		return err
	}
	return e.JSON(http.StatusOK, TelegramSettingsResponse{
		Enabled:        settings.Enabled,
		PollingEnabled: settings.PollingEnabled,
		BotUsername:    settings.BotUsername,
		HasToken:       strings.TrimSpace(settings.BotTokenEncrypted) != "",
		LastError:      settings.LastError,
		Updated:        settings.Updated,
	})
}

func (h *Hub) updateTelegramSettings(e *core.RequestEvent) error {
	input := TelegramSettingsInput{}
	if err := e.BindBody(&input); err != nil {
		return e.BadRequestError("invalid request body", err)
	}
	if err := validateTelegramSettingsInput(input); err != nil {
		return e.BadRequestError(err.Error(), nil)
	}
	current, err := h.loadTelegramSettings()
	if err != nil {
		return err
	}
	settings, err := h.saveTelegramSettings(input, current)
	if err != nil {
		return err
	}
	return e.JSON(http.StatusOK, TelegramSettingsResponse{
		Enabled:        settings.Enabled,
		PollingEnabled: settings.PollingEnabled,
		BotUsername:    settings.BotUsername,
		HasToken:       strings.TrimSpace(settings.BotTokenEncrypted) != "",
		LastError:      settings.LastError,
		Updated:        settings.Updated,
	})
}

func (h *Hub) testTelegramSettings(e *core.RequestEvent) error {
	input := TelegramSettingsTestInput{}
	if err := e.BindBody(&input); err != nil {
		return e.BadRequestError("invalid request body", err)
	}

	settings, err := h.loadTelegramSettings()
	if err != nil {
		return err
	}
	token := strings.TrimSpace(input.BotToken)
	if token == "" {
		token, err = h.decryptTelegramToken(settings)
		if err != nil {
			errorMessage := sanitizeTelegramError(err.Error()).Error()
			return e.JSON(http.StatusOK, telegramCredentialFailureResponse(errorMessage))
		}
	}
	if !telegramTokenPattern.MatchString(token) {
		return e.BadRequestError("invalid telegram bot token format", nil)
	}
	bot, err := h.telegramTransport.GetMe(context.Background(), token)
	if err != nil {
		errorMessage := sanitizeTelegramError(err.Error()).Error()
		return e.JSON(http.StatusOK, telegramCredentialFailureResponse(errorMessage))
	}
	if err := h.telegramTransport.SetMyCommands(context.Background(), token, telegramAdminBotCommands()); err != nil {
		errorMessage := sanitizeTelegramError(err.Error()).Error()
		return e.JSON(http.StatusOK, TelegramTestResponse{
			OK: false, BotUsername: bot.Username, Error: errorMessage,
			Stages: TelegramTestStages{
				Credentials: TelegramTestStage{OK: true},
				CommandMenu: TelegramTestStage{OK: false, Error: errorMessage},
			},
		})
	}
	_ = h.setTelegramSettingsRuntimeState(bot.Username, settings.LastPollOffset, "")
	return e.JSON(http.StatusOK, TelegramTestResponse{
		OK: true, BotUsername: bot.Username,
		Stages: TelegramTestStages{
			Credentials: TelegramTestStage{OK: true},
			CommandMenu: TelegramTestStage{OK: true},
		},
	})
}

func telegramCredentialFailureResponse(message string) TelegramTestResponse {
	return TelegramTestResponse{
		OK: false, Error: message,
		Stages: TelegramTestStages{
			Credentials: TelegramTestStage{OK: false, Error: message},
			CommandMenu: TelegramTestStage{OK: false, Error: "credentials were not verified"},
		},
	}
}
