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
		Warnings:           []string{},
		RequiresCredential: configBackupHasEncryptedSecrets(document),
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
	h.previewSystems(document.Systems, emailToUserID, &preview)
	h.previewAlerts(document.Alerts, emailToUserID, backupSystemIDs, &preview)
	h.previewNotifications(document.Notifications, emailToUserID, &preview)
	h.previewPublicStatus(document.PublicStatus, backupSystemIDs, &preview)
	h.previewNetworkProbes(document.NetworkProbes, backupSystemIDs, &preview)
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
	for _, item := range alerts.Definitions {
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
	for _, item := range alerts.QuietHours {
		if emailToUserID[strings.ToLower(item.UserEmail)] == "" {
			preview.addConfigBackupPreviewItem(ConfigBackupSectionAlerts, item.StableID, item.Type, configBackupActionConflict, "missing quiet-hours user")
			continue
		}
		if item.SystemStableID != "" {
			if !h.configBackupSystemExists(item.SystemStableID, backupSystemIDs) {
				preview.addConfigBackupPreviewItem(ConfigBackupSectionAlerts, item.StableID, item.Type, configBackupActionConflict, "missing quiet-hours system")
				continue
			}
		}
		if _, err := h.FindRecordById("quiet_hours", item.StableID); err == nil {
			preview.addConfigBackupPreviewItem(ConfigBackupSectionAlerts, item.StableID, item.Type, configBackupActionUpdate, "stable identifier matched existing quiet-hours window")
		} else {
			preview.addConfigBackupPreviewItem(ConfigBackupSectionAlerts, item.StableID, item.Type, configBackupActionCreate, "quiet-hours window does not exist")
		}
	}
}

func (h *Hub) previewNotifications(notifications ConfigBackupNotifications, emailToUserID map[string]string, preview *ConfigBackupPreviewResponse) {
	for _, item := range notifications.UserSettings {
		if emailToUserID[strings.ToLower(item.UserEmail)] == "" {
			preview.addConfigBackupPreviewItem(ConfigBackupSectionNotifications, item.UserEmail, item.UserEmail, configBackupActionConflict, "missing notification user")
			continue
		}
		preview.addConfigBackupPreviewItem(ConfigBackupSectionNotifications, item.UserEmail, item.UserEmail, configBackupActionUpdate, "user notification settings will be merged")
	}
	if notifications.Telegram.Settings.Enabled || notifications.Telegram.Settings.BotToken != nil || len(notifications.Telegram.Destinations) > 0 {
		preview.addConfigBackupPreviewItem(ConfigBackupSectionNotifications, "telegram_settings", "Telegram settings", configBackupActionUpdate, "telegram settings will be merged")
	}
	for _, item := range notifications.Telegram.Destinations {
		if item.UserEmail != "" && emailToUserID[strings.ToLower(item.UserEmail)] == "" {
			preview.addConfigBackupPreviewItem(ConfigBackupSectionNotifications, item.StableID, item.Name, configBackupActionConflict, "missing telegram destination user")
			continue
		}
		if _, err := h.FindRecordById(CollectionTelegramDestinations, item.StableID); err == nil {
			preview.addConfigBackupPreviewItem(ConfigBackupSectionNotifications, item.StableID, item.Name, configBackupActionUpdate, "stable identifier matched existing telegram destination")
		} else {
			preview.addConfigBackupPreviewItem(ConfigBackupSectionNotifications, item.StableID, item.Name, configBackupActionCreate, "telegram destination does not exist")
		}
	}
}

func (h *Hub) previewPublicStatus(publicStatus ConfigBackupPublicStatus, backupSystemIDs map[string]struct{}, preview *ConfigBackupPreviewResponse) {
	for _, item := range publicStatus.Systems {
		if !h.configBackupSystemExists(item.SystemStableID, backupSystemIDs) {
			preview.addConfigBackupPreviewItem(ConfigBackupSectionPublicStatus, item.SystemStableID, item.PublicName, configBackupActionConflict, "missing public system")
			continue
		}
		preview.addConfigBackupPreviewItem(ConfigBackupSectionPublicStatus, item.SystemStableID, item.PublicName, configBackupActionUpdate, "public visibility will be merged")
	}
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
	for _, item := range networkProbes.Assignments {
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
		preview.addConfigBackupPreviewItem(ConfigBackupSectionNetworkProbes, item.ProbeStableID+":"+item.SystemStableID, item.ProbeStableID, configBackupActionUpdate, "probe assignment will be merged")
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

func (h *Hub) applyConfigBackup(document ConfigBackupDocument, credential string) (ConfigBackupApplySummary, []string, error) {
	summary := ConfigBackupApplySummary{}
	warnings := []string{}
	emailToUserID, err := h.userIDByEmailMap()
	if err != nil {
		return summary, warnings, err
	}
	if err := h.applyConfigBackupSystems(document.Systems, emailToUserID, credential, &summary); err != nil {
		return summary, warnings, err
	}
	if err := h.applyConfigBackupAlerts(document.Alerts, emailToUserID, &summary); err != nil {
		return summary, warnings, err
	}
	if err := h.applyConfigBackupNotifications(document.Notifications, emailToUserID, credential, &summary); err != nil {
		return summary, warnings, err
	}
	if err := h.applyConfigBackupNetworkProbes(document.NetworkProbes, &summary); err != nil {
		return summary, warnings, err
	}
	if err := h.applyConfigBackupPublicStatus(document.PublicStatus, &summary); err != nil {
		return summary, warnings, err
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
		webhooks := make([]string, 0, len(item.Webhooks))
		for i := range item.Webhooks {
			webhook, err := decryptConfigBackupSecret(&item.Webhooks[i], credential)
			if err != nil {
				return err
			}
			if webhook != "" {
				webhooks = append(webhooks, webhook)
			}
		}
		var existing map[string]any
		_ = record.UnmarshalJSONField("settings", &existing)
		if existing == nil {
			existing = map[string]any{}
		}
		existing["emails"] = item.Emails
		existing["webhooks"] = webhooks
		record.Set("settings", existing)
		if err := h.Save(record); err != nil {
			return err
		}
		incrementApplySummary(summary, created)
	}
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
	if notifications.Telegram.Settings.Enabled || notifications.Telegram.Settings.BotToken != nil || len(notifications.Telegram.Destinations) > 0 {
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
		incrementApplySummary(summary, created)
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

func incrementApplySummary(summary *ConfigBackupApplySummary, created bool) {
	if created {
		summary.Created++
	} else {
		summary.Updated++
	}
}
