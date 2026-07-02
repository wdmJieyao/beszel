package migrations

import (
	"encoding/json"

	"github.com/pocketbase/pocketbase/core"
	m "github.com/pocketbase/pocketbase/migrations"
)

func init() {
	m.Register(func(app core.App) error {
		collection, err := app.FindCollectionByNameOrId("network_probe_results")
		if err != nil {
			return nil
		}
		if collection.Fields.GetByName("failure_category") != nil {
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
			"hidden":      false,
			"id":          "select_failure_category",
			"maxSelect":   1,
			"name":        "failure_category",
			"presentable": false,
			"required":    false,
			"system":      false,
			"type":        "select",
			"values": []string{
				"invalid_target",
				"dns_failure",
				"timeout",
				"connection_refused",
				"target_unreachable",
				"execution_node_unavailable",
				"unsupported",
				"unknown_failure",
			},
		})
		data["fields"] = fields
		importRaw, err := json.Marshal([]map[string]any{data})
		if err != nil {
			return err
		}
		return app.ImportCollectionsByMarshaledJSON(importRaw, false)
	}, func(app core.App) error {
		collection, err := app.FindCollectionByNameOrId("network_probe_results")
		if err != nil || collection.Fields.GetByName("failure_category") == nil {
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
			if ok && fieldMap["name"] == "failure_category" {
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
