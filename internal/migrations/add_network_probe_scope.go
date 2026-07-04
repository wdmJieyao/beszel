package migrations

import (
	"encoding/json"

	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase/core"
	m "github.com/pocketbase/pocketbase/migrations"
)

const (
	networkProbeScopeGlobal = "global"
	networkProbeScopeFixed  = "fixed"
)

func init() {
	m.Register(func(app core.App) error {
		collection, err := app.FindCollectionByNameOrId("network_probes")
		if err != nil {
			return nil
		}
		if collection.Fields.GetByName("scope") == nil {
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
				"hidden":      false,
				"id":          "select_probe_scope",
				"maxSelect":   1,
				"name":        "scope",
				"presentable": false,
				"required":    false,
				"system":      false,
				"type":        "select",
				"values":      []string{networkProbeScopeGlobal, networkProbeScopeFixed},
			})
			data["fields"] = fields
			importRaw, err := json.Marshal([]map[string]any{data})
			if err != nil {
				return err
			}
			if err := app.ImportCollectionsByMarshaledJSON(importRaw, false); err != nil {
				return err
			}
		}
		return backfillNetworkProbeScope(app)
	}, func(app core.App) error {
		collection, err := app.FindCollectionByNameOrId("network_probes")
		if err != nil || collection.Fields.GetByName("scope") == nil {
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
			if ok && fieldMap["name"] == "scope" {
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

func backfillNetworkProbeScope(app core.App) error {
	probes, err := app.FindRecordsByFilter("network_probes", "id != ''", "", -1, 0)
	if err != nil {
		return err
	}
	systems, err := app.FindRecordsByFilter("systems", "id != ''", "", -1, 0)
	if err != nil {
		return err
	}
	systemIDs := make(map[string]struct{}, len(systems))
	for _, system := range systems {
		systemIDs[system.Id] = struct{}{}
	}
	for _, probe := range probes {
		if probe.GetString("scope") != "" {
			continue
		}
		assignments, err := app.FindRecordsByFilter("network_probe_assignments", "probe = {:probe} && enabled = true", "", -1, 0, dbx.Params{"probe": probe.Id})
		if err != nil {
			return err
		}
		scope := networkProbeScopeFixed
		if probeAssignmentsCoverAllSystems(assignments, systemIDs) {
			scope = networkProbeScopeGlobal
		}
		probe.Set("scope", scope)
		if err := app.Save(probe); err != nil {
			return err
		}
	}
	return nil
}

func probeAssignmentsCoverAllSystems(assignments []*core.Record, systemIDs map[string]struct{}) bool {
	if len(systemIDs) == 0 || len(assignments) == 0 {
		return true
	}
	assigned := make(map[string]struct{}, len(assignments))
	for _, assignment := range assignments {
		systemID := assignment.GetString("system")
		if _, ok := systemIDs[systemID]; ok {
			assigned[systemID] = struct{}{}
		}
	}
	return len(assigned) == len(systemIDs)
}
