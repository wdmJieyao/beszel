package hub

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/url"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/henrygd/beszel/internal/common"
	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase/core"
)

const (
	CollectionNetworkProbes           = "network_probes"
	CollectionNetworkProbeAssignments = "network_probe_assignments"
	CollectionNetworkProbeResults     = "network_probe_results"

	NetworkProbeTypeTCPing   = string(common.NetworkProbeTCPing)
	NetworkProbeTypeICMPPing = string(common.NetworkProbeICMPPing)
	NetworkProbeTypeHTTPGet  = string(common.NetworkProbeHTTPGet)

	NetworkProbeScopeGlobal = "global"
	NetworkProbeScopeFixed  = "fixed"

	defaultProbeIntervalSeconds = 10
	defaultProbeTimeoutSeconds  = 5
	minProbeIntervalSeconds     = 10
	liveProbeTimeoutSeconds     = 1
)

const (
	ProbeFailureInvalidTarget            = string(common.NetworkProbeFailureInvalidTarget)
	ProbeFailureDNSFailure               = string(common.NetworkProbeFailureDNSFailure)
	ProbeFailureTimeout                  = string(common.NetworkProbeFailureTimeout)
	ProbeFailureConnectionRefused        = string(common.NetworkProbeFailureConnectionRefused)
	ProbeFailureTargetUnreachable        = string(common.NetworkProbeFailureTargetUnreachable)
	ProbeFailureExecutionNodeUnavailable = string(common.NetworkProbeFailureExecutionNodeUnavailable)
	ProbeFailureUnsupported              = string(common.NetworkProbeFailureUnsupported)
	ProbeFailureUnknown                  = string(common.NetworkProbeFailureUnknown)
)

var validNetworkProbeTypes = []common.NetworkProbeType{
	common.NetworkProbeTCPing,
	common.NetworkProbeICMPPing,
	common.NetworkProbeHTTPGet,
}

type NetworkProbeInput struct {
	Name            string   `json:"name"`
	Type            string   `json:"type"`
	Target          string   `json:"target"`
	IntervalSeconds int      `json:"intervalSeconds"`
	TimeoutSeconds  int      `json:"timeoutSeconds"`
	Enabled         *bool    `json:"enabled,omitempty"`
	PublicVisible   *bool    `json:"publicVisible,omitempty"`
	Scope           string   `json:"scope,omitempty"`
	Systems         []string `json:"systems"`
}

type NetworkProbeResponse struct {
	ID              string   `json:"id"`
	Name            string   `json:"name"`
	Type            string   `json:"type"`
	Target          string   `json:"target"`
	IntervalSeconds int      `json:"intervalSeconds"`
	TimeoutSeconds  int      `json:"timeoutSeconds"`
	Enabled         bool     `json:"enabled"`
	PublicVisible   bool     `json:"publicVisible"`
	Scope           string   `json:"scope"`
	Systems         []string `json:"systems"`
}

type NetworkProbeListResponse struct {
	Probes []NetworkProbeResponse `json:"probes"`
}

type NetworkProbeResultPoint struct {
	SystemID        string   `json:"systemId"`
	Created         string   `json:"created"`
	Success         bool     `json:"success"`
	LatencyMs       *float64 `json:"latencyMs"`
	PacketLossPct   *float64 `json:"packetLossPercent"`
	HTTPStatus      *int     `json:"httpStatus"`
	FailureCategory string   `json:"failureCategory,omitempty"`
	Error           string   `json:"error,omitempty"`
	ProbeType       string   `json:"type,omitempty"`
	ProbeTarget     string   `json:"target,omitempty"`
	RetentionBucket string   `json:"bucket,omitempty"`
}

type NetworkProbeResultsResponse struct {
	ProbeID string                    `json:"probeId"`
	Series  []NetworkProbeResultPoint `json:"series"`
}

type networkProbeConfig struct {
	Name            string
	Type            common.NetworkProbeType
	Target          string
	IntervalSeconds int
	TimeoutSeconds  int
	PublicVisible   bool
	Enabled         bool
	Scope           string
	Systems         []string
}

type networkProbeAssignment struct {
	ID       string
	ProbeID  string
	SystemID string
	Enabled  bool
}

func ValidateNetworkProbeInput(input NetworkProbeInput) error {
	cfg := networkProbeConfig{
		Name:            input.Name,
		Type:            common.NetworkProbeType(input.Type),
		Target:          input.Target,
		IntervalSeconds: input.IntervalSeconds,
		TimeoutSeconds:  input.TimeoutSeconds,
	}
	if cfg.IntervalSeconds == 0 {
		cfg.IntervalSeconds = defaultProbeIntervalSeconds
	}
	if cfg.TimeoutSeconds == 0 {
		cfg.TimeoutSeconds = defaultProbeTimeoutSeconds
	}
	if cfg.IntervalSeconds < minProbeIntervalSeconds {
		return fmt.Errorf("intervalSeconds must be at least %d", minProbeIntervalSeconds)
	}
	return validateNetworkProbeConfig(cfg)
}

func validateNetworkProbeConfig(cfg networkProbeConfig) error {
	if strings.TrimSpace(cfg.Name) == "" {
		return errors.New("probe name is required")
	}
	if cfg.IntervalSeconds <= 0 {
		return errors.New("interval must be positive")
	}
	if cfg.TimeoutSeconds <= 0 {
		return errors.New("timeout must be positive")
	}
	if cfg.TimeoutSeconds >= cfg.IntervalSeconds {
		return errors.New("timeout must be less than interval")
	}
	if !slices.Contains(validNetworkProbeTypes, cfg.Type) {
		return errors.New("unsupported probe type")
	}
	switch cfg.Type {
	case common.NetworkProbeTCPing:
		host, port, err := splitTCPingTarget(cfg.Target)
		if err != nil || strings.TrimSpace(host) == "" || strings.TrimSpace(port) == "" {
			return errors.New("tcping target must use host:port format")
		}
		portNum, err := strconv.ParseUint(port, 10, 16)
		if err != nil || portNum == 0 {
			return errors.New("tcping target must include a valid port from 1 to 65535")
		}
	case common.NetworkProbeICMPPing:
		if strings.TrimSpace(cfg.Target) == "" || strings.Contains(cfg.Target, "://") {
			return errors.New("icmp target must be host or ip")
		}
		if _, _, err := net.SplitHostPort(cfg.Target); err == nil {
			return errors.New("icmp target must not include a port")
		}
	case common.NetworkProbeHTTPGet:
		u, err := url.Parse(cfg.Target)
		if err != nil || u.Host == "" || (u.Scheme != "http" && u.Scheme != "https") {
			return errors.New("http target must be http or https url")
		}
	}
	return nil
}

func splitTCPingTarget(target string) (string, string, error) {
	target = strings.TrimSpace(target)
	host, port, err := net.SplitHostPort(target)
	if err == nil {
		return host, port, nil
	}
	if strings.Count(target, ":") == 1 && !strings.Contains(target, "://") {
		parts := strings.SplitN(target, ":", 2)
		return parts[0], parts[1], nil
	}
	return "", "", err
}

func normalizeProbeFailureCategory(category string, message string) string {
	category = strings.TrimSpace(strings.ToLower(category))
	switch category {
	case ProbeFailureInvalidTarget,
		ProbeFailureDNSFailure,
		ProbeFailureTimeout,
		ProbeFailureConnectionRefused,
		ProbeFailureTargetUnreachable,
		ProbeFailureExecutionNodeUnavailable,
		ProbeFailureUnsupported,
		ProbeFailureUnknown:
		return category
	}

	message = strings.ToLower(strings.TrimSpace(message))
	switch {
	case message == "":
		return ""
	case strings.Contains(message, "agent offline"), strings.Contains(message, "execution node unavailable"), strings.Contains(message, "no compatible probe transport"):
		return ProbeFailureExecutionNodeUnavailable
	case strings.Contains(message, "unsupported probe type"), strings.Contains(message, "unsupported"):
		return ProbeFailureUnsupported
	case strings.Contains(message, "invalid target"), strings.Contains(message, "host:port"), strings.Contains(message, "missing port"):
		return ProbeFailureInvalidTarget
	case strings.Contains(message, "dns failure"), strings.Contains(message, "no such host"):
		return ProbeFailureDNSFailure
	case strings.Contains(message, "timeout"), strings.Contains(message, "deadline exceeded"), strings.Contains(message, "i/o timeout"):
		return ProbeFailureTimeout
	case strings.Contains(message, "connection refused"):
		return ProbeFailureConnectionRefused
	case strings.Contains(message, "network is unreachable"), strings.Contains(message, "no route to host"), strings.Contains(message, "host is unreachable"), strings.Contains(message, "cannot assign requested address"):
		return ProbeFailureTargetUnreachable
	case strings.Contains(message, "invalid target"), strings.Contains(message, "host:port"), strings.Contains(message, "missing port"):
		return ProbeFailureInvalidTarget
	default:
		return ProbeFailureUnknown
	}
}

func safeProbeFailureLabel(category string) string {
	switch normalizeProbeFailureCategory(category, "") {
	case ProbeFailureInvalidTarget:
		return "invalid target"
	case ProbeFailureDNSFailure:
		return "dns failure"
	case ProbeFailureTimeout:
		return "timeout"
	case ProbeFailureConnectionRefused:
		return "connection refused"
	case ProbeFailureTargetUnreachable:
		return "target unreachable"
	case ProbeFailureExecutionNodeUnavailable:
		return "execution node unavailable"
	case ProbeFailureUnsupported:
		return "unsupported probe type"
	case ProbeFailureUnknown:
		return "probe failed"
	default:
		return ""
	}
}

func normalizeProbeInput(input NetworkProbeInput) NetworkProbeInput {
	input.Name = strings.TrimSpace(input.Name)
	input.Type = strings.TrimSpace(input.Type)
	input.Target = strings.TrimSpace(input.Target)
	input.Scope = normalizeNetworkProbeScope(input.Scope, input.Systems)
	if input.IntervalSeconds == 0 {
		input.IntervalSeconds = defaultProbeIntervalSeconds
	}
	if input.TimeoutSeconds == 0 {
		input.TimeoutSeconds = defaultProbeTimeoutSeconds
	}
	return input
}

func normalizeNetworkProbeScope(scope string, systems []string) string {
	switch strings.TrimSpace(strings.ToLower(scope)) {
	case NetworkProbeScopeGlobal:
		return NetworkProbeScopeGlobal
	case NetworkProbeScopeFixed:
		return NetworkProbeScopeFixed
	default:
		if len(systems) > 0 {
			return NetworkProbeScopeFixed
		}
		return NetworkProbeScopeGlobal
	}
}

func networkProbeScopeFromRecord(record *core.Record) string {
	return normalizeNetworkProbeScope(record.GetString("scope"), nil)
}

func boolValue(v *bool, fallback bool) bool {
	if v == nil {
		return fallback
	}
	return *v
}

func (h *Hub) listNetworkProbes(e *core.RequestEvent) error {
	probes, err := h.FindRecordsByFilter(CollectionNetworkProbes, "id != ''", "name", -1, 0)
	if err != nil {
		return err
	}
	responses := make([]NetworkProbeResponse, 0, len(probes))
	for _, probe := range probes {
		assignments, err := h.FindRecordsByFilter(CollectionNetworkProbeAssignments, "probe = {:probe}", "", -1, 0, dbx.Params{"probe": probe.Id})
		if err != nil {
			return err
		}
		responses = append(responses, networkProbeResponse(probe, assignments))
	}
	return e.JSON(200, NetworkProbeListResponse{Probes: responses})
}

func (h *Hub) createNetworkProbe(e *core.RequestEvent) error {
	var input NetworkProbeInput
	if err := e.BindBody(&input); err != nil {
		return e.BadRequestError("Invalid network probe request.", err)
	}
	input = normalizeProbeInput(input)
	if err := ValidateNetworkProbeInput(input); err != nil {
		return e.BadRequestError(err.Error(), err)
	}
	if input.Scope == NetworkProbeScopeFixed && len(input.Systems) == 0 {
		return e.BadRequestError("fixed scope requires at least one system", nil)
	}
	if input.Scope == NetworkProbeScopeFixed {
		if err := h.ensureSystemsVisibleToAuth(e.Auth, input.Systems); err != nil {
			return e.BadRequestError(err.Error(), err)
		}
	}
	if input.Scope == NetworkProbeScopeGlobal {
		input.Systems = nil
	}
	collection, err := h.FindCachedCollectionByNameOrId(CollectionNetworkProbes)
	if err != nil {
		return err
	}
	probe := core.NewRecord(collection)
	setProbeRecord(probe, input)
	if err := h.Save(probe); err != nil {
		return err
	}
	assignments, err := h.replaceProbeAssignments(probe.Id, input.Scope, input.Systems)
	if err != nil {
		return err
	}
	return e.JSON(201, networkProbeResponse(probe, assignments))
}

func (h *Hub) updateNetworkProbe(e *core.RequestEvent) error {
	probeID := e.Request.PathValue("probeId")
	probe, err := h.FindRecordById(CollectionNetworkProbes, probeID)
	if err != nil {
		return e.NotFoundError("Network probe not found.", err)
	}
	existingAssignments, err := h.FindRecordsByFilter(CollectionNetworkProbeAssignments, "probe = {:probe}", "", -1, 0, dbx.Params{"probe": probe.Id})
	if err != nil {
		return err
	}
	var input NetworkProbeInput
	if err := e.BindBody(&input); err != nil {
		return e.BadRequestError("Invalid network probe request.", err)
	}
	merged := networkProbeInputFromRecord(probe, existingAssignments)
	mergeProbeInput(&merged, input)
	merged = normalizeProbeInput(merged)
	if err := ValidateNetworkProbeInput(merged); err != nil {
		return e.BadRequestError(err.Error(), err)
	}
	if merged.Scope == NetworkProbeScopeFixed && len(merged.Systems) == 0 {
		return e.BadRequestError("fixed scope requires at least one system", nil)
	}
	if merged.Scope == NetworkProbeScopeFixed {
		if err := h.ensureSystemsVisibleToAuth(e.Auth, merged.Systems); err != nil {
			return e.BadRequestError(err.Error(), err)
		}
	}
	if merged.Scope == NetworkProbeScopeGlobal {
		merged.Systems = nil
	}
	setProbeRecord(probe, merged)
	if err := h.Save(probe); err != nil {
		return err
	}
	assignments, err := h.replaceProbeAssignments(probe.Id, merged.Scope, merged.Systems)
	if err != nil {
		return err
	}
	return e.JSON(200, networkProbeResponse(probe, assignments))
}

func (h *Hub) deleteNetworkProbe(e *core.RequestEvent) error {
	probeID := e.Request.PathValue("probeId")
	probe, err := h.FindRecordById(CollectionNetworkProbes, probeID)
	if err != nil {
		return e.NotFoundError("Network probe not found.", err)
	}
	if err := h.Delete(probe); err != nil {
		return err
	}
	return e.NoContent(204)
}

func (h *Hub) getNetworkProbeResults(e *core.RequestEvent) error {
	probeID := e.Request.PathValue("probeId")
	rangeValue := e.Request.URL.Query().Get("range")
	rangeSpec, err := parsePublicChartRange(rangeValue)
	if err != nil {
		return e.BadRequestError("Invalid network probe chart range.", map[string]string{"range": rangeValue})
	}
	if _, err := h.FindRecordById(CollectionNetworkProbes, probeID); err != nil {
		return e.NotFoundError("Network probe not found.", err)
	}
	systemID := e.Request.URL.Query().Get("system")
	results, err := h.compatibleProbeRangeRecords(probeID, systemID, rangeSpec)
	if err != nil {
		return err
	}
	series := make([]NetworkProbeResultPoint, 0, len(results))
	for _, result := range results {
		series = append(series, networkProbeResultPoint(result))
	}
	return e.JSON(200, NetworkProbeResultsResponse{ProbeID: probeID, Series: series})
}

func (h *Hub) ensureSystemsVisibleToAuth(auth *core.Record, systemIDs []string) error {
	if auth == nil {
		return errors.New("authentication required")
	}
	for _, systemID := range systemIDs {
		if err := h.ensureSystemVisibleToAuth(auth, systemID); err != nil {
			return err
		}
	}
	return nil
}

func (h *Hub) ensureSystemVisibleToAuth(auth *core.Record, systemID string) error {
	if auth == nil {
		return errors.New("authentication required")
	}
	if strings.TrimSpace(systemID) == "" {
		return errors.New("systems must contain valid system IDs")
	}
	systemRecord, err := h.FindRecordById("systems", systemID)
	if err != nil {
		return fmt.Errorf("system %s not found", systemID)
	}
	if auth.IsSuperuser() || auth.GetString("role") == "admin" {
		return nil
	}
	if !slices.Contains(systemRecord.GetStringSlice("users"), auth.Id) {
		return fmt.Errorf("system %s not found", systemID)
	}
	return nil
}

func setProbeRecord(record *core.Record, input NetworkProbeInput) {
	record.Set("name", input.Name)
	record.Set("type", input.Type)
	record.Set("target", input.Target)
	record.Set("interval_seconds", input.IntervalSeconds)
	record.Set("timeout_seconds", input.TimeoutSeconds)
	record.Set("enabled", boolValue(input.Enabled, true))
	record.Set("public_visible", boolValue(input.PublicVisible, true))
	record.Set("scope", input.Scope)
}

func networkProbeInputFromRecord(record *core.Record, assignments []*core.Record) NetworkProbeInput {
	enabled := record.GetBool("enabled")
	publicVisible := record.GetBool("public_visible")
	return NetworkProbeInput{
		Name:            record.GetString("name"),
		Type:            record.GetString("type"),
		Target:          record.GetString("target"),
		IntervalSeconds: record.GetInt("interval_seconds"),
		TimeoutSeconds:  record.GetInt("timeout_seconds"),
		Enabled:         &enabled,
		PublicVisible:   &publicVisible,
		Scope:           networkProbeScopeFromRecord(record),
		Systems:         systemIDsFromAssignments(assignments),
	}
}

func mergeProbeInput(dst *NetworkProbeInput, patch NetworkProbeInput) {
	if patch.Name != "" {
		dst.Name = patch.Name
	}
	if patch.Type != "" {
		dst.Type = patch.Type
	}
	if patch.Target != "" {
		dst.Target = patch.Target
	}
	if patch.IntervalSeconds != 0 {
		dst.IntervalSeconds = patch.IntervalSeconds
	}
	if patch.TimeoutSeconds != 0 {
		dst.TimeoutSeconds = patch.TimeoutSeconds
	}
	if patch.Enabled != nil {
		dst.Enabled = patch.Enabled
	}
	if patch.PublicVisible != nil {
		dst.PublicVisible = patch.PublicVisible
	}
	if patch.Scope != "" {
		dst.Scope = patch.Scope
	}
	if patch.Systems != nil {
		dst.Systems = patch.Systems
		if patch.Scope == "" {
			dst.Scope = normalizeNetworkProbeScope("", patch.Systems)
		}
	}
}

func (h *Hub) replaceProbeAssignments(probeID string, scope string, systemIDs []string) ([]*core.Record, error) {
	existing, err := h.FindRecordsByFilter(CollectionNetworkProbeAssignments, "probe = {:probe}", "", -1, 0, dbx.Params{"probe": probeID})
	if err != nil {
		return nil, err
	}
	for _, assignment := range existing {
		if err := h.Delete(assignment); err != nil {
			return nil, err
		}
	}
	if scope == NetworkProbeScopeGlobal {
		return nil, nil
	}
	collection, err := h.FindCachedCollectionByNameOrId(CollectionNetworkProbeAssignments)
	if err != nil {
		return nil, err
	}
	assignments := make([]*core.Record, 0, len(systemIDs))
	seen := make(map[string]struct{}, len(systemIDs))
	for _, systemID := range systemIDs {
		if _, ok := seen[systemID]; ok {
			continue
		}
		seen[systemID] = struct{}{}
		assignment := core.NewRecord(collection)
		assignment.Set("probe", probeID)
		assignment.Set("system", systemID)
		assignment.Set("enabled", true)
		if err := h.Save(assignment); err != nil {
			return nil, err
		}
		assignments = append(assignments, assignment)
	}
	return assignments, nil
}

func networkProbeResponse(probe *core.Record, assignments []*core.Record) NetworkProbeResponse {
	scope := networkProbeScopeFromRecord(probe)
	systemIDs := systemIDsFromAssignments(assignments)
	if scope == NetworkProbeScopeGlobal {
		systemIDs = []string{}
	}
	return NetworkProbeResponse{
		ID:              probe.Id,
		Name:            probe.GetString("name"),
		Type:            probe.GetString("type"),
		Target:          probe.GetString("target"),
		IntervalSeconds: probe.GetInt("interval_seconds"),
		TimeoutSeconds:  probe.GetInt("timeout_seconds"),
		Enabled:         probe.GetBool("enabled"),
		PublicVisible:   probe.GetBool("public_visible"),
		Scope:           scope,
		Systems:         systemIDs,
	}
}

func systemIDsFromAssignments(assignments []*core.Record) []string {
	systemIDs := make([]string, 0, len(assignments))
	for _, assignment := range assignments {
		if assignment.GetBool("enabled") {
			systemIDs = append(systemIDs, assignment.GetString("system"))
		}
	}
	return systemIDs
}

func networkProbeResultPoint(result *core.Record) NetworkProbeResultPoint {
	errorMessage := result.GetString("error")
	failureCategory := ""
	if !result.GetBool("success") {
		failureCategory = normalizeProbeFailureCategory(result.GetString("failure_category"), errorMessage)
	}
	point := NetworkProbeResultPoint{
		SystemID:        result.GetString("system"),
		Created:         result.GetDateTime("created").Time().UTC().Format(time.RFC3339),
		Success:         result.GetBool("success"),
		LatencyMs:       optionalFloat(result, "latency_ms"),
		PacketLossPct:   optionalFloat(result, "packet_loss_percent"),
		HTTPStatus:      optionalInt(result, "http_status"),
		FailureCategory: failureCategory,
		ProbeType:       result.GetString("type"),
		ProbeTarget:     result.GetString("target"),
		RetentionBucket: result.GetString("bucket"),
	}
	if !point.Success {
		point.Error = safeProbeResultError(errorMessage, failureCategory)
	}
	return point
}

func safeProbeResultError(message string, failureCategory string) string {
	message = strings.TrimSpace(message)
	if message != "" && len(message) <= 200 && !strings.Contains(message, "/") && !strings.Contains(message, "\\") {
		return message
	}
	return safeProbeFailureLabel(failureCategory)
}

func normalizePersistedProbeFailureCategory(category string, message string) string {
	normalized := normalizeProbeFailureCategory(category, message)
	if normalized != "" {
		return normalized
	}
	return string(common.NetworkProbeFailureUnknown)
}

func optionalFloat(record *core.Record, field string) *float64 {
	value := record.GetFloat(field)
	if value == 0 {
		return nil
	}
	return &value
}

func optionalInt(record *core.Record, field string) *int {
	value := record.GetInt(field)
	if value == 0 {
		return nil
	}
	return &value
}

func (h *Hub) persistNetworkProbeResult(probeID string, systemID string, result common.NetworkProbeResult) error {
	collection, err := h.FindCachedCollectionByNameOrId(CollectionNetworkProbeResults)
	if err != nil {
		return err
	}
	record := core.NewRecord(collection)
	record.Set("probe", probeID)
	record.Set("system", systemID)
	record.Set("type", string(result.Type))
	record.Set("target", result.Target)
	record.Set("success", result.Success)
	record.Set("latency_ms", result.LatencyMs)
	if result.PacketLossPercent != nil {
		record.Set("packet_loss_percent", *result.PacketLossPercent)
	}
	if result.HTTPStatus != nil {
		record.Set("http_status", *result.HTTPStatus)
	}
	record.Set("error", result.Error)
	if field := record.Collection().Fields.GetByName("failure_category"); field != nil {
		if result.Success {
			record.Set("failure_category", "")
		} else {
			record.Set("failure_category", normalizePersistedProbeFailureCategory(result.FailureCategory, result.Error))
		}
	}
	record.Set("bucket", "1m")
	if result.CheckedAt != "" {
		if checkedAt, err := time.Parse(time.RFC3339, result.CheckedAt); err == nil {
			record.Set("created", checkedAt.UTC())
		}
	}
	return h.Save(record)
}

// RunDueNetworkProbes executes enabled probes from their assigned agent nodes.
func (h *Hub) RunDueNetworkProbes(ctx context.Context) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	now := time.Now().UTC()
	assignments, err := h.effectiveNetworkProbeAssignments("")
	if err != nil {
		return err
	}
	for _, assignment := range assignments {
		probeID := assignment.ProbeID
		probe, err := h.FindRecordById(CollectionNetworkProbes, probeID)
		if err != nil {
			h.Logger().Warn("Network probe not found for assignment", "assignment", assignment.ID, "probe", probeID, "err", err)
			continue
		}
		latest, err := h.latestNetworkProbeResult(probe.Id, assignment.SystemID)
		if err != nil {
			h.Logger().Warn("Network probe latest result lookup failed", "assignment", assignment.ID, "probe", probe.Id, "err", err)
			continue
		}
		if !networkProbeAssignmentDue(probe, latest, now) {
			continue
		}
		if err := h.runNetworkProbeAssignment(ctx, assignment); err != nil {
			h.Logger().Warn("Network probe assignment failed", "assignment", assignment.ID, "err", err)
		}
	}
	return ctx.Err()
}

// RunLiveNetworkProbes executes enabled latency probes for one active node-detail
// 1m live observation without mutating the normal configured probe schedule.
func (h *Hub) RunLiveNetworkProbes(ctx context.Context, systemID string) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	assignments, err := h.liveNetworkProbeAssignments(systemID)
	if err != nil {
		return err
	}
	for _, assignment := range assignments {
		assignment := assignment
		if !h.liveProbeManager().beginAssignment(assignment.ID, time.Now().UTC()) {
			continue
		}
		go func() {
			defer h.liveProbeManager().endAssignment(assignment.ID)
			if err := h.runLiveNetworkProbeAssignment(ctx, assignment); err != nil {
				h.Logger().Warn("Live network probe assignment failed", "assignment", assignment.ID, "err", err)
			}
		}()
	}
	return ctx.Err()
}

func (h *Hub) liveNetworkProbeAssignments(systemID string) ([]networkProbeAssignment, error) {
	assignments, err := h.effectiveNetworkProbeAssignments(systemID)
	if err != nil {
		return nil, err
	}
	eligible := make([]networkProbeAssignment, 0, len(assignments))
	for _, assignment := range assignments {
		probe, err := h.FindRecordById(CollectionNetworkProbes, assignment.ProbeID)
		if err != nil {
			h.Logger().Warn("Network probe not found for live assignment", "assignment", assignment.ID, "err", err)
			continue
		}
		if !probe.GetBool("enabled") || !isLiveLatencyProbeType(common.NetworkProbeType(probe.GetString("type"))) {
			continue
		}
		eligible = append(eligible, assignment)
	}
	return eligible, nil
}

func (h *Hub) effectiveNetworkProbeAssignments(systemID string) ([]networkProbeAssignment, error) {
	probes, err := h.FindRecordsByFilter(CollectionNetworkProbes, "enabled = true", "name", -1, 0)
	if err != nil {
		return nil, err
	}
	assignments, err := h.FindRecordsByFilter(CollectionNetworkProbeAssignments, "enabled = true", "", -1, 0)
	if err != nil {
		return nil, err
	}
	systems, err := h.effectiveNetworkProbeSystems(systemID)
	if err != nil {
		return nil, err
	}
	systemsByID := make(map[string]*core.Record, len(systems))
	for _, system := range systems {
		systemsByID[system.Id] = system
	}
	result := make([]networkProbeAssignment, 0, len(assignments)+len(probes)*len(systems))
	seen := make(map[string]struct{})
	for _, probe := range probes {
		switch networkProbeScopeFromRecord(probe) {
		case NetworkProbeScopeGlobal:
			for _, system := range systems {
				addEffectiveNetworkProbeAssignment(&result, seen, networkProbeAssignment{
					ID:       "global:" + probe.Id + ":" + system.Id,
					ProbeID:  probe.Id,
					SystemID: system.Id,
					Enabled:  true,
				})
			}
		default:
			for _, assignment := range assignments {
				if assignment.GetString("probe") != probe.Id {
					continue
				}
				assignedSystemID := assignment.GetString("system")
				if _, ok := systemsByID[assignedSystemID]; !ok {
					continue
				}
				addEffectiveNetworkProbeAssignment(&result, seen, networkProbeAssignment{
					ID:       assignment.Id,
					ProbeID:  probe.Id,
					SystemID: assignedSystemID,
					Enabled:  assignment.GetBool("enabled"),
				})
			}
		}
	}
	return result, nil
}

func (h *Hub) effectiveNetworkProbeSystems(systemID string) ([]*core.Record, error) {
	if strings.TrimSpace(systemID) != "" {
		system, err := h.FindRecordById("systems", systemID)
		if err != nil {
			return nil, err
		}
		return []*core.Record{system}, nil
	}
	return h.FindRecordsByFilter("systems", "id != ''", "", -1, 0)
}

func addEffectiveNetworkProbeAssignment(result *[]networkProbeAssignment, seen map[string]struct{}, assignment networkProbeAssignment) bool {
	if !assignment.Enabled {
		return false
	}
	key := assignment.ProbeID + ":" + assignment.SystemID
	if _, ok := seen[key]; ok {
		return false
	}
	seen[key] = struct{}{}
	*result = append(*result, assignment)
	return true
}

func (h *Hub) latestNetworkProbeResult(probeID string, systemID string) (*core.Record, error) {
	results, err := h.FindRecordsByFilter(
		CollectionNetworkProbeResults,
		"probe = {:probe} && system = {:system}",
		"-created",
		1,
		0,
		dbx.Params{"probe": probeID, "system": systemID},
	)
	if err != nil || len(results) == 0 {
		return nil, err
	}
	return results[0], nil
}

func networkProbeAssignmentDue(probe *core.Record, latest *core.Record, now time.Time) bool {
	if probe == nil || !probe.GetBool("enabled") {
		return false
	}
	if latest == nil {
		return true
	}
	interval := time.Duration(probe.GetInt("interval_seconds")) * time.Second
	if interval <= 0 {
		interval = time.Duration(defaultProbeIntervalSeconds) * time.Second
	}
	latestAt := latest.GetDateTime("created").Time().UTC()
	if latestAt.IsZero() {
		return true
	}
	return !now.Before(latestAt.Add(interval))
}

func (h *Hub) runNetworkProbeAssignment(ctx context.Context, assignment networkProbeAssignment) error {
	probeID := assignment.ProbeID
	systemID := assignment.SystemID
	probe, err := h.FindRecordById(CollectionNetworkProbes, probeID)
	if err != nil {
		return err
	}
	if !probe.GetBool("enabled") {
		return nil
	}
	timeout := time.Duration(probe.GetInt("timeout_seconds")) * time.Second
	if timeout <= 0 {
		timeout = defaultProbeTimeoutSeconds * time.Second
	}
	probeCtx, cancel := context.WithTimeout(ctx, timeout+5*time.Second)
	defer cancel()

	request := common.NetworkProbeRequest{
		ProbeID:        probe.Id,
		Type:           common.NetworkProbeType(probe.GetString("type")),
		Target:         probe.GetString("target"),
		TimeoutSeconds: uint16(probe.GetInt("timeout_seconds")),
	}
	system, err := h.sm.GetSystem(systemID)
	if err != nil || system.WsConn == nil || !system.WsConn.IsConnected() {
		return h.persistNetworkProbeResult(probeID, systemID, failedNetworkProbeResult(request, "agent offline or unsupported"))
	}
	result, err := system.WsConn.RunNetworkProbe(probeCtx, request)
	if err != nil {
		return h.persistNetworkProbeResult(probeID, systemID, failedNetworkProbeResult(request, err.Error()))
	}
	return h.persistNetworkProbeResult(probeID, systemID, result)
}

func (h *Hub) runLiveNetworkProbeAssignment(ctx context.Context, assignment networkProbeAssignment) error {
	probeID := assignment.ProbeID
	systemID := assignment.SystemID
	probe, err := h.FindRecordById(CollectionNetworkProbes, probeID)
	if err != nil {
		return err
	}
	if !probe.GetBool("enabled") || !isLiveLatencyProbeType(common.NetworkProbeType(probe.GetString("type"))) {
		return nil
	}
	request := liveNetworkProbeRequest(probe)
	probeCtx, cancel := context.WithTimeout(ctx, (time.Duration(request.TimeoutSeconds)+1)*time.Second)
	defer cancel()
	system, err := h.sm.GetSystem(systemID)
	if err != nil || system.WsConn == nil || !system.WsConn.IsConnected() {
		return h.persistNetworkProbeResult(probeID, systemID, failedNetworkProbeResult(request, "agent offline or unsupported"))
	}
	result, err := system.WsConn.RunNetworkProbe(probeCtx, request)
	if err != nil {
		return h.persistNetworkProbeResult(probeID, systemID, failedNetworkProbeResult(request, err.Error()))
	}
	return h.persistNetworkProbeResult(probeID, systemID, result)
}

func liveNetworkProbeRequest(probe *core.Record) common.NetworkProbeRequest {
	timeout := probe.GetInt("timeout_seconds")
	if timeout <= 0 || timeout > liveProbeTimeoutSeconds {
		timeout = liveProbeTimeoutSeconds
	}
	return common.NetworkProbeRequest{
		ProbeID:        probe.Id,
		Type:           common.NetworkProbeType(probe.GetString("type")),
		Target:         probe.GetString("target"),
		TimeoutSeconds: uint16(timeout),
	}
}

func isLiveLatencyProbeType(probeType common.NetworkProbeType) bool {
	return probeType == common.NetworkProbeTCPing || probeType == common.NetworkProbeICMPPing
}

func failedNetworkProbeResult(request common.NetworkProbeRequest, message string) common.NetworkProbeResult {
	category := normalizeProbeFailureCategory("", message)
	if category == "" {
		category = string(common.NetworkProbeFailureUnknown)
	}
	return common.NetworkProbeResult{
		ProbeID:         request.ProbeID,
		Type:            request.Type,
		Target:          request.Target,
		Success:         false,
		Error:           safeProbeResultError(message, category),
		FailureCategory: category,
		CheckedAt:       time.Now().UTC().Format(time.RFC3339),
	}
}
