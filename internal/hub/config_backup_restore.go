package hub

import (
	"fmt"
	"strings"

	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/tools/types"
)

const (
	configBackupActionCreate   = "create"
	configBackupActionUpdate   = "update"
	configBackupActionPreserve = "preserve"
	configBackupActionSkip     = "skip"
	configBackupActionConflict = "conflict"
	configBackupActionError    = "error"
)

func (h *Hub) previewConfigBackup(document ConfigBackupDocument, content string, credential string) (ConfigBackupPreviewResponse, error) {
	preview := ConfigBackupPreviewResponse{
		PreviewID:          configBackupPreviewID(content, credential),
		Mode:               ConfigBackupMode,
		BackupMeta:         document.Meta,
		Items:              []ConfigBackupPreviewItem{},
		Warnings:           append([]string{}, document.CompatibilityWarnings...),
		RequiresCredential: configBackupHasEnabledEncryptedSecrets(document),
	}
	for section, reason := range document.SkippedSections {
		preview.Warnings = append(preview.Warnings, fmt.Sprintf("%s skipped: %s", section, reason))
		preview.addConfigBackupPreviewItem(section, "", section, configBackupActionSkip, reason)
	}
	if preview.RequiresCredential && credential == "" {
		preview.addConfigBackupPreviewItem(ConfigBackupSectionNotifications, "", "encrypted secrets", configBackupActionError, "decryption credential is required")
		return preview, nil
	}
	if _, err := decryptDocumentSecretsForValidation(document, credential); err != nil {
		preview.addConfigBackupPreviewItem(ConfigBackupSectionNotifications, "", "encrypted secrets", configBackupActionError, err.Error())
		return preview, nil
	}

	emailToUserID, err := h.userIDByEmailMap()
	if err != nil {
		return preview, err
	}
	backupSystemIDs := configBackupSystemIDSet(document.Systems)
	if document.configBackupSectionEnabled(ConfigBackupSectionSystems) {
		h.previewSystems(document.Systems, emailToUserID, &preview)
	}
	if document.configBackupSectionEnabled(ConfigBackupSectionAlerts) {
		h.previewAlerts(document.Alerts, emailToUserID, backupSystemIDs, &preview)
	}
	if document.configBackupSectionEnabled(ConfigBackupSectionNotifications) {
		h.previewNotifications(document.Notifications, emailToUserID, backupSystemIDs, &preview)
	}
	if document.configBackupSectionEnabled(ConfigBackupSectionPublicStatus) {
		h.previewPublicStatus(document.PublicStatus, backupSystemIDs, &preview)
	}
	if document.configBackupSectionEnabled(ConfigBackupSectionNetworkProbes) {
		h.previewNetworkProbes(document.NetworkProbes, backupSystemIDs, &preview)
	}
	return preview, nil
}

func (h *Hub) previewSystems(items []ConfigBackupSystem, emailToUserID map[string]string, preview *ConfigBackupPreviewResponse) {
	existing, err := h.existingIDs("systems")
	if err != nil {
		preview.addConfigBackupPreviewItem(ConfigBackupSectionSystems, "", "systems", configBackupActionError, err.Error())
		return
	}
	seen := map[string]struct{}{}
	for _, item := range items {
		seen[item.StableID] = struct{}{}
		_, missing := configBackupEmailRefsToUserIDs(item.Users, emailToUserID)
		if len(missing) > 0 {
			preview.addConfigBackupPreviewItem(ConfigBackupSectionSystems, item.StableID, item.Name, configBackupActionConflict, "missing user: "+strings.Join(missing, ", "))
			continue
		}
		if _, ok := existing[item.StableID]; ok {
			preview.addConfigBackupPreviewItem(ConfigBackupSectionSystems, item.StableID, item.Name, configBackupActionUpdate, "stable identifier matched existing system")
		} else {
			preview.addConfigBackupPreviewItem(ConfigBackupSectionSystems, item.StableID, item.Name, configBackupActionCreate, "system does not exist")
		}
	}
	for id, record := range existing {
		if _, ok := seen[id]; !ok {
			preview.addConfigBackupPreviewItem(ConfigBackupSectionSystems, id, record.GetString("name"), configBackupActionPreserve, "target-only system will be preserved")
		}
	}
}

func (h *Hub) previewAlerts(alerts ConfigBackupAlerts, emailToUserID map[string]string, backupSystemIDs map[string]struct{}, preview *ConfigBackupPreviewResponse) {
	seenAlerts := make(map[string]struct{}, len(alerts.Definitions))
	for _, item := range alerts.Definitions {
		seenAlerts[item.StableID] = struct{}{}
		if emailToUserID[strings.ToLower(item.UserEmail)] == "" {
			preview.addConfigBackupPreviewItem(ConfigBackupSectionAlerts, item.StableID, item.Name, configBackupActionConflict, "missing alert user")
			continue
		}
		if !h.configBackupSystemExists(item.SystemStableID, backupSystemIDs) {
			preview.addConfigBackupPreviewItem(ConfigBackupSectionAlerts, item.StableID, item.Name, configBackupActionConflict, "missing alert system")
			continue
		}
		if _, err := h.FindRecordById("alerts", item.StableID); err == nil {
			preview.addConfigBackupPreviewItem(ConfigBackupSectionAlerts, item.StableID, item.Name, configBackupActionUpdate, "stable identifier matched existing alert")
		} else {
			preview.addConfigBackupPreviewItem(ConfigBackupSectionAlerts, item.StableID, item.Name, configBackupActionCreate, "alert does not exist")
		}
	}
	seenQuietHours := make(map[string]struct{}, len(alerts.QuietHours))
	for _, item := range alerts.QuietHours {
		seenQuietHours[item.StableID] = struct{}{}
		if emailToUserID[strings.ToLower(item.UserEmail)] == "" {
			preview.addConfigBackupPreviewItem(ConfigBackupSectionAlerts, item.StableID, item.Type, configBackupActionConflict, "missing quiet-hours user")
			continue
		}
		if item.SystemStableID != "" && !h.configBackupSystemExists(item.SystemStableID, backupSystemIDs) {
			preview.addConfigBackupPreviewItem(ConfigBackupSectionAlerts, item.StableID, item.Type, configBackupActionConflict, "missing quiet-hours system")
			continue
		}
		if _, err := h.FindRecordById("quiet_hours", item.StableID); err == nil {
			preview.addConfigBackupPreviewItem(ConfigBackupSectionAlerts, item.StableID, item.Type, configBackupActionUpdate, "stable identifier matched existing quiet-hours window")
		} else {
			preview.addConfigBackupPreviewItem(ConfigBackupSectionAlerts, item.StableID, item.Type, configBackupActionCreate, "quiet-hours window does not exist")
		}
	}
	h.previewTargetOnlyIDs(ConfigBackupSectionAlerts, "alerts", seenAlerts, "name", "target-only alert will be preserved", preview)
	h.previewTargetOnlyIDs(ConfigBackupSectionAlerts, "quiet_hours", seenQuietHours, "type", "target-only quiet-hours window will be preserved", preview)
}

func (h *Hub) previewNotifications(notifications ConfigBackupNotifications, emailToUserID map[string]string, backupSystemIDs map[string]struct{}, preview *ConfigBackupPreviewResponse) {
	seenUsers := make(map[string]struct{}, len(notifications.UserSettings))
	for _, item := range notifications.UserSettings {
		email := strings.ToLower(item.UserEmail)
		seenUsers[email] = struct{}{}
		if emailToUserID[email] == "" {
			preview.addConfigBackupPreviewItem(ConfigBackupSectionNotifications, item.UserEmail, item.UserEmail, configBackupActionConflict, "missing notification user")
			continue
		}
		preview.addConfigBackupPreviewItem(ConfigBackupSectionNotifications, item.UserEmail, item.UserEmail, configBackupActionUpdate, "user notification settings will be merged")
	}
	if configBackupHasTelegramSettings(notifications.Telegram.Settings) {
		preview.addConfigBackupPreviewItem(ConfigBackupSectionNotifications, "telegram_settings", "Telegram settings", configBackupActionUpdate, "telegram settings will be merged")
	}
	seenDestinations := make(map[string]struct{}, len(notifications.Telegram.Destinations))
	seenChatIDs := make(map[string]string, len(notifications.Telegram.Destinations))
	for _, item := range notifications.Telegram.Destinations {
		seenDestinations[item.StableID] = struct{}{}
		if item.UserEmail != "" && emailToUserID[strings.ToLower(item.UserEmail)] == "" {
			preview.addConfigBackupPreviewItem(ConfigBackupSectionNotifications, item.StableID, item.Name, configBackupActionConflict, "missing telegram destination user")
			continue
		}
		chatID := strings.TrimSpace(item.ChatID)
		if priorID, duplicate := seenChatIDs[chatID]; duplicate && priorID != item.StableID {
			preview.addConfigBackupPreviewItem(ConfigBackupSectionNotifications, item.StableID, item.Name, configBackupActionConflict, "duplicate Telegram Chat ID in backup")
			continue
		}
		seenChatIDs[chatID] = item.StableID
		if existing, err := h.findTelegramDestinationByChatID(chatID); err == nil && existing.Id != item.StableID {
			preview.addConfigBackupPreviewItem(ConfigBackupSectionNotifications, item.StableID, item.Name, configBackupActionConflict, "Telegram Chat ID belongs to a different target channel")
			continue
		}
		if _, err := h.FindRecordById(CollectionTelegramDestinations, item.StableID); err == nil {
			preview.addConfigBackupPreviewItem(ConfigBackupSectionNotifications, item.StableID, item.Name, configBackupActionUpdate, "stable identifier matched existing telegram destination")
		} else {
			preview.addConfigBackupPreviewItem(ConfigBackupSectionNotifications, item.StableID, item.Name, configBackupActionCreate, "telegram destination does not exist")
		}
	}
	seenPolicies := make(map[string]struct{}, len(notifications.Telegram.Policies))
	seenPolicyNames := make(map[string]string, len(notifications.Telegram.Policies))
	for _, item := range notifications.Telegram.Policies {
		seenPolicies[item.StableID] = struct{}{}
		if _, inBackup := seenDestinations[item.DestinationStableID]; !inBackup {
			if _, err := h.FindRecordById(CollectionTelegramDestinations, item.DestinationStableID); err != nil {
				preview.addConfigBackupPreviewItem(ConfigBackupSectionNotifications, item.StableID, item.Name, configBackupActionConflict, "missing telegram policy destination")
				continue
			}
		}
		if item.NodeScopeMode == TelegramNodeScopeSelected {
			unknownSystem := ""
			for _, systemID := range normalizeTelegramStringSlice(item.NodeScope) {
				if !h.configBackupSystemExists(systemID, backupSystemIDs) {
					unknownSystem = systemID
					break
				}
			}
			if unknownSystem != "" {
				preview.addConfigBackupPreviewItem(ConfigBackupSectionNotifications, item.StableID, item.Name, configBackupActionConflict, "unknown system in telegram policy scope: "+unknownSystem)
				continue
			}
		}
		nameKey := item.DestinationStableID + "\x00" + strings.ToLower(strings.TrimSpace(item.Name))
		if priorID, duplicate := seenPolicyNames[nameKey]; duplicate && priorID != item.StableID {
			preview.addConfigBackupPreviewItem(ConfigBackupSectionNotifications, item.StableID, item.Name, configBackupActionConflict, "duplicate policy name in backup destination")
			continue
		}
		seenPolicyNames[nameKey] = item.StableID
		existingPolicy, existingErr := h.FindRecordById(CollectionTelegramNotificationPolicies, item.StableID)
		if existingErr == nil && existingPolicy.GetString("destination") != item.DestinationStableID {
			preview.addConfigBackupPreviewItem(ConfigBackupSectionNotifications, item.StableID, item.Name, configBackupActionConflict, "telegram policy belongs to a different destination")
			continue
		}
		matchingNames, err := h.FindRecordsByFilter(CollectionTelegramNotificationPolicies, "destination = {:destination} && name = {:name}", "", 1, 0, dbx.Params{
			"destination": item.DestinationStableID,
			"name":        strings.TrimSpace(item.Name),
		})
		if err != nil {
			preview.addConfigBackupPreviewItem(ConfigBackupSectionNotifications, item.StableID, item.Name, configBackupActionError, err.Error())
			continue
		}
		if len(matchingNames) > 0 && matchingNames[0].Id != item.StableID {
			preview.addConfigBackupPreviewItem(ConfigBackupSectionNotifications, item.StableID, item.Name, configBackupActionConflict, "duplicate policy name in target destination")
			continue
		}
		if existingErr == nil {
			preview.addConfigBackupPreviewItem(ConfigBackupSectionNotifications, item.StableID, item.Name, configBackupActionUpdate, "stable identifier matched existing telegram policy")
		} else {
			preview.addConfigBackupPreviewItem(ConfigBackupSectionNotifications, item.StableID, item.Name, configBackupActionCreate, "telegram policy does not exist")
		}
	}
	settings, err := h.FindRecordsByFilter("user_settings", "id != ''", "", -1, 0)
	if err != nil {
		preview.addConfigBackupPreviewItem(ConfigBackupSectionNotifications, "", "user notification settings", configBackupActionError, err.Error())
	} else {
		for _, record := range settings {
			user, err := h.FindRecordById("users", record.GetString("user"))
			if err != nil {
				preview.addConfigBackupPreviewItem(ConfigBackupSectionNotifications, record.Id, "user notification settings", configBackupActionError, err.Error())
				continue
			}
			email := strings.ToLower(user.GetString("email"))
			if _, ok := seenUsers[email]; !ok {
				preview.addConfigBackupPreviewItem(ConfigBackupSectionNotifications, record.Id, user.GetString("email"), configBackupActionPreserve, "target-only user notification settings will be preserved")
			}
		}
	}
	settingsRecord, err := h.getTelegramSettingsRecord()
	if err != nil {
		preview.addConfigBackupPreviewItem(ConfigBackupSectionNotifications, "", "Telegram settings", configBackupActionError, err.Error())
	} else if settingsRecord != nil && !configBackupHasTelegramSettings(notifications.Telegram.Settings) {
		preview.addConfigBackupPreviewItem(ConfigBackupSectionNotifications, settingsRecord.Id, "Telegram settings", configBackupActionPreserve, "target-only Telegram settings will be preserved")
	}
	h.previewTargetOnlyIDs(ConfigBackupSectionNotifications, CollectionTelegramDestinations, seenDestinations, "name", "target-only Telegram destination will be preserved", preview)
	h.previewTargetOnlyIDs(ConfigBackupSectionNotifications, CollectionTelegramNotificationPolicies, seenPolicies, "name", "target-only Telegram policy will be preserved", preview)
}

func (h *Hub) previewPublicStatus(publicStatus ConfigBackupPublicStatus, backupSystemIDs map[string]struct{}, preview *ConfigBackupPreviewResponse) {
	seen := make(map[string]struct{}, len(publicStatus.Systems))
	for _, item := range publicStatus.Systems {
		seen[item.SystemStableID] = struct{}{}
		if !h.configBackupSystemExists(item.SystemStableID, backupSystemIDs) {
			preview.addConfigBackupPreviewItem(ConfigBackupSectionPublicStatus, item.SystemStableID, item.PublicName, configBackupActionConflict, "missing public system")
			continue
		}
		preview.addConfigBackupPreviewItem(ConfigBackupSectionPublicStatus, item.SystemStableID, item.PublicName, configBackupActionUpdate, "public visibility will be merged")
	}
	records, err := h.FindRecordsByFilter(CollectionPublicSystemVisibility, "id != ''", "", -1, 0)
	if err != nil {
		preview.addConfigBackupPreviewItem(ConfigBackupSectionPublicStatus, "", "public visibility", configBackupActionError, err.Error())
		return
	}
	for _, record := range records {
		systemID := record.GetString("system")
		if _, ok := seen[systemID]; !ok {
			preview.addConfigBackupPreviewItem(ConfigBackupSectionPublicStatus, systemID, record.GetString("public_name"), configBackupActionPreserve, "target-only public visibility will be preserved")
		}
	}
}

func configBackupHasTelegramSettings(settings ConfigBackupTelegramSettings) bool {
	return settings.Present || settings.Enabled || settings.PollingEnabled || settings.BotUsername != "" || settings.BotToken != nil
}

func configBackupSystemIDSet(systems []ConfigBackupSystem) map[string]struct{} {
	result := make(map[string]struct{}, len(systems))
	for _, system := range systems {
		if system.StableID != "" {
			result[system.StableID] = struct{}{}
		}
	}
	return result
}

func (h *Hub) configBackupSystemExists(systemID string, backupSystemIDs map[string]struct{}) bool {
	if systemID == "" {
		return false
	}
	if _, ok := backupSystemIDs[systemID]; ok {
		return true
	}
	_, err := h.FindRecordById("systems", systemID)
	return err == nil
}

func (h *Hub) previewNetworkProbes(networkProbes ConfigBackupNetworkProbes, backupSystemIDs map[string]struct{}, preview *ConfigBackupPreviewResponse) {
	backupProbeIDs := make(map[string]struct{}, len(networkProbes.Probes))
	for _, item := range networkProbes.Probes {
		backupProbeIDs[item.StableID] = struct{}{}
		if _, err := h.FindRecordById(CollectionNetworkProbes, item.StableID); err == nil {
			preview.addConfigBackupPreviewItem(ConfigBackupSectionNetworkProbes, item.StableID, item.Name, configBackupActionUpdate, "stable identifier matched existing network probe")
		} else {
			preview.addConfigBackupPreviewItem(ConfigBackupSectionNetworkProbes, item.StableID, item.Name, configBackupActionCreate, "network probe does not exist")
		}
	}
	seenAssignments := make(map[string]struct{}, len(networkProbes.Assignments))
	for _, item := range networkProbes.Assignments {
		key := item.ProbeStableID + ":" + item.SystemStableID
		seenAssignments[key] = struct{}{}
		if _, ok := backupProbeIDs[item.ProbeStableID]; !ok {
			if _, err := h.FindRecordById(CollectionNetworkProbes, item.ProbeStableID); err != nil {
				preview.addConfigBackupPreviewItem(ConfigBackupSectionNetworkProbes, item.ProbeStableID, item.ProbeStableID, configBackupActionConflict, "missing probe assignment probe")
				continue
			}
		}
		if !h.configBackupSystemExists(item.SystemStableID, backupSystemIDs) {
			preview.addConfigBackupPreviewItem(ConfigBackupSectionNetworkProbes, item.SystemStableID, item.SystemStableID, configBackupActionConflict, "missing probe assignment system")
			continue
		}
		preview.addConfigBackupPreviewItem(ConfigBackupSectionNetworkProbes, key, item.ProbeStableID, configBackupActionUpdate, "probe assignment will be merged")
	}
	h.previewTargetOnlyIDs(ConfigBackupSectionNetworkProbes, CollectionNetworkProbes, backupProbeIDs, "name", "target-only network probe will be preserved", preview)
	assignments, err := h.FindRecordsByFilter(CollectionNetworkProbeAssignments, "id != ''", "", -1, 0)
	if err != nil {
		preview.addConfigBackupPreviewItem(ConfigBackupSectionNetworkProbes, "", "probe assignments", configBackupActionError, err.Error())
		return
	}
	for _, record := range assignments {
		key := record.GetString("probe") + ":" + record.GetString("system")
		if _, ok := seenAssignments[key]; !ok {
			preview.addConfigBackupPreviewItem(ConfigBackupSectionNetworkProbes, key, record.GetString("probe"), configBackupActionPreserve, "target-only probe assignment will be preserved")
		}
	}
}

func (h *Hub) previewTargetOnlyIDs(section, collection string, seen map[string]struct{}, displayField, reason string, preview *ConfigBackupPreviewResponse) {
	existing, err := h.existingIDs(collection)
	if err != nil {
		preview.addConfigBackupPreviewItem(section, "", collection, configBackupActionError, err.Error())
		return
	}
	for id, record := range existing {
		if _, ok := seen[id]; !ok {
			preview.addConfigBackupPreviewItem(section, id, record.GetString(displayField), configBackupActionPreserve, reason)
		}
	}
}

func (preview *ConfigBackupPreviewResponse) addConfigBackupPreviewItem(section string, stableID string, displayName string, action string, reason string) {
	preview.Items = append(preview.Items, ConfigBackupPreviewItem{
		Section:     section,
		StableID:    stableID,
		DisplayName: displayName,
		Action:      action,
		Reason:      reason,
	})
	switch action {
	case configBackupActionCreate:
		preview.Summary.Create++
	case configBackupActionUpdate:
		preview.Summary.Update++
	case configBackupActionPreserve:
		preview.Summary.Preserve++
	case configBackupActionSkip:
		preview.Summary.Skip++
	case configBackupActionConflict:
		preview.Summary.Conflict++
	case configBackupActionError:
		preview.Summary.Error++
	}
}

func decryptDocumentSecretsForValidation(document ConfigBackupDocument, credential string) (ConfigBackupDocument, error) {
	if !document.configBackupSectionEnabled(ConfigBackupSectionSystems) {
		document.Systems = nil
	}
	if !document.configBackupSectionEnabled(ConfigBackupSectionNotifications) {
		document.Notifications = ConfigBackupNotifications{}
	}
	for _, system := range document.Systems {
		if _, err := decryptConfigBackupSecret(system.Token, credential); err != nil {
			return document, err
		}
	}
	for _, settings := range document.Notifications.UserSettings {
		for i := range settings.Webhooks {
			if _, err := decryptConfigBackupSecret(&settings.Webhooks[i], credential); err != nil {
				return document, err
			}
		}
	}
	if _, err := decryptConfigBackupSecret(document.Notifications.Telegram.Settings.BotToken, credential); err != nil {
		return document, err
	}
	return document, nil
}

func (document ConfigBackupDocument) configBackupSectionEnabled(section string) bool {
	if _, skipped := document.SkippedSections[section]; skipped {
		return false
	}
	if len(document.Meta.Sections) == 0 {
		return true
	}
	return configBackupIncludesSection(document.Meta.Sections, section)
}

func mergeConfigBackupApplySummary(target *ConfigBackupApplySummary, source ConfigBackupApplySummary) {
	target.Created += source.Created
	target.Updated += source.Updated
	target.Preserved += source.Preserved
	target.Skipped += source.Skipped
}

type configBackupSectionRestoreError struct {
	Section           string
	CompletedSections []string
	Applied           ConfigBackupApplySummary
	Err               error
}

func (e *configBackupSectionRestoreError) Error() string {
	return fmt.Sprintf("restore section %s failed and was rolled back: %v", e.Section, e.Err)
}

func (e *configBackupSectionRestoreError) Unwrap() error { return e.Err }

func (h *Hub) applyConfigBackup(document ConfigBackupDocument, credential string) (ConfigBackupApplySummary, []string, error) {
	summary := ConfigBackupApplySummary{}
	completedSections := []string{}
	warnings := []string{}
	for section, reason := range document.SkippedSections {
		warnings = append(warnings, fmt.Sprintf("%s skipped: %s", section, reason))
		summary.Skipped++
	}
	emailToUserID, err := h.userIDByEmailMap()
	if err != nil {
		return summary, warnings, err
	}
	type sectionApply struct {
		name  string
		apply func(*Hub, *ConfigBackupApplySummary) error
	}
	sections := []sectionApply{
		{ConfigBackupSectionSystems, func(txHub *Hub, sectionSummary *ConfigBackupApplySummary) error {
			return txHub.applyConfigBackupSystems(document.Systems, emailToUserID, credential, sectionSummary)
		}},
		{ConfigBackupSectionAlerts, func(txHub *Hub, sectionSummary *ConfigBackupApplySummary) error {
			return txHub.applyConfigBackupAlerts(document.Alerts, emailToUserID, sectionSummary)
		}},
		{ConfigBackupSectionNotifications, func(txHub *Hub, sectionSummary *ConfigBackupApplySummary) error {
			return txHub.applyConfigBackupNotifications(document.Notifications, emailToUserID, credential, sectionSummary)
		}},
		{ConfigBackupSectionNetworkProbes, func(txHub *Hub, sectionSummary *ConfigBackupApplySummary) error {
			return txHub.applyConfigBackupNetworkProbes(document.NetworkProbes, sectionSummary)
		}},
		{ConfigBackupSectionPublicStatus, func(txHub *Hub, sectionSummary *ConfigBackupApplySummary) error {
			return txHub.applyConfigBackupPublicStatus(document.PublicStatus, sectionSummary)
		}},
	}
	for _, section := range sections {
		if !document.configBackupSectionEnabled(section.name) {
			continue
		}
		sectionSummary := ConfigBackupApplySummary{}
		err := h.RunInTransaction(func(txApp core.App) error {
			txHub := h.configBackupTransactionHub(txApp)
			return section.apply(txHub, &sectionSummary)
		})
		if err != nil {
			return summary, warnings, &configBackupSectionRestoreError{
				Section: section.name, CompletedSections: completedSections, Applied: summary, Err: err,
			}
		}
		mergeConfigBackupApplySummary(&summary, sectionSummary)
		completedSections = append(completedSections, section.name)
	}
	return summary, warnings, nil
}

func (h *Hub) applyConfigBackupSystems(items []ConfigBackupSystem, emailToUserID map[string]string, credential string, summary *ConfigBackupApplySummary) error {
	collection, err := h.FindCachedCollectionByNameOrId("systems")
	if err != nil {
		return err
	}
	for _, item := range items {
		users, missing := configBackupEmailRefsToUserIDs(item.Users, emailToUserID)
		if len(missing) > 0 {
			return fmt.Errorf("missing users for system %s: %s", item.Name, strings.Join(missing, ", "))
		}
		record, err := h.FindRecordById("systems", item.StableID)
		created := false
		if err != nil {
			record = core.NewRecord(collection)
			record.Id = item.StableID
			created = true
		}
		record.Set("name", item.Name)
		record.Set("host", item.Host)
		record.Set("port", item.Port)
		record.Set("users", users)
		record.Set("status", item.Status)
		if item.Info != nil {
			record.Set("info", item.Info)
		}
		if err := h.Save(record); err != nil {
			return err
		}
		token, err := decryptConfigBackupSecret(item.Token, credential)
		if err != nil {
			return err
		}
		if token != "" {
			if err := h.upsertConfigBackupFingerprint(item.StableID, token); err != nil {
				return err
			}
		}
		if created {
			summary.Created++
		} else {
			summary.Updated++
		}
	}
	return nil
}

func (h *Hub) upsertConfigBackupFingerprint(systemID string, token string) error {
	record, err := h.fingerprintBySystem(systemID)
	if err != nil {
		collection, err := h.FindCachedCollectionByNameOrId("fingerprints")
		if err != nil {
			return err
		}
		record = core.NewRecord(collection)
		record.Set("system", systemID)
		record.Set("fingerprint", "")
	}
	record.Set("token", token)
	return h.Save(record)
}

func (h *Hub) applyConfigBackupAlerts(alerts ConfigBackupAlerts, emailToUserID map[string]string, summary *ConfigBackupApplySummary) error {
	alertCollection, err := h.FindCachedCollectionByNameOrId("alerts")
	if err != nil {
		return err
	}
	for _, item := range alerts.Definitions {
		userID := emailToUserID[strings.ToLower(item.UserEmail)]
		if userID == "" {
			return fmt.Errorf("missing alert user: %s", item.UserEmail)
		}
		record, err := h.FindRecordById("alerts", item.StableID)
		created := false
		if err != nil {
			record = core.NewRecord(alertCollection)
			record.Id = item.StableID
			created = true
		}
		record.Set("user", userID)
		record.Set("system", item.SystemStableID)
		record.Set("name", item.Name)
		record.Set("min", item.Min)
		record.Set("value", item.Value)
		record.Set("triggered", item.Triggered)
		if err := h.Save(record); err != nil {
			return err
		}
		incrementApplySummary(summary, created)
	}
	quietCollection, err := h.FindCachedCollectionByNameOrId("quiet_hours")
	if err != nil {
		return err
	}
	for _, item := range alerts.QuietHours {
		userID := emailToUserID[strings.ToLower(item.UserEmail)]
		if userID == "" {
			return fmt.Errorf("missing quiet-hours user: %s", item.UserEmail)
		}
		record, err := h.FindRecordById("quiet_hours", item.StableID)
		created := false
		if err != nil {
			record = core.NewRecord(quietCollection)
			record.Id = item.StableID
			created = true
		}
		record.Set("user", userID)
		record.Set("system", item.SystemStableID)
		record.Set("type", item.Type)
		if start, err := types.ParseDateTime(item.Start); err == nil {
			record.Set("start", start)
		}
		if end, err := types.ParseDateTime(item.End); err == nil {
			record.Set("end", end)
		}
		if err := h.Save(record); err != nil {
			return err
		}
		incrementApplySummary(summary, created)
	}
	return nil
}

func (h *Hub) applyConfigBackupNotifications(notifications ConfigBackupNotifications, emailToUserID map[string]string, credential string, summary *ConfigBackupApplySummary) error {
	for _, item := range notifications.UserSettings {
		userID := emailToUserID[strings.ToLower(item.UserEmail)]
		if userID == "" {
			return fmt.Errorf("missing notification user: %s", item.UserEmail)
		}
		record, err := h.FindFirstRecordByFilter("user_settings", "user = {:user}", dbx.Params{"user": userID})
		created := false
		if err != nil {
			collection, err := h.FindCachedCollectionByNameOrId("user_settings")
			if err != nil {
				return err
			}
			record = core.NewRecord(collection)
			record.Set("user", userID)
			created = true
		}
		var existing map[string]any
		_ = record.UnmarshalJSONField("settings", &existing)
		if existing == nil {
			existing = map[string]any{}
		}
		existing["emails"] = item.Emails
		if len(item.Webhooks) > 0 {
			current := configBackupStringSlice(existing["webhooks"])
			webhooks := make([]string, 0, len(item.Webhooks))
			for i := range item.Webhooks {
				secret := &item.Webhooks[i]
				if secret.Redacted || secret.Encrypted == "" {
					if i < len(current) {
						webhooks = append(webhooks, current[i])
					}
					continue
				}
				webhook, err := decryptConfigBackupSecret(secret, credential)
				if err != nil {
					return err
				}
				if webhook != "" {
					webhooks = append(webhooks, webhook)
				}
			}
			existing["webhooks"] = webhooks
		}
		record.Set("settings", existing)
		if err := h.Save(record); err != nil {
			return err
		}
		incrementApplySummary(summary, created)
	}
	if configBackupHasTelegramSettings(notifications.Telegram.Settings) {
		token, err := decryptConfigBackupSecret(notifications.Telegram.Settings.BotToken, credential)
		if err != nil {
			return err
		}
		_, err = h.saveTelegramSettings(TelegramSettingsInput{
			Enabled:        notifications.Telegram.Settings.Enabled,
			PollingEnabled: notifications.Telegram.Settings.PollingEnabled,
			BotToken:       token,
		}, telegramSettingsRecord{})
		if err != nil {
			return err
		}
		summary.Updated++
	}
	for _, item := range notifications.Telegram.Destinations {
		userID := ""
		if item.UserEmail != "" {
			userID = emailToUserID[strings.ToLower(item.UserEmail)]
			if userID == "" {
				return fmt.Errorf("missing telegram destination user: %s", item.UserEmail)
			}
		}
		record, err := h.FindRecordById(CollectionTelegramDestinations, item.StableID)
		created := false
		if err != nil {
			collection, err := h.FindCachedCollectionByNameOrId(CollectionTelegramDestinations)
			if err != nil {
				return err
			}
			record = core.NewRecord(collection)
			record.Id = item.StableID
			created = true
		}
		input := TelegramDestinationInput{
			UserID:          userID,
			Name:            item.Name,
			ChatID:          item.ChatID,
			ChatType:        item.ChatType,
			Role:            item.Role,
			Enabled:         &item.Enabled,
			NodeScope:       item.NodeScope,
			AlertLevelScope: item.AlertLevelScope,
		}
		if item.MuteUntil != "" {
			if parsed, err := types.ParseDateTime(item.MuteUntil); err == nil {
				t := parsed.Time()
				input.MuteUntil = &t
			}
		}
		if _, err := h.upsertTelegramDestination(record, input); err != nil {
			return err
		}
		if notifications.SectionVersion == "" || notifications.SectionVersion == "1" {
			if err := h.upsertTelegramDefaultPolicyFromLegacy(record, input); err != nil {
				return err
			}
		}
		incrementApplySummary(summary, created)
	}
	if notifications.SectionVersion == ConfigBackupNotificationsVersion {
		for _, item := range notifications.Telegram.Policies {
			record, err := h.FindRecordById(CollectionTelegramNotificationPolicies, item.StableID)
			created := false
			if err != nil {
				collection, err := h.FindCachedCollectionByNameOrId(CollectionTelegramNotificationPolicies)
				if err != nil {
					return err
				}
				record = core.NewRecord(collection)
				record.Id = item.StableID
				created = true
			}
			if _, err := h.upsertTelegramNotificationPolicy(record, item.DestinationStableID, TelegramNotificationPolicyInput{
				Name: item.Name, Enabled: &item.Enabled, NodeScopeMode: item.NodeScopeMode,
				NodeScope: item.NodeScope, AlertLevelScope: item.AlertLevelScope,
			}); err != nil {
				return err
			}
			incrementApplySummary(summary, created)
		}
	}
	return nil
}

func (h *Hub) applyConfigBackupPublicStatus(publicStatus ConfigBackupPublicStatus, summary *ConfigBackupApplySummary) error {
	for _, item := range publicStatus.Systems {
		_, record := h.findPublicVisibility(item.SystemStableID)
		created := record == nil
		if _, err := h.upsertPublicVisibility(item.SystemStableID, PublicSystemVisibility{
			SystemID:       item.SystemStableID,
			PublicEnabled:  item.PublicEnabled,
			PublicName:     item.PublicName,
			ShowCPU:        item.ShowCPU,
			ShowMemory:     item.ShowMemory,
			ShowDisk:       item.ShowDisk,
			PublicProbeIDs: item.PublicProbeStableIDs,
		}); err != nil {
			return err
		}
		incrementApplySummary(summary, created)
	}
	return nil
}

func (h *Hub) applyConfigBackupNetworkProbes(networkProbes ConfigBackupNetworkProbes, summary *ConfigBackupApplySummary) error {
	collection, err := h.FindCachedCollectionByNameOrId(CollectionNetworkProbes)
	if err != nil {
		return err
	}
	for _, item := range networkProbes.Probes {
		record, err := h.FindRecordById(CollectionNetworkProbes, item.StableID)
		created := false
		if err != nil {
			record = core.NewRecord(collection)
			record.Id = item.StableID
			created = true
		}
		enabled := item.Enabled
		input := NetworkProbeInput{
			Name:            item.Name,
			Type:            item.Type,
			Target:          item.Target,
			IntervalSeconds: item.IntervalSeconds,
			TimeoutSeconds:  item.TimeoutSeconds,
			Enabled:         &enabled,
			Scope:           item.Scope,
		}
		setProbeRecord(record, normalizeProbeInput(input))
		if err := h.Save(record); err != nil {
			return err
		}
		incrementApplySummary(summary, created)
	}
	assignmentCollection, err := h.FindCachedCollectionByNameOrId(CollectionNetworkProbeAssignments)
	if err != nil {
		return err
	}
	for _, item := range networkProbes.Assignments {
		record, err := h.FindFirstRecordByFilter(CollectionNetworkProbeAssignments, "probe = {:probe} && system = {:system}", dbx.Params{
			"probe":  item.ProbeStableID,
			"system": item.SystemStableID,
		})
		created := false
		if err != nil {
			record = core.NewRecord(assignmentCollection)
			record.Set("probe", item.ProbeStableID)
			record.Set("system", item.SystemStableID)
			created = true
		}
		record.Set("enabled", item.Enabled)
		if err := h.Save(record); err != nil {
			return err
		}
		incrementApplySummary(summary, created)
	}
	return nil
}

func configBackupStringSlice(value any) []string {
	values, ok := value.([]any)
	if !ok {
		if strings, ok := value.([]string); ok {
			return strings
		}
		return nil
	}
	result := make([]string, 0, len(values))
	for _, value := range values {
		if text, ok := value.(string); ok {
			result = append(result, text)
		}
	}
	return result
}

func incrementApplySummary(summary *ConfigBackupApplySummary, created bool) {
	if created {
		summary.Created++
	} else {
		summary.Updated++
	}
}
