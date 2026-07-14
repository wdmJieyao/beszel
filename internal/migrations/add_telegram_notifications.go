package migrations

import (
	"github.com/pocketbase/pocketbase/core"
	m "github.com/pocketbase/pocketbase/migrations"
)

func init() {
	m.Register(func(app core.App) error {
		jsonData := `[
			{
				"id": "tg_settings_cfg",
				"name": "telegram_settings",
				"type": "base",
				"system": false,
				"listRule": "@request.auth.id != \"\" && @request.auth.role = \"admin\"",
				"viewRule": "@request.auth.id != \"\" && @request.auth.role = \"admin\"",
				"createRule": "@request.auth.id != \"\" && @request.auth.role = \"admin\"",
				"updateRule": "@request.auth.id != \"\" && @request.auth.role = \"admin\"",
				"deleteRule": null,
				"fields": [
					{"autogeneratePattern":"[a-z0-9]{15}","hidden":false,"id":"text3208210256","max":15,"min":15,"name":"id","pattern":"^[a-z0-9]+$","presentable":false,"primaryKey":true,"required":true,"system":true,"type":"text"},
					{"hidden":false,"id":"bool_tg_enabled","name":"enabled","presentable":false,"required":false,"system":false,"type":"bool"},
					{"hidden":false,"id":"bool_tg_polling","name":"polling_enabled","presentable":false,"required":false,"system":false,"type":"bool"},
					{"autogeneratePattern":"","hidden":true,"id":"text_tg_token","max":4096,"min":0,"name":"bot_token_encrypted","pattern":"","presentable":false,"primaryKey":false,"required":false,"system":false,"type":"text"},
					{"autogeneratePattern":"","hidden":false,"id":"text_tg_username","max":255,"min":0,"name":"bot_username","pattern":"","presentable":false,"primaryKey":false,"required":false,"system":false,"type":"text"},
					{"hidden":false,"id":"num_tg_offset","max":null,"min":0,"name":"last_poll_offset","onlyInt":true,"presentable":false,"required":false,"system":false,"type":"number"},
					{"autogeneratePattern":"","hidden":false,"id":"text_tg_error","max":500,"min":0,"name":"last_error","pattern":"","presentable":false,"primaryKey":false,"required":false,"system":false,"type":"text"},
					{"hidden":false,"id":"autodate2990389176","name":"created","onCreate":true,"onUpdate":false,"presentable":false,"system":false,"type":"autodate"},
					{"hidden":false,"id":"autodate3332085495","name":"updated","onCreate":true,"onUpdate":true,"presentable":false,"system":false,"type":"autodate"}
				],
				"indexes": ["CREATE UNIQUE INDEX ` + "`" + `idx_telegram_settings_singleton` + "`" + ` ON ` + "`" + `telegram_settings` + "`" + ` (` + "`" + `id` + "`" + `)"]
			},
			{
				"id": "tg_destinations",
				"name": "telegram_destinations",
				"type": "base",
				"system": false,
				"listRule": "@request.auth.id != \"\" && @request.auth.role = \"admin\"",
				"viewRule": "@request.auth.id != \"\" && @request.auth.role = \"admin\"",
				"createRule": "@request.auth.id != \"\" && @request.auth.role = \"admin\"",
				"updateRule": "@request.auth.id != \"\" && @request.auth.role = \"admin\"",
				"deleteRule": "@request.auth.id != \"\" && @request.auth.role = \"admin\"",
				"fields": [
					{"autogeneratePattern":"[a-z0-9]{15}","hidden":false,"id":"text3208210256","max":15,"min":15,"name":"id","pattern":"^[a-z0-9]+$","presentable":false,"primaryKey":true,"required":true,"system":true,"type":"text"},
					{"cascadeDelete":false,"collectionId":"_pb_users_auth_","hidden":false,"id":"rel_tg_user","maxSelect":1,"minSelect":0,"name":"user","presentable":false,"required":false,"system":false,"type":"relation"},
					{"autogeneratePattern":"","hidden":false,"id":"text_tg_name","max":120,"min":1,"name":"name","pattern":"","presentable":false,"primaryKey":false,"required":true,"system":false,"type":"text"},
					{"autogeneratePattern":"","hidden":false,"id":"text_tg_chat_id","max":64,"min":1,"name":"chat_id","pattern":"","presentable":false,"primaryKey":false,"required":true,"system":false,"type":"text"},
					{"hidden":false,"id":"select_tg_chat_type","maxSelect":1,"name":"chat_type","presentable":false,"required":false,"system":false,"type":"select","values":["private","group","supergroup","channel","unknown"]},
					{"hidden":false,"id":"select_tg_role","maxSelect":1,"name":"role","presentable":false,"required":true,"system":false,"type":"select","values":["admin","read_only"]},
					{"hidden":false,"id":"bool_tg_dest_enabled","name":"enabled","presentable":false,"required":false,"system":false,"type":"bool"},
					{"hidden":false,"id":"json_tg_nodes","maxSize":2000000,"name":"node_scope","presentable":false,"required":false,"system":false,"type":"json"},
					{"hidden":false,"id":"json_tg_alerts","maxSize":2000000,"name":"alert_level_scope","presentable":false,"required":false,"system":false,"type":"json"},
					{"hidden":false,"id":"date_tg_mute","max":"","min":"","name":"mute_until","presentable":false,"required":false,"system":false,"type":"date"},
					{"hidden":false,"id":"date_tg_test","max":"","min":"","name":"last_test_at","presentable":false,"required":false,"system":false,"type":"date"},
					{"hidden":false,"id":"date_tg_delivery","max":"","min":"","name":"last_delivery_at","presentable":false,"required":false,"system":false,"type":"date"},
					{"autogeneratePattern":"","hidden":false,"id":"text_tg_last_error","max":500,"min":0,"name":"last_error","pattern":"","presentable":false,"primaryKey":false,"required":false,"system":false,"type":"text"},
					{"hidden":false,"id":"autodate2990389176","name":"created","onCreate":true,"onUpdate":false,"presentable":false,"system":false,"type":"autodate"},
					{"hidden":false,"id":"autodate3332085495","name":"updated","onCreate":true,"onUpdate":true,"presentable":false,"system":false,"type":"autodate"}
				],
				"indexes": ["CREATE UNIQUE INDEX ` + "`" + `idx_telegram_destination_chat` + "`" + ` ON ` + "`" + `telegram_destinations` + "`" + ` (` + "`" + `chat_id` + "`" + `)"]
			}
		]`
		return app.ImportCollectionsByMarshaledJSON([]byte(jsonData), false)
	}, func(app core.App) error {
		for _, name := range []string{"telegram_destinations", "telegram_settings"} {
			collection, err := app.FindCollectionByNameOrId(name)
			if err == nil {
				if err := app.Delete(collection); err != nil {
					return err
				}
			}
		}
		return nil
	})
	registerTelegramNotificationPoliciesMigration()
}
