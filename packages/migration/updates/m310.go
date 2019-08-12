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

var M310 = `
INSERT INTO "1_system_parameters" (id, name, value, conditions) VALUES
	(next_id('1_system_parameters'), 'price_exec_app_param', '10', 'ContractAccess("@1UpdateSysParam")'),
	(next_id('1_system_parameters'), 'price_exec_to_upper', '10', 'ContractAccess("@1UpdateSysParam")'),
	(next_id('1_system_parameters'), 'price_exec_pub_to_hex', '20', 'ContractAccess("@1UpdateSysParam")'),
	(next_id('1_system_parameters'), 'price_exec_check_condition', '10', 'ContractAccess("@1UpdateSysParam")'),
	(next_id('1_system_parameters'), 'price_exec_contract_conditions', '50', 'ContractAccess("@1UpdateSysParam")'),
	(next_id('1_system_parameters'), 'price_exec_log10', '15', 'ContractAccess("@1UpdateSysParam")'),
	(next_id('1_system_parameters'), 'price_exec_send_external_transaction', '100', 'ContractAccess("@1UpdateSysParam")'),
	(next_id('1_system_parameters'), 'price_exec_create_language', '50', 'ContractAccess("@1UpdateSysParam")'),
	(next_id('1_system_parameters'), 'price_exec_edit_language', '50', 'ContractAccess("@1UpdateSysParam")'),
	(next_id('1_system_parameters'), 'price_exec_hmac', '50', 'ContractAccess("@1UpdateSysParam")'),
	(next_id('1_system_parameters'), 'price_exec_split', '50', 'ContractAccess("@1UpdateSysParam")'),
	(next_id('1_system_parameters'), 'price_exec_sqrt', '15', 'ContractAccess("@1UpdateSysParam")'),
	(next_id('1_system_parameters'), 'price_exec_get_contract_by_name', '20', 'ContractAccess("@1UpdateSysParam")'),
	(next_id('1_system_parameters'), 'price_exec_jsonto_map', '50', 'ContractAccess("@1UpdateSysParam")'),
	(next_id('1_system_parameters'), 'price_exec_role_access', '50', 'ContractAccess("@1UpdateSysParam")'),
	(next_id('1_system_parameters'), 'price_exec_del_column', '100', 'ContractAccess("@1UpdateSysParam")'),
	(next_id('1_system_parameters'), 'price_exec_pow', '15', 'ContractAccess("@1UpdateSysParam")'),
	(next_id('1_system_parameters'), 'price_exec_round', '15', 'ContractAccess("@1UpdateSysParam")'),
	(next_id('1_system_parameters'), 'price_exec_get_contract_by_id', '20', 'ContractAccess("@1UpdateSysParam")'),
	(next_id('1_system_parameters'), 'price_exec_trim_space', '10', 'ContractAccess("@1UpdateSysParam")'),
	(next_id('1_system_parameters'), 'price_exec_transaction_info', '100', 'ContractAccess("@1UpdateSysParam")'),
	(next_id('1_system_parameters'), 'price_exec_log', '15', 'ContractAccess("@1UpdateSysParam")'),
	(next_id('1_system_parameters'), 'price_exec_validate_edit_contract_new_value', '10', 'ContractAccess("@1UpdateSysParam")'),
	(next_id('1_system_parameters'), 'price_exec_hex_to_pub', '20', 'ContractAccess("@1UpdateSysParam")'),
	(next_id('1_system_parameters'), 'price_exec_update_contract', '60', 'ContractAccess("@1UpdateSysParam")'),
	(next_id('1_system_parameters'), 'price_exec_create_contract', '60', 'ContractAccess("@1UpdateSysParam")'),
	(next_id('1_system_parameters'), 'price_exec_to_lower', '10', 'ContractAccess("@1UpdateSysParam")'),
	(next_id('1_system_parameters'), 'price_exec_del_table', '100', 'ContractAccess("@1UpdateSysParam")'),
	(next_id('1_system_parameters'), 'price_exec_floor', '15', 'ContractAccess("@1UpdateSysParam")'),
	(next_id('1_system_parameters'), 'price_exec_contract_access', '50', 'ContractAccess("@1UpdateSysParam")'),
	(next_id('1_system_parameters'), 'price_exec_contract_name', '10', 'ContractAccess("@1UpdateSysParam")'),
	(next_id('1_system_parameters'), 'access_exec_compile_contract', 'ContractAccess("@1NewContract", "@1EditContract", "@1Import")', 'ContractAccess("@1UpdateSysParam")'),
	(next_id('1_system_parameters'), 'access_exec_update_contract', 'ContractAccess("@1EditContract", "@1Import")', 'ContractAccess("@1UpdateSysParam")'),
	(next_id('1_system_parameters'), 'access_exec_create_contract', 'ContractAccess("@1NewContract", "@1Import")', 'ContractAccess("@1UpdateSysParam")'),
	(next_id('1_system_parameters'), 'access_exec_create_table', 'ContractAccess("@1NewTable", "@1NewTableJoint", "@1Import")', 'ContractAccess("@1UpdateSysParam")'),
	(next_id('1_system_parameters'), 'access_exec_flush_contract', 'ContractAccess("@1NewContract", "@1EditContract", "@1Import")', 'ContractAccess("@1UpdateSysParam")'),
	(next_id('1_system_parameters'), 'access_exec_perm_table', 'ContractAccess("@1EditTable")', 'ContractAccess("@1UpdateSysParam")'),
	(next_id('1_system_parameters'), 'access_exec_table_conditions', 'ContractAccess("@1NewTable", "@1Import", "@1NewTableJoint", "@1EditTable")', 'ContractAccess("@1UpdateSysParam")'),
	(next_id('1_system_parameters'), 'access_exec_column_condition', 'ContractAccess("@1NewColumn", "@1EditColumn")', 'ContractAccess("@1UpdateSysParam")'),
	(next_id('1_system_parameters'), 'access_exec_create_column', 'ContractAccess("@1NewColumn")', 'ContractAccess("@1UpdateSysParam")'),
	(next_id('1_system_parameters'), 'access_exec_perm_column', 'ContractAccess("@1EditColumn")', 'ContractAccess("@1UpdateSysParam")'),
	(next_id('1_system_parameters'), 'access_exec_create_language', 'ContractAccess("@1NewLang", "@1NewLangJoint", "@1Import")', 'ContractAccess("@1UpdateSysParam")'),
	(next_id('1_system_parameters'), 'access_exec_edit_language', 'ContractAccess("@1EditLang", "@1EditLangJoint", "@1Import")', 'ContractAccess("@1UpdateSysParam")'),
	(next_id('1_system_parameters'), 'access_exec_create_ecosystem', 'ContractAccess("@1NewEcosystem")', 'ContractAccess("@1UpdateSysParam")'),
	(next_id('1_system_parameters'), 'access_exec_edit_ecosys_name', 'ContractAccess("@1EditEcosystemName")', 'ContractAccess("@1UpdateSysParam")'),
	(next_id('1_system_parameters'), 'access_exec_bind_wallet', 'ContractAccess("@1BindWallet")', 'ContractAccess("@1UpdateSysParam")'),
	(next_id('1_system_parameters'), 'access_exec_unbind_wallet', 'ContractAccess("@1UnbindWallet")', 'ContractAccess("@1UpdateSysParam")'),
	(next_id('1_system_parameters'), 'access_exec_set_contract_wallet', 'ContractAccess("@1BindWallet", "@1UnbindWallet")', 'ContractAccess("@1UpdateSysParam")');
`
