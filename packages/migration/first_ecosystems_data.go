package migration

var firstEcosystemDataSQL = `
INSERT INTO "1_ecosystems" ("id", "name", "is_valued") VALUES ('1', 'platform ecosystem', 0);

INSERT INTO "1_roles" ("id", "default_page", "role_name", "deleted", "role_type",
	"date_created","creator","roles_access") VALUES
	('3','', 'Apla Consensus asbl', '0', '3', NOW(), '{}', '{}');

INSERT INTO "1_applications" (id, name, conditions) VALUES (2, 'System parameters', 
	'ContractConditions("MainCondition")');
`
