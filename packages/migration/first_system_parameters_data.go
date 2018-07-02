package migration

var firstSystemParametersDataSQL = `
INSERT INTO "1_system_parameters" ("id","name", "value", "conditions") VALUES 
	('1','default_ecosystem_page', 'Div(content-wrapper){
		Div(panel panel-primary){
			Div(list-group-item text-center){
				P(Class: h3 m0 text-bold, Body: Congratulations! You created your own ecosystem.)
			}
			Div(list-group-item){
				Span(Class: h3, Body: "You as Founder hold a complete set of rights for controlling the ecosystem – creating and editing applications, modifying ecosystem parameters, etc. ")
				Span(Class: h3, Body: "To get started, you can download the finished applications from the")
				Span(Class: h3 text-primary, Body: " https://github.com/GenesisKernel/apps ")
				Span(Class: h3, Body: "and install them using the Import service. ")
				Span(Class: h3, Body: "The Strong(basic.json) contains applications for managing roles, creating notifications and votings. ")
				Span(Class: h3, Body: "Or you can create your own apps using the tools in the Admin tab. ")
				Span(Class: h3, Body: "Documentation ")
				Span(Class: h3 text-primary, Body: "https://genesiskernel.readthedocs.io")
			}
			Div(panel-footer text-right clearfix){
				Div(pull-left){
					Button(Body: Ecosystem parameters, Class: btn btn-default, Page: params_list)
				}.Style(margin-right: 20px;)
				Div(pull-left){
					Button(Body: Dashboard, Class: btn btn-default, Page: admin_dashboard)          
				}
				Button(Body: Import, Class: btn btn-primary, Page: import_upload)
			}
		}
	}', 'true'),
	('2','default_ecosystem_menu', '', 'true'),
	('3','default_ecosystem_contract', '', 'true'),
	('4','gap_between_blocks', '2', 'true'),
	('5','rb_blocks_1', '60', 'true'),
	('7','new_version_url', 'upd.apla.io', 'true'),
	('8','full_nodes', '', 'true'),
	('9','number_of_nodes', '101', 'true'),
	('10','ecosystem_price', '1000', 'true'),
	('11','contract_price', '200', 'true'),
	('12','column_price', '200', 'true'),
	('13','table_price', '200', 'true'),
	('14','menu_price', '100', 'true'),
	('15','page_price', '100', 'true'),
	('16','blockchain_url', '', 'true'),
	('17','max_block_size', '67108864', 'true'),
	('18','max_tx_size', '33554432', 'true'),
	('19','max_tx_count', '1000', 'true'),
	('20','max_columns', '50', 'true'),
	('21','max_indexes', '5', 'true'),
	('22','max_block_user_tx', '100', 'true'),
	('23','max_fuel_tx', '20000', 'true'),
	('24','max_fuel_block', '100000', 'true'),
	('25','commission_size', '3', 'true'),
	('26','commission_wallet', '', 'true'),
	('27','fuel_rate', '[["1","1000000000000000"]]', 'true'),
	('28','extend_cost_address_to_id', '10', 'true'),
	('29','extend_cost_id_to_address', '10', 'true'),
	('30','extend_cost_new_state', '1000', 'true'), -- What cost must be?
	('31','extend_cost_sha256', '50', 'true'),
	('32','extend_cost_pub_to_id', '10', 'true'),
	('33','extend_cost_ecosys_param', '10', 'true'),
	('34','extend_cost_sys_param_string', '10', 'true'),
	('35','extend_cost_sys_param_int', '10', 'true'),
	('36','extend_cost_sys_fuel', '10', 'true'),
	('37','extend_cost_validate_condition', '30', 'true'),
	('38','extend_cost_eval_condition', '20', 'true'),
	('39','extend_cost_has_prefix', '10', 'true'),
	('40','extend_cost_contains', '10', 'true'),
	('41','extend_cost_replace', '10', 'true'),
	('42','extend_cost_join', '10', 'true'),
	('43','extend_cost_update_lang', '10', 'true'),
	('44','extend_cost_size', '10', 'true'),
	('45','extend_cost_substr', '10', 'true'),
	('46','extend_cost_contracts_list', '10', 'true'),
	('47','extend_cost_is_object', '10', 'true'),
	('48','extend_cost_compile_contract', '100', 'true'),
	('49','extend_cost_flush_contract', '50', 'true'),
	('50','extend_cost_eval', '10', 'true'),
	('51','extend_cost_len', '5', 'true'),
	('52','extend_cost_activate', '10', 'true'),
	('53','extend_cost_deactivate', '10', 'true'),
	('54','extend_cost_create_ecosystem', '100', 'true'),
	('55','extend_cost_table_conditions', '100', 'true'),
	('56','extend_cost_create_table', '100', 'true'),
	('57','extend_cost_perm_table', '100', 'true'),
	('58','extend_cost_column_condition', '50', 'true'),
	('59','extend_cost_create_column', '50', 'true'),
	('60','extend_cost_perm_column', '50', 'true'),
	('61','extend_cost_json_to_map', '50', 'true'),
	('62','max_block_generation_time', '2000', 'true'),
	('63','block_reward','1000','true'),
	('64','incorrect_blocks_per_day','10','true'),
	('65','node_ban_time','86400000','true'),
	('66','local_node_ban_time','1800000','true');
`
