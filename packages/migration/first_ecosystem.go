package migration

// SchemaFirstEcosystem contains SQL queries for creating first ecosystem
var firstEcosystemSchema = `
DROP TABLE IF EXISTS "1_ecosystems";
CREATE TABLE "1_ecosystems" (
		"id" bigint NOT NULL DEFAULT '0',
		"name"	varchar(255) NOT NULL DEFAULT '',
		"is_valued" bigint NOT NULL DEFAULT '0'
);
ALTER TABLE ONLY "1_ecosystems" ADD CONSTRAINT "1_ecosystems_pkey" PRIMARY KEY ("id");


DROP TABLE IF EXISTS "1_system_parameters";
	CREATE TABLE "1_system_parameters" (
	"id" bigint NOT NULL DEFAULT '0',
	"name" varchar(255)  NOT NULL DEFAULT '',
	"value" text NOT NULL DEFAULT '',
	"conditions" text  NOT NULL DEFAULT ''
	);
	ALTER TABLE ONLY "1_system_parameters" ADD CONSTRAINT "1_system_parameters_pkey" PRIMARY KEY (id);
	CREATE INDEX "1_system_parameters_index_name" ON "1_system_parameters" (name);
	
	
	DROP TABLE IF EXISTS "1_delayed_contracts";
	CREATE TABLE "1_delayed_contracts" (
		"id" int NOT NULL default 0,
		"contract" varchar(255) NOT NULL DEFAULT '',
		"key_id" bigint NOT NULL DEFAULT '0',
		"block_id" int NOT NULL DEFAULT '0',
		"every_block" int NOT NULL DEFAULT '0',
		"counter" int NOT NULL DEFAULT '0',
		"limit" int NOT NULL DEFAULT '0',
		"deleted" boolean NOT NULL DEFAULT 'false',
		"conditions" text NOT NULL DEFAULT ''
	);
	ALTER TABLE ONLY "1_delayed_contracts" ADD CONSTRAINT "1_delayed_contracts_pkey" PRIMARY KEY ("id");
	CREATE INDEX "1_delayed_contracts_index_block_id" ON "1_delayed_contracts" ("block_id");

	DROP TABLE IF EXISTS "1_metrics";
	CREATE TABLE "1_metrics" (
		"id" int NOT NULL default 0,
		"time" bigint NOT NULL DEFAULT '0',
		"metric" varchar(255) NOT NULL,
		"key" varchar(255) NOT NULL,
		"value" bigint NOT NULL
	);
	ALTER TABLE ONLY "1_metrics" ADD CONSTRAINT "1_metrics_pkey" PRIMARY KEY (id);
	CREATE INDEX "1_metrics_unique_index" ON "1_metrics" (metric, time, "key");

	DROP TABLE IF EXISTS "1_bad_blocks"; CREATE TABLE "1_bad_blocks" (
		"id" bigint NOT NULL DEFAULT '0',
		"producer_node_id" bigint NOT NULL,
		"block_id" int NOT NULL,
		"consumer_node_id" bigint NOT NULL,
		"block_time" timestamp NOT NULL,
		"reason" TEXT NOT NULL DEFAULT '',
		"deleted" boolean NOT NULL DEFAULT 'false'
	);
	ALTER TABLE ONLY "1_bad_blocks" ADD CONSTRAINT "1_bad_blocks_pkey" PRIMARY KEY ("id");

	DROP TABLE IF EXISTS "1_node_ban_logs"; CREATE TABLE "1_node_ban_logs" (
		"id" bigint NOT NULL DEFAULT '0',
		"node_id" bigint NOT NULL,
		"banned_at" timestamp NOT NULL,
		"ban_time" bigint NOT NULL,
		"reason" TEXT NOT NULL DEFAULT ''
	);
	ALTER TABLE ONLY "1_node_ban_logs" ADD CONSTRAINT "1_node_ban_logs_pkey" PRIMARY KEY ("id");
`
var firstEcosystemCommon = `DROP TABLE IF EXISTS "1_keys"; CREATE TABLE "1_keys" (
	"id" bigint  NOT NULL DEFAULT '0',
	"pub" bytea  NOT NULL DEFAULT '',
	"amount" decimal(30) NOT NULL DEFAULT '0' CHECK (amount >= 0),
	"maxpay" decimal(30) NOT NULL DEFAULT '0' CHECK (maxpay >= 0),
	"multi" bigint NOT NULL DEFAULT '0',
	"deleted" bigint NOT NULL DEFAULT '0',
	"blocked" bigint NOT NULL DEFAULT '0',
	"ecosystem" bigint NOT NULL DEFAULT '1'
	);
	ALTER TABLE ONLY "1_keys" ADD CONSTRAINT "1_keys_pkey" PRIMARY KEY (ecosystem,id);

	DROP TABLE IF EXISTS "1_menu";
	CREATE TABLE "1_menu" (
		"id" bigint  NOT NULL DEFAULT '0',
		"name" character varying(255) NOT NULL DEFAULT '',
		"title" character varying(255) NOT NULL DEFAULT '',
		"value" text NOT NULL DEFAULT '',
		"conditions" text NOT NULL DEFAULT '',
		"ecosystem" bigint NOT NULL DEFAULT '1',
		UNIQUE (ecosystem, name)
	);
	ALTER TABLE ONLY "1_menu" ADD CONSTRAINT "1_menu_pkey" PRIMARY KEY (id);
	CREATE INDEX "1_menu_index_name" ON "1_menu" (ecosystem,name);

	DROP TABLE IF EXISTS "1_pages"; 
	CREATE TABLE "1_pages" (
		"id" bigint  NOT NULL DEFAULT '0',
		"name" character varying(255) NOT NULL DEFAULT '',
		"value" text NOT NULL DEFAULT '',
		"menu" character varying(255) NOT NULL DEFAULT '',
		"validate_count" bigint NOT NULL DEFAULT '1',
		"conditions" text NOT NULL DEFAULT '',
		"app_id" bigint NOT NULL DEFAULT '1',
		"validate_mode" character(1) NOT NULL DEFAULT '0',
		"ecosystem" bigint NOT NULL DEFAULT '1',
		UNIQUE (ecosystem, name)
	);
	ALTER TABLE ONLY "1_pages" ADD CONSTRAINT "1_pages_pkey" PRIMARY KEY (id);
	CREATE INDEX "1_pages_index_name" ON "1_pages" (ecosystem,name);

		
	DROP TABLE IF EXISTS "1_blocks"; CREATE TABLE "1_blocks" (
		"id" bigint  NOT NULL DEFAULT '0',
		"name" character varying(255) NOT NULL DEFAULT '',
		"value" text NOT NULL DEFAULT '',
		"conditions" text NOT NULL DEFAULT '',
		"app_id" bigint NOT NULL DEFAULT '1',
		"ecosystem" bigint NOT NULL DEFAULT '1',
		UNIQUE (ecosystem, name)
	);
	ALTER TABLE ONLY "1_blocks" ADD CONSTRAINT "1_blocks_pkey" PRIMARY KEY (id);
	CREATE INDEX "1_blocks_index_name" ON "1_blocks" (ecosystem,name);

	DROP TABLE IF EXISTS "1_languages"; CREATE TABLE "1_languages" (
		"id" bigint  NOT NULL DEFAULT '0',
		"name" character varying(100) NOT NULL DEFAULT '',
		"res" text NOT NULL DEFAULT '',
		"conditions" text NOT NULL DEFAULT '',
		"app_id" bigint NOT NULL DEFAULT '1',
		"ecosystem" bigint NOT NULL DEFAULT '1'
	  );
	  ALTER TABLE ONLY "1_languages" ADD CONSTRAINT "1_languages_pkey" PRIMARY KEY (id);
	  CREATE INDEX "1_languages_index_name" ON "1_languages" (ecosystem, name);

	  CREATE TABLE "1_contracts" (
		"id" bigint NOT NULL  DEFAULT '0',
		"name" text NOT NULL DEFAULT '',
		"value" text  NOT NULL DEFAULT '',
		"wallet_id" bigint NOT NULL DEFAULT '0',
		"token_id" bigint NOT NULL DEFAULT '1',
		"active" character(1) NOT NULL DEFAULT '0',
		"conditions" text  NOT NULL DEFAULT '',
		"app_id" bigint NOT NULL DEFAULT '1',
		"ecosystem" bigint NOT NULL DEFAULT '1',
		UNIQUE(ecosystem,name)
		);
		ALTER TABLE ONLY "1_contracts" ADD CONSTRAINT "1_contracts_pkey" PRIMARY KEY (id);
		CREATE INDEX "1_contracts_index_ecosystem" ON "1_contracts" (ecosystem);

	DROP TABLE IF EXISTS "1_tables";
	CREATE TABLE "1_tables" (
	"id" bigint NOT NULL  DEFAULT '0',
	"name" varchar(100) NOT NULL DEFAULT '',
	"permissions" jsonb,
	"columns" jsonb,
	"conditions" text  NOT NULL DEFAULT '',
	"app_id" bigint NOT NULL DEFAULT '1',
	"ecosystem" bigint NOT NULL DEFAULT '1',
	UNIQUE(ecosystem,name)
    );
	ALTER TABLE ONLY "1_tables" ADD CONSTRAINT "1_tables_pkey" PRIMARY KEY ("id");
	CREATE INDEX "1_tables_index_name" ON "1_tables" (ecosystem, name);

`
