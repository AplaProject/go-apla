// Copyright (C) 2017, 2018, 2019 EGAAS S.A.
//
// This program is free software; you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation; either version 2 of the License, or (at
// your option) any later version.
//
// This program is distributed in the hope that it will be useful, but
// WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the GNU
// General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program; if not, write to the Free Software
// Foundation, Inc., 51 Franklin Street, Fifth Floor, Boston, MA 02110-1301, USA.

package obs

import "github.com/AplaProject/go-apla/packages/consts"

var rolesDataSQL = `
INSERT INTO "1_roles" ("id", "default_page", "role_name", "deleted", "role_type", "creator","roles_access", "ecosystem") VALUES
	(next_id('1_roles'),'', 'Admin', '0', '3', '{}', '{}', '%[1]d'),
	(next_id('1_roles'),'', 'Developer', '0', '3', '{}', '{}', '%[1]d'),
	(next_id('1_roles'),'', 'Apla Consensus asbl', '0', '3', '{}', '{"rids": "1"}', '%[1]d'),
	(next_id('1_roles'),'', 'Candidate for validators', '0', '3', '{}', '{}', '%[1]d'),
	(next_id('1_roles'),'', 'Validator', '0', '3', '{}', '{}', '%[1]d'),
	(next_id('1_roles'),'', 'Investor with voting rights', '0', '3', '{}', '{}', '%[1]d'),
	(next_id('1_roles'),'', 'Delegate', '0', '3', '{}', '{}', '%[1]d');

	INSERT INTO "1_roles_participants" ("id","role" ,"member", "date_created", "ecosystem")
	VALUES (next_id('1_roles_participants'), '{"id": "1", "type": "3", "name": "Admin", "image_id":"0"}', '{"member_id": "%[2]d", "member_name": "founder", "image_id": "0"}', NOW(), '%[1]d'),
	(next_id('1_roles_participants'), '{"id": "2", "type": "3", "name": "Developer", "image_id":"0"}', '{"member_id": "%[2]d", "member_name": "founder", "image_id": "0"}', NOW(), '%[1]d');

	INSERT INTO "1_members" ("id", "account", "member_name", "ecosystem") 
	VALUES
		(next_id('1_members'), '%[3]s', 'founder', '%[1]d'),
		(next_id('1_members'), '` + consts.GuestAddress + `', 'guest', '%[1]d');

`
