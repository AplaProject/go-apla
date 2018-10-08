package migration

import (
	"strings"
)

const (
	// GuestKey is the guest id
	GuestKey = `4544233900443112470`
	// GuestPublic is the public guest key
	GuestPublic = `489347a1205c818d9a02f285faaedd0122a56138e3d985f5e1b4f6a9470f90f692a00a3453771dd7feea388ceb7aefeaf183e299c70ad1aecb7f870bfada3b86`
)

// GetEcosystemScript returns script to create ecosystem
func GetEcosystemScript() string {
	scripts := []string{
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
		firstKeysDataSQL,
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
