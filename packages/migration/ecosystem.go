package migration

import (
	"strings"

	"github.com/GenesisKernel/go-genesis/packages/types"
)

type Row struct {
	Registry   *types.Registry
	PrimaryKey string
	Data       interface{}
}

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

func GetSystemContractsScript() string {
	return systemContractsDataSQL
}

func GetSystemParametersScript() string {
	return firstSystemParametersDataSQL
}

func GetNewFirstEcosystemData() []Row {
	return firstEcosystemData
}
