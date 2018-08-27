package migration

import (
	"github.com/GenesisKernel/go-genesis/packages/migration/vde"
)

var rolesDataSQL = `
INSERT INTO "1_roles" ("id", "default_page", "role_name", "deleted", "role_type",
	"date_created","creator","roles_access", "ecosystem") VALUES
	(next_id('1_roles'),'', 'Admin', '0', '3', NOW(), '{}', '{"rids": "1"}', '%[1]d'),
	(next_id('1_roles'),'', 'Developer', '0', '3', NOW(), '{}', '{"rids": "1"}', '%[1]d');

	INSERT INTO "1_roles_participants" ("id","role" ,"member", "date_created", "ecosystem")
	VALUES (next_id('1_roles_participants'), '{"id": "1", "type": "3", "name": "Admin", "image_id":"0"}', '{"member_id": "%[2]d", "member_name": "founder", "image_id": "0"}', NOW(), '%[1]d'),
	(next_id('1_roles_participants'), '{"id": "2", "type": "3", "name": "Developer", "image_id":"0"}', '{"member_id": "%[2]d", "member_name": "founder", "image_id": "0"}', NOW(), '%[1]d');

	INSERT INTO "1_members" ("id", "member_name", "ecosystem") VALUES('%[2]d', 'founder', '%[1]d'),
	('` + vde.GuestKey + `', 'guest', '%[1]d');

`
