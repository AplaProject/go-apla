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

package migration

var sectionsDataSQL = `
INSERT INTO "1_sections" ("id","title","urlname","page","roles_access", "status", "ecosystem") VALUES
(next_id('1_sections'), 'Home', 'home', 'default_page', '[]', 2, '%[1]d'),
(next_id('1_sections'), 'Admin', 'admin', 'admin_index', '[]', 1, '%[1]d'),
(next_id('1_sections'), 'Developer', 'developer', 'developer_index', '[]', 1, '%[1]d');
`
