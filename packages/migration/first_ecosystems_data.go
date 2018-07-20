package migration

var firstEcosystemDataSQL = `
INSERT INTO "1_ecosystems" ("id", "name", "is_valued") VALUES ('1', 'platform ecosystem', 0);

UPDATE "1_roles"
	SET role_name = 'Platform Administrator'
	WHERE id = 1;

UPDATE "1_roles_participants"
	SET role = '{"id": "1", "type": "3", "name": "Platform Administrator", "image_id":"0"}'
	WHERE id = 1;

INSERT INTO "1_applications" (id, name, conditions) VALUES (2, 'System parameters', 
	'ContractConditions("MainCondition")');
`
