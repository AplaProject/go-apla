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
		
		DROP TABLE IF EXISTS "%[1]d_buffer_data";
		CREATE TABLE "%[1]d_buffer_data" (
			"id" bigint NOT NULL DEFAULT '0',
			"key" varchar(255) NOT NULL DEFAULT '',
			"value" jsonb,
			"member_id" bigint NOT NULL DEFAULT '0'
		);
		ALTER TABLE ONLY "%[1]d_buffer_data" ADD CONSTRAINT "%[1]d_buffer_data_pkey" PRIMARY KEY ("id");
`
