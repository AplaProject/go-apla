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

// SchemaFirstEcosystem contains SQL queries for creating first ecosystem
var firstEcosystemSchema = `
DROP TABLE IF EXISTS "1_ecosystems";
CREATE TABLE "1_ecosystems" (
		"id" bigint NOT NULL DEFAULT '0',
		"name"	varchar(255) NOT NULL DEFAULT '',
		"info" jsonb,
		"is_valued" bigint NOT NULL DEFAULT '0',
		"emission_amount" jsonb,
		"token_title" varchar(255),
		"type_emission" bigint NOT NULL DEFAULT '0',
		"type_withdraw" bigint NOT NULL DEFAULT '0'
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
		"block_id" bigint NOT NULL DEFAULT '0',
		"every_block" bigint NOT NULL DEFAULT '0',
		"counter" bigint NOT NULL DEFAULT '0',
		"limit" bigint NOT NULL DEFAULT '0',
		"deleted" bigint NOT NULL DEFAULT '0',
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
		"block_id" bigint NOT NULL,
		"consumer_node_id" bigint NOT NULL,
		"block_time" timestamp NOT NULL,
		"reason" TEXT NOT NULL DEFAULT '',
		"deleted" bigint NOT NULL DEFAULT '0'
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
var sqlFirstEcosystemCommon = `
	{{head "1_keys"}}
		t.Column("id", "bigint", {"default": "0"})
		t.Column("pub", "bytea", {"default": ""})
		t.Column("amount", "decimal(30)", {"default_raw": "'0' CHECK (amount >= 0)"})
		t.Column("maxpay", "decimal(30)", {"default_raw": "'0' CHECK (maxpay >= 0)"})
		t.Column("deposit", "decimal(30)", {"default_raw": "'0' CHECK (deposit >= 0)"})
		t.Column("multi", "bigint", {"default": "0"})
		t.Column("deleted", "bigint", {"default": "0"})
		t.Column("blocked", "bigint", {"default": "0"})
		t.Column("ecosystem", "bigint", {"default": "1"})
		t.Column("account", "char(24)", {})
		t.PrimaryKey("ecosystem", "id")
	{{footer}}

	{{head "1_menu"}}
		t.Column("id", "bigint", {"default": "0"})
		t.Column("name", "string", {"default": "", "size":255})
		t.Column("title", "string", {"default": "", "size":255})
		t.Column("value", "text", {"default": ""})
		t.Column("conditions", "text", {"default": ""})
		t.Column("permissions", "jsonb", {"null": true})
		t.Column("ecosystem", "bigint", {"default": "1"})
	{{footer "primary" "unique(ecosystem, name)" "index(ecosystem, name)"}}

	{{head "1_pages"}}
		t.Column("id", "bigint", {"default": "0"})
		t.Column("name", "string", {"default": "", "size":255})
		t.Column("value", "text", {"default": ""})
		t.Column("menu", "string", {"default": "", "size":255})
		t.Column("validate_count", "bigint", {"default": "1"})
		t.Column("conditions", "text", {"default": ""})
		t.Column("permissions", "jsonb", {"null": true})
		t.Column("app_id", "bigint", {"default": "1"})
		t.Column("validate_mode", "character(1)", {"default": "0"})
		t.Column("ecosystem", "bigint", {"default": "1"})
	{{footer "primary" "unique(ecosystem, name)" "index(ecosystem, name)"}}

	{{head "1_blocks"}}
		t.Column("id", "bigint", {"default": "0"})
		t.Column("name", "string", {"default": "", "size":255})
		t.Column("value", "text", {"default": ""})
		t.Column("conditions", "text", {"default": ""})
		t.Column("permissions", "jsonb", {"null": true})
		t.Column("app_id", "bigint", {"default": "1"})
		t.Column("ecosystem", "bigint", {"default": "1"})
	{{footer "primary" "unique(ecosystem, name)" "index(ecosystem, name)"}}

	{{head "1_languages"}}
		t.Column("id", "bigint", {"default": "0"})
		t.Column("name", "string", {"default": "", "size":100})
		t.Column("res", "text", {"default": ""})
		t.Column("conditions", "text", {"default": ""})
		t.Column("permissions", "jsonb", {"null": true})
		t.Column("ecosystem", "bigint", {"default": "1"})
	{{footer "primary" "index(ecosystem, name)"}}

	{{head "1_contracts"}}
		t.Column("id", "bigint", {"default": "0"})
		t.Column("name", "text", {"default": ""})
		t.Column("value", "text", {"default": ""})
		t.Column("wallet_id", "bigint", {"default": "0"})
		t.Column("token_id", "bigint", {"default": "1"})
		t.Column("conditions", "text", {"default": ""})
		t.Column("permissions", "jsonb", {"null": true})
		t.Column("app_id", "bigint", {"default": "1"})
		t.Column("ecosystem", "bigint", {"default": "1"})
	{{footer "primary" "unique(ecosystem, name)" "index(ecosystem)"}}

	{{head "1_tables"}}
		t.Column("id", "bigint", {"default": "0"})
		t.Column("name", "string", {"default": "", "size": 100})
		t.Column("permissions", "jsonb", {"null": true})
		t.Column("columns", "jsonb", {"null": true})
		t.Column("conditions", "text", {"default": ""})
		t.Column("app_id", "bigint", {"default": "1"})
		t.Column("ecosystem", "bigint", {"default": "1"})
	{{footer "primary" "unique(ecosystem, name)" "index(ecosystem, name)"}}

	{{head "1_parameters"}}
		t.Column("id", "bigint", {"default": "0"})
		t.Column("name", "string", {"default": "", "size": 255})
		t.Column("value", "text", {"default": ""})
		t.Column("conditions", "text", {"default": ""})
		t.Column("permissions", "jsonb", {"null": true})
		t.Column("ecosystem", "bigint", {"default": "1"})
	{{footer "primary" "unique(ecosystem, name)" "index(ecosystem, name)"}}

	{{head "1_history"}}
		t.Column("id", "bigint", {"default": "0"})
		t.Column("sender_id", "bigint", {"default": "0"})
		t.Column("recipient_id", "bigint", {"default": "0"})
		t.Column("amount", "decimal(30)", {"default": "0"})
		t.Column("comment", "text", {"default": ""})
		t.Column("block_id", "bigint", {"default": "0"})
		t.Column("txhash", "bytea", {"default": ""})
		t.Column("created_at", "bigint", {"default": "0"})
		t.Column("ecosystem", "bigint", {"default": "1"})
		t.Column("type", "bigint", {"default": "1"})
	{{footer "primary" "index(ecosystem, sender_id)"}}
	add_index("1_history", ["ecosystem", "recipient_id"], {})
	add_index("1_history", ["block_id", "txhash"], {})

	{{head "1_sections"}}
		t.Column("id", "bigint", {"default": "0"})
		t.Column("title", "string", {"default": "", "size": 255})
		t.Column("urlname", "string", {"default": "", "size": 255})
		t.Column("page", "string", {"default": "", "size": 255})
		t.Column("roles_access", "jsonb", {"null": true})
		t.Column("status", "bigint", {"default": "0"})
		t.Column("ecosystem", "bigint", {"default": "1"})
	{{footer "primary" "index(ecosystem)"}}

	{{head "1_members"}}
		t.Column("id", "bigint", {"default": "0"})
		t.Column("member_name", "string", {"default": "", "size": 255})
		t.Column("image_id", "bigint", {"default": "0"})
		t.Column("member_info", "jsonb", {"null": true})
		t.Column("ecosystem", "bigint", {"default": "1"})
		t.Column("account", "char(24)", {})
	{{footer "primary" "index(ecosystem)"}}
	add_index("1_members", ["account", "ecosystem"], {"unique": true})

	{{head "1_roles"}}
		t.Column("id", "bigint", {"default": "0"})
		t.Column("default_page", "string", {"default": "", "size": 255})
		t.Column("role_name", "string", {"default": "", "size": 255})
		t.Column("deleted", "bigint", {"default": "0"})
		t.Column("role_type", "bigint", {"default": "0"})
		t.Column("creator", "jsonb", {"default": "{}"})
		t.Column("date_created", "bigint", {"default": "0"})
		t.Column("date_deleted", "bigint", {"default": "0"})
		t.Column("company_id", "bigint", {"default": "0"})
		t.Column("roles_access", "jsonb", {"null": true})
		t.Column("image_id", "bigint", {"default": "0"})
		t.Column("ecosystem", "bigint", {"default": "1"})
	{{footer "primary" "index(ecosystem, deleted)"}}
	add_index("1_roles", ["ecosystem", "role_type"], {})

	{{head "1_roles_participants"}}
		t.Column("id", "bigint", {"default": "0"})
		t.Column("role", "jsonb", {"null": true})
		t.Column("member", "jsonb", {"null": true})
		t.Column("appointed", "jsonb", {"null": true})
		t.Column("date_created", "bigint", {"default": "0"})
		t.Column("date_deleted", "bigint", {"default": "0"})
		t.Column("deleted", "bigint", {"default": "0"})
		t.Column("ecosystem", "bigint", {"default": "1"})
	{{footer "primary" "index(ecosystem)"}}

	{{head "1_notifications"}}
		t.Column("id", "bigint", {"default": "0"})
		t.Column("recipient", "jsonb", {"null": true})
		t.Column("sender", "jsonb", {"null": true})
		t.Column("notification", "jsonb", {"null": true})
		t.Column("page_params", "jsonb", {"null": true})
		t.Column("processing_info", "jsonb", {"null": true})
		t.Column("page_name", "string", {"default": "", "size": 255})
		t.Column("date_created", "bigint", {"default": "0"})
		t.Column("date_start_processing", "bigint", {"default": "0"})
		t.Column("date_closed", "bigint", {"default": "0"})
		t.Column("closed", "bigint", {"default": "0"})
		t.Column("ecosystem", "bigint", {"default": "1"})
	{{footer "primary" "index(ecosystem)"}}

	{{head "1_applications"}}
		t.Column("id", "bigint", {"default": "0"})
		t.Column("name", "string", {"default": "", "size": 255})
		t.Column("uuid", "uuid", {"default": "00000000-0000-0000-0000-000000000000"})
		t.Column("conditions", "text", {"default": ""})
		t.Column("deleted", "bigint", {"default": "0"})
		t.Column("ecosystem", "bigint", {"default": "1"})
	{{footer "primary" "index(ecosystem)"}}

	{{head "1_binaries"}}
		t.Column("id", "bigint", {"default": "0"})
		t.Column("app_id", "bigint", {"default": "1"})
		t.Column("name", "string", {"default": "", "size": 255})
		t.Column("data", "bytea", {"default": ""})
		t.Column("hash", "string", {"default": "", "size": 64})
		t.Column("mime_type", "string", {"default": "", "size": 255})
		t.Column("ecosystem", "bigint", {"default": "1"})
		t.Column("account", "char(24)", {})
	{{footer "primary"}}
	add_index("1_binaries", ["account", "ecosystem", "app_id", "name"], {"unique": true})

	{{head "1_app_params"}}
		t.Column("id", "bigint", {"default": "0"})
		t.Column("app_id", "bigint", {"default": "0"})
		t.Column("name", "string", {"default": "", "size": 255})
		t.Column("value", "text", {"default": ""})
		t.Column("conditions", "text", {"default": ""})
		t.Column("permissions", "jsonb", {"null": true})
		t.Column("ecosystem", "bigint", {"default": "1"})
	{{footer "primary" "unique(ecosystem, app_id, name)" "index(ecosystem,app_id,name)"}}

	{{head "1_buffer_data"}}
		t.Column("id", "bigint", {"default": "0"})
		t.Column("key", "string", {"default": "", "size": 255})
		t.Column("value", "jsonb", {"null": true})
		t.Column("ecosystem", "bigint", {"default": "1"})
		t.Column("account", "char(24)", {})
	{{footer "primary" "index(ecosystem)"}}
`
