//go:build testing

package hub

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDefaultTelegramHTTPTimeoutExceedsLongPollingWindow(t *testing.T) {
	transport, ok := newTelegramHTTPTransport(nil).(*telegramHTTPTransport)
	require.True(t, ok)
	minimum := time.Duration(telegramPollingTimeoutSeconds+5) * time.Second
	assert.GreaterOrEqual(t, transport.client.Timeout, minimum)
}
