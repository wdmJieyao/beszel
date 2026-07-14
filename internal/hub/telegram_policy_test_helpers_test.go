//go:build testing

package hub

import (
	"fmt"
	"testing"

	"github.com/pocketbase/pocketbase/core"
)

func mustCreateTelegramTestSystems(t *testing.T, app core.App, userID string, count int) []*core.Record {
	t.Helper()
	systems := make([]*core.Record, 0, count)
	for index := 0; index < count; index++ {
		systems = append(systems, mustCreateTelegramRecord(t, app, "systems", map[string]any{
			"name":   fmt.Sprintf("node-%03d", index+1),
			"host":   fmt.Sprintf("127.0.1.%d", index+1),
			"port":   "45876",
			"status": "up",
			"users":  []string{userID},
		}))
	}
	return systems
}

func telegramRecordIDs(records []*core.Record) []string {
	ids := make([]string, 0, len(records))
	for _, record := range records {
		ids = append(ids, record.Id)
	}
	return ids
}
