package hub

const (
	ConfigBackupVersion = "1"
	ConfigBackupMode    = "merge"

	ConfigBackupSectionSystems       = "systems"
	ConfigBackupSectionAlerts        = "alerts"
	ConfigBackupSectionNotifications = "notifications"
	ConfigBackupSectionPublicStatus  = "publicStatus"
	ConfigBackupSectionNetworkProbes = "networkProbes"
)

var defaultConfigBackupSections = []string{
	ConfigBackupSectionSystems,
	ConfigBackupSectionAlerts,
	ConfigBackupSectionNotifications,
	ConfigBackupSectionPublicStatus,
	ConfigBackupSectionNetworkProbes,
}

type ConfigBackupExportRequest struct {
	IncludeSecrets       bool     `json:"includeSecrets"`
	EncryptionCredential string   `json:"encryptionCredential"`
	Sections             []string `json:"sections"`
}

type ConfigBackupExportResponse struct {
	Filename      string   `json:"filename"`
	ContentType   string   `json:"contentType"`
	BackupVersion string   `json:"backupVersion"`
	Warnings      []string `json:"warnings"`
	Content       string   `json:"content"`
}

type ConfigBackupValidationRequest struct {
	Content              string `json:"content"`
	DecryptionCredential string `json:"decryptionCredential"`
}

type ConfigBackupRestoreRequest struct {
	Content              string `json:"content"`
	PreviewID            string `json:"previewId"`
	Mode                 string `json:"mode"`
	DecryptionCredential string `json:"decryptionCredential"`
}

type ConfigBackupRestoreResponse struct {
	Mode     string                   `json:"mode"`
	Applied  ConfigBackupApplySummary `json:"applied"`
	Warnings []string                 `json:"warnings"`
}

type ConfigBackupApplySummary struct {
	Created   int `json:"created"`
	Updated   int `json:"updated"`
	Preserved int `json:"preserved"`
	Skipped   int `json:"skipped"`
}

type ConfigBackupDocument struct {
	Meta          ConfigBackupMeta          `json:"meta" yaml:"meta"`
	Encryption    ConfigBackupEncryption    `json:"encryption,omitempty" yaml:"encryption,omitempty"`
	Users         []ConfigBackupUser        `json:"users,omitempty" yaml:"users,omitempty"`
	Systems       []ConfigBackupSystem      `json:"systems,omitempty" yaml:"systems,omitempty"`
	Alerts        ConfigBackupAlerts        `json:"alerts,omitempty" yaml:"alerts,omitempty"`
	Notifications ConfigBackupNotifications `json:"notifications,omitempty" yaml:"notifications,omitempty"`
	PublicStatus  ConfigBackupPublicStatus  `json:"publicStatus,omitempty" yaml:"publicStatus,omitempty"`
	NetworkProbes ConfigBackupNetworkProbes `json:"networkProbes,omitempty" yaml:"networkProbes,omitempty"`
}

type ConfigBackupMeta struct {
	BackupVersion string   `json:"backupVersion" yaml:"backupVersion"`
	SourceVersion string   `json:"sourceVersion" yaml:"sourceVersion"`
	CreatedAt     string   `json:"createdAt" yaml:"createdAt"`
	Mode          string   `json:"mode" yaml:"mode"`
	Sections      []string `json:"sections" yaml:"sections"`
}

type ConfigBackupEncryption struct {
	Enabled   bool   `json:"enabled" yaml:"enabled"`
	Algorithm string `json:"algorithm" yaml:"algorithm"`
	KDF       string `json:"kdf" yaml:"kdf"`
}

type ConfigBackupSecret struct {
	Encrypted   string `json:"encrypted,omitempty" yaml:"encrypted,omitempty"`
	Nonce       string `json:"nonce,omitempty" yaml:"nonce,omitempty"`
	Salt        string `json:"salt,omitempty" yaml:"salt,omitempty"`
	Algorithm   string `json:"algorithm,omitempty" yaml:"algorithm,omitempty"`
	KDF         string `json:"kdf,omitempty" yaml:"kdf,omitempty"`
	ContentType string `json:"contentType,omitempty" yaml:"contentType,omitempty"`
	Redacted    bool   `json:"redacted,omitempty" yaml:"redacted,omitempty"`
}

type ConfigBackupUser struct {
	StableID string `json:"stableId" yaml:"stableId"`
	Email    string `json:"email" yaml:"email"`
}

type ConfigBackupSystem struct {
	StableID string                `json:"stableId" yaml:"stableId"`
	Name     string                `json:"name" yaml:"name"`
	Host     string                `json:"host" yaml:"host"`
	Port     string                `json:"port,omitempty" yaml:"port,omitempty"`
	Users    []ConfigBackupUserRef `json:"users,omitempty" yaml:"users,omitempty"`
	Token    *ConfigBackupSecret   `json:"token,omitempty" yaml:"token,omitempty"`
	Info     map[string]any        `json:"info,omitempty" yaml:"info,omitempty"`
	Status   string                `json:"status,omitempty" yaml:"status,omitempty"`
}

type ConfigBackupUserRef struct {
	Email string `json:"email" yaml:"email"`
}

type ConfigBackupAlerts struct {
	Definitions []ConfigBackupAlertDefinition `json:"definitions,omitempty" yaml:"definitions,omitempty"`
	QuietHours  []ConfigBackupQuietHour       `json:"quietHours,omitempty" yaml:"quietHours,omitempty"`
}

type ConfigBackupAlertDefinition struct {
	StableID       string  `json:"stableId" yaml:"stableId"`
	SystemStableID string  `json:"systemStableId" yaml:"systemStableId"`
	UserEmail      string  `json:"userEmail" yaml:"userEmail"`
	Name           string  `json:"name" yaml:"name"`
	Min            float64 `json:"min" yaml:"min"`
	Value          float64 `json:"value" yaml:"value"`
	Triggered      bool    `json:"triggered,omitempty" yaml:"triggered,omitempty"`
}

type ConfigBackupQuietHour struct {
	StableID       string `json:"stableId" yaml:"stableId"`
	UserEmail      string `json:"userEmail" yaml:"userEmail"`
	SystemStableID string `json:"systemStableId,omitempty" yaml:"systemStableId,omitempty"`
	Type           string `json:"type" yaml:"type"`
	Start          string `json:"start" yaml:"start"`
	End            string `json:"end" yaml:"end"`
}

type ConfigBackupNotifications struct {
	UserSettings []ConfigBackupUserNotificationSettings `json:"userSettings,omitempty" yaml:"userSettings,omitempty"`
	Telegram     ConfigBackupTelegramNotifications      `json:"telegram,omitempty" yaml:"telegram,omitempty"`
}

type ConfigBackupUserNotificationSettings struct {
	UserEmail string               `json:"userEmail" yaml:"userEmail"`
	Emails    []string             `json:"emails,omitempty" yaml:"emails,omitempty"`
	Webhooks  []ConfigBackupSecret `json:"webhooks,omitempty" yaml:"webhooks,omitempty"`
}

type ConfigBackupTelegramNotifications struct {
	Settings     ConfigBackupTelegramSettings      `json:"settings,omitempty" yaml:"settings,omitempty"`
	Destinations []ConfigBackupTelegramDestination `json:"destinations,omitempty" yaml:"destinations,omitempty"`
}

type ConfigBackupTelegramSettings struct {
	Enabled        bool                `json:"enabled" yaml:"enabled"`
	PollingEnabled bool                `json:"pollingEnabled" yaml:"pollingEnabled"`
	BotUsername    string              `json:"botUsername,omitempty" yaml:"botUsername,omitempty"`
	BotToken       *ConfigBackupSecret `json:"botToken,omitempty" yaml:"botToken,omitempty"`
}

type ConfigBackupTelegramDestination struct {
	StableID        string   `json:"stableId" yaml:"stableId"`
	UserEmail       string   `json:"userEmail,omitempty" yaml:"userEmail,omitempty"`
	Name            string   `json:"name" yaml:"name"`
	ChatID          string   `json:"chatId" yaml:"chatId"`
	ChatType        string   `json:"chatType" yaml:"chatType"`
	Role            string   `json:"role" yaml:"role"`
	Enabled         bool     `json:"enabled" yaml:"enabled"`
	NodeScope       []string `json:"nodeScope,omitempty" yaml:"nodeScope,omitempty"`
	AlertLevelScope []string `json:"alertLevelScope,omitempty" yaml:"alertLevelScope,omitempty"`
	MuteUntil       string   `json:"muteUntil,omitempty" yaml:"muteUntil,omitempty"`
}

type ConfigBackupPublicStatus struct {
	Systems []ConfigBackupPublicSystem `json:"systems,omitempty" yaml:"systems,omitempty"`
}

type ConfigBackupPublicSystem struct {
	SystemStableID       string   `json:"systemStableId" yaml:"systemStableId"`
	PublicEnabled        bool     `json:"publicEnabled" yaml:"publicEnabled"`
	PublicName           string   `json:"publicName,omitempty" yaml:"publicName,omitempty"`
	ShowCPU              bool     `json:"showCpu" yaml:"showCpu"`
	ShowMemory           bool     `json:"showMemory" yaml:"showMemory"`
	ShowDisk             bool     `json:"showDisk" yaml:"showDisk"`
	PublicProbeStableIDs []string `json:"publicProbeStableIds,omitempty" yaml:"publicProbeStableIds,omitempty"`
}

type ConfigBackupNetworkProbes struct {
	Probes      []ConfigBackupNetworkProbe           `json:"probes,omitempty" yaml:"probes,omitempty"`
	Assignments []ConfigBackupNetworkProbeAssignment `json:"assignments,omitempty" yaml:"assignments,omitempty"`
}

type ConfigBackupNetworkProbe struct {
	StableID        string `json:"stableId" yaml:"stableId"`
	Name            string `json:"name" yaml:"name"`
	Type            string `json:"type" yaml:"type"`
	Target          string `json:"target" yaml:"target"`
	IntervalSeconds int    `json:"intervalSeconds" yaml:"intervalSeconds"`
	TimeoutSeconds  int    `json:"timeoutSeconds" yaml:"timeoutSeconds"`
	Enabled         bool   `json:"enabled" yaml:"enabled"`
	Scope           string `json:"scope" yaml:"scope"`
}

type ConfigBackupNetworkProbeAssignment struct {
	ProbeStableID  string `json:"probeStableId" yaml:"probeStableId"`
	SystemStableID string `json:"systemStableId" yaml:"systemStableId"`
	Enabled        bool   `json:"enabled" yaml:"enabled"`
}

type ConfigBackupPreviewResponse struct {
	PreviewID          string                     `json:"previewId"`
	Mode               string                     `json:"mode"`
	BackupMeta         ConfigBackupMeta           `json:"backupMeta"`
	Summary            ConfigBackupPreviewSummary `json:"summary"`
	Items              []ConfigBackupPreviewItem  `json:"items"`
	Warnings           []string                   `json:"warnings"`
	RequiresCredential bool                       `json:"requiresCredential"`
}

type ConfigBackupPreviewSummary struct {
	Create   int `json:"create"`
	Update   int `json:"update"`
	Preserve int `json:"preserve"`
	Skip     int `json:"skip"`
	Conflict int `json:"conflict"`
	Error    int `json:"error"`
}

type ConfigBackupPreviewItem struct {
	Section     string `json:"section"`
	StableID    string `json:"stableId,omitempty"`
	DisplayName string `json:"displayName,omitempty"`
	Action      string `json:"action"`
	Reason      string `json:"reason"`
}
