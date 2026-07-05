package migrations

import (
	"encoding/json"
	"strings"

	"github.com/pocketbase/pocketbase/core"
	m "github.com/pocketbase/pocketbase/migrations"
)

const publicProbeIDsFieldName = "public_probe_ids"

func init() {
	m.Register(func(app core.App) error {
		if err := EnsurePublicProbeVisibilityField(app); err != nil {
			return err
		}
		return SeedPublicProbeVisibility(app)
	}, func(app core.App) error {
		collection, err := app.FindCollectionByNameOrId("public_system_visibility")
		if err != nil || collection.Fields.GetByName(publicProbeIDsFieldName) == nil {
			return nil
		}
		raw, err := collection.MarshalJSON()
		if err != nil {
			return err
		}
		var data map[string]any
		if err := json.Unmarshal(raw, &data); err != nil {
			return err
		}
		fields, _ := data["fields"].([]any)
		filtered := make([]any, 0, len(fields))
		for _, field := range fields {
			fieldMap, ok := field.(map[string]any)
			if ok && fieldMap["name"] == publicProbeIDsFieldName {
				continue
			}
			filtered = append(filtered, field)
		}
		data["fields"] = filtered
		importRaw, err := json.Marshal([]map[string]any{data})
		if err != nil {
			return err
		}
		return app.ImportCollectionsByMarshaledJSON(importRaw, false)
	})
}

func EnsurePublicProbeVisibilityField(app core.App) error {
	collection, err := app.FindCollectionByNameOrId("public_system_visibility")
	if err != nil {
		return nil
	}
	if collection.Fields.GetByName(publicProbeIDsFieldName) != nil {
		return nil
	}
	raw, err := collection.MarshalJSON()
	if err != nil {
		return err
	}
	var data map[string]any
	if err := json.Unmarshal(raw, &data); err != nil {
		return err
	}
	fields, _ := data["fields"].([]any)
	fields = append(fields, map[string]any{
		"cascadeDelete": false,
		"collectionId":  "np_network_probe",
		"hidden":        false,
		"id":            "rel_public_probe_ids",
		"maxSelect":     999,
		"minSelect":     0,
		"name":          publicProbeIDsFieldName,
		"presentable":   false,
		"required":      false,
		"system":        false,
		"type":          "relation",
	})
	data["fields"] = fields
	importRaw, err := json.Marshal([]map[string]any{data})
	if err != nil {
		return err
	}
	return app.ImportCollectionsByMarshaledJSON(importRaw, false)
}

func SeedPublicProbeVisibility(app core.App) error {
	if _, err := app.FindCollectionByNameOrId("public_system_visibility"); err != nil {
		return nil
	}
	if _, err := app.FindCollectionByNameOrId("network_probes"); err != nil {
		return nil
	}
	if _, err := app.FindCollectionByNameOrId("network_probe_assignments"); err != nil {
		return nil
	}
	visibilityRecords, err := app.FindRecordsByFilter("public_system_visibility", "public_enabled = true", "", -1, 0)
	if err != nil {
		return err
	}
	probes, err := app.FindRecordsByFilter("network_probes", "enabled = true && public_visible = true", "name", -1, 0)
	if err != nil {
		return err
	}
	assignments, err := app.FindRecordsByFilter("network_probe_assignments", "enabled = true", "", -1, 0)
	if err != nil {
		return err
	}
	assignmentMap := make(map[string][]string, len(probes))
	for _, assignment := range assignments {
		probeID := assignment.GetString("probe")
		assignmentMap[probeID] = append(assignmentMap[probeID], assignment.GetString("system"))
	}

	for _, visibilityRecord := range visibilityRecords {
		if len(publicProbeIDsFromVisibilityRecord(visibilityRecord)) > 0 {
			continue
		}
		systemID := visibilityRecord.GetString("system")
		selectedProbeIDs := make([]string, 0, len(probes))
		for _, probe := range probes {
			scope := probe.GetString("scope")
			if scope == "" || scope == "global" {
				selectedProbeIDs = append(selectedProbeIDs, probe.Id)
				continue
			}
			if containsString(assignmentMap[probe.Id], systemID) {
				selectedProbeIDs = append(selectedProbeIDs, probe.Id)
			}
		}
		visibilityRecord.Set(publicProbeIDsFieldName, selectedProbeIDs)
		if err := app.Save(visibilityRecord); err != nil {
			return err
		}
	}
	return nil
}

func containsString(values []string, target string) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}

func publicProbeIDsFromVisibilityRecord(record *core.Record) []string {
	if record == nil {
		return []string{}
	}
	raw := record.Get(publicProbeIDsFieldName)
	switch value := raw.(type) {
	case []string:
		return normalizeSeededProbeIDs(value)
	case []any:
		probeIDs := make([]string, 0, len(value))
		for _, item := range value {
			if text, ok := item.(string); ok {
				probeIDs = append(probeIDs, text)
			}
		}
		return normalizeSeededProbeIDs(probeIDs)
	case string:
		if strings.TrimSpace(value) == "" {
			return []string{}
		}
	}
	return normalizeSeededProbeIDs(record.GetStringSlice(publicProbeIDsFieldName))
}

func normalizeSeededProbeIDs(probeIDs []string) []string {
	if len(probeIDs) == 0 {
		return []string{}
	}
	seen := make(map[string]struct{}, len(probeIDs))
	normalized := make([]string, 0, len(probeIDs))
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
