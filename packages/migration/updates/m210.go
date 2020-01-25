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

package updates

var M210 = `
	sql("insert into \"1_system_parameters\" (id, name, value, conditions) values
(next_id('1_system_parameters'), 'external_blockchain', '', 'ContractAccess(\"@1UpdateSysParam\")');")

	{{head "external_blockchain"}}
		t.Column("id", "bigint", {"default": "0"})
		t.Column("netname", "string", {"default": "", "size":255})
		t.Column("value", "text", {"default": ""})
	{{footer "primary"}}
`
