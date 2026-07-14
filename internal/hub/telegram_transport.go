package hub

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"
	"unicode/utf8"
)

const (
	telegramHTTPTimeout            = time.Duration(telegramPollingTimeoutSeconds+10) * time.Second
	telegramMessageMaxRunes        = 4096
	telegramMessageTruncatedSuffix = "\n…内容已截断"
)

func truncateTelegramMessage(message string) string {
	if utf8.RuneCountInString(message) <= telegramMessageMaxRunes {
		return message
	}
	suffix := []rune(telegramMessageTruncatedSuffix)
	limit := telegramMessageMaxRunes - len(suffix)
	return string([]rune(message)[:limit]) + telegramMessageTruncatedSuffix
}

type telegramBotCommand struct {
	Command     string `json:"command"`
	Description string `json:"description"`
}

type TelegramTransport interface {
	GetMe(ctx context.Context, botToken string) (*TelegramBotIdentity, error)
	SetMyCommands(ctx context.Context, botToken string, commands []telegramBotCommand) error
	SendMessage(ctx context.Context, botToken string, chatID string, text string, options *TelegramSendOptions) error
	GetUpdates(ctx context.Context, botToken string, offset int64, timeoutSeconds int) ([]TelegramUpdate, error)
	AnswerCallbackQuery(ctx context.Context, botToken string, callbackQueryID string, text string) error
}

type telegramHTTPTransport struct {
	client  *http.Client
	baseURL string
}

type telegramSendOptions struct {
	ReplyMarkup any
}

type telegramBotIdentity struct {
	ID       int64  `json:"id"`
	Username string `json:"username"`
}

type telegramChat struct {
	ID   int64  `json:"id"`
	Type string `json:"type"`
}

type telegramMessage struct {
	Text string       `json:"text"`
	Chat telegramChat `json:"chat"`
}

type telegramCallbackQuery struct {
	ID      string           `json:"id"`
	Data    string           `json:"data"`
	Message *telegramMessage `json:"message"`
}

type telegramUpdate struct {
	UpdateID      int64                  `json:"update_id"`
	Message       *telegramMessage       `json:"message,omitempty"`
	CallbackQuery *telegramCallbackQuery `json:"callback_query,omitempty"`
}

type telegramAPIResponse[T any] struct {
	OK          bool   `json:"ok"`
	Result      T      `json:"result"`
	Description string `json:"description"`
}

type TelegramSendOptions = telegramSendOptions
type TelegramBotIdentity = telegramBotIdentity
type TelegramUpdate = telegramUpdate

func newTelegramHTTPTransport(client *http.Client) TelegramTransport {
	if client == nil {
		client = &http.Client{Timeout: telegramHTTPTimeout}
	}
	return &telegramHTTPTransport{
		client:  client,
		baseURL: "https://api.telegram.org",
	}
}

func (t *telegramHTTPTransport) GetMe(ctx context.Context, botToken string) (*TelegramBotIdentity, error) {
	var response telegramAPIResponse[telegramBotIdentity]
	if err := t.doJSON(ctx, botToken, http.MethodGet, "getMe", nil, &response); err != nil {
		return nil, err
	}
	return &response.Result, nil
}

func (t *telegramHTTPTransport) SetMyCommands(ctx context.Context, botToken string, commands []telegramBotCommand) error {
	body := map[string]any{
		"commands": commands,
		"scope":    map[string]string{"type": "all_private_chats"},
	}
	return t.doJSON(ctx, botToken, http.MethodPost, "setMyCommands", body, nil)
}

func (t *telegramHTTPTransport) SendMessage(ctx context.Context, botToken string, chatID string, text string, options *TelegramSendOptions) error {
	text = truncateTelegramMessage(text)
	body := map[string]any{
		"chat_id":                  chatID,
		"text":                     text,
		"disable_web_page_preview": true,
	}
	if options != nil && options.ReplyMarkup != nil {
		body["reply_markup"] = options.ReplyMarkup
	}
	return t.doJSON(ctx, botToken, http.MethodPost, "sendMessage", body, nil)
}

func (t *telegramHTTPTransport) GetUpdates(ctx context.Context, botToken string, offset int64, timeoutSeconds int) ([]TelegramUpdate, error) {
	body := map[string]any{
		"offset":          offset,
		"timeout":         timeoutSeconds,
		"allowed_updates": []string{"message", "callback_query"},
	}
	var response telegramAPIResponse[[]telegramUpdate]
	if err := t.doJSON(ctx, botToken, http.MethodPost, "getUpdates", body, &response); err != nil {
		return nil, err
	}
	return response.Result, nil
}

func (t *telegramHTTPTransport) AnswerCallbackQuery(ctx context.Context, botToken string, callbackQueryID string, text string) error {
	body := map[string]any{
		"callback_query_id": callbackQueryID,
		"text":              text,
	}
	return t.doJSON(ctx, botToken, http.MethodPost, "answerCallbackQuery", body, nil)
}

func (t *telegramHTTPTransport) doJSON(ctx context.Context, botToken string, method string, apiMethod string, payload any, result any) error {
	endpoint := fmt.Sprintf("%s/bot%s/%s", strings.TrimRight(t.baseURL, "/"), url.PathEscape(botToken), apiMethod)
	var body *bytes.Reader
	if payload != nil {
		data, err := json.Marshal(payload)
		if err != nil {
			return err
		}
		body = bytes.NewReader(data)
	} else {
		body = bytes.NewReader(nil)
	}
	req, err := http.NewRequestWithContext(ctx, method, endpoint, body)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := t.client.Do(req)
	if err != nil {
		return sanitizeTelegramError(err.Error())
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if result == nil {
		var apiResp telegramAPIResponse[json.RawMessage]
		if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
			return sanitizeTelegramError(err.Error())
		}
		if !apiResp.OK {
			return sanitizeTelegramError(apiResp.Description)
		}
		return nil
	}

	if err := json.NewDecoder(resp.Body).Decode(result); err != nil {
		return sanitizeTelegramError(err.Error())
	}
	switch typed := result.(type) {
	case *telegramAPIResponse[telegramBotIdentity]:
		if !typed.OK {
			return sanitizeTelegramError(typed.Description)
		}
	case *telegramAPIResponse[[]telegramUpdate]:
		if !typed.OK {
			return sanitizeTelegramError(typed.Description)
		}
	default:
	}
	return nil
}

func sanitizeTelegramError(message string) error {
	message = strings.TrimSpace(message)
	message = strings.ReplaceAll(message, "\n", " ")
	message = regexp.MustCompile(`\d+:[A-Za-z0-9_-]{10,}`).ReplaceAllString(message, "[telegram token hidden]")
	switch {
	case message == "":
		return fmt.Errorf("telegram request failed")
	case strings.Contains(strings.ToLower(message), "chat not found"):
		return fmt.Errorf("chat not found")
	case strings.Contains(strings.ToLower(message), "bot was blocked"):
		return fmt.Errorf("bot was blocked by the target chat")
	case strings.Contains(strings.ToLower(message), "forbidden"):
		return fmt.Errorf("bot cannot post to this chat")
	default:
		return errors.New(message)
	}
}
