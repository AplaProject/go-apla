package migration

var firstTablesDataSQL = `
INSERT INTO "1_tables" ("id", "name", "permissions","columns", "conditions") VALUES
		('20', 'delayed_contracts',
		'{"insert": "ContractConditions(\"MainCondition\")", "update": "ContractConditions(\"MainCondition\")",
		"new_column": "ContractConditions(\"MainCondition\")"}',
		'{"contract": "ContractConditions(\"MainCondition\")",
			"key_id": "ContractConditions(\"MainCondition\")",
			"block_id": "ContractConditions(\"MainCondition\")",
			"every_block": "ContractConditions(\"MainCondition\")",
			"counter": "ContractConditions(\"MainCondition\")",
			"limit": "ContractConditions(\"MainCondition\")",
			"deleted": "ContractConditions(\"MainCondition\")",
			"conditions": "ContractConditions(\"MainCondition\")"}',
			'ContractConditions("MainCondition")'
		),
		(
			'21',
			'ecosystems',
			'{"insert": "ContractConditions(\"MainCondition\")", "update": "ContractConditions(\"MainCondition\")", "new_column": "ContractConditions(\"MainCondition\")"}',
			'{"name": "ContractConditions(\"MainCondition\")"}',
			'ContractConditions("MainCondition")'
		),
		(
			'22',
			'metrics',
			'{"insert": "ContractConditions(\"MainCondition\")", "update": "ContractConditions(\"MainCondition\")","new_column": "ContractConditions(\"MainCondition\")"}',
			'{"time": "ContractConditions(\"MainCondition\")",
				"metric": "ContractConditions(\"MainCondition\")","key": "ContractConditions(\"MainCondition\")",
				"value": "ContractConditions(\"MainCondition\")"}',
			'ContractConditions("MainCondition")'
		),
		(
			'23',
			'system_parameters',
			'{"insert": "false", "update": "ContractAccess(\"1@UpdateSysParam\")","new_column": "ContractConditions(\"MainCondition\")"}',
			'{"value": "ContractConditions(\"MainCondition\")"}',
			'ContractConditions("MainCondition")'
		),
		(
			'24',
			'bad_blocks',
			'{
				"insert": "ContractAccess(\"NewBadBlock\")",
				"update": "ContractAccess(\"NewBadBlock\", \"CheckNodesBan\")",
				"new_column": "ContractConditions(\"MainCondition\")"
			}',
			'{
				"contract": "ContractConditions(\"MainCondition\")",
				"id": "ContractConditions(\"MainCondition\")",
				"block_id": "ContractConditions(\"MainCondition\")",
				"producer_node_id": "ContractConditions(\"MainCondition\")",
				"consumer_node_id": "ContractConditions(\"MainCondition\")",
				"block_time": "ContractConditions(\"MainCondition\")",
				"reason": "ContractConditions(\"MainCondition\")",
				"deleted": "ContractAccess(\"NewBadBlock\", \"CheckNodesBan\")"
			}',
			'ContractConditions(\"MainCondition\")'
		),
		(
			'25',
			'node_ban_logs',
			'{
				"insert": "ContractAccess(\"CheckNodesBan\")",
				"update": "ContractConditions(\"MainCondition\")",
				"new_column": "ContractConditions(\"MainCondition\")"
			}',
			'{
				"contract": "ContractConditions(\"MainCondition\")",
				"id": "ContractConditions(\"MainCondition\")",
				"node_id": "ContractConditions(\"MainCondition\")",
				"banned_at": "ContractConditions(\"MainCondition\")",
				"ban_time": "ContractConditions(\"MainCondition\")",
				"reason": "ContractConditions(\"MainCondition\")"
			}',
			'ContractConditions(\"MainCondition\")'
		);
`
