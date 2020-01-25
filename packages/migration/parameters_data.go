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

var parametersDataSQL = `
INSERT INTO "1_parameters" ("id","name", "value", "conditions", "ecosystem") VALUES
	(next_id('1_parameters'),'founder_account', '{{.Wallet}}', 'ContractConditions("@1DeveloperCondition")', '{{.Ecosystem}}'),
	(next_id('1_parameters'),'new_table', 'ContractConditions("MainCondition")', 'ContractConditions("@1DeveloperCondition")', '{{.Ecosystem}}'),
	(next_id('1_parameters'),'changing_tables', 'ContractConditions("MainCondition")', 'ContractConditions("@1DeveloperCondition")', '{{.Ecosystem}}'),
	(next_id('1_parameters'),'changing_language', 'ContractConditions("MainCondition")', 'ContractConditions("@1DeveloperCondition")', '{{.Ecosystem}}'),
	(next_id('1_parameters'),'changing_signature', 'ContractConditions("MainCondition")', 'ContractConditions("@1DeveloperCondition")', '{{.Ecosystem}}'),
	(next_id('1_parameters'),'changing_page', 'ContractConditions("MainCondition")', 'ContractConditions("@1DeveloperCondition")', '{{.Ecosystem}}'),
	(next_id('1_parameters'),'changing_menu', 'ContractConditions("MainCondition")', 'ContractConditions("@1DeveloperCondition")', '{{.Ecosystem}}'),
	(next_id('1_parameters'),'changing_contracts', 'ContractConditions("MainCondition")', 'ContractConditions("@1DeveloperCondition")', '{{.Ecosystem}}'),
	(next_id('1_parameters'),'max_sum', '1000000', 'ContractConditions("@1DeveloperCondition")', '{{.Ecosystem}}'),
	(next_id('1_parameters'),'stylesheet', 'body {
		  /* You can define your custom styles here or create custom CSS rules */
	}', 'ContractConditions("@1DeveloperCondition")', '{{.Ecosystem}}'),
	(next_id('1_parameters'),'stylesheet_print', 'body {
		/* You can define your custom styles here or create custom CSS rules */
	}', 'ContractConditions("@1DeveloperCondition")', '{{.Ecosystem}}'),
	(next_id('1_parameters'),'max_tx_block_per_user', '100', 'ContractConditions("@1DeveloperCondition")', '{{.Ecosystem}}'),
	(next_id('1_parameters'),'min_page_validate_count', '1', 'ContractConditions("@1DeveloperCondition")', '{{.Ecosystem}}'),
	(next_id('1_parameters'),'max_page_validate_count', '6', 'ContractConditions("@1DeveloperCondition")', '{{.Ecosystem}}'),
	(next_id('1_parameters'),'changing_blocks', 'ContractConditions("MainCondition")', 'ContractConditions("@1DeveloperCondition")', '{{.Ecosystem}}'),
	(next_id('1_parameters'),'ecosystem_wallet', '', 'ContractConditions("@1DeveloperCondition")', '{{.Ecosystem}}'),
	(next_id('1_parameters'),'error_page', '@1error_page', 'ContractConditions("@1DeveloperCondition")', '{{.Ecosystem}}');
`
