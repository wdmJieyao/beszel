package hub

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/henrygd/beszel"
	"gopkg.in/yaml.v3"
)

func newConfigBackupDocument(sections []string) ConfigBackupDocument {
	now := time.Now().UTC().Format(time.RFC3339)
	return ConfigBackupDocument{
		Meta: ConfigBackupMeta{
			BackupVersion:   ConfigBackupVersion,
			SourceVersion:   beszel.Version,
			CreatedAt:       now,
			Mode:            ConfigBackupMode,
			Sections:        sections,
			SectionVersions: configBackupSectionVersions(sections),
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
	var raw map[string]yaml.Node
	if err := yaml.Unmarshal([]byte(content), &raw); err != nil {
		return ConfigBackupDocument{}, err
	}
	var document ConfigBackupDocument
	strictDecoder := yaml.NewDecoder(strings.NewReader(content))
	strictDecoder.KnownFields(true)
	var strictDocument ConfigBackupDocument
	if err := strictDecoder.Decode(&strictDocument); err != nil {
		document.CompatibilityWarnings = append(document.CompatibilityWarnings, "unknown or incompatible YAML fields: "+err.Error())
	}
	warnings := document.CompatibilityWarnings
	if err := yaml.Unmarshal([]byte(content), &document); err != nil {
		return ConfigBackupDocument{}, err
	}
	document.CompatibilityWarnings = warnings
	if settingsNode := configBackupMappingValue(raw[ConfigBackupSectionNotifications], "telegram", "settings"); settingsNode != nil {
		document.Notifications.Telegram.Settings.Present = true
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
	if len(document.Meta.Sections) == 0 {
		document.Meta.Sections = configBackupPresentSections(raw)
	}
	if document.Meta.SectionVersions == nil {
		document.Meta.SectionVersions = map[string]string{}
	}
	document.SkippedSections = map[string]string{}
	known := configBackupKnownSectionSet()
	for key := range raw {
		if key == "meta" || key == "encryption" || key == "users" {
			continue
		}
		if _, ok := known[key]; !ok {
			document.UnknownSections = append(document.UnknownSections, key)
			document.SkippedSections[key] = "unknown backup section"
		}
	}
	for _, section := range document.Meta.Sections {
		if _, ok := known[section]; !ok {
			if _, exists := document.SkippedSections[section]; !exists {
				document.UnknownSections = append(document.UnknownSections, section)
				document.SkippedSections[section] = "unknown backup section"
			}
			continue
		}
		version := strings.TrimSpace(document.Meta.SectionVersions[section])
		if version == "" {
			version = ConfigBackupSectionVersion
			document.Meta.SectionVersions[section] = version
		}
		supportedVersion := configBackupSupportedSectionVersion(section)
		if configBackupVersionNewer(version, supportedVersion) {
			document.SkippedSections[section] = fmt.Sprintf("section version %s is newer than supported version %s", version, supportedVersion)
		}
		if section == ConfigBackupSectionNotifications {
			document.Notifications.SectionVersion = version
		}
	}
	return document, nil
}

func configBackupMappingValue(root yaml.Node, keys ...string) *yaml.Node {
	node := &root
	for _, key := range keys {
		if node.Kind == yaml.DocumentNode && len(node.Content) > 0 {
			node = node.Content[0]
		}
		if node.Kind != yaml.MappingNode {
			return nil
		}
		var next *yaml.Node
		for i := 0; i+1 < len(node.Content); i += 2 {
			if node.Content[i].Value == key {
				next = node.Content[i+1]
				break
			}
		}
		if next == nil {
			return nil
		}
		node = next
	}
	return node
}

func configBackupSectionVersions(sections []string) map[string]string {
	versions := make(map[string]string, len(sections))
	for _, section := range sections {
		versions[section] = configBackupSupportedSectionVersion(section)
	}
	return versions
}

func configBackupSupportedSectionVersion(section string) string {
	if section == ConfigBackupSectionNotifications {
		return ConfigBackupNotificationsVersion
	}
	return ConfigBackupSectionVersion
}

func configBackupKnownSectionSet() map[string]struct{} {
	known := make(map[string]struct{}, len(defaultConfigBackupSections))
	for _, section := range defaultConfigBackupSections {
		known[section] = struct{}{}
	}
	return known
}

func configBackupPresentSections(raw map[string]yaml.Node) []string {
	sections := make([]string, 0, len(defaultConfigBackupSections))
	for _, section := range defaultConfigBackupSections {
		if _, ok := raw[section]; ok {
			sections = append(sections, section)
		}
	}
	return sections
}

func configBackupVersionNewer(candidate string, supported string) bool {
	candidateNumber, candidateErr := strconv.Atoi(candidate)
	supportedNumber, supportedErr := strconv.Atoi(supported)
	if candidateErr == nil && supportedErr == nil {
		return candidateNumber > supportedNumber
	}
	return candidate != supported
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
