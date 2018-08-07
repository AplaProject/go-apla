package vde

import (
	"strings"
)

// GetVDEScript returns script for VDE schema
func GetVDEScript() string {
	scripts := []string{
		schemaVDE,
		membersDataSQL,
		menuDataSQL,
		pagesDataSQL,
		parametersDataSQL,
		tablesDataSQL,
		contractsDataSQL,
		keysDataSQL,
	}

	return strings.Join(scripts, "\r\n")
}

var schemaVDE = `
	DROP TABLE IF EXISTS "%[1]d_keys"; CREATE TABLE "%[1]d_keys" (
	"id" bigint  NOT NULL DEFAULT '0',
	"pub" bytea  NOT NULL DEFAULT '',
	"multi" bigint NOT NULL DEFAULT '0',
	"deleted" bigint NOT NULL DEFAULT '0',
	"blocked" bigint NOT NULL DEFAULT '0'
	);
	ALTER TABLE ONLY "%[1]d_keys" ADD CONSTRAINT "%[1]d_keys_pkey" PRIMARY KEY (id);

	DROP TABLE IF EXISTS "%[1]d_members";
		CREATE TABLE "%[1]d_members" (
			"id" bigint NOT NULL DEFAULT '0',
			"member_name"	varchar(255) NOT NULL DEFAULT '',
			"image_id"	bigint,
			"member_info" jsonb
		);
		ALTER TABLE ONLY "%[1]d_members" ADD CONSTRAINT "%[1]d_members_pkey" PRIMARY KEY ("id");

		DROP TABLE IF EXISTS "%[1]d_languages"; CREATE TABLE "%[1]d_languages" (
		"id" bigint  NOT NULL DEFAULT '0',
		"name" character varying(100) NOT NULL DEFAULT '',
		"res" text NOT NULL DEFAULT ''
	  );
	  ALTER TABLE ONLY "%[1]d_languages" ADD CONSTRAINT "%[1]d_languages_pkey" PRIMARY KEY (id);
	  CREATE INDEX "%[1]d_languages_index_name" ON "%[1]d_languages" (name);
	  
	  DROP TABLE IF EXISTS "%[1]d_menu"; CREATE TABLE "%[1]d_menu" (
		  "id" bigint  NOT NULL DEFAULT '0',
		  "name" character varying(255) UNIQUE NOT NULL DEFAULT '',
		  "title" character varying(255) NOT NULL DEFAULT '',
		  "value" text NOT NULL DEFAULT '',
		  "conditions" text NOT NULL DEFAULT ''
	  );
	  ALTER TABLE ONLY "%[1]d_menu" ADD CONSTRAINT "%[1]d_menu_pkey" PRIMARY KEY (id);
	  CREATE INDEX "%[1]d_menu_index_name" ON "%[1]d_menu" (name);

	  DROP TABLE IF EXISTS "%[1]d_pages"; CREATE TABLE "%[1]d_pages" (
		  "id" bigint  NOT NULL DEFAULT '0',
		  "name" character varying(255) UNIQUE NOT NULL DEFAULT '',
		  "value" text NOT NULL DEFAULT '',
		  "menu" character varying(255) NOT NULL DEFAULT '',
		  "conditions" text NOT NULL DEFAULT '',
		  "validate_count" bigint NOT NULL DEFAULT '1',
		  "app_id" bigint NOT NULL DEFAULT '0',
		  "validate_mode" character(1) NOT NULL DEFAULT '0'
	  );
	  ALTER TABLE ONLY "%[1]d_pages" ADD CONSTRAINT "%[1]d_pages_pkey" PRIMARY KEY (id);
	  CREATE INDEX "%[1]d_pages_index_name" ON "%[1]d_pages" (name);

	  DROP TABLE IF EXISTS "%[1]d_blocks"; CREATE TABLE "%[1]d_blocks" (
		  "id" bigint  NOT NULL DEFAULT '0',
		  "name" character varying(255) UNIQUE NOT NULL DEFAULT '',
		  "value" text NOT NULL DEFAULT '',
		  "conditions" text NOT NULL DEFAULT ''
	  );
	  ALTER TABLE ONLY "%[1]d_blocks" ADD CONSTRAINT "%[1]d_blocks_pkey" PRIMARY KEY (id);
	  CREATE INDEX "%[1]d_blocks_index_name" ON "%[1]d_blocks" (name);
	  
	  DROP TABLE IF EXISTS "%[1]d_signatures"; CREATE TABLE "%[1]d_signatures" (
		  "id" bigint  NOT NULL DEFAULT '0',
		  "name" character varying(100) NOT NULL DEFAULT '',
		  "value" jsonb,
		  "conditions" text NOT NULL DEFAULT ''
	  );
	  ALTER TABLE ONLY "%[1]d_signatures" ADD CONSTRAINT "%[1]d_signatures_pkey" PRIMARY KEY (name);
	  
	  CREATE TABLE "%[1]d_contracts" (
	  "id" bigint NOT NULL  DEFAULT '0',
	  "name" text NOT NULL DEFAULT '',
	  "value" text  NOT NULL DEFAULT '',
	  "conditions" text  NOT NULL DEFAULT '',
	  "app_id" bigint NOT NULL DEFAULT '1'
	  );
	  ALTER TABLE ONLY "%[1]d_contracts" ADD CONSTRAINT "%[1]d_contracts_pkey" PRIMARY KEY (id);
	  
	  DROP TABLE IF EXISTS "%[1]d_parameters";
	  CREATE TABLE "%[1]d_parameters" (
	  "id" bigint NOT NULL  DEFAULT '0',
	  "name" varchar(255) UNIQUE NOT NULL DEFAULT '',
	  "value" text NOT NULL DEFAULT '',
	  "conditions" text  NOT NULL DEFAULT ''
	  );
	  ALTER TABLE ONLY "%[1]d_parameters" ADD CONSTRAINT "%[1]d_parameters_pkey" PRIMARY KEY ("id");
	  CREATE INDEX "%[1]d_parameters_index_name" ON "%[1]d_parameters" (name);
	  
	  DROP TABLE IF EXISTS "%[1]d_cron";
	  CREATE TABLE "%[1]d_cron" (
		  "id"        bigint NOT NULL DEFAULT '0',
		  "owner"	  bigint NOT NULL DEFAULT '0',
		  "cron"      varchar(255) NOT NULL DEFAULT '',
		  "contract"  varchar(255) NOT NULL DEFAULT '',
		  "counter"   bigint NOT NULL DEFAULT '0',
		  "till"      timestamp NOT NULL DEFAULT timestamp '1970-01-01 00:00:00',
		  "conditions" text  NOT NULL DEFAULT ''
	  );
	  ALTER TABLE ONLY "%[1]d_cron" ADD CONSTRAINT "%[1]d_cron_pkey" PRIMARY KEY ("id");

		DROP TABLE IF EXISTS "%[1]d_binaries";
		CREATE TABLE "%[1]d_binaries" (
			"id" bigint NOT NULL DEFAULT '0',
			"app_id" bigint NOT NULL DEFAULT '1',
			"member_id" bigint NOT NULL DEFAULT '0',
			"name" varchar(255) NOT NULL DEFAULT '',
			"data" bytea NOT NULL DEFAULT '',
			"hash" varchar(32) NOT NULL DEFAULT '',
			"mime_type" varchar(255) NOT NULL DEFAULT ''
		);
		ALTER TABLE ONLY "%[1]d_binaries" ADD CONSTRAINT "%[1]d_binaries_pkey" PRIMARY KEY (id);
		CREATE UNIQUE INDEX "%[1]d_binaries_index_app_id_member_id_name" ON "%[1]d_binaries" (app_id, member_id, name);

	  CREATE TABLE "%[1]d_tables" (
	  "id" bigint NOT NULL  DEFAULT '0',
	  "name" varchar(100) UNIQUE NOT NULL DEFAULT '',
	  "permissions" jsonb,
	  "columns" jsonb,
	  "conditions" text  NOT NULL DEFAULT '',
	  "app_id" bigint NOT NULL DEFAULT '1'
	  );
	  ALTER TABLE ONLY "%[1]d_tables" ADD CONSTRAINT "%[1]d_tables_pkey" PRIMARY KEY ("id");
	  CREATE INDEX "%[1]d_tables_index_name" ON "%[1]d_tables" (name); 

	  DROP TABLE IF EXISTS "%[1]d_notifications";
		CREATE TABLE "%[1]d_notifications" (
			"id"    bigint NOT NULL DEFAULT '0',
			"recipient" jsonb,
			"sender" jsonb,
			"notification" jsonb,
			"page_params"	jsonb,
			"processing_info" jsonb,
			"page_name"	varchar(255) NOT NULL DEFAULT '',
			"date_created"	timestamp,
			"date_start_processing" timestamp,
			"date_closed" timestamp,
			"closed" bigint NOT NULL DEFAULT '0'
		);
		ALTER TABLE ONLY "%[1]d_notifications" ADD CONSTRAINT "%[1]d_notifications_pkey" PRIMARY KEY ("id");

		DROP TABLE IF EXISTS "%[1]d_roles_participants";
		CREATE TABLE "%[1]d_roles_participants" (
			"id" bigint NOT NULL DEFAULT '0',
			"role" jsonb,
			"member" jsonb,
			"appointed" jsonb,
			"date_created" timestamp,
			"date_deleted" timestamp,
			"deleted" bigint NOT NULL DEFAULT '0'
		);
		ALTER TABLE ONLY "%[1]d_roles_participants" ADD CONSTRAINT "%[1]d_roles_participants_pkey" PRIMARY KEY ("id");

	`
