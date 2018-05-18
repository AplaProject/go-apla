package migration

var firstDelayedContractsDataSQL = `INSERT INTO "1_delayed_contracts"
		("id", "contract", "key_id", "block_id", "every_block", "conditions")
	VALUES
		(1, '@1UpdateMetrics', '%[1]d', '100', '100', 'ContractConditions("MainCondition")'),
		(2, '@1CheckNodesBan', '%[1]d', '10', '10', 'ContractConditions("MainCondition")');`
