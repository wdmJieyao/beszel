//go:build testing

package hub

import (
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTelegramAndBackupRoutesDoNotChangeAgentConnectRoute(t *testing.T) {
	source, err := os.ReadFile("api.go")
	require.NoError(t, err)
	text := string(source)

	assert.Contains(t, text, `apiNoAuth.GET("/agent-connect", h.handleAgentConnect)`)
	assert.Contains(t, text, `apiAuth.GET("/telegram/settings", h.getTelegramSettings)`)
	assert.Contains(t, text, `apiAuth.POST("/config-backups/exports", h.exportConfigBackup)`)
	assert.Less(t, strings.Index(text, `apiAuth.POST("/config-backups/exports"`), strings.Index(text, `apiNoAuth.GET("/agent-connect"`))
}
