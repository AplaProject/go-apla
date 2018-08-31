package migration

var pagesDataSQL = `INSERT INTO "1_pages" (id, name, value, menu, conditions, ecosystem) VALUES
	(next_id('1_pages'), 'admin_index', '', 'admin_menu','ContractAccess("@1EditPage")', '%[1]d'),
	(next_id('1_pages'), 'notifications', '', 'default_menu','ContractAccess("@1EditPage")', '%[1]d');
`
