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
('default_ecosystem_page', 'P(class, Default Ecosystem Page)', 'ContractAccess("@0UpdSysParam")'),    
('default_ecosystem_menu', 'MenuItem(main, Default Ecosystem Menu)', 'ContractAccess("@0UpdSysParam")'),
('default_ecosystem_contract', '', 'ContractAccess("@0UpdSysParam")'),
('gap_between_blocks', '2', 'ContractAccess("@0UpdSysParam")'),
('rb_blocks_1', '60', 'ContractAccess("@0UpdSysParam")'),
('rb_blocks_2', '3600', 'ContractAccess("@0UpdSysParam")'),
('new_version_url', 'upd.apla.io', 'ContractAccess("@0UpdSysParam")'),
('full_nodes', '', 'ContractAccess("@0UpdFullNodes")'),
('number_of_nodes', '101', 'ContractAccess("@0UpdSysParam")'),
('op_price', '', 'ContractAccess("@0UpdSysParam")'),
('ecosystem_price', '1000', 'ContractAccess("@0UpdSysParam")'),
('contract_price', '200', 'ContractAccess("@0UpdSysParam")'),
('column_price', '200', 'ContractAccess("@0UpdSysParam")'),
('table_price', '200', 'ContractAccess("@0UpdSysParam")'),
('menu_price', '100', 'ContractAccess("@0UpdSysParam")'),
('page_price', '100', 'ContractAccess("@0UpdSysParam")'),
('blockchain_url', '', 'ContractAccess("@0UpdSysParam")'),
('max_block_size', '67108864', 'ContractAccess("@0UpdSysParam")'),
('max_tx_size', '33554432', 'ContractAccess("@0UpdSysParam")'),
('max_tx_count', '1000', 'ContractAccess("@0UpdSysParam")'),
('max_columns', '50', 'ContractAccess("@0UpdSysParam")'),
('max_indexes', '1', 'ContractAccess("@0UpdSysParam")'),
('max_block_user_tx', '100', 'ContractAccess("@0UpdSysParam")'),
('max_fuel_tx', '1000', 'ContractAccess("@0UpdSysParam")'),
('max_fuel_block', '100000', 'ContractAccess("@0UpdSysParam")'),
('upd_full_nodes_period', '3600', 'ContractAccess("@0UpdSysParam")'),
('last_upd_full_nodes', '23672372', 'ContractAccess("@0UpdSysParam")'),
('size_price', '100', 'ContractAccess("@0UpdSysParam")'),
('commission_size', '3', 'ContractAccess("@0UpdSysParam")'),
('commission_wallet', '[["1","8275283526439353759"]]', 'ContractAccess("@0UpdSysParam")'),
('sys_currencies', '[1]', 'ContractAccess("@0UpdSysParam")'),
('fuel_rate', '[["1","1000000000000000"]]', 'ContractAccess("@0UpdSysParam")'),
('recovery_address', '[["1","8275283526439353759"]]', 'ContractAccess("@0UpdSysParam")');

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

INSERT INTO system_contracts ("id","value", "active", "conditions") VALUES 
('1','contract UpdSysParam {
    data {
    }
    conditions {
    }
    action {
    }
}', '0','ContractAccess("@0UpdSysContract")'),
('2','contract UpdSysContract {
    data {
    }
    conditions {
    }
    action {
    }
}', '0','ContractAccess("@0UpdSysContract")'),
('3','contract UpdFullNodes {
     data {
    }
    conditions {
      var prev int
      var nodekey bytes
      prev = DBInt(`upd_full_nodes`, `time`, 1)
	    if $time-prev < SysParamInt(`upd_full_nodes_period`) {
		    warning Sprintf("txTime - upd_full_nodes < UPD_FULL_NODES_PERIOD")
	    }
/*	    nodekey = bytes(DBStringExt(`dlt_wallets`, `node_public_key`, $key_id, `wallet_id`))
	    if !nodekey {
	        error `len(node_key) == 0`
	    }*/
    }
    action {
/*      var list array
        list = DBGetList("dlt_wallets", "address_vote", 0, SysParamInt(`number_of_dlt_nodes`), "sum(amount) DESC", "address_vote != ? and amount > ? GROUP BY address_vote", ``, `10000000000000000000000`)
        var i int
        var out string
        while i<Len(list) {
            var row, item map
            item = list[i]
            row = DBRowExt(`dlt_wallets`, `host, wallet_id`, item[`address_vote`], `wallet_id`)
            if i > 0 {
                out = out + `,`
            }
            out = out + Sprintf(`[%q,%q]`, row[`host`],row[`wallet_id`])
            i = i+1
        }
        UpdateSysParam(`full_nodes`, `[`+out+`]`, ``)*/
    }
}', '0','ContractAccess("@0UpdSysContract")');

CREATE TABLE "upd_contracts" (
"id" bigint NOT NULL  DEFAULT '0',
"id_contract" bigint  NOT NULL DEFAULT '0',
"value" text  NOT NULL DEFAULT '',
"votes" bigint  NOT NULL DEFAULT '0',
"rb_id" bigint NOT NULL DEFAULT '0'
);
ALTER TABLE ONLY "upd_contracts" ADD CONSTRAINT upd_contracts_pkey PRIMARY KEY (id);

CREATE TABLE "upd_system_parameters" (
"id" bigint NOT NULL DEFAULT '0',
"name" varchar(255)  NOT NULL DEFAULT '',
"value" text  NOT NULL DEFAULT '',
"votes" bigint  NOT NULL DEFAULT '0',
"rb_id" bigint NOT NULL DEFAULT '0'
);
ALTER TABLE ONLY "upd_system_parameters" ADD CONSTRAINT upd_system_parameters_pkey PRIMARY KEY (id);

CREATE TABLE "system_tables" (
"name" varchar(100)  NOT NULL DEFAULT '',
"permissions" jsonb,
"columns" jsonb,
"conditions" text  NOT NULL DEFAULT '',
"rb_id" bigint NOT NULL DEFAULT '0'
);
ALTER TABLE ONLY "system_tables" ADD CONSTRAINT system_tables_pkey PRIMARY KEY (name);

INSERT INTO system_tables ("name", "permissions","columns", "conditions") VALUES ('upd_contracts', 
        '{"insert": "ContractAccess(\"@0UpdSysContract\")", "update": "ContractAccess(\"@0UpdSysContract\")", 
          "new_column": "ContractAccess(\"@0UpdSysContract\")"}',
        '{"id_contract": "ContractAccess(\"@0UpdSysContract\")", "value": "ContractAccess(\"@0UpdSysContract\")", 
          "votes": "ContractAccess(\"@0UpdSysContract\")"}',          
        'ContractAccess(\"@0UpdSysContract\")'),
        ('upd_system_parameters', 
        '{"insert": "ContractAccess(\"@0UpdSysContract\")", "update": "ContractAccess(\"@0UpdSysContract\")", 
          "new_column": "ContractAccess(\"@0UpdSysContract\")"}',
        '{"name": "ContractAccess(\"@0UpdSysContract\")", "value": "ContractAccess(\"@0UpdSysContract\")", 
          "votes": "ContractAccess(\"@0UpdSysContract\")"}',          
        'ContractAccess(\"@0UpdSysContract\")'),
        ('system_states', 
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

