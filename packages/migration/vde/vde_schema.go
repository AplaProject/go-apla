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
	
	DROP TABLE IF EXISTS "%[1]d_members";
		CREATE TABLE "%[1]d_members" (
			"id" bigint NOT NULL DEFAULT '0',
			"member_name"	varchar(255) NOT NULL DEFAULT '',
			"image_id"	bigint,
			"member_info" jsonb
		);
		ALTER TABLE ONLY "%[1]d_members" ADD CONSTRAINT "%[1]d_members_pkey" PRIMARY KEY ("id");

	  DROP TABLE IF EXISTS "%[1]d_signatures"; CREATE TABLE "%[1]d_signatures" (
		  "id" bigint  NOT NULL DEFAULT '0',
		  "name" character varying(100) NOT NULL DEFAULT '',
		  "value" jsonb,
		  "conditions" text NOT NULL DEFAULT ''
	  );
	  ALTER TABLE ONLY "%[1]d_signatures" ADD CONSTRAINT "%[1]d_signatures_pkey" PRIMARY KEY (name);
	  
	    
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
