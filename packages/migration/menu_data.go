package migration

var menuDataSQL = `INSERT INTO "1_menu" (id, name, value, conditions, ecosystem) VALUES
(next_id('1_menu'), 'admin_menu', 'MenuItem(Title:"Import", Page:import_upload, Icon:"icon-cloud-upload")', 'ContractConditions("MainCondition")','%[1]d');
`
