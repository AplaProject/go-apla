package migration

var menuDataSQL = `INSERT INTO "%[1]d_menu" (id, name, value, conditions) VALUES
(2, 'admin_menu', 'MenuItem(Title:"Import", Page:import_upload, Icon:"icon-cloud-upload")', 'ContractConditions("MainCondition")');
`