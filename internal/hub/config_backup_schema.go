package hub

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"slices"
	"time"

	"github.com/henrygd/beszel"
	"gopkg.in/yaml.v3"
)

func newConfigBackupDocument(sections []string) ConfigBackupDocument {
	now := time.Now().UTC().Format(time.RFC3339)
	return ConfigBackupDocument{
		Meta: ConfigBackupMeta{
			BackupVersion: ConfigBackupVersion,
			SourceVersion: beszel.Version,
			CreatedAt:     now,
			Mode:          ConfigBackupMode,
			Sections:      sections,
		},
		Encryption: ConfigBackupEncryption{
			Enabled:   false,
			Algorithm: configBackupCryptoAlgorithm,
			KDF:       configBackupCryptoKDF,
		},
	}
}

func marshalConfigBackupDocument(document ConfigBackupDocument) (string, error) {
	data, err := yaml.Marshal(&document)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func parseConfigBackupDocument(content string) (ConfigBackupDocument, error) {
	var document ConfigBackupDocument
	if err := yaml.Unmarshal([]byte(content), &document); err != nil {
		return ConfigBackupDocument{}, err
	}
	if document.Meta.BackupVersion != ConfigBackupVersion {
		return ConfigBackupDocument{}, fmt.Errorf("unsupported backup version")
	}
	if document.Meta.Mode != "" && document.Meta.Mode != ConfigBackupMode {
		return ConfigBackupDocument{}, fmt.Errorf("unsupported restore mode")
	}
	if document.Meta.Mode == "" {
		document.Meta.Mode = ConfigBackupMode
	}
	return document, nil
}

func normalizeConfigBackupSections(sections []string) ([]string, error) {
	if len(sections) == 0 {
		return slices.Clone(defaultConfigBackupSections), nil
	}
	allowed := map[string]struct{}{
		ConfigBackupSectionSystems:       {},
		ConfigBackupSectionAlerts:        {},
		ConfigBackupSectionNotifications: {},
		ConfigBackupSectionPublicStatus:  {},
		ConfigBackupSectionNetworkProbes: {},
	}
	normalized := make([]string, 0, len(sections))
	seen := make(map[string]struct{}, len(sections))
	for _, section := range sections {
		if _, ok := allowed[section]; !ok {
			return nil, fmt.Errorf("unsupported backup section: %s", section)
		}
		if _, ok := seen[section]; ok {
			continue
		}
		seen[section] = struct{}{}
		normalized = append(normalized, section)
	}
	return normalized, nil
}

func configBackupIncludesSection(sections []string, section string) bool {
	return slices.Contains(sections, section)
}

func configBackupPreviewID(content string, credential string) string {
	sum := sha256.Sum256([]byte(content + "\x00" + credential))
	return "sha256:" + hex.EncodeToString(sum[:])
}
