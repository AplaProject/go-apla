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
