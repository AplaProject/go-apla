package migration

var rolesDataSQL = `
INSERT INTO "%[1]d_roles" ("id", "default_page", "role_name", "deleted", "role_type",
	"date_created","creator") VALUES
	('1','default_ecosystem_page', 'Admin', '0', '3', NOW(), '{}'),
	('2','', 'Candidate for validators', '0', '3', NOW(), '{}'),
	('3','', 'Validator', '0', '3', NOW(), '{}'),
	('4','', 'Investor with voting rights', '0', '3', NOW(), '{}'),
	('5','', 'Delegate', '0', '3', NOW(), '{}'),
	('6','', 'Developer', '0', '3', NOW(), '{}');

	INSERT INTO "%[1]d_roles_participants" ("id","role" ,"member", "date_created")
	VALUES ('1', '{"id": "1", "type": "3", "name": "Admin", "image_id":"0"}', '{"member_id": "%[4]d", "member_name": "founder", "image_id": "0"}', NOW()),
	('2', '{"id": "6", "type": "3", "name": "Developer", "image_id":"0"}', '{"member_id": "%[4]d", "member_name": "founder", "image_id": "0"}', NOW());

	INSERT INTO "%[1]d_members" ("id", "member_name") VALUES('%[4]d', 'founder');

`
