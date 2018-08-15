package vde

var parametersDataSQL = `
INSERT INTO "%[1]d_parameters" ("id","name", "value", "conditions") VALUES 
		('1','founder_account', '%[2]d', 'ContractConditions("MainCondition")'),
		('2','new_table', 'ContractConditions("MainCondition")', 'ContractConditions("MainCondition")'),
		('3','changing_tables', 'ContractConditions("MainCondition")', 'ContractConditions("MainCondition")'),
		('4','changing_language', 'ContractConditions("MainCondition")', 'ContractConditions("MainCondition")'),
		('5','changing_signature', 'ContractConditions("MainCondition")', 'ContractConditions("MainCondition")'),
		('6','changing_page', 'ContractConditions("MainCondition")', 'ContractConditions("MainCondition")'),
		('7','changing_menu', 'ContractConditions("MainCondition")', 'ContractConditions("MainCondition")'),
		('8','changing_contracts', 'ContractConditions("MainCondition")', 'ContractConditions("MainCondition")'),
		('9','max_sum', '1000000', 'ContractConditions("MainCondition")'),
		('10','money_digit', '18', 'ContractConditions("MainCondition")'),
		('11','stylesheet', 'body {
		  /* You can define your custom styles here or create custom CSS rules */
		}', 'ContractConditions("MainCondition")'),
		('12','max_block_user_tx', '100', 'ContractConditions("MainCondition")'),
		('13','min_page_validate_count', '1', 'ContractConditions("MainCondition")'),
		('14','max_page_validate_count', '6', 'ContractConditions("MainCondition")'),
		('15','changing_blocks', 'ContractConditions("MainCondition")', 'ContractConditions("MainCondition")');
`
