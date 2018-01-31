package migration

var (
	migrationInitial = `
		DROP SEQUENCE IF EXISTS migration_history_id_seq CASCADE;
		CREATE SEQUENCE migration_history_id_seq START WITH 1;
		DROP TABLE IF EXISTS "migration_history";
		CREATE TABLE "migration_history" (
			"id" int NOT NULL default nextval('migration_history_id_seq'),
			"version" varchar(255) NOT NULL,
			"date_applied" int NOT NULL
		);
		ALTER SEQUENCE migration_history_id_seq owned by migration_history.id;
		ALTER TABLE ONLY "migration_history" ADD CONSTRAINT migration_history_pkey PRIMARY KEY (id);`

	migrationInitialSchema = `DROP TABLE IF EXISTS "transactions_status"; CREATE TABLE "transactions_status" (
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
		"rollbacks_hash" bytea NOT NULL DEFAULT '',
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
		
		DROP TABLE IF EXISTS "queue_tx"; CREATE TABLE "queue_tx" (
		"hash" bytea  NOT NULL DEFAULT '',
		"data" bytea NOT NULL DEFAULT '',
		"from_gate" int NOT NULL DEFAULT '0'
		);
		ALTER TABLE ONLY "queue_tx" ADD CONSTRAINT queue_tx_pkey PRIMARY KEY (hash);
		
		DROP TABLE IF EXISTS "system_states"; CREATE TABLE "system_states" (
		"id" bigint NOT NULL DEFAULT '0'
		);
		ALTER TABLE ONLY "system_states" ADD CONSTRAINT system_states_pkey PRIMARY KEY (id);
		
		DROP TABLE IF EXISTS "system_parameters";
		CREATE TABLE "system_parameters" (
		"id" bigint NOT NULL DEFAULT '0',
		"name" varchar(255)  NOT NULL DEFAULT '',
		"value" text NOT NULL DEFAULT '',
		"conditions" text  NOT NULL DEFAULT ''
		);
		ALTER TABLE ONLY "system_parameters" ADD CONSTRAINT system_parameters_pkey PRIMARY KEY (id);
		CREATE INDEX "system_parameters_index_name" ON "system_parameters" (name);
		
		INSERT INTO system_parameters ("id","name", "value", "conditions") VALUES 
		('1','default_ecosystem_page', 'P(class, Default Ecosystem Page)', 'true'),
		('2','default_ecosystem_menu', 'MenuItem(main, Default Ecosystem Menu)', 'true'),
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
		('61','extend_cost_json_to_map', '50', 'true'),
		('62','max_block_generation_time', '2000', 'true');
		
		CREATE TABLE "system_contracts" (
		"id" bigint NOT NULL  DEFAULT '0',
		"value" text  NOT NULL DEFAULT '',
		"wallet_id" bigint NOT NULL DEFAULT '0',
		"token_id" bigint NOT NULL DEFAULT '0',
		"active" character(1) NOT NULL DEFAULT '0',
		"conditions" text  NOT NULL DEFAULT ''
		);
		ALTER TABLE ONLY "system_contracts" ADD CONSTRAINT system_contracts_pkey PRIMARY KEY (id);
		
		
		CREATE TABLE "system_tables" (
		"name" varchar(100)  NOT NULL DEFAULT '',
		"permissions" jsonb,
		"columns" jsonb,
		"conditions" text  NOT NULL DEFAULT ''
		);
		ALTER TABLE ONLY "system_tables" ADD CONSTRAINT system_tables_pkey PRIMARY KEY (name);
		
		INSERT INTO system_tables ("name", "permissions","columns", "conditions") VALUES  ('system_states',
				'{"insert": "false", "update": "ContractAccess(\"@1EditParameter\")",
				  "new_column": "false"}','{}', 'ContractAccess(\"@0UpdSysContract\")');
		
		
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
		"attempt" smallint NOT NULL DEFAULT '0',
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
		"table_id" varchar(255) NOT NULL DEFAULT '',
		"data" TEXT NOT NULL DEFAULT ''
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
		"block_id" int NOT NULL DEFAULT '0'
		);
		ALTER SEQUENCE my_node_keys_id_seq owned by my_node_keys.id;
		ALTER TABLE ONLY "my_node_keys" ADD CONSTRAINT my_node_keys_pkey PRIMARY KEY (id);
		
		DROP TABLE IF EXISTS "stop_daemons"; CREATE TABLE "stop_daemons" (
		"stop_time" int NOT NULL DEFAULT '0'
		);
		`
)
