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
 
	`
