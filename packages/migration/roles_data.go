package migration

import (
	"github.com/GenesisKernel/go-genesis/packages/migration/vde"
)

var rolesDataSQL = `
INSERT INTO "%[1]d_roles" ("id", "default_page", "role_name", "deleted", "role_type",
	"date_created","creator","roles_access") VALUES
	('1','', 'Admin', '0', '3', NOW(), '{}', '{"rids": "1"}'),
	('2','', 'Developer', '0', '3', NOW(), '{}', '{"rids": "1"}');

	INSERT INTO "%[1]d_roles_participants" ("id","role" ,"member", "date_created")
	VALUES ('1', '{"id": "1", "type": "3", "name": "Admin", "image_id":"0"}', '{"member_id": "%[2]d", "member_name": "founder", "image_id": "0"}', NOW()),
	('2', '{"id": "2", "type": "3", "name": "Developer", "image_id":"0"}', '{"member_id": "%[2]d", "member_name": "founder", "image_id": "0"}', NOW());

	INSERT INTO "%[1]d_members" ("id", "member_name") VALUES('%[2]d', 'founder'),
	('` + vde.GuestKey + `', 'guest');

`
