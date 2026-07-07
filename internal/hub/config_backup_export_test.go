//go:build testing

package hub

import (
	"testing"

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

	_, err = normalizeConfigBackupSections([]string{"unknown"})
	require.Error(t, err)
}
