package hub

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"slices"
	"strings"
	"time"

	"github.com/henrygd/beszel/internal/entities/system"
	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase/core"
)

const (
	CollectionPublicSystemVisibility = "public_system_visibility"
	publicNameMaxLength              = 120
)

type publicChartRange struct {
	Name      string
	Duration  time.Duration
	StatsType string
}

var publicChartRanges = map[string]publicChartRange{
	"1m":  {Name: "1m", Duration: time.Minute, StatsType: "1m"},
	"30m": {Name: "30m", Duration: 30 * time.Minute, StatsType: "1m"},
	"1h":  {Name: "1h", Duration: time.Hour, StatsType: "1m"},
	"12h": {Name: "12h", Duration: 12 * time.Hour, StatsType: "10m"},
	"24h": {Name: "24h", Duration: 24 * time.Hour, StatsType: "20m"},
	"1w":  {Name: "1w", Duration: 7 * 24 * time.Hour, StatsType: "120m"},
	"30d": {Name: "30d", Duration: 30 * 24 * time.Hour, StatsType: "480m"},
}

type PublicSystemVisibility struct {
	SystemID       string
	PublicEnabled  bool
	PublicName     string
	ShowCPU        bool
	ShowMemory     bool
	ShowDisk       bool
	PublicProbeIDs []string
}

type PublicVisibilityInput struct {
	PublicEnabled  bool     `json:"publicEnabled"`
	PublicName     string   `json:"publicName"`
	ShowCPU        *bool    `json:"showCpu,omitempty"`
	ShowMemory     *bool    `json:"showMemory,omitempty"`
	ShowDisk       *bool    `json:"showDisk,omitempty"`
	PublicProbeIDs []string `json:"publicProbeIds,omitempty"`
}

type PublicMetrics struct {
	CPUPercent    *float64 `json:"cpuPercent,omitempty"`
	MemoryPercent *float64 `json:"memoryPercent,omitempty"`
	DiskPercent   *float64 `json:"diskPercent,omitempty"`
	Unavailable   []string `json:"unavailable,omitempty"`
}

type PublicMetricPoint struct {
	Created       string   `json:"created"`
	CPUPercent    *float64 `json:"cpuPercent,omitempty"`
	MemoryPercent *float64 `json:"memoryPercent,omitempty"`
	DiskPercent   *float64 `json:"diskPercent,omitempty"`
}

type PublicProbeLatest struct {
	Success         bool     `json:"success"`
	LatencyMs       *float64 `json:"latencyMs,omitempty"`
	Error           string   `json:"error,omitempty"`
	FailureCategory string   `json:"failureCategory,omitempty"`
	Created         string   `json:"created"`
}

type PublicProbeSeriesPoint struct {
	Created         string   `json:"created"`
	Success         bool     `json:"success"`
	LatencyMs       *float64 `json:"latencyMs,omitempty"`
	Error           string   `json:"error,omitempty"`
	FailureCategory string   `json:"failureCategory,omitempty"`
}

type PublicProbeSummary struct {
	ID     string                   `json:"id"`
	Name   string                   `json:"name"`
	Type   string                   `json:"type"`
	Latest *PublicProbeLatest       `json:"latest,omitempty"`
	Series []PublicProbeSeriesPoint `json:"series"`
}

type PublicSystemSummary struct {
	ID        string               `json:"id"`
	Name      string               `json:"name"`
	Status    string               `json:"status"`
	Freshness string               `json:"freshness,omitempty"`
	Updated   string               `json:"updated,omitempty"`
	Metrics   PublicMetrics        `json:"metrics"`
	History   []PublicMetricPoint  `json:"history,omitempty"`
	Probes    []PublicProbeSummary `json:"probes"`
}

type PublicStatusResponse struct {
	GeneratedAt string                `json:"generatedAt"`
	Systems     []PublicSystemSummary `json:"systems"`
}

type AdminPublicSystemResponse struct {
	ID             string   `json:"id"`
	Name           string   `json:"name"`
	Status         string   `json:"status"`
	PublicEnabled  bool     `json:"publicEnabled"`
	PublicName     string   `json:"publicName"`
	ShowCPU        bool     `json:"showCpu"`
	ShowMemory     bool     `json:"showMemory"`
	ShowDisk       bool     `json:"showDisk"`
	PublicProbeIDs []string `json:"publicProbeIds"`
}

type AdminPublicSystemsResponse struct {
	Systems []AdminPublicSystemResponse `json:"systems"`
}

type PublicVisibilityResponse struct {
	ID             string   `json:"id"`
	PublicEnabled  bool     `json:"publicEnabled"`
	PublicName     string   `json:"publicName"`
	ShowCPU        bool     `json:"showCpu"`
	ShowMemory     bool     `json:"showMemory"`
	ShowDisk       bool     `json:"showDisk"`
	PublicProbeIDs []string `json:"publicProbeIds"`
}

func normalizePublicVisibilityInput(input PublicVisibilityInput) (PublicSystemVisibility, error) {
	publicName := strings.TrimSpace(input.PublicName)
	if len(publicName) > publicNameMaxLength {
		return PublicSystemVisibility{}, errors.New("publicName is too long")
	}
	return PublicSystemVisibility{
		PublicEnabled:  input.PublicEnabled,
		PublicName:     publicName,
		ShowCPU:        boolValue(input.ShowCPU, true),
		ShowMemory:     boolValue(input.ShowMemory, true),
		ShowDisk:       boolValue(input.ShowDisk, true),
		PublicProbeIDs: normalizePublicProbeIDs(input.PublicProbeIDs),
	}, nil
}

func normalizePublicProbeIDs(probeIDs []string) []string {
	if len(probeIDs) == 0 {
		return []string{}
	}
	normalized := make([]string, 0, len(probeIDs))
	seen := make(map[string]struct{}, len(probeIDs))
	for _, probeID := range probeIDs {
		probeID = strings.TrimSpace(probeID)
		if probeID == "" {
			continue
		}
		if _, ok := seen[probeID]; ok {
			continue
		}
		seen[probeID] = struct{}{}
		normalized = append(normalized, probeID)
	}
	return normalized
}

func publicProbeIDsFromRecord(record *core.Record) []string {
	if record == nil {
		return []string{}
	}
	raw := record.Get("public_probe_ids")
	switch value := raw.(type) {
	case []string:
		return normalizePublicProbeIDs(value)
	case []any:
		probeIDs := make([]string, 0, len(value))
		for _, item := range value {
			if text, ok := item.(string); ok {
				probeIDs = append(probeIDs, text)
			}
		}
		return normalizePublicProbeIDs(probeIDs)
	case string:
		if strings.TrimSpace(value) == "" {
			return []string{}
		}
	}
	return normalizePublicProbeIDs(record.GetStringSlice("public_probe_ids"))
}

func sanitizePublicSystem(record *core.Record, visibility PublicSystemVisibility, probes []PublicProbeSummary, history []PublicMetricPoint) PublicSystemSummary {
	name := strings.TrimSpace(visibility.PublicName)
	if name == "" {
		name = record.GetString("name")
	}
	info, hasInfo := systemInfoFromRecord(record)
	status := record.GetString("status")
	if status == "pending" || status == "" {
		status = "stale"
	}
	freshness := publicFreshnessFromRecord(record)
	summary := PublicSystemSummary{
		ID:        record.Id,
		Name:      name,
		Status:    status,
		Freshness: freshness,
		Updated:   freshness,
		Metrics:   PublicMetrics{},
		History:   history,
		Probes:    probes,
	}
	summary.Metrics = publicMetricsFromRecord(info, visibility, hasInfo)
	if summary.Metrics.CPUPercent == nil || summary.Metrics.MemoryPercent == nil || summary.Metrics.DiskPercent == nil || freshness == "" {
		summary.Metrics.Unavailable = publicMetricUnavailable(summary.Metrics, freshness == "")
	}
	return summary
}

func publicMetricsFromRecord(info system.Info, visibility PublicSystemVisibility, hasInfo bool) PublicMetrics {
	metrics := PublicMetrics{}
	if !hasInfo {
		return metrics
	}
	if visibility.ShowCPU {
		metrics.CPUPercent = &info.Cpu
	}
	if visibility.ShowMemory {
		metrics.MemoryPercent = &info.MemPct
	}
	if visibility.ShowDisk {
		metrics.DiskPercent = &info.DiskPct
	}
	return metrics
}

func publicMetricUnavailable(metrics PublicMetrics, freshnessUnavailable bool) []string {
	unavailable := make([]string, 0, 3)
	if metrics.CPUPercent == nil {
		unavailable = append(unavailable, "cpu")
	}
	if metrics.MemoryPercent == nil {
		unavailable = append(unavailable, "memory")
	}
	if metrics.DiskPercent == nil {
		unavailable = append(unavailable, "disk")
	}
	if freshnessUnavailable {
		unavailable = append(unavailable, "freshness")
	}
	return unavailable
}

func publicFreshnessFromRecord(record *core.Record) string {
	raw := strings.TrimSpace(record.GetString("updated"))
	if raw != "" {
		if parsed, err := time.Parse("2006-01-02 15:04:05.000Z", raw); err == nil {
			return parsed.UTC().Format(time.RFC3339)
		}
		if parsed, err := time.Parse(time.RFC3339Nano, raw); err == nil {
			return parsed.UTC().Format(time.RFC3339)
		}
	}

	updated := record.GetDateTime("updated")
	if updated.IsZero() {
		return ""
	}
	return updated.Time().UTC().Format(time.RFC3339)
}

func systemInfoFromRecord(record *core.Record) (system.Info, bool) {
	var info system.Info
	raw := record.Get("info")
	switch value := raw.(type) {
	case system.Info:
		return value, true
	case map[string]any:
		if len(value) == 0 {
			return info, false
		}
		data, _ := json.Marshal(value)
		if err := json.Unmarshal(data, &info); err != nil {
			return info, false
		}
		return info, true
	case []byte:
		if len(value) == 0 {
			return info, false
		}
		if err := json.Unmarshal(value, &info); err != nil {
			return info, false
		}
		return info, true
	case string:
		if strings.TrimSpace(value) == "" {
			return info, false
		}
		if err := json.Unmarshal([]byte(value), &info); err != nil {
			return info, false
		}
		return info, true
	}
	if value := strings.TrimSpace(record.GetString("info")); value != "" {
		if err := json.Unmarshal([]byte(value), &info); err == nil {
			return info, true
		}
	}
	return info, false
}

func parsePublicChartRange(value string) (publicChartRange, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		value = "30m"
	}
	if rangeSpec, ok := publicChartRanges[value]; ok {
		return rangeSpec, nil
	}
	return publicChartRange{}, fmt.Errorf("unsupported public chart range %q", value)
}

func (h *Hub) getPublicStatus(e *core.RequestEvent) error {
	rangeValue := e.Request.URL.Query().Get("range")
	rangeSpec, err := parsePublicChartRange(rangeValue)
	if err != nil {
		return e.BadRequestError("Invalid public chart range.", map[string]string{"range": rangeValue})
	}
	visibilityRecords, err := h.FindRecordsByFilter(CollectionPublicSystemVisibility, "public_enabled = true", "", -1, 0)
	if err != nil {
		return err
	}
	systems := make([]PublicSystemSummary, 0, len(visibilityRecords))
	for _, visibilityRecord := range visibilityRecords {
		visibility := publicVisibilityFromRecord(visibilityRecord)
		systemRecord, err := h.FindRecordById("systems", visibility.SystemID)
		if err != nil {
			continue
		}
		probes, err := h.publicProbeSummaries(systemRecord.Id, visibility, rangeSpec)
		if err != nil {
			return err
		}
		history, err := h.publicMetricHistory(systemRecord.Id, visibility, rangeSpec)
		if err != nil {
			return err
		}
		systems = append(systems, sanitizePublicSystem(systemRecord, visibility, probes, history))
	}
	return e.JSON(http.StatusOK, PublicStatusResponse{
		GeneratedAt: time.Now().UTC().Format(time.RFC3339),
		Systems:     systems,
	})
}

func (h *Hub) listPublicSystems(e *core.RequestEvent) error {
	systemRecords, err := h.FindRecordsByFilter("systems", "id != ''", "name", -1, 0)
	if err != nil {
		return err
	}
	systems := make([]AdminPublicSystemResponse, 0, len(systemRecords))
	for _, systemRecord := range systemRecords {
		visibility, _ := h.findPublicVisibility(systemRecord.Id)
		systems = append(systems, AdminPublicSystemResponse{
			ID:             systemRecord.Id,
			Name:           systemRecord.GetString("name"),
			Status:         systemRecord.GetString("status"),
			PublicEnabled:  visibility.PublicEnabled,
			PublicName:     visibility.PublicName,
			ShowCPU:        visibility.ShowCPU,
			ShowMemory:     visibility.ShowMemory,
			ShowDisk:       visibility.ShowDisk,
			PublicProbeIDs: visibility.PublicProbeIDs,
		})
	}
	return e.JSON(http.StatusOK, AdminPublicSystemsResponse{Systems: systems})
}

func (h *Hub) updatePublicSystem(e *core.RequestEvent) error {
	systemID := e.Request.PathValue("systemId")
	systemRecord, err := h.FindRecordById("systems", systemID)
	if err != nil {
		return e.NotFoundError("System not found.", err)
	}
	if e.Auth.GetString("role") != "admin" && !slices.Contains(systemRecord.GetStringSlice("users"), e.Auth.Id) {
		return e.NotFoundError("System not found.", nil)
	}
	var input PublicVisibilityInput
	if err := e.BindBody(&input); err != nil {
		return e.BadRequestError("Invalid public visibility request.", err)
	}
	visibility, err := normalizePublicVisibilityInput(input)
	if err != nil {
		return e.BadRequestError(err.Error(), err)
	}
	if input.PublicProbeIDs == nil {
		existing, _ := h.findPublicVisibility(systemID)
		visibility.PublicProbeIDs = existing.PublicProbeIDs
	}
	if _, err := h.validatePublicProbeSelection(systemID, visibility.PublicProbeIDs); err != nil {
		return e.BadRequestError(err.Error(), err)
	}
	record, err := h.upsertPublicVisibility(systemID, visibility)
	if err != nil {
		return err
	}
	visibility = publicVisibilityFromRecord(record)
	return e.JSON(http.StatusOK, PublicVisibilityResponse{
		ID:             systemID,
		PublicEnabled:  visibility.PublicEnabled,
		PublicName:     visibility.PublicName,
		ShowCPU:        visibility.ShowCPU,
		ShowMemory:     visibility.ShowMemory,
		ShowDisk:       visibility.ShowDisk,
		PublicProbeIDs: visibility.PublicProbeIDs,
	})
}

func (h *Hub) findPublicVisibility(systemID string) (PublicSystemVisibility, *core.Record) {
	record, err := h.FindFirstRecordByFilter(CollectionPublicSystemVisibility, "system = {:system}", dbx.Params{"system": systemID})
	if err != nil {
		return PublicSystemVisibility{SystemID: systemID, ShowCPU: true, ShowMemory: true, ShowDisk: true, PublicProbeIDs: []string{}}, nil
	}
	return publicVisibilityFromRecord(record), record
}

func publicVisibilityFromRecord(record *core.Record) PublicSystemVisibility {
	return PublicSystemVisibility{
		SystemID:       record.GetString("system"),
		PublicEnabled:  record.GetBool("public_enabled"),
		PublicName:     record.GetString("public_name"),
		ShowCPU:        record.GetBool("show_cpu"),
		ShowMemory:     record.GetBool("show_memory"),
		ShowDisk:       record.GetBool("show_disk"),
		PublicProbeIDs: publicProbeIDsFromRecord(record),
	}
}

func (h *Hub) upsertPublicVisibility(systemID string, visibility PublicSystemVisibility) (*core.Record, error) {
	_, record := h.findPublicVisibility(systemID)
	if record == nil {
		collection, err := h.FindCachedCollectionByNameOrId(CollectionPublicSystemVisibility)
		if err != nil {
			return nil, err
		}
		record = core.NewRecord(collection)
		record.Set("system", systemID)
	}
	record.Set("public_enabled", visibility.PublicEnabled)
	record.Set("public_name", visibility.PublicName)
	record.Set("show_cpu", visibility.ShowCPU)
	record.Set("show_memory", visibility.ShowMemory)
	record.Set("show_disk", visibility.ShowDisk)
	record.Set("public_probe_ids", normalizePublicProbeIDs(visibility.PublicProbeIDs))
	if err := h.Save(record); err != nil {
		return nil, err
	}
	return record, nil
}

func (h *Hub) validatePublicProbeSelection(systemID string, probeIDs []string) ([]string, error) {
	if len(probeIDs) == 0 {
		return []string{}, nil
	}
	assignments, err := h.effectiveNetworkProbeAssignments(systemID)
	if err != nil {
		return nil, err
	}
	available := make(map[string]struct{}, len(assignments))
	for _, assignment := range assignments {
		available[assignment.ProbeID] = struct{}{}
	}
	normalized := normalizePublicProbeIDs(probeIDs)
	for _, probeID := range normalized {
		if _, ok := available[probeID]; !ok {
			return nil, fmt.Errorf("probe %s does not cover system %s", probeID, systemID)
		}
		if _, err := h.FindRecordById(CollectionNetworkProbes, probeID); err != nil {
			return nil, fmt.Errorf("probe %s not found", probeID)
		}
	}
	return normalized, nil
}

func (h *Hub) publicProbeSummaries(systemID string, visibility PublicSystemVisibility, rangeSpec publicChartRange) ([]PublicProbeSummary, error) {
	if len(visibility.PublicProbeIDs) == 0 {
		return []PublicProbeSummary{}, nil
	}
	assignments, err := h.effectiveNetworkProbeAssignments(systemID)
	if err != nil {
		return nil, err
	}
	selected := make(map[string]struct{}, len(visibility.PublicProbeIDs))
	for _, probeID := range visibility.PublicProbeIDs {
		selected[probeID] = struct{}{}
	}
	covered := make(map[string]struct{}, len(assignments))
	for _, assignment := range assignments {
		if _, ok := selected[assignment.ProbeID]; ok {
			covered[assignment.ProbeID] = struct{}{}
		}
	}
	summaries := make([]PublicProbeSummary, 0, len(visibility.PublicProbeIDs))
	for _, probeID := range visibility.PublicProbeIDs {
		if _, ok := covered[probeID]; !ok {
			continue
		}
		probe, err := h.FindRecordById(CollectionNetworkProbes, probeID)
		if err != nil || !probe.GetBool("enabled") {
			continue
		}
		latestResults, err := h.FindRecordsByFilter(CollectionNetworkProbeResults, "probe = {:probe} && system = {:system}", "-created", 1, 0, dbx.Params{
			"probe":  probe.Id,
			"system": systemID,
		})
		if err != nil {
			return nil, err
		}
		results, err := h.FindRecordsByFilter(
			CollectionNetworkProbeResults,
			"probe = {:probe} && system = {:system} && created >= {:created}",
			"created",
			-1,
			0,
			dbx.Params{
				"probe":   probe.Id,
				"system":  systemID,
				"created": time.Now().UTC().Add(-rangeSpec.Duration),
			},
		)
		if err != nil {
			return nil, err
		}
		summary := PublicProbeSummary{
			ID:     probe.Id,
			Name:   probe.GetString("name"),
			Type:   probe.GetString("type"),
			Series: make([]PublicProbeSeriesPoint, 0, len(results)),
		}
		for _, result := range results {
			point := publicProbeSeriesPoint(result)
			summary.Series = append(summary.Series, point)
		}
		if len(latestResults) > 0 {
			summary.Latest = publicProbeLatest(latestResults[0])
		}
		summaries = append(summaries, summary)
	}
	return summaries, nil
}

func publicProbeSeriesPoint(result *core.Record) PublicProbeSeriesPoint {
	errorMessage := result.GetString("error")
	failureCategory := ""
	if !result.GetBool("success") {
		failureCategory = normalizeProbeFailureCategory(result.GetString("failure_category"), errorMessage)
	}
	point := PublicProbeSeriesPoint{
		Created:         result.GetDateTime("created").Time().UTC().Format(time.RFC3339),
		Success:         result.GetBool("success"),
		LatencyMs:       optionalFloat(result, "latency_ms"),
		FailureCategory: failureCategory,
	}
	if !point.Success {
		point.Error = safeProbeResultError(errorMessage, failureCategory)
	}
	return point
}

func publicProbeLatest(result *core.Record) *PublicProbeLatest {
	point := publicProbeSeriesPoint(result)
	latest := &PublicProbeLatest{
		Success:         point.Success,
		LatencyMs:       point.LatencyMs,
		FailureCategory: point.FailureCategory,
		Created:         point.Created,
	}
	if !latest.Success {
		latest.Error = point.Error
	}
	return latest
}

func (h *Hub) publicMetricHistory(systemID string, visibility PublicSystemVisibility, rangeSpec publicChartRange) ([]PublicMetricPoint, error) {
	records, err := h.FindRecordsByFilter(
		"system_stats",
		"system = {:system} && type = {:type} && created >= {:created}",
		"created",
		-1,
		0,
		dbx.Params{
			"system":  systemID,
			"type":    rangeSpec.StatsType,
			"created": time.Now().UTC().Add(-rangeSpec.Duration),
		},
	)
	if err != nil {
		return nil, err
	}
	history := make([]PublicMetricPoint, 0, len(records))
	for _, record := range records {
		stats, ok := systemStatsFromRecord(record)
		if !ok {
			continue
		}
		point := PublicMetricPoint{
			Created: record.GetDateTime("created").Time().UTC().Format(time.RFC3339),
		}
		if visibility.ShowCPU {
			point.CPUPercent = &stats.Cpu
		}
		if visibility.ShowMemory {
			point.MemoryPercent = &stats.MemPct
		}
		if visibility.ShowDisk {
			point.DiskPercent = &stats.DiskPct
		}
		history = append(history, point)
	}
	return history, nil
}

func systemStatsFromRecord(record *core.Record) (system.Stats, bool) {
	var stats system.Stats
	raw := record.Get("stats")
	switch value := raw.(type) {
	case system.Stats:
		return value, true
	case map[string]any:
		if len(value) == 0 {
			return stats, false
		}
		data, _ := json.Marshal(value)
		if err := json.Unmarshal(data, &stats); err != nil {
			return stats, false
		}
		return stats, true
	case []byte:
		if len(value) == 0 {
			return stats, false
		}
		if err := json.Unmarshal(value, &stats); err != nil {
			return stats, false
		}
		return stats, true
	case string:
		if strings.TrimSpace(value) == "" {
			return stats, false
		}
		if err := json.Unmarshal([]byte(value), &stats); err != nil {
			return stats, false
		}
		return stats, true
	}
	if value := strings.TrimSpace(record.GetString("stats")); value != "" {
		if err := json.Unmarshal([]byte(value), &stats); err == nil {
			return stats, true
		}
	}
	return stats, false
}
