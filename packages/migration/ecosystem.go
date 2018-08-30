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
		membersDataSQL,
		sectionsDataSQL,
		tablesDataSQL,
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
`
