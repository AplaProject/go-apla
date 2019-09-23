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

import (
	"strings"

	"github.com/gobuffalo/fizz"
	"github.com/gobuffalo/fizz/translators"
)

var _ fizz.Translator = (*translators.Postgres)(nil)
var pgt = translators.NewPostgres()

func sqlConvert(in []string) (ret string, err error) {
	var item string
	for _, sql := range in {
		item, err = fizz.AString(sql, pgt)
		if err != nil {
			return
		}
		ret += item + "\r\n"
	}
	return
}

// GetEcosystemScript returns script to create ecosystem
func GetEcosystemScript() string {
	scripts := []string{
		contractsDataSQL,
		menuDataSQL,
		pagesDataSQL,
		parametersDataSQL,
		membersDataSQL,
		sectionsDataSQL,
		keysDataSQL,
	}

	return strings.Join(scripts, "\r\n")
}

// GetFirstEcosystemScript returns script to update with additional data for first ecosystem
func GetFirstEcosystemScript() string {
	scripts := []string{
		firstEcosystemSchema,
		firstDelayedContractsDataSQL,
		firstEcosystemContractsSQL,
		firstEcosystemPagesDataSQL,
		firstEcosystemBlocksDataSQL,
		firstEcosystemDataSQL,
		firstSystemParametersDataSQL,
		firstTablesDataSQL,
	}

	return strings.Join(scripts, "\r\n")
}

// GetFirstTableScript returns script to update _tables for first ecosystem
func GetFirstTableScript() string {
	scripts := []string{
		tablesDataSQL,
	}
	return strings.Join(scripts, "\r\n")
}

// GetCommonEcosystemScript returns script with common tables
func GetCommonEcosystemScript() (string, error) {
	sql, err := sqlConvert([]string{
		sqlFirstEcosystemCommon,
	})
	if err != nil {
		return ``, err
	}
	scripts := []string{
		sql,
		firstEcosystemCommon,
		timeZonesSQL,
	}
	return strings.Join(scripts, "\r\n"), nil
}
