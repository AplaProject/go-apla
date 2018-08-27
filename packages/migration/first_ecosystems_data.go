package migration

var firstEcosystemDataSQL = `
INSERT INTO "1_ecosystems" ("id", "name", "is_valued") VALUES ('1', 'platform ecosystem', 0);

INSERT INTO "1_roles" ("id", "default_page", "role_name", "deleted", "role_type",
	"date_created","creator","roles_access", "ecosystem") VALUES
	(next_id('1_roles'),'', 'Apla Consensus asbl', '0', '3', NOW(), '{}', '{"rids": "1"}', '1'),
	(next_id('1_roles'),'', 'Candidate for validators', '0', '3', NOW(), '{}', '{"rids": "1"}', '1'),
	(next_id('1_roles'),'', 'Validator', '0', '3', NOW(), '{}', '{"rids": "1"}', '1'),
	(next_id('1_roles'),'', 'Investor with voting rights', '0', '3', NOW(), '{}', '{"rids": "1"}', '1'),
	(next_id('1_roles'),'', 'Delegate', '0', '3', NOW(), '{}', '{"rids": "1"}', '1');
`
