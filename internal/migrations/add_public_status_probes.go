package migrations

import (
	"github.com/pocketbase/pocketbase/core"
	m "github.com/pocketbase/pocketbase/migrations"
)

func init() {
	m.Register(func(app core.App) error {
		jsonData := `[
			{
				"id": "psv_public_sys",
				"name": "public_system_visibility",
				"type": "base",
				"system": false,
				"listRule": "@request.auth.id != \"\" && system.users.id ?= @request.auth.id",
				"viewRule": "@request.auth.id != \"\" && system.users.id ?= @request.auth.id",
				"createRule": "@request.auth.id != \"\" && system.users.id ?= @request.auth.id && @request.auth.role != \"readonly\"",
				"updateRule": "@request.auth.id != \"\" && system.users.id ?= @request.auth.id && @request.auth.role != \"readonly\"",
				"deleteRule": "@request.auth.id != \"\" && system.users.id ?= @request.auth.id && @request.auth.role != \"readonly\"",
				"fields": [
					{"autogeneratePattern":"[a-z0-9]{15}","hidden":false,"id":"text3208210256","max":15,"min":15,"name":"id","pattern":"^[a-z0-9]+$","presentable":false,"primaryKey":true,"required":true,"system":true,"type":"text"},
					{"cascadeDelete":true,"collectionId":"2hz5ncl8tizk5nx","hidden":false,"id":"rel_public_system","maxSelect":1,"minSelect":0,"name":"system","presentable":false,"required":true,"system":false,"type":"relation"},
					{"hidden":false,"id":"bool_public_enabled","name":"public_enabled","presentable":false,"required":false,"system":false,"type":"bool"},
					{"autogeneratePattern":"","hidden":false,"id":"text_public_name","max":120,"min":0,"name":"public_name","pattern":"","presentable":false,"primaryKey":false,"required":false,"system":false,"type":"text"},
					{"hidden":false,"id":"bool_show_cpu","name":"show_cpu","presentable":false,"required":false,"system":false,"type":"bool"},
					{"hidden":false,"id":"bool_show_memory","name":"show_memory","presentable":false,"required":false,"system":false,"type":"bool"},
					{"hidden":false,"id":"bool_show_disk","name":"show_disk","presentable":false,"required":false,"system":false,"type":"bool"},
					{"hidden":false,"id":"autodate2990389176","name":"created","onCreate":true,"onUpdate":false,"presentable":false,"system":false,"type":"autodate"},
					{"hidden":false,"id":"autodate3332085495","name":"updated","onCreate":true,"onUpdate":true,"presentable":false,"system":false,"type":"autodate"}
				],
				"indexes": ["CREATE UNIQUE INDEX ` + "`" + `idx_public_system_visibility_system` + "`" + ` ON ` + "`" + `public_system_visibility` + "`" + ` (` + "`" + `system` + "`" + `)"]
			},
			{
				"id": "np_network_probe",
				"name": "network_probes",
				"type": "base",
				"system": false,
				"listRule": "@request.auth.id != \"\"",
				"viewRule": "@request.auth.id != \"\"",
				"createRule": "@request.auth.id != \"\" && @request.auth.role != \"readonly\"",
				"updateRule": "@request.auth.id != \"\" && @request.auth.role != \"readonly\"",
				"deleteRule": "@request.auth.id != \"\" && @request.auth.role != \"readonly\"",
				"fields": [
					{"autogeneratePattern":"[a-z0-9]{15}","hidden":false,"id":"text3208210256","max":15,"min":15,"name":"id","pattern":"^[a-z0-9]+$","presentable":false,"primaryKey":true,"required":true,"system":true,"type":"text"},
					{"autogeneratePattern":"","hidden":false,"id":"text_probe_name","max":120,"min":1,"name":"name","pattern":"","presentable":false,"primaryKey":false,"required":true,"system":false,"type":"text"},
					{"hidden":false,"id":"select_probe_type","maxSelect":1,"name":"type","presentable":false,"required":true,"system":false,"type":"select","values":["tcping","icmp_ping","http_get"]},
					{"hidden":false,"id":"select_probe_scope","maxSelect":1,"name":"scope","presentable":false,"required":false,"system":false,"type":"select","values":["global","fixed"]},
					{"autogeneratePattern":"","hidden":false,"id":"text_probe_target","max":500,"min":1,"name":"target","pattern":"","presentable":false,"primaryKey":false,"required":true,"system":false,"type":"text"},
					{"hidden":false,"id":"num_interval_seconds","max":86400,"min":10,"name":"interval_seconds","onlyInt":true,"presentable":false,"required":true,"system":false,"type":"number"},
					{"hidden":false,"id":"num_timeout_seconds","max":300,"min":1,"name":"timeout_seconds","onlyInt":true,"presentable":false,"required":true,"system":false,"type":"number"},
					{"hidden":false,"id":"bool_public_visible","name":"public_visible","presentable":false,"required":false,"system":false,"type":"bool"},
					{"hidden":false,"id":"bool_enabled","name":"enabled","presentable":false,"required":false,"system":false,"type":"bool"},
					{"hidden":false,"id":"autodate2990389176","name":"created","onCreate":true,"onUpdate":false,"presentable":false,"system":false,"type":"autodate"},
					{"hidden":false,"id":"autodate3332085495","name":"updated","onCreate":true,"onUpdate":true,"presentable":false,"system":false,"type":"autodate"}
				],
				"indexes": []
			},
			{
				"id": "npa_probe_assign",
				"name": "network_probe_assignments",
				"type": "base",
				"system": false,
				"listRule": "@request.auth.id != \"\" && system.users.id ?= @request.auth.id",
				"viewRule": "@request.auth.id != \"\" && system.users.id ?= @request.auth.id",
				"createRule": "@request.auth.id != \"\" && system.users.id ?= @request.auth.id && @request.auth.role != \"readonly\"",
				"updateRule": "@request.auth.id != \"\" && system.users.id ?= @request.auth.id && @request.auth.role != \"readonly\"",
				"deleteRule": "@request.auth.id != \"\" && system.users.id ?= @request.auth.id && @request.auth.role != \"readonly\"",
				"fields": [
					{"autogeneratePattern":"[a-z0-9]{15}","hidden":false,"id":"text3208210256","max":15,"min":15,"name":"id","pattern":"^[a-z0-9]+$","presentable":false,"primaryKey":true,"required":true,"system":true,"type":"text"},
					{"cascadeDelete":true,"collectionId":"np_network_probe","hidden":false,"id":"rel_assignment_probe","maxSelect":1,"minSelect":0,"name":"probe","presentable":false,"required":true,"system":false,"type":"relation"},
					{"cascadeDelete":true,"collectionId":"2hz5ncl8tizk5nx","hidden":false,"id":"rel_assignment_system","maxSelect":1,"minSelect":0,"name":"system","presentable":false,"required":true,"system":false,"type":"relation"},
					{"hidden":false,"id":"bool_enabled","name":"enabled","presentable":false,"required":false,"system":false,"type":"bool"},
					{"hidden":false,"id":"autodate2990389176","name":"created","onCreate":true,"onUpdate":false,"presentable":false,"system":false,"type":"autodate"},
					{"hidden":false,"id":"autodate3332085495","name":"updated","onCreate":true,"onUpdate":true,"presentable":false,"system":false,"type":"autodate"}
				],
				"indexes": ["CREATE UNIQUE INDEX ` + "`" + `idx_network_probe_assignment` + "`" + ` ON ` + "`" + `network_probe_assignments` + "`" + ` (` + "`" + `probe` + "`" + `, ` + "`" + `system` + "`" + `)"]
			},
			{
				"id": "npr_probe_result",
				"name": "network_probe_results",
				"type": "base",
				"system": false,
				"listRule": "@request.auth.id != \"\" && system.users.id ?= @request.auth.id",
				"viewRule": "@request.auth.id != \"\" && system.users.id ?= @request.auth.id",
				"createRule": null,
				"updateRule": null,
				"deleteRule": "@request.auth.id != \"\" && system.users.id ?= @request.auth.id && @request.auth.role != \"readonly\"",
				"fields": [
					{"autogeneratePattern":"[a-z0-9]{15}","hidden":false,"id":"text3208210256","max":15,"min":15,"name":"id","pattern":"^[a-z0-9]+$","presentable":false,"primaryKey":true,"required":true,"system":true,"type":"text"},
					{"cascadeDelete":true,"collectionId":"np_network_probe","hidden":false,"id":"rel_result_probe","maxSelect":1,"minSelect":0,"name":"probe","presentable":false,"required":true,"system":false,"type":"relation"},
					{"cascadeDelete":true,"collectionId":"2hz5ncl8tizk5nx","hidden":false,"id":"rel_result_system","maxSelect":1,"minSelect":0,"name":"system","presentable":false,"required":true,"system":false,"type":"relation"},
					{"hidden":false,"id":"select_result_type","maxSelect":1,"name":"type","presentable":false,"required":true,"system":false,"type":"select","values":["tcping","icmp_ping","http_get"]},
					{"autogeneratePattern":"","hidden":false,"id":"text_result_target","max":500,"min":0,"name":"target","pattern":"","presentable":false,"primaryKey":false,"required":false,"system":false,"type":"text"},
					{"hidden":false,"id":"bool_success","name":"success","presentable":false,"required":false,"system":false,"type":"bool"},
					{"hidden":false,"id":"num_latency_ms","max":null,"min":null,"name":"latency_ms","onlyInt":false,"presentable":false,"required":false,"system":false,"type":"number"},
					{"hidden":false,"id":"num_packet_loss","max":100,"min":0,"name":"packet_loss_percent","onlyInt":false,"presentable":false,"required":false,"system":false,"type":"number"},
					{"hidden":false,"id":"num_http_status","max":599,"min":100,"name":"http_status","onlyInt":true,"presentable":false,"required":false,"system":false,"type":"number"},
					{"autogeneratePattern":"","hidden":false,"id":"text_error","max":200,"min":0,"name":"error","pattern":"","presentable":false,"primaryKey":false,"required":false,"system":false,"type":"text"},
					{"hidden":false,"id":"select_failure_category","maxSelect":1,"name":"failure_category","presentable":false,"required":false,"system":false,"type":"select","values":["invalid_target","dns_failure","timeout","connection_refused","target_unreachable","execution_node_unavailable","unsupported","unknown_failure"]},
					{"hidden":false,"id":"select_result_bucket","maxSelect":1,"name":"bucket","presentable":false,"required":false,"system":false,"type":"select","values":["1m","10m","20m","120m","480m"]},
					{"hidden":false,"id":"autodate2990389176","name":"created","onCreate":true,"onUpdate":false,"presentable":false,"system":false,"type":"autodate"}
				],
				"indexes": ["CREATE INDEX ` + "`" + `idx_network_probe_results_lookup` + "`" + ` ON ` + "`" + `network_probe_results` + "`" + ` (` + "`" + `probe` + "`" + `, ` + "`" + `system` + "`" + `, ` + "`" + `created` + "`" + `)"]
			}
		]`
		return app.ImportCollectionsByMarshaledJSON([]byte(jsonData), false)
	}, func(app core.App) error {
		for _, name := range []string{"network_probe_results", "network_probe_assignments", "network_probes", "public_system_visibility"} {
			collection, err := app.FindCollectionByNameOrId(name)
			if err == nil {
				if err := app.Delete(collection); err != nil {
					return err
				}
			}
		}
		return nil
	})
}
