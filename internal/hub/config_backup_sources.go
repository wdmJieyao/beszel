package hub

import (
	"encoding/json"
	"fmt"
	"slices"
	"strings"

	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase/core"
)

type configBackupExportOptions struct {
	IncludeSecrets bool
	Credential     string
	Sections       []string
}

func (h *Hub) configBackupTransactionHub(app core.App) *Hub {
	return &Hub{
		App:                   app,
		AlertManager:          h.AlertManager,
		um:                    h.um,
		rm:                    h.rm,
		sm:                    h.sm,
		hb:                    h.hb,
		hbStop:                h.hbStop,
		telegramDeliveryState: h.telegramDeliveryState,
		telegramTransport:     h.telegramTransport,
		telegramPollingCancel: h.telegramPollingCancel,
		networkProbeLive:      h.networkProbeLive,
		pubKey:                h.pubKey,
		signer:                h.signer,
		appURL:                h.appURL,
	}
}

func (h *Hub) buildConfigBackupDocument(options configBackupExportOptions) (ConfigBackupDocument, []string, error) {
	var document ConfigBackupDocument
	var warnings []string
	err := h.RunInTransaction(func(txApp core.App) error {
		txHub := h.configBackupTransactionHub(txApp)
		var err error
		document, warnings, err = txHub.buildConfigBackupDocumentSnapshot(options)
		return err
	})
	return document, warnings, err
}

func (h *Hub) buildConfigBackupDocumentSnapshot(options configBackupExportOptions) (ConfigBackupDocument, []string, error) {
	sections, err := normalizeConfigBackupSections(options.Sections)
	if err != nil {
		return ConfigBackupDocument{}, nil, err
	}
	if options.IncludeSecrets && options.Credential == "" {
		return ConfigBackupDocument{}, nil, fmt.Errorf("encryptionCredential is required when includeSecrets is true")
	}
	document := newConfigBackupDocument(sections)
	document.Encryption.Enabled = options.IncludeSecrets
	warnings := []string{}

	users, userIDToEmail, err := h.configBackupUsers()
	if err != nil {
		return ConfigBackupDocument{}, nil, err
	}
	document.Users = users

	if configBackupIncludesSection(sections, ConfigBackupSectionSystems) {
		document.Systems, err = h.configBackupSystems(userIDToEmail, options)
		if err != nil {
			return ConfigBackupDocument{}, nil, err
		}
	}
	if configBackupIncludesSection(sections, ConfigBackupSectionAlerts) {
		document.Alerts, err = h.configBackupAlerts(userIDToEmail)
		if err != nil {
			return ConfigBackupDocument{}, nil, err
		}
	}
	if configBackupIncludesSection(sections, ConfigBackupSectionNotifications) {
		document.Notifications, err = h.configBackupNotifications(userIDToEmail, options)
		if err != nil {
			return ConfigBackupDocument{}, nil, err
		}
	}
	if configBackupIncludesSection(sections, ConfigBackupSectionPublicStatus) {
		document.PublicStatus, err = h.configBackupPublicStatus()
		if err != nil {
			return ConfigBackupDocument{}, nil, err
		}
	}
	if configBackupIncludesSection(sections, ConfigBackupSectionNetworkProbes) {
		document.NetworkProbes, err = h.configBackupNetworkProbes()
		if err != nil {
			return ConfigBackupDocument{}, nil, err
		}
	}
	return document, warnings, nil
}

func (h *Hub) configBackupUsers() ([]ConfigBackupUser, map[string]string, error) {
	records, err := h.FindRecordsByFilter("users", "id != ''", "email", -1, 0)
	if err != nil {
		return nil, nil, err
	}
	users := make([]ConfigBackupUser, 0, len(records))
	userIDToEmail := make(map[string]string, len(records))
	for _, record := range records {
		email := record.GetString("email")
		userIDToEmail[record.Id] = email
		users = append(users, ConfigBackupUser{StableID: record.Id, Email: email})
	}
	return users, userIDToEmail, nil
}

func (h *Hub) configBackupSystems(userIDToEmail map[string]string, options configBackupExportOptions) ([]ConfigBackupSystem, error) {
	systems, err := h.FindRecordsByFilter("systems", "id != ''", "name", -1, 0)
	if err != nil {
		return nil, err
	}
	fingerprints, err := h.FindRecordsByFilter("fingerprints", "id != ''", "", -1, 0)
	if err != nil {
		return nil, err
	}
	tokenBySystem := make(map[string]string, len(fingerprints))
	for _, fingerprint := range fingerprints {
		tokenBySystem[fingerprint.GetString("system")] = fingerprint.GetString("token")
	}
	result := make([]ConfigBackupSystem, 0, len(systems))
	for _, record := range systems {
		users := make([]ConfigBackupUserRef, 0, len(record.GetStringSlice("users")))
		for _, userID := range record.GetStringSlice("users") {
			if email := userIDToEmail[userID]; email != "" {
				users = append(users, ConfigBackupUserRef{Email: email})
			}
		}
		item := ConfigBackupSystem{
			StableID: record.Id,
			Name:     record.GetString("name"),
			Host:     record.GetString("host"),
			Port:     record.GetString("port"),
			Users:    users,
			Status:   record.GetString("status"),
			Info:     recordJSONMap(record.Get("info")),
		}
		if token := tokenBySystem[record.Id]; token != "" {
			secret, err := configBackupSecretValue(token, options, "system.token")
			if err != nil {
				return nil, err
			}
			item.Token = secret
		}
		result = append(result, item)
	}
	return result, nil
}

func (h *Hub) configBackupAlerts(userIDToEmail map[string]string) (ConfigBackupAlerts, error) {
	alertRecords, err := h.FindRecordsByFilter("alerts", "id != ''", "name", -1, 0)
	if err != nil {
		return ConfigBackupAlerts{}, err
	}
	quietRecords, err := h.FindRecordsByFilter("quiet_hours", "id != ''", "user", -1, 0)
	if err != nil {
		return ConfigBackupAlerts{}, err
	}
	result := ConfigBackupAlerts{
		Definitions: make([]ConfigBackupAlertDefinition, 0, len(alertRecords)),
		QuietHours:  make([]ConfigBackupQuietHour, 0, len(quietRecords)),
	}
	for _, record := range alertRecords {
		result.Definitions = append(result.Definitions, ConfigBackupAlertDefinition{
			StableID:       record.Id,
			SystemStableID: record.GetString("system"),
			UserEmail:      userIDToEmail[record.GetString("user")],
			Name:           record.GetString("name"),
			Min:            record.GetFloat("min"),
			Value:          record.GetFloat("value"),
			Triggered:      record.GetBool("triggered"),
		})
	}
	for _, record := range quietRecords {
		result.QuietHours = append(result.QuietHours, ConfigBackupQuietHour{
			StableID:       record.Id,
			UserEmail:      userIDToEmail[record.GetString("user")],
			SystemStableID: record.GetString("system"),
			Type:           record.GetString("type"),
			Start:          record.GetDateTime("start").String(),
			End:            record.GetDateTime("end").String(),
		})
	}
	return result, nil
}

func (h *Hub) configBackupNotifications(userIDToEmail map[string]string, options configBackupExportOptions) (ConfigBackupNotifications, error) {
	settingsRecords, err := h.FindRecordsByFilter("user_settings", "id != ''", "", -1, 0)
	if err != nil {
		return ConfigBackupNotifications{}, err
	}
	result := ConfigBackupNotifications{
		SectionVersion: ConfigBackupNotificationsVersion,
		UserSettings:   make([]ConfigBackupUserNotificationSettings, 0, len(settingsRecords)),
	}
	for _, record := range settingsRecords {
		var settings struct {
			Emails   []string `json:"emails"`
			Webhooks []string `json:"webhooks"`
		}
		_ = record.UnmarshalJSONField("settings", &settings)
		item := ConfigBackupUserNotificationSettings{
			UserEmail: userIDToEmail[record.GetString("user")],
			Emails:    settings.Emails,
			Webhooks:  []ConfigBackupSecret{},
		}
		for _, webhook := range settings.Webhooks {
			secret, err := configBackupSecretValue(webhook, options, "notification.webhook")
			if err != nil {
				return ConfigBackupNotifications{}, err
			}
			if secret != nil {
				item.Webhooks = append(item.Webhooks, *secret)
			}
		}
		result.UserSettings = append(result.UserSettings, item)
	}

	telegramSettings, err := h.loadTelegramSettings()
	if err != nil {
		return ConfigBackupNotifications{}, err
	}
	result.Telegram.Settings = ConfigBackupTelegramSettings{
		Present:        true,
		Enabled:        telegramSettings.Enabled,
		PollingEnabled: telegramSettings.PollingEnabled,
		BotUsername:    telegramSettings.BotUsername,
	}
	if telegramSettings.BotTokenEncrypted != "" {
		token, err := h.decryptTelegramToken(telegramSettings)
		if err != nil {
			return ConfigBackupNotifications{}, err
		}
		secret, err := configBackupSecretValue(token, options, "telegram.botToken")
		if err != nil {
			return ConfigBackupNotifications{}, err
		}
		result.Telegram.Settings.BotToken = secret
	}

	destinations, err := h.listTelegramDestinations()
	if err != nil {
		return ConfigBackupNotifications{}, err
	}
	result.Telegram.Destinations = make([]ConfigBackupTelegramDestination, 0, len(destinations))
	for _, destination := range destinations {
		item := ConfigBackupTelegramDestination{
			StableID:  destination.ID,
			UserEmail: userIDToEmail[destination.UserID],
			Name:      destination.Name,
			ChatID:    destination.ChatID,
			ChatType:  destination.ChatType,
			Role:      destination.Role,
			Enabled:   destination.Enabled,
		}
		if destination.MuteUntil != nil {
			item.MuteUntil = destination.MuteUntil.UTC().Format(configBackupDateLayout)
		}
		result.Telegram.Destinations = append(result.Telegram.Destinations, item)
	}
	policiesByDestination, err := h.listAllTelegramNotificationPolicies()
	if err != nil {
		return ConfigBackupNotifications{}, err
	}
	result.Telegram.Policies = []ConfigBackupTelegramPolicy{}
	for _, destination := range destinations {
		for _, policy := range policiesByDestination[destination.ID] {
			result.Telegram.Policies = append(result.Telegram.Policies, ConfigBackupTelegramPolicy{
				StableID: policy.ID, DestinationStableID: destination.ID, Name: policy.Name,
				Enabled: policy.Enabled, NodeScopeMode: policy.NodeScopeMode,
				NodeScope: policy.NodeScope, AlertLevelScope: policy.AlertLevelScope,
			})
		}
	}
	return result, nil
}

func (h *Hub) configBackupPublicStatus() (ConfigBackupPublicStatus, error) {
	records, err := h.FindRecordsByFilter(CollectionPublicSystemVisibility, "id != ''", "system", -1, 0)
	if err != nil {
		return ConfigBackupPublicStatus{}, err
	}
	result := ConfigBackupPublicStatus{Systems: make([]ConfigBackupPublicSystem, 0, len(records))}
	for _, record := range records {
		visibility := publicVisibilityFromRecord(record)
		result.Systems = append(result.Systems, ConfigBackupPublicSystem{
			SystemStableID:       visibility.SystemID,
			PublicEnabled:        visibility.PublicEnabled,
			PublicName:           visibility.PublicName,
			ShowCPU:              visibility.ShowCPU,
			ShowMemory:           visibility.ShowMemory,
			ShowDisk:             visibility.ShowDisk,
			PublicProbeStableIDs: visibility.PublicProbeIDs,
		})
	}
	return result, nil
}

func (h *Hub) configBackupNetworkProbes() (ConfigBackupNetworkProbes, error) {
	probes, err := h.FindRecordsByFilter(CollectionNetworkProbes, "id != ''", "name", -1, 0)
	if err != nil {
		return ConfigBackupNetworkProbes{}, err
	}
	assignments, err := h.FindRecordsByFilter(CollectionNetworkProbeAssignments, "id != ''", "", -1, 0)
	if err != nil {
		return ConfigBackupNetworkProbes{}, err
	}
	result := ConfigBackupNetworkProbes{
		Probes:      make([]ConfigBackupNetworkProbe, 0, len(probes)),
		Assignments: make([]ConfigBackupNetworkProbeAssignment, 0, len(assignments)),
	}
	for _, record := range probes {
		result.Probes = append(result.Probes, ConfigBackupNetworkProbe{
			StableID:        record.Id,
			Name:            record.GetString("name"),
			Type:            record.GetString("type"),
			Target:          record.GetString("target"),
			IntervalSeconds: record.GetInt("interval_seconds"),
			TimeoutSeconds:  record.GetInt("timeout_seconds"),
			Enabled:         record.GetBool("enabled"),
			Scope:           networkProbeScopeFromRecord(record),
		})
	}
	for _, record := range assignments {
		result.Assignments = append(result.Assignments, ConfigBackupNetworkProbeAssignment{
			ProbeStableID:  record.GetString("probe"),
			SystemStableID: record.GetString("system"),
			Enabled:        record.GetBool("enabled"),
		})
	}
	return result, nil
}

func configBackupSecretValue(value string, options configBackupExportOptions, contentType string) (*ConfigBackupSecret, error) {
	if value == "" {
		return nil, nil
	}
	if !options.IncludeSecrets {
		return redactedConfigBackupSecret(contentType), nil
	}
	return encryptConfigBackupSecret(value, options.Credential, contentType)
}

func recordJSONMap(value any) map[string]any {
	if value == nil {
		return nil
	}
	data, err := json.Marshal(value)
	if err != nil {
		return nil
	}
	var out map[string]any
	if err := json.Unmarshal(data, &out); err != nil {
		return nil
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

func (h *Hub) userIDByEmailMap() (map[string]string, error) {
	users, err := h.FindRecordsByFilter("users", "id != ''", "email", -1, 0)
	if err != nil {
		return nil, err
	}
	result := make(map[string]string, len(users))
	for _, user := range users {
		result[strings.ToLower(user.GetString("email"))] = user.Id
	}
	return result, nil
}

func (h *Hub) existingIDs(collection string) (map[string]*core.Record, error) {
	records, err := h.FindRecordsByFilter(collection, "id != ''", "", -1, 0)
	if err != nil {
		return nil, err
	}
	result := make(map[string]*core.Record, len(records))
	for _, record := range records {
		result[record.Id] = record
	}
	return result, nil
}

func configBackupEmailRefsToUserIDs(refs []ConfigBackupUserRef, emailToUserID map[string]string) ([]string, []string) {
	userIDs := make([]string, 0, len(refs))
	missing := []string{}
	for _, ref := range refs {
		email := strings.ToLower(strings.TrimSpace(ref.Email))
		if email == "" {
			continue
		}
		userID := emailToUserID[email]
		if userID == "" {
			missing = append(missing, ref.Email)
			continue
		}
		if !slices.Contains(userIDs, userID) {
			userIDs = append(userIDs, userID)
		}
	}
	return userIDs, missing
}

const configBackupDateLayout = "2006-01-02T15:04:05Z07:00"

func (h *Hub) fingerprintBySystem(systemID string) (*core.Record, error) {
	return h.FindFirstRecordByFilter("fingerprints", "system = {:system}", dbx.Params{"system": systemID})
}
