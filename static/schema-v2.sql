DROP TABLE IF EXISTS "transactions_status"; CREATE TABLE "transactions_status" (
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
"name" varchar(255)  NOT NULL DEFAULT '',
"value" text NOT NULL DEFAULT '',
"conditions" text  NOT NULL DEFAULT '',
"rb_id" bigint  NOT NULL DEFAULT '0'
);
ALTER TABLE ONLY "system_parameters" ADD CONSTRAINT system_parameters_pkey PRIMARY KEY ("name");

INSERT INTO system_parameters ("name", "value", "conditions") VALUES 
('default_ecosystem_page', 'P(class, Default Ecosystem Page)', 'true'),
('default_ecosystem_menu', 'MenuItem(main, Default Ecosystem Menu)', 'true'),
('default_ecosystem_contract', '', 'true'),
('gap_between_blocks', '2', 'true'),
('rb_blocks_1', '60', 'true'),
('rb_blocks_2', '3600', 'true'),
('new_version_url', 'upd.apla.io', 'true'),
('full_nodes', '', 'true'),
('number_of_nodes', '101', 'true'),
('op_price', '', 'true'),
('ecosystem_price', '1000', 'true'),
('contract_price', '200', 'true'),
('column_price', '200', 'true'),
('table_price', '200', 'true'),
('menu_price', '100', 'true'),
('page_price', '100', 'true'),
('blockchain_url', '', 'true'),
('max_block_size', '67108864', 'true'),
('max_tx_size', '33554432', 'true'),
('max_tx_count', '1000', 'true'),
('max_columns', '50', 'true'),
('max_indexes', '5', 'true'),
('max_block_user_tx', '100', 'true'),
('max_fuel_tx', '1000', 'true'),
('max_fuel_block', '100000', 'true'),
('size_price', '100', 'true'),
('commission_size', '3', 'true'),
('commission_wallet', '', 'true'),
('fuel_rate', '[["1","1000000000000000"]]', 'true');

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
