package migration

var (
	Schema = `DROP TABLE IF EXISTS "transactions_status"; CREATE TABLE "transactions_status" (
		"hash" bytea  NOT NULL DEFAULT '',
		"time" int NOT NULL DEFAULT '0',
		"type" int NOT NULL DEFAULT '0',
		"ecosystem" int NOT NULL DEFAULT '1',
		"wallet_id" bigint NOT NULL DEFAULT '0',
		"block_id" int NOT NULL DEFAULT '0',
		"error" varchar(255) NOT NULL DEFAULT ''
		);
		ALTER TABLE ONLY "transactions_status" ADD CONSTRAINT transactions_status_pkey PRIMARY KEY (hash);
		
		DROP TABLE IF EXISTS "confirmations"; CREATE TABLE "confirmations" (
		"block_id" bigint  NOT NULL DEFAULT '0',
		"good" int  NOT NULL DEFAULT '0',
		"bad" int  NOT NULL DEFAULT '0',
		"time" int  NOT NULL DEFAULT '0'
		);
		ALTER TABLE ONLY "confirmations" ADD CONSTRAINT confirmations_pkey PRIMARY KEY (block_id);
		
		DROP TABLE IF EXISTS "block_chain"; CREATE TABLE "block_chain" (
		"id" int NOT NULL DEFAULT '0',
		"hash" bytea  NOT NULL DEFAULT '',
		"data" bytea NOT NULL DEFAULT '',
		"ecosystem_id" int  NOT NULL DEFAULT '0',
		"key_id" bigint  NOT NULL DEFAULT '0',
		"node_position" bigint  NOT NULL DEFAULT '0',
		"time" int NOT NULL DEFAULT '0',
		"tx" int NOT NULL DEFAULT '0'
		);
		ALTER TABLE ONLY "block_chain" ADD CONSTRAINT block_chain_pkey PRIMARY KEY (id);
		
		DROP TABLE IF EXISTS "log_transactions"; CREATE TABLE "log_transactions" (
		"hash" bytea  NOT NULL DEFAULT '',
		"time" int NOT NULL DEFAULT '0'
		);
		ALTER TABLE ONLY "log_transactions" ADD CONSTRAINT log_transactions_pkey PRIMARY KEY (hash);
		
		DROP TABLE IF EXISTS "migration_history"; CREATE TABLE "migration_history" (
		"id" int NOT NULL  DEFAULT '0',
		"version" int NOT NULL DEFAULT '0',
		"date_applied" int NOT NULL DEFAULT '0'
		);
		ALTER TABLE ONLY "migration_history" ADD CONSTRAINT migration_history_pkey PRIMARY KEY (id);
		
		DROP TABLE IF EXISTS "queue_tx"; CREATE TABLE "queue_tx" (
		"hash" bytea  NOT NULL DEFAULT '',
		"data" bytea NOT NULL DEFAULT '',
		"from_gate" int NOT NULL DEFAULT '0'
		);
		ALTER TABLE ONLY "queue_tx" ADD CONSTRAINT queue_tx_pkey PRIMARY KEY (hash);
		
		DROP TABLE IF EXISTS "config"; CREATE TABLE "config" (
		"my_block_id" int NOT NULL DEFAULT '0',
		"ecosystem_id" int NOT NULL DEFAULT '0',
		"key_id" bigint NOT NULL DEFAULT '0',
		"bad_blocks" text NOT NULL DEFAULT '',
		"auto_reload" int NOT NULL DEFAULT '0',
		"first_load_blockchain_url" varchar(255)  NOT NULL DEFAULT '',
		"first_load_blockchain"  varchar(255)  NOT NULL DEFAULT '',
		"current_load_blockchain"  varchar(255)  NOT NULL DEFAULT ''
		);
		
		DROP SEQUENCE IF EXISTS rollback_rb_id_seq CASCADE;
		CREATE SEQUENCE rollback_rb_id_seq START WITH 1;
		DROP TABLE IF EXISTS "rollback"; CREATE TABLE "rollback" (
		"rb_id" bigint NOT NULL  default nextval('rollback_rb_id_seq'),
		"block_id" bigint NOT NULL DEFAULT '0',
		"data" text NOT NULL DEFAULT ''
		);
		ALTER SEQUENCE rollback_rb_id_seq owned by rollback.rb_id;
		ALTER TABLE ONLY "rollback" ADD CONSTRAINT rollback_pkey PRIMARY KEY (rb_id);
		
		DROP TABLE IF EXISTS "system_states"; CREATE TABLE "system_states" (
		"id" bigint NOT NULL DEFAULT '0',
		"name" varchar(255) NOT NULL DEFAULT '',
		"rb_id" bigint NOT NULL DEFAULT '0'
		);
		ALTER TABLE ONLY "system_states" ADD CONSTRAINT system_states_pkey PRIMARY KEY (id);
		CREATE INDEX "system_states_index_name" ON "system_states" (name);
		
		DROP TABLE IF EXISTS "system_parameters";
		CREATE TABLE "system_parameters" (
		"id" bigint NOT NULL DEFAULT '0',
		"name" varchar(255)  NOT NULL DEFAULT '',
		"value" text NOT NULL DEFAULT '',
		"conditions" text  NOT NULL DEFAULT '',
		"rb_id" bigint  NOT NULL DEFAULT '0'
		);
		ALTER TABLE ONLY "system_parameters" ADD CONSTRAINT system_parameters_pkey PRIMARY KEY (id);
		CREATE INDEX "system_parameters_index_name" ON "system_parameters" (name);
		
		INSERT INTO system_parameters ("id","name", "value", "conditions") VALUES 
		('1','default_ecosystem_page', 'P(class, Default Ecosystem Page)', 'true'),
		('2','default_ecosystem_menu', 'MenuItem(main, Default Ecosystem Menu)', 'true'),
		('3','default_ecosystem_contract', '', 'true'),
		('4','gap_between_blocks', '2', 'true'),
		('5','rb_blocks_1', '60', 'true'),
		('6','rb_blocks_2', '3600', 'true'),
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
		('23','max_fuel_tx', '1000', 'true'),
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
		('61','extend_cost_json_to_map', '50', 'true');
		
		CREATE TABLE "system_contracts" (
		"id" bigint NOT NULL  DEFAULT '0',
		"value" text  NOT NULL DEFAULT '',
		"wallet_id" bigint NOT NULL DEFAULT '0',
		"token_id" bigint NOT NULL DEFAULT '0',
		"active" character(1) NOT NULL DEFAULT '0',
		"conditions" text  NOT NULL DEFAULT '',
		"rb_id" bigint NOT NULL DEFAULT '0'
		);
		ALTER TABLE ONLY "system_contracts" ADD CONSTRAINT system_contracts_pkey PRIMARY KEY (id);
		
		
		CREATE TABLE "system_tables" (
		"name" varchar(100)  NOT NULL DEFAULT '',
		"permissions" jsonb,
		"columns" jsonb,
		"conditions" text  NOT NULL DEFAULT '',
		"rb_id" bigint NOT NULL DEFAULT '0'
		);
		ALTER TABLE ONLY "system_tables" ADD CONSTRAINT system_tables_pkey PRIMARY KEY (name);
		
		INSERT INTO system_tables ("name", "permissions","columns", "conditions") VALUES  ('system_states',
				'{"insert": "false", "update": "ContractAccess(\"@1EditParameter\")",
				  "new_column": "false"}',
				'{"name": "ContractAccess(\"@1EditParameter\")"}',
				'ContractAccess(\"@0UpdSysContract\")');
		
		
		DROP TABLE IF EXISTS "info_block"; CREATE TABLE "info_block" (
		"hash" bytea  NOT NULL DEFAULT '',
		"block_id" int NOT NULL DEFAULT '0',
		"node_position" int  NOT NULL DEFAULT '0',
		"ecosystem_id" bigint NOT NULL DEFAULT '0',
		"key_id" bigint NOT NULL DEFAULT '0',
		"time" int  NOT NULL DEFAULT '0',
		"current_version" varchar(50) NOT NULL DEFAULT '0.0.1',
		"sent" smallint NOT NULL DEFAULT '0'
		);
		
		DROP TABLE IF EXISTS "queue_blocks"; CREATE TABLE "queue_blocks" (
		"hash" bytea  NOT NULL DEFAULT '',
		"full_node_id" bigint NOT NULL DEFAULT '0',
		"block_id" int NOT NULL DEFAULT '0'
		);
		ALTER TABLE ONLY "queue_blocks" ADD CONSTRAINT queue_blocks_pkey PRIMARY KEY (hash);
		
		DROP TABLE IF EXISTS "transactions"; CREATE TABLE "transactions" (
		"hash" bytea  NOT NULL DEFAULT '',
		"data" bytea NOT NULL DEFAULT '',
		"used" smallint NOT NULL DEFAULT '0',
		"high_rate" smallint NOT NULL DEFAULT '0',
		"type" smallint NOT NULL DEFAULT '0',
		"key_id" bigint NOT NULL DEFAULT '0',
		"counter" smallint NOT NULL DEFAULT '0',
		"sent" smallint NOT NULL DEFAULT '0',
		"verified" smallint NOT NULL DEFAULT '1'
		);
		ALTER TABLE ONLY "transactions" ADD CONSTRAINT transactions_pkey PRIMARY KEY (hash);
		
		DROP SEQUENCE IF EXISTS rollback_tx_id_seq CASCADE;
		CREATE SEQUENCE rollback_tx_id_seq START WITH 1;
		DROP TABLE IF EXISTS "rollback_tx"; CREATE TABLE "rollback_tx" (
		"id" bigint NOT NULL  default nextval('rollback_tx_id_seq'),
		"block_id" bigint NOT NULL DEFAULT '0',
		"tx_hash" bytea  NOT NULL DEFAULT '',
		"table_name" varchar(255) NOT NULL DEFAULT '',
		"table_id" varchar(255) NOT NULL DEFAULT ''
		);
		ALTER SEQUENCE rollback_tx_id_seq owned by rollback_tx.id;
		ALTER TABLE ONLY "rollback_tx" ADD CONSTRAINT rollback_tx_pkey PRIMARY KEY (id);
		
		DROP TABLE IF EXISTS "install"; CREATE TABLE "install" (
		"progress" varchar(10) NOT NULL DEFAULT ''
		);
		
		
		DROP TYPE IF EXISTS "my_node_keys_enum_status" CASCADE;
		CREATE TYPE "my_node_keys_enum_status" AS ENUM ('my_pending','approved');
		DROP SEQUENCE IF EXISTS my_node_keys_id_seq CASCADE;
		CREATE SEQUENCE my_node_keys_id_seq START WITH 1;
		DROP TABLE IF EXISTS "my_node_keys"; CREATE TABLE "my_node_keys" (
		"id" int NOT NULL  default nextval('my_node_keys_id_seq'),
		"add_time" int NOT NULL DEFAULT '0',
		"public_key" bytea  NOT NULL DEFAULT '',
		"private_key" varchar(3096) NOT NULL DEFAULT '',
		"status" my_node_keys_enum_status  NOT NULL DEFAULT 'my_pending',
		"my_time" int NOT NULL DEFAULT '0',
		"time" bigint NOT NULL DEFAULT '0',
		"block_id" int NOT NULL DEFAULT '0',
		"rb_id" int NOT NULL DEFAULT '0'
		);
		ALTER SEQUENCE my_node_keys_id_seq owned by my_node_keys.id;
		ALTER TABLE ONLY "my_node_keys" ADD CONSTRAINT my_node_keys_pkey PRIMARY KEY (id);
		
		DROP TABLE IF EXISTS "stop_daemons"; CREATE TABLE "stop_daemons" (
		"stop_time" int NOT NULL DEFAULT '0'
		);
		`

	SchemaVDE = `DROP TABLE IF EXISTS "%[1]d_vde_languages"; CREATE TABLE "%[1]d_vde_languages" (
		"id" bigint  NOT NULL DEFAULT '0',
		"name" character varying(100) NOT NULL DEFAULT '',
		"res" text NOT NULL DEFAULT ''
	  );
	  ALTER TABLE ONLY "%[1]d_vde_languages" ADD CONSTRAINT "%[1]d_vde_languages_pkey" PRIMARY KEY (id);
	  CREATE INDEX "%[1]d_vde_languages_index_name" ON "%[1]d_vde_languages" (name);
	  
	  DROP TABLE IF EXISTS "%[1]d_vde_menu"; CREATE TABLE "%[1]d_vde_menu" (
		  "id" bigint  NOT NULL DEFAULT '0',
		  "name" character varying(255) UNIQUE NOT NULL DEFAULT '',
		  "title" character varying(255) NOT NULL DEFAULT '',
		  "value" text NOT NULL DEFAULT '',
		  "conditions" text NOT NULL DEFAULT ''
	  );
	  ALTER TABLE ONLY "%[1]d_vde_menu" ADD CONSTRAINT "%[1]d_vde_menu_pkey" PRIMARY KEY (id);
	  CREATE INDEX "%[1]d_vde_menu_index_name" ON "%[1]d_vde_menu" (name);
	  
	  DROP TABLE IF EXISTS "%[1]d_vde_pages"; CREATE TABLE "%[1]d_vde_pages" (
		  "id" bigint  NOT NULL DEFAULT '0',
		  "name" character varying(255) UNIQUE NOT NULL DEFAULT '',
		  "value" text NOT NULL DEFAULT '',
		  "menu" character varying(255) NOT NULL DEFAULT '',
		  "conditions" text NOT NULL DEFAULT ''
	  );
	  ALTER TABLE ONLY "%[1]d_vde_pages" ADD CONSTRAINT "%[1]d_vde_pages_pkey" PRIMARY KEY (id);
	  CREATE INDEX "%[1]d_vde_pages_index_name" ON "%[1]d_vde_pages" (name);
	  
	  DROP TABLE IF EXISTS "%[1]d_vde_blocks"; CREATE TABLE "%[1]d_vde_blocks" (
		  "id" bigint  NOT NULL DEFAULT '0',
		  "name" character varying(255) UNIQUE NOT NULL DEFAULT '',
		  "value" text NOT NULL DEFAULT '',
		  "conditions" text NOT NULL DEFAULT ''
	  );
	  ALTER TABLE ONLY "%[1]d_vde_blocks" ADD CONSTRAINT "%[1]d_vde_blocks_pkey" PRIMARY KEY (id);
	  CREATE INDEX "%[1]d_vde_blocks_index_name" ON "%[1]d_vde_blocks" (name);
	  
	  DROP TABLE IF EXISTS "%[1]d_vde_signatures"; CREATE TABLE "%[1]d_vde_signatures" (
		  "id" bigint  NOT NULL DEFAULT '0',
		  "name" character varying(100) NOT NULL DEFAULT '',
		  "value" jsonb,
		  "conditions" text NOT NULL DEFAULT ''
	  );
	  ALTER TABLE ONLY "%[1]d_vde_signatures" ADD CONSTRAINT "%[1]d_vde_signatures_pkey" PRIMARY KEY (name);
	  
	  CREATE TABLE "%[1]d_vde_contracts" (
	  "id" bigint NOT NULL  DEFAULT '0',
	  "value" text  NOT NULL DEFAULT '',
	  "conditions" text  NOT NULL DEFAULT ''
	  );
	  ALTER TABLE ONLY "%[1]d_vde_contracts" ADD CONSTRAINT "%[1]d_vde_contracts_pkey" PRIMARY KEY (id);
	  
	  DROP TABLE IF EXISTS "%[1]d_vde_parameters";
	  CREATE TABLE "%[1]d_vde_parameters" (
	  "id" bigint NOT NULL  DEFAULT '0',
	  "name" varchar(255) UNIQUE NOT NULL DEFAULT '',
	  "value" text NOT NULL DEFAULT '',
	  "conditions" text  NOT NULL DEFAULT ''
	  );
	  ALTER TABLE ONLY "%[1]d_vde_parameters" ADD CONSTRAINT "%[1]d_vde_parameters_pkey" PRIMARY KEY ("id");
	  CREATE INDEX "%[1]d_vde_parameters_index_name" ON "%[1]d_vde_parameters" (name);
	  
	  INSERT INTO "%[1]d_vde_parameters" ("id","name", "value", "conditions") VALUES 
	  ('1','founder_account', '%[2]d', 'ContractConditions("MainCondition")'),
	  ('2','new_table', 'ContractConditions("MainCondition")', 'ContractConditions("MainCondition")'),
	  ('3','new_column', 'ContractConditions("MainCondition")', 'ContractConditions("MainCondition")'),
	  ('4','changing_tables', 'ContractConditions("MainCondition")', 'ContractConditions("MainCondition")'),
	  ('5','changing_language', 'ContractConditions("MainCondition")', 'ContractConditions("MainCondition")'),
	  ('6','changing_signature', 'ContractConditions("MainCondition")', 'ContractConditions("MainCondition")'),
	  ('7','changing_page', 'ContractConditions("MainCondition")', 'ContractConditions("MainCondition")'),
	  ('8','changing_menu', 'ContractConditions("MainCondition")', 'ContractConditions("MainCondition")'),
	  ('9','changing_contracts', 'ContractConditions("MainCondition")', 'ContractConditions("MainCondition")'),
	  ('10','stylesheet', 'body { 
		/* You can define your custom styles here or create custom CSS rules */
	  }', 'ContractConditions("MainCondition")');
	  
	  CREATE TABLE "%[1]d_vde_tables" (
	  "id" bigint NOT NULL  DEFAULT '0',
	  "name" varchar(100) UNIQUE NOT NULL DEFAULT '',
	  "permissions" jsonb,
	  "columns" jsonb,
	  "conditions" text  NOT NULL DEFAULT ''
	  );
	  ALTER TABLE ONLY "%[1]d_vde_tables" ADD CONSTRAINT "%[1]d_vde_tables_pkey" PRIMARY KEY ("id");
	  CREATE INDEX "%[1]d_vde_tables_index_name" ON "%[1]d_vde_tables" (name);
	  
	  INSERT INTO "%[1]d_vde_tables" ("id", "name", "permissions","columns", "conditions") VALUES ('1', 'contracts', 
			  '{"insert": "ContractAccess(\"NewContract\")", "update": "ContractAccess(\"EditContract\")", 
				"new_column": "ContractAccess(\"NewColumn\")"}',
			  '{"value": "ContractAccess(\"EditContract\")",
				"conditions": "ContractAccess(\"EditContract\")"}', 'ContractAccess("EditTable")'),
			  ('2', 'languages', 
			  '{"insert": "ContractAccess(\"NewLang\")", "update": "ContractAccess(\"EditLang\")", 
				"new_column": "ContractAccess(\"NewColumn\")"}',
			  '{ "name": "ContractAccess(\"EditLang\")",
				"res": "ContractAccess(\"EditLang\")",
				"conditions": "ContractAccess(\"EditLang\")"}', 'ContractAccess("EditTable")'),
			  ('3', 'menu', 
			  '{"insert": "ContractAccess(\"NewMenu\")", "update": "ContractAccess(\"EditMenu\", \"AppendMenu\")", 
				"new_column": "ContractAccess(\"NewColumn\")"}',
			  '{"name": "ContractAccess(\"EditMenu\")",
		  "value": "ContractAccess(\"EditMenu\", \"AppendMenu\")",
		  "conditions": "ContractAccess(\"EditMenu\")"
			  }', 'ContractAccess("EditTable")'),
			  ('4', 'pages', 
			  '{"insert": "ContractAccess(\"NewPage\")", "update": "ContractAccess(\"EditPage\", \"AppendPage\")", 
				"new_column": "ContractAccess(\"NewColumn\")"}',
			  '{"name": "ContractAccess(\"EditPage\")",
		  "value": "ContractAccess(\"EditPage\", \"AppendPage\")",
		  "menu": "ContractAccess(\"EditPage\")",
		  "conditions": "ContractAccess(\"EditPage\")"
			  }', 'ContractAccess("EditTable")'),
			  ('5', 'blocks', 
			  '{"insert": "ContractAccess(\"NewBlock\")", "update": "ContractAccess(\"EditBlock\")", 
				"new_column": "ContractAccess(\"NewColumn\")"}',
			  '{"name": "ContractAccess(\"EditBlock\")",
		  "value": "ContractAccess(\"EditBlock\")",
		  "conditions": "ContractAccess(\"EditBlock\")"
			  }', 'ContractAccess("EditTable")'),
			  ('6', 'signatures', 
			  '{"insert": "ContractAccess(\"NewSign\")", "update": "ContractAccess(\"EditSign\")", 
				"new_column": "ContractAccess(\"NewColumn\")"}',
			  '{"name": "ContractAccess(\"EditSign\")",
		  "value": "ContractAccess(\"EditSign\")",
		  "conditions": "ContractAccess(\"EditSign\")"
			  }', 'ContractAccess("EditTable")');
	  
	  INSERT INTO "%[1]d_vde_contracts" ("id", "value", "conditions") VALUES 
	  ('1','contract MainCondition {
		conditions {
		  if EcosysParam("founder_account")!=$key_id
		  {
			warning "Sorry, you do not have access to this action."
		  }
		}
	  }', 'ContractConditions("MainCondition")'),
	  ('2','contract VDEFunctions {}
	  
		func DBFind(table string).Columns(columns string).Where(where string, params ...)
			.WhereId(id int).Order(order string).Limit(limit int).Offset(offset int).Ecosystem(ecosystem int) array {
			return DBSelect(table, columns, id, order, offset, limit, ecosystem, where, params)
		}

		func DBRow(table string).Columns(columns string).Where(where string, params ...)
			.WhereId(id int).Order(order string).Ecosystem(ecosystem int) map {

			var result array
			result = DBFind(table).Columns(columns).Where(where, params ...).WhereId(id).Order(order).Ecosystem(ecosystem)

			var row map
			if Len(result) > 0 {
				row = result[0]
			}

			return row
		}

		func ConditionById(table string, validate bool) {
			var row map
			row = DBFind(table).Columns("conditions").WhereId($Id)
			if !row["conditions"] {
				error Sprintf("Item %%d has not been found", $Id)
			}

			Eval(row["conditions"])

			if validate {
				ValidateCondition($Conditions,$ecosystem_id)
			}
		}
	  ', 'ContractConditions("MainCondition")'),
	  ('3','contract NewContract {
		  data {
			  Value      string
			  Conditions string
		  }
		  conditions {
			ValidateCondition($Conditions,$ecosystem_id)
			  var list array
			  list = ContractsList($Value)
			  var i int
			  while i < Len(list) {
				  if IsObject(list[i], $ecosystem_id) {
					  warning Sprintf("Contract or function %%s exists", list[i] )
				  }
				  i = i + 1
			  }
		  }
		  action {
			  var root, id int
			  root = CompileContract($Value, $ecosystem_id, 0, 0)
			  id = DBInsert("contracts", "value,conditions", $Value, $Conditions )
			  FlushContract(root, id, false)
			  $result = id
		  }
	  }', 'ContractConditions("MainCondition")'),
	  ('4','contract EditContract {
		  data {
			  Id         int
			  Value      string
			  Conditions string
		  }
		  conditions {
			  var row array
			  row = DBFind("contracts").Columns("id,value,conditions").WhereId($Id)
			  if !Len(row) {
				  error Sprintf("Contract %%d does not exist", $Id)
			  }
			  $cur = row[0]
			  Eval($cur["conditions"])
			  ValidateCondition($Conditions,$ecosystem_id)
			  var list, curlist array
			  list = ContractsList($Value)
			  curlist = ContractsList($cur["value"])
			  if Len(list) != Len(curlist) {
				  error "Contracts cannot be removed or inserted"
			  }
			  var i int
			  while i < Len(list) {
				  var j int
				  var ok bool
				  while j < Len(curlist) {
					  if curlist[j] == list[i] {
						  ok = true
						  break
					  }
					  j = j + 1 
				  }
				  if !ok {
					  error "Contracts or functions names cannot be changed"
				  }
				  i = i + 1
			  }
		  }
		  action {
			  var root int
			  root = CompileContract($Value, $ecosystem_id, 0, 0)
			  DBUpdate("contracts", $Id, "value,conditions", $Value, $Conditions)
			  FlushContract(root, $Id, false)
		  }
	  }', 'ContractConditions("MainCondition")'),
	  ('5','contract NewParameter {
		  data {
			  Name string
			  Value string
			  Conditions string
		  }
		  conditions {
			  var ret array
			  ValidateCondition($Conditions, $ecosystem_id)
			  ret = DBFind("parameters").Columns("id").Where("name=?", $Name).Limit(1)
			  if Len(ret) > 0 {
				  warning Sprintf( "Parameter %%s already exists", $Name)
			  }
		  }
		  action {
			  $result = DBInsert("parameters", "name,value,conditions", $Name, $Value, $Conditions )
		  }
	  }', 'ContractConditions("MainCondition")'),
	  ('6','contract EditParameter {
		  data {
			  Id int
			  Value string
			  Conditions string
		  }
		  conditions {
			  ConditionById("parameters", true)
		  }
		  action {
			  DBUpdate("parameters", $Id, "value,conditions", $Value, $Conditions )
		  }
	  }', 'ContractConditions("MainCondition")'),
	  ('7', 'contract NewMenu {
		  data {
			  Name       string
			  Value      string
			  Title      string "optional"
			  Conditions string
		  }
		  conditions {
			  var ret int
			  ValidateCondition($Conditions,$ecosystem_id)
			  ret = DBFind("menu").Columns("id").Where("name=?", $Name).Limit(1)
			  if Len(ret) > 0 {
				  warning Sprintf( "Menu %%s already exists", $Name)
			  }
		  }
		  action {
			  $result = DBInsert("menu", "name,value,title,conditions", $Name, $Value, $Title, $Conditions )
		  }
	  }', 'ContractConditions("MainCondition")'),
	  ('8','contract EditMenu {
		  data {
			  Id         int
			  Value      string
			  Title      string "optional"
			  Conditions string
		  }
		  conditions {
			  ConditionById("menu", true)
		  }
		  action {
			  DBUpdate("menu", $Id, "value,title,conditions", $Value, $Title, $Conditions)
		  }
	  }', 'ContractConditions("MainCondition")'),
	  ('9','contract AppendMenu {
		data {
			Id     int
			Value  string
		}
		conditions {
			ConditionById("menu", false)
		}
		action {
			var row map
			row = DBRow("menu").Columns("value").WhereId($Id)
			DBUpdate("menu", $Id, "value", row["value"] + "\r\n" + $Value)
		}
	  }', 'ContractConditions("MainCondition")'),
	  ('10','contract NewPage {
		  data {
			  Name       string
			  Value      string
			  Menu       string
			  Conditions string
		  }
		  conditions {
			  var ret int
			  ValidateCondition($Conditions,$ecosystem_id)
			  ret = DBFind("pages").Columns("id").Where("name=?", $Name).Limit(1)
			  if Len(ret) > 0 {
				  warning Sprintf( "Page %%s already exists", $Name)
			  }
		  }
		  action {
			  $result = DBInsert("pages", "name,value,menu,conditions", $Name, $Value, $Menu, $Conditions )
		  }
	  }', 'ContractConditions("MainCondition")'),
	  ('11','contract EditPage {
		  data {
			  Id         int
			  Value      string
			  Menu      string
			  Conditions string
		  }
		  conditions {
			  ConditionById("pages", true)
		  }
		  action {
			  DBUpdate("pages", $Id, "value,menu,conditions", $Value, $Menu, $Conditions)
		  }
	  }', 'ContractConditions("MainCondition")'),
	  ('12','contract AppendPage {
		  data {
			  Id         int
			  Value      string
		  }
		  conditions {
			  ConditionById("pages", false)
		  }
		  action {
			  var row map
			  row = DBRow("pages").Columns("value").WhereId($Id)
			  DBUpdate("pages", $Id, "value", row["value"] + "\r\n" + $Value)
		  }
	  }', 'ContractConditions("MainCondition")'),
	  ('13','contract NewBlock {
		  data {
			  Name       string
			  Value      string
			  Conditions string
		  }
		  conditions {
			  var ret int
			  ValidateCondition($Conditions,$ecosystem_id)
			  ret = DBFind("blocks").Columns("id").Where("name=?", $Name).Limit(1)
			  if Len(ret) > 0 {
				  warning Sprintf( "Block %%s already exists", $Name)
			  }
		  }
		  action {
			  $result = DBInsert("blocks", "name,value,conditions", $Name, $Value, $Conditions )
		  }
	  }', 'ContractConditions("MainCondition")'),
	  ('14','contract EditBlock {
		  data {
			  Id         int
			  Value      string
			  Conditions string
		  }
		  conditions {
			  ConditionById("blocks", true)
		  }
		  action {
			  DBUpdate("blocks", $Id, "value,conditions", $Value, $Conditions)
		  }
	  }', 'ContractConditions("MainCondition")'),
	  ('15','contract NewTable {
		  data {
			  Name       string
			  Columns      string
			  Permissions string
		  }
		  conditions {
			  TableConditions($Name, $Columns, $Permissions)
		  }
		  action {
			  CreateTable($Name, $Columns, $Permissions)
		  }
	  }', 'ContractConditions("MainCondition")'),
	  ('16','contract EditTable {
		  data {
			  Name       string
			  Permissions string
		  }
		  conditions {
			  TableConditions($Name, "", $Permissions)
		  }
		  action {
			  PermTable($Name, $Permissions )
		  }
	  }', 'ContractConditions("MainCondition")'),
	  ('17','contract NewColumn {
		  data {
			  TableName   string
			  Name        string
			  Type        string
			  Permissions string
			  Index       string "optional"
		  }
		  conditions {
			  ColumnCondition($TableName, $Name, $Type, $Permissions, $Index)
		  }
		  action {
			  CreateColumn($TableName, $Name, $Type, $Permissions, $Index)
		  }
	  }', 'ContractConditions("MainCondition")'),
	  ('18','contract EditColumn {
		  data {
			  TableName   string
			  Name        string
			  Permissions string
		  }
		  conditions {
			  ColumnCondition($TableName, $Name, "", $Permissions, "")
		  }
		  action {
			  PermColumn($TableName, $Name, $Permissions)
		  }
	  }', 'ContractConditions("MainCondition")'),
	  ('19','contract NewLang {
		data {
			Name  string
			Trans string
		}
		conditions {
			EvalCondition("parameters", "changing_language", "value")
			var row array
			row = DBFind("languages").Columns("name").Where("name=?", $Name).Limit(1)
			if Len(row) > 0 {
				error Sprintf("The language resource %%s already exists", $Name)
			}
		}
		action {
			DBInsert("languages", "name,res", $Name, $Trans )
			UpdateLang($Name, $Trans)
		}
	}', 'ContractConditions("MainCondition")'),
	('20','contract EditLang {
		data {
			Name  string
			Trans string
		}
		conditions {
			EvalCondition("parameters", "changing_language", "value")
		}
		action {
			DBUpdateExt("languages", "name", $Name, "res", $Trans )
			UpdateLang($Name, $Trans)
		}
	}', 'ContractConditions("MainCondition")'),
	('21','func ImportList(row array, cnt string) {
		if !row {
			return
		}
		var i int
		while i < Len(row) {
			var idata map
			idata = row[i]
			CallContract(cnt, idata)
			i = i + 1
		}
	}
	
	func ImportData(row array) {
		if !row {
			return
		}
		var i int
		while i < Len(row) {
			var idata map
			var list array
			var tblname, columns string
			idata = row[i]
			i = i + 1
			tblname = idata["Table"]
			columns = Join(idata["Columns"], ",")
			list = idata["Data"] 
			if !list {
				continue
			}
			var j int
			while j < Len(list) {
				var ilist array
				ilist = list[j]
				DBInsert(tblname, columns, ilist)
				j=j+1
			}
		}
	}
	
	contract Import {
		data {
			Data string
		}
		conditions {
			$list = JSONToMap($Data)
		}
		action {
			ImportList($list["pages"], "NewPage")
			ImportList($list["blocks"], "NewBlock")
			ImportList($list["menus"], "NewMenu")
			ImportList($list["parameters"], "NewParameter")
			ImportList($list["languages"], "NewLang")
			ImportList($list["contracts"], "NewContract")
			ImportList($list["tables"], "NewTable")
			ImportData($list["data"])
		}
	}', 'ContractConditions("MainCondition")');
	  `

	SchemaEcosystem = `DROP TABLE IF EXISTS "%[1]d_keys"; CREATE TABLE "%[1]d_keys" (
		"id" bigint  NOT NULL DEFAULT '0',
		"pub" bytea  NOT NULL DEFAULT '',
		"amount" decimal(30) NOT NULL DEFAULT '0',
		"rb_id" bigint NOT NULL DEFAULT '0'
		);
		ALTER TABLE ONLY "%[1]d_keys" ADD CONSTRAINT "%[1]d_keys_pkey" PRIMARY KEY (id);
		
		DROP TABLE IF EXISTS "%[1]d_history"; CREATE TABLE "%[1]d_history" (
		"id" bigint NOT NULL  DEFAULT '0',
		"sender_id" bigint NOT NULL DEFAULT '0',
		"recipient_id" bigint NOT NULL DEFAULT '0',
		"amount" decimal(30) NOT NULL DEFAULT '0',
		"comment" text NOT NULL DEFAULT '',
		"block_id" int  NOT NULL DEFAULT '0',
		"txhash" bytea  NOT NULL DEFAULT '',
		"rb_id" int  NOT NULL DEFAULT '0'
		);
		ALTER TABLE ONLY "%[1]d_history" ADD CONSTRAINT "%[1]d_history_pkey" PRIMARY KEY (id);
		CREATE INDEX "%[1]d_history_index_sender" ON "%[1]d_history" (sender_id);
		CREATE INDEX "%[1]d_history_index_recipient" ON "%[1]d_history" (recipient_id);
		CREATE INDEX "%[1]d_history_index_block" ON "%[1]d_history" (block_id, txhash);
		
		
		DROP TABLE IF EXISTS "%[1]d_languages"; CREATE TABLE "%[1]d_languages" (
		  "id" bigint  NOT NULL DEFAULT '0',
		  "name" character varying(100) NOT NULL DEFAULT '',
		  "res" text NOT NULL DEFAULT '',
		  "conditions" text NOT NULL DEFAULT '',
		  "rb_id" bigint NOT NULL DEFAULT '0'
		);
		ALTER TABLE ONLY "%[1]d_languages" ADD CONSTRAINT "%[1]d_languages_pkey" PRIMARY KEY (id);
		CREATE INDEX "%[1]d_languages_index_name" ON "%[1]d_languages" (name);
		
		DROP TABLE IF EXISTS "%[1]d_menu"; CREATE TABLE "%[1]d_menu" (
			"id" bigint  NOT NULL DEFAULT '0',
			"name" character varying(255) UNIQUE NOT NULL DEFAULT '',
			"title" character varying(255) NOT NULL DEFAULT '',
			"value" text NOT NULL DEFAULT '',
			"conditions" text NOT NULL DEFAULT '',
			"rb_id" bigint NOT NULL DEFAULT '0'
		);
		ALTER TABLE ONLY "%[1]d_menu" ADD CONSTRAINT "%[1]d_menu_pkey" PRIMARY KEY (id);
		CREATE INDEX "%[1]d_menu_index_name" ON "%[1]d_menu" (name);
		
		DROP TABLE IF EXISTS "%[1]d_pages"; CREATE TABLE "%[1]d_pages" (
			"id" bigint  NOT NULL DEFAULT '0',
			"name" character varying(255) UNIQUE NOT NULL DEFAULT '',
			"value" text NOT NULL DEFAULT '',
			"menu" character varying(255) NOT NULL DEFAULT '',
			"conditions" text NOT NULL DEFAULT '',
			"rb_id" bigint NOT NULL DEFAULT '0'
		);
		ALTER TABLE ONLY "%[1]d_pages" ADD CONSTRAINT "%[1]d_pages_pkey" PRIMARY KEY (id);
		CREATE INDEX "%[1]d_pages_index_name" ON "%[1]d_pages" (name);
		
		DROP TABLE IF EXISTS "%[1]d_blocks"; CREATE TABLE "%[1]d_blocks" (
			"id" bigint  NOT NULL DEFAULT '0',
			"name" character varying(255) UNIQUE NOT NULL DEFAULT '',
			"value" text NOT NULL DEFAULT '',
			"conditions" text NOT NULL DEFAULT '',
			"rb_id" bigint NOT NULL DEFAULT '0'
		);
		ALTER TABLE ONLY "%[1]d_blocks" ADD CONSTRAINT "%[1]d_blocks_pkey" PRIMARY KEY (id);
		CREATE INDEX "%[1]d_blocks_index_name" ON "%[1]d_blocks" (name);
		
		DROP TABLE IF EXISTS "%[1]d_signatures"; CREATE TABLE "%[1]d_signatures" (
			"id" bigint  NOT NULL DEFAULT '0',
			"name" character varying(100) NOT NULL DEFAULT '',
			"value" jsonb,
			"conditions" text NOT NULL DEFAULT '',
			"rb_id" bigint NOT NULL DEFAULT '0'
		);
		ALTER TABLE ONLY "%[1]d_signatures" ADD CONSTRAINT "%[1]d_signatures_pkey" PRIMARY KEY (name);
		
		CREATE TABLE "%[1]d_contracts" (
		"id" bigint NOT NULL  DEFAULT '0',
		"value" text  NOT NULL DEFAULT '',
		"wallet_id" bigint NOT NULL DEFAULT '0',
		"token_id" bigint NOT NULL DEFAULT '1',
		"active" character(1) NOT NULL DEFAULT '0',
		"conditions" text  NOT NULL DEFAULT '',
		"rb_id" bigint NOT NULL DEFAULT '0'
		);
		ALTER TABLE ONLY "%[1]d_contracts" ADD CONSTRAINT "%[1]d_contracts_pkey" PRIMARY KEY (id);
		
		INSERT INTO "%[1]d_contracts" ("id", "value", "wallet_id","active", "conditions") VALUES 
		('1','contract MainCondition {
		  conditions {
			if EcosysParam("founder_account")!=$key_id
			{
			  warning "Sorry, you do not have access to this action."
			}
		  }
		}', '%[2]d', '0', 'ContractConditions("MainCondition")');
		
		DROP TABLE IF EXISTS "%[1]d_parameters";
		CREATE TABLE "%[1]d_parameters" (
		"id" bigint NOT NULL  DEFAULT '0',
		"name" varchar(255) UNIQUE NOT NULL DEFAULT '',
		"value" text NOT NULL DEFAULT '',
		"conditions" text  NOT NULL DEFAULT '',
		"rb_id" bigint  NOT NULL DEFAULT '0'
		);
		ALTER TABLE ONLY "%[1]d_parameters" ADD CONSTRAINT "%[1]d_parameters_pkey" PRIMARY KEY ("id");
		CREATE INDEX "%[1]d_parameters_index_name" ON "%[1]d_parameters" (name);
		
		INSERT INTO "%[1]d_parameters" ("id","name", "value", "conditions") VALUES 
		('1','founder_account', '%[2]d', 'ContractConditions("MainCondition")'),
		('2','new_table', 'ContractConditions("MainCondition")', 'ContractConditions("MainCondition")'),
		('3','changing_tables', 'ContractConditions("MainCondition")', 'ContractConditions("MainCondition")'),
		('4','changing_language', 'ContractConditions("MainCondition")', 'ContractConditions("MainCondition")'),
		('5','changing_signature', 'ContractConditions("MainCondition")', 'ContractConditions("MainCondition")'),
		('6','changing_page', 'ContractConditions("MainCondition")', 'ContractConditions("MainCondition")'),
		('7','changing_menu', 'ContractConditions("MainCondition")', 'ContractConditions("MainCondition")'),
		('8','changing_contracts', 'ContractConditions("MainCondition")', 'ContractConditions("MainCondition")'),
		('9','ecosystem_name', '%[3]s', 'ContractConditions("MainCondition")'),
		('10','max_sum', '1000000', 'ContractConditions("MainCondition")'),
		('11','money_digit', '2', 'ContractConditions("MainCondition")'),
		('12','stylesheet', 'body {
		  /* You can define your custom styles here or create custom CSS rules */
		}', 'ContractConditions("MainCondition")');
		
		CREATE TABLE "%[1]d_tables" (
		"id" bigint NOT NULL  DEFAULT '0',
		"name" varchar(100) UNIQUE NOT NULL DEFAULT '',
		"permissions" jsonb,
		"columns" jsonb,
		"conditions" text  NOT NULL DEFAULT '',
		"rb_id" bigint NOT NULL DEFAULT '0'
		);
		ALTER TABLE ONLY "%[1]d_tables" ADD CONSTRAINT "%[1]d_tables_pkey" PRIMARY KEY ("id");
		CREATE INDEX "%[1]d_tables_index_name" ON "%[1]d_tables" (name);
		
		INSERT INTO "%[1]d_tables" ("id", "name", "permissions","columns", "conditions") VALUES ('1', 'contracts', 
			'{"insert": "ContractAccess(\"@1NewContract\")", "update": "ContractAccess(\"@1EditContract\",\"@1ActivateContract\", \"@1DeactivateContract\")",
				  "new_column": "ContractAccess(\"@1NewColumn\")"}',
				'{"value": "ContractAccess(\"@1EditContract\", \"@1ActivateContract\", \"@1DeactivateContract\")",
				  "wallet_id": "ContractAccess(\"@1EditContract\", \"@1ActivateContract\", \"@1DeactivateContract\")",
				  "token_id": "ContractAccess(\"@1EditContract\", \"@1ActivateContract\", \"@1DeactivateContract\")",
				  "active": "ContractAccess(\"@1EditContract\", \"@1ActivateContract\", \"@1DeactivateContract\")",
				  "conditions": "ContractAccess(\"@1EditContract\", \"@1ActivateContract\", \"@1DeactivateContract\")"}', 'ContractAccess("@1EditTable")'),
				('2', 'keys', 
				'{"insert": "ContractAccess(\"@1MoneyTransfer\", \"@1NewEcosystem\")", "update": "ContractAccess(\"@1MoneyTransfer\")", 
				  "new_column": "ContractAccess(\"@1NewColumn\")"}',
				'{"pub": "ContractAccess(\"@1MoneyTransfer\")",
				  "amount": "ContractAccess(\"@1MoneyTransfer\")"}', 'ContractAccess("@1EditTable")'),
				('3', 'history', 
				'{"insert": "ContractAccess(\"@1MoneyTransfer\")", "update": "false", 
				  "new_column": "false"}',
				'{"sender_id": "ContractAccess(\"@1MoneyTransfer\")",
				  "recipient_id": "ContractAccess(\"@1MoneyTransfer\")",
				  "amount":  "ContractAccess(\"@1MoneyTransfer\")",
				  "comment": "ContractAccess(\"@1MoneyTransfer\")",
				  "block_id":  "ContractAccess(\"@1MoneyTransfer\")",
				  "txhash": "ContractAccess(\"@1MoneyTransfer\")"}', 'ContractAccess("@1EditTable")'),        
				('4', 'languages', 
				'{"insert": "ContractAccess(\"@1NewLang\")", "update": "ContractAccess(\"@1EditLang\")", 
				  "new_column": "ContractAccess(\"@1NewColumn\")"}',
				'{ "name": "ContractAccess(\"@1EditLang\")",
				  "res": "ContractAccess(\"@1EditLang\")",
				  "conditions": "ContractAccess(\"@1EditLang\")"}', 'ContractAccess("@1EditTable")'),
				('5', 'menu', 
					'{"insert": "ContractAccess(\"@1NewMenu\", \"@1NewEcosystem\")", "update": "ContractAccess(\"@1EditMenu\",\"@1AppendMenu\")", 
				  "new_column": "ContractAccess(\"@1NewColumn\")"}',
				'{"name": "ContractAccess(\"@1EditMenu\")",
			"value": "ContractAccess(\"@1EditMenu\",\"@1AppendMenu\")",
			"conditions": "ContractAccess(\"@1EditMenu\")"
				}', 'ContractAccess("@1EditTable")'),
				('6', 'pages', 
					'{"insert": "ContractAccess(\"@1NewPage\", \"@1NewEcosystem\")", "update": "ContractAccess(\"@1EditPage\",\"@1AppendPage\")", 
				  "new_column": "ContractAccess(\"@1NewColumn\")"}',
				'{"name": "ContractAccess(\"@1EditPage\")",
			"value": "ContractAccess(\"@1EditPage\",\"@1AppendPage\")",
			"menu": "ContractAccess(\"@1EditPage\")",
			"conditions": "ContractAccess(\"@1EditPage\")"
				}', 'ContractAccess("@1EditTable")'),
				('7', 'blocks', 
				'{"insert": "ContractAccess(\"@1NewBlock\")", "update": "ContractAccess(\"@1EditBlock\")", 
				  "new_column": "ContractAccess(\"@1NewColumn\")"}',
				'{"name": "ContractAccess(\"@1EditBlock\")",
			"value": "ContractAccess(\"@1EditBlock\")",
			"conditions": "ContractAccess(\"@1EditBlock\")"
				}', 'ContractAccess("@1EditTable")'),
				('8', 'signatures', 
				'{"insert": "ContractAccess(\"@1NewSign\")", "update": "ContractAccess(\"@1EditSign\")", 
				  "new_column": "ContractAccess(\"@1NewColumn\")"}',
				'{"name": "ContractAccess(\"@1EditSign\")",
			"value": "ContractAccess(\"@1EditSign\")",
			"conditions": "ContractAccess(\"@1EditSign\")"
				}', 'ContractAccess("@1EditTable")');
		
		`

	SchemaFirstEcosystem = `INSERT INTO "system_states" ("id","rb_id") VALUES ('1','0');
	
	INSERT INTO "1_contracts" ("id","value", "wallet_id", "conditions") VALUES 
	('2','contract SystemFunctions {
	}
	
	func DBFind(table string).Columns(columns string).Where(where string, params ...)
		 .WhereId(id int).Order(order string).Limit(limit int).Offset(offset int).Ecosystem(ecosystem int) array {
		return DBSelect(table, columns, id, order, offset, limit, ecosystem, where, params)
	}

	func DBRow(table string).Columns(columns string).Where(where string, params ...)
		.WhereId(id int).Order(order string).Ecosystem(ecosystem int) map {
		
		var result array
		result = DBFind(table).Columns(columns).Where(where, params ...).WhereId(id).Order(order).Ecosystem(ecosystem)

		var row map
		if Len(result) > 0 {
			row = result[0]
		}

		return row
	}
	
	func ConditionById(table string, validate bool) {
		var row map
		row = DBRow(table).Columns("conditions").WhereId($Id)
		if !row["conditions"] {
			error Sprintf("Item %%d has not been found", $Id)
		}

		Eval(row["conditions"])

		if validate {
			ValidateCondition($Conditions,$ecosystem_id)
		}
	}
	
	', '%[1]d','ContractConditions("MainCondition")'),
	('3','contract MoneyTransfer {
		data {
			Recipient string
			Amount    string
			Comment     string "optional"
		}
		conditions {
			$recipient = AddressToId($Recipient)
			if $recipient == 0 {
				error Sprintf("Recipient %%s is invalid", $Recipient)
			}
			var total money
			$amount = Money($Amount) 
			if $amount == 0 {
				error "Amount is zero"
			}
			var row map
			row = DBRow("keys").Columns("amount").WhereId($key_id)
			total = Money(row["amount"])
			if $amount >= total {
				error Sprintf("Money is not enough %%v < %%v",total, $amount)
			}
		}
		action {
			DBUpdate("keys", $key_id,"-amount", $amount)
			DBUpdate("keys", $recipient,"+amount", $amount)
			DBInsert("history", "sender_id,recipient_id,amount,comment,block_id,txhash", 
				$key_id, $recipient, $amount, $Comment, $block, $txhash)
		}
	}', '%[1]d', 'ContractConditions("MainCondition")'),
	('4','contract NewContract {
		data {
			Value      string
			Conditions string
			Wallet         string "optional"
			TokenEcosystem int "optional"
		}
		conditions {
			ValidateCondition($Conditions,$ecosystem_id)
			$walletContract = $key_id
			   if $Wallet {
				$walletContract = AddressToId($Wallet)
				if $walletContract == 0 {
				   error Sprintf("wrong wallet %%s", $Wallet)
				}
			}
			var list array
			list = ContractsList($Value)
			var i int
			while i < Len(list) {
				if IsObject(list[i], $ecosystem_id) {
					warning Sprintf("Contract or function %%s exists", list[i] )
				}
				i = i + 1
			}
			if !$TokenEcosystem {
				$TokenEcosystem = 1
			} else {
				if !SysFuel($TokenEcosystem) {
					warning Sprintf("Ecosystem %%d is not system", $TokenEcosystem )
				}
			}
		}
		action {
			var root, id int
			root = CompileContract($Value, $ecosystem_id, $walletContract, $TokenEcosystem)
			id = DBInsert("contracts", "value,conditions, wallet_id, token_id", 
				   $Value, $Conditions, $walletContract, $TokenEcosystem)
			FlushContract(root, id, false)
			$result = id
		}
		func price() int {
			return  SysParamInt("contract_price")
		}
	}', '%[1]d', 'ContractConditions("MainCondition")'),
	('5','contract EditContract {
		data {
			Id         int
			Value      string
			Conditions string
		}
		conditions {
			$cur = DBRow("contracts").Columns("id,value,conditions,active,wallet_id,token_id").WhereId($Id)
			if !$cur {
				error Sprintf("Contract %%d does not exist", $Id)
			}
			Eval($cur["conditions"])
			ValidateCondition($Conditions,$ecosystem_id)
			var list, curlist array
			list = ContractsList($Value)
			curlist = ContractsList($cur["value"])
			if Len(list) != Len(curlist) {
				error "Contracts cannot be removed or inserted"
			}
			var i int
			while i < Len(list) {
				var j int
				var ok bool
				while j < Len(curlist) {
					if curlist[j] == list[i] {
						ok = true
						break
					}
					j = j + 1 
				}
				if !ok {
					error "Contracts or functions names cannot be changed"
				}
				i = i + 1
			}
		}
		action {
			var root int
			root = CompileContract($Value, $ecosystem_id, Int($cur["wallet_id"]), Int($cur["token_id"]))
			DBUpdate("contracts", $Id, "value,conditions", $Value, $Conditions)
			FlushContract(root, $Id, Int($cur["active"]) == 1)
		}
	}', '%[1]d','ContractConditions("MainCondition")'),
	('6','contract ActivateContract {
		data {
			Id  int
		}
		conditions {
			$cur = DBRow("contracts").Columns("id,conditions,active,wallet_id").WhereId($Id)
			if !$cur {
				error Sprintf("Contract %%d does not exist", $Id)
			}
			if Int($cur["active"]) == 1 {
				error Sprintf("The contract %%d has been already activated", $Id)
			}
			Eval($cur["conditions"])
			if $key_id != Int($cur["wallet_id"]) {
				error Sprintf("Wallet %%d cannot activate the contract", $key_id)
			}
		}
		action {
			DBUpdate("contracts", $Id, "active", 1)
			Activate($Id, $ecosystem_id)
		}
	}', '%[1]d','ContractConditions("MainCondition")'),
	('7','contract NewEcosystem {
		data {
			Name  string "optional"
		}
		conditions {
			if $Name && FindEcosystem($Name) {
				error Sprintf("Ecosystem %%s is already existed", $Name)
			}
		}
		action {
			$result = CreateEcosystem($key_id, $Name)
		}
		func price() int {
			return  SysParamInt("ecosystem_price")
		}
		func rollback() {
			RollbackEcosystem()
		}
	}', '%[1]d','ContractConditions("MainCondition")'),
	('8','contract NewParameter {
		data {
			Name string
			Value string
			Conditions string
		}
		conditions {
			ValidateCondition($Conditions, $ecosystem_id)

			var row map
			row = DBRow("parameters").Columns("id").Where("name = ?", $Name)

			if row {
				warning Sprintf( "Parameter %%s already exists", $Name)
			}
		}
		action {
			DBInsert("parameters", "name,value,conditions", $Name, $Value, $Conditions )
		}
	}', '%[1]d','ContractConditions("MainCondition")'),
	('9','contract EditParameter {
		data {
			Id int
			Value string
			Conditions string
		}
		conditions {
			ConditionById("parameters", true)
			ValidateCondition($Conditions, $ecosystem_id)
			var exist int
			var row map
			row = DBRow("parameters").Columns("name").WhereId($Id)
			if row["name"] == "ecosystem_name" {
				exist = FindEcosystem($Value)
				if exist > 0 && exist != $ecosystem_id {
					warning Sprintf("Ecosystem %%s already exists", $Value)
				}
			}
		}
		action {
			DBUpdate("parameters", $Id, "value,conditions", $Value, $Conditions )
			var row map
			row = DBRow("parameters").Columns("name").WhereId($Id)
			if row["name"] == "ecosystem_name" {
				DBUpdate("system_states", $ecosystem_id, "name", $Value)
			}
		}
	}', '%[1]d','ContractConditions("MainCondition")'),
	('10', 'contract NewMenu {
		data {
			Name       string
			Value      string
			Title      string "optional"
			Conditions string
		}
		conditions {
			ValidateCondition($Conditions,$ecosystem_id)

			var row map
			row = DBRow("menu").Columns("id").Where("name = ?", $Name)

			if row {
				warning Sprintf( "Menu %%s already exists", $Name)
			}
		}
		action {
			DBInsert("menu", "name,value,title,conditions", $Name, $Value, $Title, $Conditions )
		}
		func price() int {
			return  SysParamInt("menu_price")
		}
	}', '%[1]d','ContractConditions("MainCondition")'),
	('11','contract EditMenu {
		data {
			Id         int
			Value      string
			Title      string "optional"
			Conditions string
		}
		conditions {
			ConditionById("menu", true)
		}
		action {
			DBUpdate("menu", $Id, "value,title,conditions", $Value, $Title, $Conditions)
		}
	}', '%[1]d','ContractConditions("MainCondition")'),
	('12','contract AppendMenu {
		data {
			Id     int
			Value      string
		}
		conditions {
			ConditionById("menu", false)
		}
		action {
			var row map
			row = DBRow("menu").Columns("value").WhereId($Id)
			DBUpdate("menu", $Id, "value", row["value"] + "\r\n" + $Value)
		}
	}', '%[1]d','ContractConditions("MainCondition")'),
	('13','contract NewPage {
		data {
			Name       string
			Value      string
			Menu       string
			Conditions string
		}
		conditions {
			ValidateCondition($Conditions,$ecosystem_id)

			var row map
			row = DBRow("pages").Columns("id").Where("name = ?", $Name)

			if row {
				warning Sprintf( "Page %%s already exists", $Name)
			}
		}
		action {
			DBInsert("pages", "name,value,menu,conditions", $Name, $Value, $Menu, $Conditions )
		}
		func price() int {
			return  SysParamInt("page_price")
		}
	}', '%[1]d','ContractConditions("MainCondition")'),
	('14','contract EditPage {
		data {
			Id         int
			Value      string
			Menu      string
			Conditions string
		}
		conditions {
			ConditionById("pages", true)
		}
		action {
			DBUpdate("pages", $Id, "value,menu,conditions", $Value, $Menu, $Conditions)
		}
	}', '%[1]d','ContractConditions("MainCondition")'),
	('15','contract AppendPage {
		data {
			Id         int
			Value      string
		}
		conditions {
			ConditionById("pages", false)
		}
		action {
			var value string
			var row map
			row = DBRow("pages").Columns("value").WhereId($Id)
			value = row["value"]
			if Contains(value, "PageEnd:") {
				value = Replace(value, "PageEnd:", $Value) + "\r\nPageEnd:"
			} else {
				value = value + "\r\n" + $Value
			}
			DBUpdate("pages", $Id, "value",  value )
		}
	}', '%[1]d','ContractConditions("MainCondition")'),
	('16','contract NewLang {
		data {
			Name  string
			Trans string
		}
		conditions {
			EvalCondition("parameters", "changing_language", "value")

			var row map
			row = DBRow("languages").Columns("id").Where("name = ?", $Name)

			if row {
				error Sprintf("The language resource %%s already exists", $Name)
			}
		}
		action {
			DBInsert("languages", "name,res", $Name, $Trans )
			UpdateLang($Name, $Trans)
		}
	}', '%[1]d','ContractConditions("MainCondition")'),
	('17','contract EditLang {
		data {
			Name  string
			Trans string
		}
		conditions {
			EvalCondition("parameters", "changing_language", "value")
		}
		action {
			DBUpdateExt("languages", "name", $Name, "res", $Trans )
			UpdateLang($Name, $Trans)
		}
	}', '%[1]d','ContractConditions("MainCondition")'),
	('18','contract NewSign {
		data {
			Name       string
			Value      string
			Conditions string
		}
		conditions {
			ValidateCondition($Conditions,$ecosystem_id)
			var exist string

			var row map
			row = DBRow("signatures").Columns("id").Where("name = ?", $Name)

			if row {
				error Sprintf("The signature %%s already exists", $Name)
			}
		}
		action {
			DBInsert("signatures", "name,value,conditions", $Name, $Value, $Conditions )
		}
	}', '%[1]d','ContractConditions("MainCondition")'),
	('19','contract EditSign {
		data {
			Id         int
			Value      string
			Conditions string
		}
		conditions {
			ConditionById("signatures", true)
		}
		action {
			DBUpdate("signatures", $Id, "value,conditions", $Value, $Conditions)
		}
	}', '%[1]d','ContractConditions("MainCondition")'),
	('20','contract NewBlock {
		data {
			Name       string
			Value      string
			Conditions string
		}
		conditions {
			ValidateCondition($Conditions,$ecosystem_id)

			var row map
			row = DBRow("blocks").Columns("id").Where("name = ?", $Name)

			if row {
				warning Sprintf( "Block %%s already exists", $Name)
			}
		}
		action {
			DBInsert("blocks", "name,value,conditions", $Name, $Value, $Conditions )
		}
	}', '%[1]d','ContractConditions("MainCondition")'),
	('21','contract EditBlock {
		data {
			Id         int
			Value      string
			Conditions string
		}
		conditions {
			ConditionById("blocks", true)
		}
		action {
			DBUpdate("blocks", $Id, "value,conditions", $Value, $Conditions)
		}
	}', '%[1]d','ContractConditions("MainCondition")'),
	('22','contract NewTable {
		data {
			Name       string
			Columns      string
			Permissions string
		}
		conditions {
			TableConditions($Name, $Columns, $Permissions)
		}
		action {
			CreateTable($Name, $Columns, $Permissions)
		}
		func rollback() {
			RollbackTable($Name)
		}
		func price() int {
			return  SysParamInt("table_price")
		}
	}', '%[1]d','ContractConditions("MainCondition")'),
	('23','contract EditTable {
		data {
			Name       string
			Permissions string
		}
		conditions {
			TableConditions($Name, "", $Permissions)
		}
		action {
			PermTable($Name, $Permissions )
		}
	}', '%[1]d','ContractConditions("MainCondition")'),
	('24','contract NewColumn {
		data {
			TableName   string
			Name        string
			Type        string
			Permissions string
			Index       string "optional"
		}
		conditions {
			ColumnCondition($TableName, $Name, $Type, $Permissions, $Index)
		}
		action {
			CreateColumn($TableName, $Name, $Type, $Permissions, $Index)
		}
		func rollback() {
			RollbackColumn($TableName, $Name)
		}
		func price() int {
			return  SysParamInt("column_price")
		}
	}', '%[1]d','ContractConditions("MainCondition")'),
	('25','contract EditColumn {
		data {
			TableName   string
			Name        string
			Permissions string
		}
		conditions {
			ColumnCondition($TableName, $Name, "", $Permissions, "")
		}
		action {
			PermColumn($TableName, $Name, $Permissions)
		}
	}', '%[1]d','ContractConditions("MainCondition")'),
	('26','func ImportList(row array, cnt string) {
		if !row {
			return
		}
		var i int
		while i < Len(row) {
			var idata map
			idata = row[i]
			CallContract(cnt, idata)
			i = i + 1
		}
	}
	
	func ImportData(row array) {
		if !row {
			return
		}
		var i int
		while i < Len(row) {
			var idata map
			var list array
			var tblname, columns string
			idata = row[i]
			i = i + 1
			tblname = idata["Table"]
			columns = Join(idata["Columns"], ",")
			list = idata["Data"] 
			if !list {
				continue
			}
			var j int
			while j < Len(list) {
				var ilist array
				ilist = list[j]
				DBInsert(tblname, columns, ilist)
				j=j+1
			}
		}
	}
	
	contract Import {
		data {
			Data string
		}
		conditions {
			$list = JSONToMap($Data)
		}
		action {
			ImportList($list["pages"], "NewPage")
			ImportList($list["blocks"], "NewBlock")
			ImportList($list["menus"], "NewMenu")
			ImportList($list["parameters"], "NewParameter")
			ImportList($list["languages"], "NewLang")
			ImportList($list["contracts"], "NewContract")
			ImportList($list["tables"], "NewTable")
			ImportData($list["data"])
		}
	}', '%[1]d','ContractConditions("MainCondition")'),
	('27','contract DeactivateContract {
		data {
			Id         int
		}
		conditions {
			$cur = DBRow("contracts").Columns("id,conditions,active,wallet_id").WhereId($Id)
			if !$cur {
				error Sprintf("Contract %%d does not exist", $Id)
			}
			if Int($cur["active"]) == 0 {
				error Sprintf("The contract %%d has been already deactivated", $Id)
			}
			Eval($cur["conditions"])
			if $key_id != Int($cur["wallet_id"]) {
				error Sprintf("Wallet %%d cannot deactivate the contract", $key_id)
			}
		}
		action {
			DBUpdate("contracts", $Id, "active", 0)
			Deactivate($Id, $ecosystem_id)
		}
	}', '%[1]d','ContractConditions("MainCondition")'),
	('28','contract UpdateSysParam {
		data {
			Name  string
			Value string
			Conditions string "optional"
		}
		action {
			DBUpdateSysParam($Name, $Value, $Conditions )
		}
	}', '%[1]d','ContractConditions("MainCondition")');`
)
