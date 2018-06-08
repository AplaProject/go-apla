package migration

var applicationsDataSQL = `INSERT INTO "%[1]d_applications" (id, name, conditions) VALUES (1, 'System', 'ContractConditions("MainCondition")');`
