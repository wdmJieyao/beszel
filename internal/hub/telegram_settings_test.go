//go:build testing

package hub

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/pocketbase/pocketbase/core"
	pbTests "github.com/pocketbase/pocketbase/tests"
	"github.com/pocketbase/pocketbase/tools/router"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUpdateTelegramSettingsStoresEncryptedTokenAndRedactsResponse(t *testing.T) {
	hub, admin := newTelegramHubWithAdmin(t)
	fake := &fakeTelegramTransport{
		bot: &telegramBotIdentity{ID: 1, Username: "beszel_bot"},
	}
	hub.telegramTransport = fake

	body := bytes.NewBufferString(`{"enabled":true,"pollingEnabled":true,"botToken":"123456:abcde_token_valid"}`)
	req := httptest.NewRequest(http.MethodPut, "/api/beszel/telegram/settings", body)
	req.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()

	require.NoError(t, hub.updateTelegramSettings(&core.RequestEvent{
		App:  hub,
		Auth: admin,
		Event: router.Event{
			Request:  req,
			Response: recorder,
		},
	}))

	assert.NotContains(t, recorder.Body.String(), "123456:abcde_token_valid")

	settingsRecord, err := hub.getTelegramSettingsRecord()
	require.NoError(t, err)
	require.NotNil(t, settingsRecord)
	assert.NotEqual(t, "123456:abcde_token_valid", settingsRecord.GetString("bot_token_encrypted"))

	settings, err := hub.loadTelegramSettings()
	require.NoError(t, err)
	token, err := hub.decryptTelegramToken(settings)
	require.NoError(t, err)
	assert.Equal(t, "123456:abcde_token_valid", token)

	getReq := httptest.NewRequest(http.MethodGet, "/api/beszel/telegram/settings", nil)
	getRecorder := httptest.NewRecorder()
	require.NoError(t, hub.getTelegramSettings(&core.RequestEvent{
		App:  hub,
		Auth: admin,
		Event: router.Event{
			Request:  getReq,
			Response: getRecorder,
		},
	}))
	assert.Contains(t, getRecorder.Body.String(), `"hasToken":true`)
	assert.NotContains(t, getRecorder.Body.String(), "123456:abcde_token_valid")
}

func TestTestTelegramSettingsRejectsInvalidToken(t *testing.T) {
	hub, admin := newTelegramHubWithAdmin(t)

	body := bytes.NewBufferString(`{"botToken":"bad-token"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/beszel/telegram/settings/test", body)
	req.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()

	err := hub.testTelegramSettings(&core.RequestEvent{
		App:  hub,
		Auth: admin,
		Event: router.Event{
			Request:  req,
			Response: recorder,
		},
	})
	require.Error(t, err)
	apiErr := router.ToApiError(err)
	assert.Equal(t, http.StatusBadRequest, apiErr.Status)
}

func TestTestTelegramSettingsRejectsMalformedRequestBody(t *testing.T) {
	hub, admin := newTelegramHubWithAdmin(t)
	_, err := hub.saveTelegramSettings(TelegramSettingsInput{
		Enabled: true, BotToken: "123456:saved_token_value",
	}, telegramSettingsRecord{})
	require.NoError(t, err)
	fake := &fakeTelegramTransport{}
	hub.telegramTransport = fake

	req := httptest.NewRequest(http.MethodPost, "/api/beszel/telegram/settings/test", bytes.NewBufferString(`{"botToken":`))
	req.Header.Set("Content-Type", "application/json")
	err = hub.testTelegramSettings(&core.RequestEvent{
		App: hub, Auth: admin,
		Event: router.Event{Request: req, Response: httptest.NewRecorder()},
	})

	require.Error(t, err)
	assert.Equal(t, http.StatusBadRequest, router.ToApiError(err).Status)
	assert.Empty(t, fake.tokens)
}

func TestTestTelegramSettingsUsesTransportAndPersistsBotUsername(t *testing.T) {
	hub, admin := newTelegramHubWithAdmin(t)
	hub.telegramTransport = &fakeTelegramTransport{
		bot: &telegramBotIdentity{ID: 99, Username: "beszel_alert_bot"},
	}
	_, err := hub.saveTelegramSettings(TelegramSettingsInput{
		Enabled:        true,
		PollingEnabled: true,
		BotToken:       "123456:abcde_token_valid",
	}, telegramSettingsRecord{})
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost, "/api/beszel/telegram/settings/test", nil)
	recorder := httptest.NewRecorder()
	require.NoError(t, hub.testTelegramSettings(&core.RequestEvent{
		App:  hub,
		Auth: admin,
		Event: router.Event{
			Request:  req,
			Response: recorder,
		},
	}))

	var response TelegramTestResponse
	require.NoError(t, json.Unmarshal(recorder.Body.Bytes(), &response))
	assert.True(t, response.OK)
	assert.Equal(t, "beszel_alert_bot", response.BotUsername)
	require.NotEmpty(t, hub.telegramTransport.(*fakeTelegramTransport).commands)
	assert.Equal(t, "systems", hub.telegramTransport.(*fakeTelegramTransport).commands[1].Command)
	assert.Contains(t, hub.telegramTransport.(*fakeTelegramTransport).commands[1].Description, "节点")

	settings, err := hub.loadTelegramSettings()
	require.NoError(t, err)
	assert.Equal(t, "beszel_alert_bot", settings.BotUsername)
}

func TestTestTelegramSettingsUsesEnteredTokenWithoutSavingAndReturnsStages(t *testing.T) {
	hub, admin := newTelegramHubWithAdmin(t)
	fake := &fakeTelegramTransport{bot: &telegramBotIdentity{ID: 99, Username: "staged_bot"}}
	hub.telegramTransport = fake
	_, err := hub.saveTelegramSettings(TelegramSettingsInput{Enabled: true, BotToken: "123456:saved_token_value"}, telegramSettingsRecord{})
	require.NoError(t, err)

	body := bytes.NewBufferString(`{"botToken":"654321:entered_token_value"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/beszel/telegram/settings/test", body)
	req.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()
	require.NoError(t, hub.testTelegramSettings(&core.RequestEvent{App: hub, Auth: admin, Event: router.Event{Request: req, Response: recorder}}))

	var response TelegramTestResponse
	require.NoError(t, json.Unmarshal(recorder.Body.Bytes(), &response))
	assert.True(t, response.OK)
	assert.True(t, response.Stages.Credentials.OK)
	assert.True(t, response.Stages.CommandMenu.OK)
	assert.Equal(t, []string{"654321:entered_token_value", "654321:entered_token_value"}, fake.tokens)
	settings, err := hub.loadTelegramSettings()
	require.NoError(t, err)
	savedToken, err := hub.decryptTelegramToken(settings)
	require.NoError(t, err)
	assert.Equal(t, "123456:saved_token_value", savedToken)
}

func TestTestTelegramSettingsReportsMenuFailureSeparatelyAndSanitizesToken(t *testing.T) {
	hub, admin := newTelegramHubWithAdmin(t)
	token := "654321:entered_secret_token"
	hub.telegramTransport = &fakeTelegramTransport{
		bot:         &telegramBotIdentity{ID: 99, Username: "staged_bot"},
		commandsErr: errors.New("request /bot" + token + "/setMyCommands failed"),
	}
	body := bytes.NewBufferString(`{"botToken":"` + token + `"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/beszel/telegram/settings/test", body)
	req.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()
	require.NoError(t, hub.testTelegramSettings(&core.RequestEvent{App: hub, Auth: admin, Event: router.Event{Request: req, Response: recorder}}))

	var response TelegramTestResponse
	require.NoError(t, json.Unmarshal(recorder.Body.Bytes(), &response))
	assert.False(t, response.OK)
	assert.True(t, response.Stages.Credentials.OK)
	assert.False(t, response.Stages.CommandMenu.OK)
	assert.NotEmpty(t, response.Stages.CommandMenu.Error)
	assert.NotContains(t, recorder.Body.String(), token)
}

func newTelegramHubWithAdmin(t *testing.T) (*Hub, *core.Record) {
	testApp, err := pbTests.NewTestAppWithConfig(core.BaseAppConfig{
		DataDir:       t.TempDir(),
		EncryptionEnv: "pb_test_env",
	})
	require.NoError(t, err)
	hub := NewHub(testApp)
	admin := mustCreateTelegramRecord(t, hub, "users", map[string]any{
		"email":    "admin@example.com",
		"password": "password123",
		"role":     "admin",
	})
	t.Cleanup(func() {
		hub.Stop()
		hub.sm.RemoveAllSystems()
		testApp.Cleanup()
	})
	return hub, admin
}

type fakeTelegramTransport struct {
	bot         *telegramBotIdentity
	commands    []telegramBotCommand
	sendCalls   []fakeTelegramMessage
	getUpdates  []telegramUpdate
	sendErr     error
	getMeErr    error
	commandsErr error
	tokens      []string
	sendStarted chan struct{}
	sendRelease chan struct{}
}

type fakeTelegramMessage struct {
	ChatID  string
	Text    string
	Options *telegramSendOptions
}

func (f *fakeTelegramTransport) GetMe(_ context.Context, token string) (*telegramBotIdentity, error) {
	f.tokens = append(f.tokens, token)
	if f.getMeErr != nil {
		return nil, f.getMeErr
	}
	if f.bot == nil {
		return &telegramBotIdentity{ID: 1, Username: "bot"}, nil
	}
	return f.bot, nil
}

func (f *fakeTelegramTransport) SendMessage(_ context.Context, _ string, chatID string, text string, options *telegramSendOptions) error {
	f.sendCalls = append(f.sendCalls, fakeTelegramMessage{ChatID: chatID, Text: text, Options: options})
	if f.sendStarted != nil {
		close(f.sendStarted)
	}
	if f.sendRelease != nil {
		<-f.sendRelease
	}
	return f.sendErr
}

func (f *fakeTelegramTransport) SetMyCommands(_ context.Context, token string, commands []telegramBotCommand) error {
	f.tokens = append(f.tokens, token)
	f.commands = append([]telegramBotCommand(nil), commands...)
	return f.commandsErr
}

func (f *fakeTelegramTransport) GetUpdates(_ context.Context, _ string, _ int64, _ int) ([]telegramUpdate, error) {
	return f.getUpdates, nil
}

func (f *fakeTelegramTransport) AnswerCallbackQuery(_ context.Context, _ string, _ string, _ string) error {
	return nil
}

func mustCreateTelegramRecord(t *testing.T, app core.App, collectionName string, fields map[string]any) *core.Record {
	collection, err := app.FindCachedCollectionByNameOrId(collectionName)
	require.NoError(t, err)
	record := core.NewRecord(collection)
	record.Load(fields)
	require.NoError(t, app.Save(record))
	return record
}
