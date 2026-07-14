//go:build testing

package hub

import (
	"github.com/pocketbase/pocketbase/core"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConfigBackupSchemaMetadataAndSectionSelection(t *testing.T) {
	sections, err := normalizeConfigBackupSections([]string{
		ConfigBackupSectionSystems,
		ConfigBackupSectionSystems,
		ConfigBackupSectionNotifications,
	})
	require.NoError(t, err)
	assert.Equal(t, []string{ConfigBackupSectionSystems, ConfigBackupSectionNotifications}, sections)

	document := newConfigBackupDocument(sections)
	assert.Equal(t, ConfigBackupVersion, document.Meta.BackupVersion)
	assert.Equal(t, ConfigBackupMode, document.Meta.Mode)
	assert.Equal(t, sections, document.Meta.Sections)
	for _, section := range sections {
		assert.Equal(t, configBackupSupportedSectionVersion(section), document.Meta.SectionVersions[section])
	}

	content, err := marshalConfigBackupDocument(document)
	require.NoError(t, err)
	assert.Contains(t, content, "sectionVersions:")

	_, err = normalizeConfigBackupSections([]string{"unknown"})
	require.Error(t, err)
}

func TestTelegramBackupCompatibilityFixturesParse(t *testing.T) {
	for _, test := range []struct {
		name    string
		version string
	}{
		{name: "notifications-v1.yml", version: "1"},
		{name: "notifications-v2.yml", version: ConfigBackupNotificationsVersion},
	} {
		content, err := os.ReadFile("testdata/telegram_backup/" + test.name)
		require.NoError(t, err)
		document, err := parseConfigBackupDocument(string(content))
		require.NoError(t, err)
		assert.Equal(t, test.version, document.Notifications.SectionVersion)
	}
}

func TestConfigBackupExportUsesConsistentCrossCollectionSnapshot(t *testing.T) {
	hub, admin := newTelegramHubWithAdmin(t)
	system := mustCreateTelegramRecord(t, hub, "systems", map[string]any{
		"name": "before-system", "host": "127.0.0.1", "port": "45876", "status": "up", "users": []string{admin.Id},
	})
	visibility := mustCreateTelegramRecord(t, hub, CollectionPublicSystemVisibility, map[string]any{
		"system": system.Id, "public_enabled": true, "public_name": "before-public",
	})

	firstRead := make(chan struct{})
	continueRead := make(chan struct{})
	writeDone := make(chan error, 1)
	var snapshotSystems []ConfigBackupSystem
	var snapshotPublic ConfigBackupPublicStatus
	var snapshotErr error
	var wait sync.WaitGroup
	wait.Add(1)
	go func() {
		defer wait.Done()
		snapshotErr = hub.RunInTransaction(func(txApp core.App) error {
			txHub := hub.configBackupTransactionHub(txApp)
			_, userMap, err := txHub.configBackupUsers()
			if err != nil {
				return err
			}
			snapshotSystems, err = txHub.configBackupSystems(userMap, configBackupExportOptions{})
			if err != nil {
				return err
			}
			close(firstRead)
			<-continueRead
			snapshotPublic, err = txHub.configBackupPublicStatus()
			return err
		})
	}()
	<-firstRead
	go func() {
		system.Set("name", "after-system")
		if err := hub.Save(system); err != nil {
			writeDone <- err
			return
		}
		visibility.Set("public_name", "after-public")
		writeDone <- hub.Save(visibility)
	}()
	select {
	case err := <-writeDone:
		require.NoError(t, err)
		t.Fatal("concurrent write committed before export transaction completed")
	case <-time.After(50 * time.Millisecond):
	}
	close(continueRead)
	wait.Wait()
	require.NoError(t, snapshotErr)
	require.NoError(t, <-writeDone)
	require.Len(t, snapshotSystems, 1)
	require.Len(t, snapshotPublic.Systems, 1)
	assert.Equal(t, "before-system", snapshotSystems[0].Name)
	assert.Equal(t, "before-public", snapshotPublic.Systems[0].PublicName)
}

func TestParseConfigBackupDocumentPreservesLegacyDocumentsWithoutSectionVersions(t *testing.T) {
	content := strings.TrimSpace(`
meta:
  backupVersion: "1"
  sourceVersion: test
  createdAt: "2026-07-10T00:00:00Z"
  mode: merge
  sections:
    - systems
systems: []
`) + "\n"

	document, err := parseConfigBackupDocument(content)
	require.NoError(t, err)
	assert.Equal(t, ConfigBackupSectionVersion, document.Meta.SectionVersions[ConfigBackupSectionSystems])
}

func TestParseConfigBackupDocumentWarnsAboutUnknownFieldsInKnownSections(t *testing.T) {
	content := strings.TrimSpace(`
meta:
  backupVersion: "1"
  sourceVersion: test
  createdAt: "2026-07-10T00:00:00Z"
  mode: merge
  sections: [notifications]
notifications:
  telegram:
    settings:
      enabled: true
      pollingEnabledd: true
`) + "\n"
	document, err := parseConfigBackupDocument(content)
	require.NoError(t, err)
	assert.Contains(t, strings.Join(document.CompatibilityWarnings, "\n"), "pollingEnabledd")
}
