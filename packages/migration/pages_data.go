package migration

var pagesDataSQL = `INSERT INTO "1_pages" (id, name, value, menu, conditions, app_id, ecosystem) VALUES
	(next_id('1_pages'), 'admin_index', '', 'admin_menu', 'ContractConditions("@1DeveloperCondition")', '%[5]d', '%[1]d'),
	(next_id('1_pages'), 'developer_index', '', 'developer_menu', 'ContractConditions("@1DeveloperCondition")', '%[5]d', '%[1]d');
`
