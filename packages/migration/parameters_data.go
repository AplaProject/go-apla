package migration

var parametersDataSQL = `
INSERT INTO "1_parameters" ("id","name", "value", "conditions", "ecosystem") VALUES 
	(next_id('1_parameters'),'founder_account', '%[2]d', 'ContractConditions("MainCondition")', '%[1]d'),
	(next_id('1_parameters'),'new_table', 'ContractConditions("MainCondition")', 'ContractConditions("MainCondition")', '%[1]d'),
	(next_id('1_parameters'),'changing_tables', 'ContractConditions("MainCondition")', 'ContractConditions("MainCondition")', '%[1]d'),
	(next_id('1_parameters'),'changing_language', 'ContractConditions("MainCondition")', 'ContractConditions("MainCondition")', '%[1]d'),
	(next_id('1_parameters'),'changing_signature', 'ContractConditions("MainCondition")', 'ContractConditions("MainCondition")', '%[1]d'),
	(next_id('1_parameters'),'changing_page', 'ContractConditions("MainCondition")', 'ContractConditions("MainCondition")', '%[1]d'),
	(next_id('1_parameters'),'changing_menu', 'ContractConditions("MainCondition")', 'ContractConditions("MainCondition")', '%[1]d'),
	(next_id('1_parameters'),'changing_contracts', 'ContractConditions("MainCondition")', 'ContractConditions("MainCondition")', '%[1]d'),
	(next_id('1_parameters'),'max_sum', '1000000', 'ContractConditions("MainCondition")', '%[1]d'),
	(next_id('1_parameters'),'money_digit', '18', 'ContractConditions("MainCondition")', '%[1]d'),
	(next_id('1_parameters'),'stylesheet', 'body {
		  /* You can define your custom styles here or create custom CSS rules */
	}', 'ContractConditions("MainCondition")', '%[1]d'),
	(next_id('1_parameters'),'max_block_user_tx', '100', 'ContractConditions("MainCondition")', '%[1]d'),
	(next_id('1_parameters'),'min_page_validate_count', '1', 'ContractConditions("MainCondition")', '%[1]d'),
	(next_id('1_parameters'),'max_page_validate_count', '6', 'ContractConditions("MainCondition")', '%[1]d'),
	(next_id('1_parameters'),'changing_blocks', 'ContractConditions("MainCondition")', 'ContractConditions("MainCondition")', '%[1]d');
`
