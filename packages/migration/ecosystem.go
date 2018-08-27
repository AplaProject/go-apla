package migration

import (
	"strings"
)

// GetEcosystemScript returns script to create ecosystem
func GetEcosystemScript() string {
	scripts := []string{
		schemaEcosystem,
		blocksDataSQL,
		contractsDataSQL,
		menuDataSQL,
		pagesDataSQL,
		parametersDataSQL,
		rolesDataSQL,
		sectionsDataSQL,
		tablesDataSQL,
		applicationsDataSQL,
	}

	return strings.Join(scripts, "\r\n")
}

// GetFirstEcosystemScript returns script to update with additional data for first ecosystem
func GetFirstEcosystemScript() string {
	scripts := []string{
		firstEcosystemSchema,
		firstDelayedContractsDataSQL,
		firstEcosystemContractsSQL,
		firstEcosystemDataSQL,
		firstSystemParametersDataSQL,
		firstTablesDataSQL,
	}

	return strings.Join(scripts, "\r\n")
}

// GetCommonEcosystemScript returns script with common tables
func GetCommonEcosystemScript() string {
	scripts := []string{
		firstEcosystemCommon,
	}
	return strings.Join(scripts, "\r\n")
}

// SchemaEcosystem contains SQL queries for creating ecosystem
var schemaEcosystem = `		

		DROP TABLE IF EXISTS "%[1]d_signatures"; CREATE TABLE "%[1]d_signatures" (
			"id" bigint  NOT NULL DEFAULT '0',
			"name" character varying(100) NOT NULL DEFAULT '',
			"value" jsonb,
			"conditions" text NOT NULL DEFAULT ''
		);
		ALTER TABLE ONLY "%[1]d_signatures" ADD CONSTRAINT "%[1]d_signatures_pkey" PRIMARY KEY (name);
		
		
		DROP TABLE IF EXISTS "%[1]d_app_params";
		CREATE TABLE "%[1]d_app_params" (
		"id" bigint NOT NULL  DEFAULT '0',
		"app_id" bigint NOT NULL  DEFAULT '0',
		"name" varchar(255) UNIQUE NOT NULL DEFAULT '',
		"value" text NOT NULL DEFAULT '',
		"conditions" text  NOT NULL DEFAULT ''
		);
		ALTER TABLE ONLY "%[1]d_app_params" ADD CONSTRAINT "%[1]d_app_params_pkey" PRIMARY KEY ("id");
		CREATE INDEX "%[1]d_app_params_index_name" ON "%[1]d_app_params" (name);
		CREATE INDEX "%[1]d_app_params_index_app" ON "%[1]d_app_params" (app_id);
		
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


		DROP TABLE IF EXISTS "%[1]d_roles";
		CREATE TABLE "%[1]d_roles" (
			"id" 	bigint NOT NULL DEFAULT '0',
			"default_page"	varchar(255) NOT NULL DEFAULT '',
			"role_name"	varchar(255) NOT NULL DEFAULT '',
			"deleted"    bigint NOT NULL DEFAULT '0',
			"role_type" bigint NOT NULL DEFAULT '0',
			"creator" jsonb NOT NULL DEFAULT '{}',
			"date_created" timestamp,
			"date_deleted" timestamp,
			"company_id" bigint NOT NULL DEFAULT '0',
			"roles_access" jsonb, 
			"image_id" bigint NOT NULL DEFAULT '0'
		);
		ALTER TABLE ONLY "%[1]d_roles" ADD CONSTRAINT "%[1]d_roles_pkey" PRIMARY KEY ("id");
		CREATE INDEX "%[1]d_roles_index_deleted" ON "%[1]d_roles" (deleted);
		CREATE INDEX "%[1]d_roles_index_type" ON "%[1]d_roles" (role_type);


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

		DROP TABLE IF EXISTS "%[1]d_applications";
		CREATE TABLE "%[1]d_applications" (
			"id" bigint NOT NULL DEFAULT '0',
			"name" varchar(255) NOT NULL DEFAULT '',
			"uuid" uuid NOT NULL DEFAULT '00000000-0000-0000-0000-000000000000',
			"conditions" text NOT NULL DEFAULT '',
			"deleted" bigint NOT NULL DEFAULT '0'
		);
		ALTER TABLE ONLY "%[1]d_applications" ADD CONSTRAINT "%[1]d_application_pkey" PRIMARY KEY ("id");

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
		
		DROP TABLE IF EXISTS "%[1]d_buffer_data";
		CREATE TABLE "%[1]d_buffer_data" (
			"id" bigint NOT NULL DEFAULT '0',
			"key" varchar(255) NOT NULL DEFAULT '',
			"value" jsonb,
			"member_id" bigint NOT NULL DEFAULT '0'
		);
		ALTER TABLE ONLY "%[1]d_buffer_data" ADD CONSTRAINT "%[1]d_buffer_data_pkey" PRIMARY KEY ("id");
`
