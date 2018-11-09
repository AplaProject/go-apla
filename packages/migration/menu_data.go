package migration

var menuDataSQL = `INSERT INTO "1_menu" (id, name, value, conditions, ecosystem) VALUES
	(next_id('1_menu'), 'admin_menu', '', 'ContractConditions("@1DeveloperCondition")', '%[1]d'),
	(next_id('1_menu'), 'developer_menu', 'MenuItem(Title:"Import", Page:@1import_upload, Icon:"icon-cloud-upload")', 'ContractConditions("@1DeveloperCondition")', '%[1]d');
`
