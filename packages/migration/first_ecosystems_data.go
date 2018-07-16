package migration

var firstEcosystemDataSQL = `
INSERT INTO "1_ecosystems" ("id", "name", "is_valued") VALUES ('1', 'platform ecosystem', 0);

UPDATE "1_roles"
	SET role_name = 'Platform Administrator'
	WHERE id = 1;

INSERT INTO "1_applications" (id, name, conditions) VALUES (2, 'System parameters', 
	'ContractConditions("MainCondition")');
`
