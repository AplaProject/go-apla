// Copyright 2016 The go-daylight Authors
// This file is part of the go-daylight library.
//
// The go-daylight library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-daylight library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-daylight library. If not, see <http://www.gnu.org/licenses/>.

package utils

import (
	"github.com/DayLightProject/go-daylight/packages/script"
	"github.com/DayLightProject/go-daylight/packages/smart"
)

func init() {
	smart.Extend(&script.ExtendData{map[string]interface{}{
		"Balance":    Balance,
		"StateParam": StateParam,
		/*		"DBInsert":   DBInsert,
		 */
	}, map[string]string{
	//		`*parser.Parser`: `parser`,
	}})
}

// Reading and compiling contracts from smart_contracts tables
func LoadContracts() (err error) {
	var contracts, states []map[string]string
	prefix := []string{`global`}
	states, err = DB.GetAll(`select id from system_states order by id`, -1)
	if err != nil {
		return err
	}
	for _, istate := range states {
		prefix = append(prefix, istate[`id`])
	}
	for _, ipref := range prefix {
		contracts, err = DB.GetAll(`select * from "`+ipref+`_smart_contracts" order by id`, -1)
		if err != nil {
			return err
		}
		for _, item := range contracts {
			if err = smart.Compile(item[`value`]); err != nil {
				return
			}
		}
	}
	return
}

func Balance(wallet_id int64) (float64, error) {
	return DB.Single("SELECT amount FROM dlt_wallets WHERE wallet_id = ?", wallet_id).Float64()
}

func StateParam(idstate int64, name string) (string, error) {
	return DB.Single(`SELECT value FROM "`+Int64ToStr(idstate)+`_state_parameters" WHERE name = ?`, name).String()
}
