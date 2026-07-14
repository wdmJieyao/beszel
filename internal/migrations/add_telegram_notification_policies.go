package migrations

import (
	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase/core"
	m "github.com/pocketbase/pocketbase/migrations"
)

const (
	telegramPoliciesCollection = "telegram_notification_policies"
	telegramDestinations       = "telegram_destinations"
	telegramDefaultPolicyName  = "默认规则"
)

func registerTelegramNotificationPoliciesMigration() {
	m.Register(func(app core.App) error {
		jsonData := `[
			{
				"id": "tg_notify_policy",
				"name": "telegram_notification_policies",
				"type": "base",
				"system": false,
				"listRule": "@request.auth.id != \"\" && @request.auth.role = \"admin\"",
				"viewRule": "@request.auth.id != \"\" && @request.auth.role = \"admin\"",
				"createRule": "@request.auth.id != \"\" && @request.auth.role = \"admin\"",
				"updateRule": "@request.auth.id != \"\" && @request.auth.role = \"admin\"",
				"deleteRule": "@request.auth.id != \"\" && @request.auth.role = \"admin\"",
				"fields": [
					{"autogeneratePattern":"[a-z0-9]{15}","hidden":false,"id":"text3208210256","max":15,"min":15,"name":"id","pattern":"^[a-z0-9]+$","presentable":false,"primaryKey":true,"required":true,"system":true,"type":"text"},
					{"cascadeDelete":true,"collectionId":"tg_destinations","hidden":false,"id":"rel_tg_policy_dest","maxSelect":1,"minSelect":0,"name":"destination","presentable":false,"required":true,"system":false,"type":"relation"},
					{"autogeneratePattern":"","hidden":false,"id":"text_tg_policy_name","max":120,"min":1,"name":"name","pattern":"","presentable":false,"primaryKey":false,"required":true,"system":false,"type":"text"},
					{"hidden":false,"id":"bool_tg_policy_enabled","name":"enabled","presentable":false,"required":false,"system":false,"type":"bool"},
					{"hidden":false,"id":"select_tg_scope_mode","maxSelect":1,"name":"node_scope_mode","presentable":false,"required":true,"system":false,"type":"select","values":["all","selected"]},
					{"hidden":false,"id":"json_tg_policy_nodes","maxSize":2000000,"name":"node_scope","presentable":false,"required":false,"system":false,"type":"json"},
					{"hidden":false,"id":"json_tg_policy_alerts","maxSize":2000000,"name":"alert_level_scope","presentable":false,"required":false,"system":false,"type":"json"},
					{"hidden":false,"id":"autodate_tg_policy_created","name":"created","onCreate":true,"onUpdate":false,"presentable":false,"system":false,"type":"autodate"},
					{"hidden":false,"id":"autodate_tg_policy_updated","name":"updated","onCreate":true,"onUpdate":true,"presentable":false,"system":false,"type":"autodate"}
				],
				"indexes": ["CREATE UNIQUE INDEX ` + "`" + `idx_tg_policy_destination_name` + "`" + ` ON ` + "`" + `telegram_notification_policies` + "`" + ` (` + "`" + `destination` + "`" + `, ` + "`" + `name` + "`" + `)"]
			}
		]`
		if err := app.ImportCollectionsByMarshaledJSON([]byte(jsonData), false); err != nil {
			return err
		}
		return backfillTelegramNotificationPolicies(app)
	}, func(app core.App) error {
		collection, err := app.FindCollectionByNameOrId(telegramPoliciesCollection)
		if err != nil {
			return nil
		}
		return app.Delete(collection)
	}, "zz_add_telegram_notification_policies.go")
}

func backfillTelegramNotificationPolicies(app core.App) error {
	collection, err := app.FindCollectionByNameOrId(telegramPoliciesCollection)
	if err != nil {
		return err
	}
	destinations, err := app.FindRecordsByFilter(telegramDestinations, "id != ''", "id", -1, 0)
	if err != nil {
		return err
	}
	for _, destination := range destinations {
		existing, err := app.FindRecordsByFilter(telegramPoliciesCollection, "destination = {:destination} && name = {:name}", "", 1, 0, dbx.Params{
			"destination": destination.Id,
			"name":        telegramDefaultPolicyName,
		})
		if err != nil {
			return err
		}
		if len(existing) != 0 {
			continue
		}
		nodes := []string{}
		alerts := []string{}
		_ = destination.UnmarshalJSONField("node_scope", &nodes)
		_ = destination.UnmarshalJSONField("alert_level_scope", &alerts)
		mode := "all"
		if len(nodes) != 0 {
			mode = "selected"
		}
		policy := core.NewRecord(collection)
		policy.Set("destination", destination.Id)
		policy.Set("name", telegramDefaultPolicyName)
		policy.Set("enabled", destination.GetBool("enabled"))
		policy.Set("node_scope_mode", mode)
		policy.Set("node_scope", nodes)
		policy.Set("alert_level_scope", alerts)
		if err := app.Save(policy); err != nil {
			return err
		}
	}
	return nil
}
