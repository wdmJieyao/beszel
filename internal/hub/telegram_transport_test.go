//go:build testing

package hub

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestTelegramTransportAcceptsScalarBooleanResultWhenIgnored(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		response.Header().Set("Content-Type", "application/json")
		_, _ = response.Write([]byte(`{"ok":true,"result":true}`))
	}))
	defer server.Close()

	transport := newTelegramHTTPTransport(server.Client()).(*telegramHTTPTransport)
	transport.baseURL = server.URL
	require.NoError(t, transport.SetMyCommands(context.Background(), "123456:valid_token_value", telegramAdminBotCommands()))
}
