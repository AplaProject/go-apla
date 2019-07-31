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

var menuDataSQL = `INSERT INTO "1_menu" (id, name, value, conditions, ecosystem) VALUES
(next_id('1_menu'), 'admin_menu', 'MenuItem(Title:"Application", Page:apps_list, Icon:"icon-folder")
MenuItem(Title:"Ecosystem parameters", Page:params_list, Icon:"icon-settings")
MenuItem(Title:"Menu", Page:menus_list, Icon:"icon-list")
MenuItem(Title:"Confirmations", Page:confirmations, Icon:"icon-check")
MenuItem(Title:"Import", Page:import_upload, Icon:"icon-cloud-upload")
MenuItem(Title:"Export", Page:export_resources, Icon:"icon-cloud-download")
MenuGroup(Title:"Resources", Icon:"icon-share"){
	MenuItem(Title:"Pages", Page:app_pages, Icon:"icon-screen-desktop")
	MenuItem(Title:"Blocks", Page:app_blocks, Icon:"icon-grid")
	MenuItem(Title:"Tables", Page:app_tables, Icon:"icon-docs")
	MenuItem(Title:"Contracts", Page:app_contracts, Icon:"icon-briefcase")
	MenuItem(Title:"Application parameters", Page:app_params, Icon:"icon-wrench")
	MenuItem(Title:"Language resources", Page:app_langres, Icon:"icon-globe")
	MenuItem(Title:"Binary data", Page:app_binary, Icon:"icon-layers")
}
MenuItem(Title:"Dashboard", Page:admin_dashboard, Icon:"icon-wrench")', 'ContractConditions("MainCondition")', '%[1]d');
`
