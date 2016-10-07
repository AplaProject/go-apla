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

package parser

import (
	"fmt"

	"github.com/DayLightProject/go-daylight/packages/consts"
)

func (p *Parser) NewCitizenInit() error {
	fmt.Println(`NEW Citizen`, p.TxHash)
	/*
		fields := []map[string]string{{"public_key": "bytes"}, {"state_id": "int64"}}
		err := p.GetTxMaps(fields)
		if err != nil {
			return p.ErrInfo(err)
		}
		p.TxMap["public_key_hex"] = utils.BinToHex(p.TxMap["public_key"])
		p.TxMaps.Bytes["public_key_hex"] = utils.BinToHex(p.TxMaps.Bytes["public_key"])
	*/
	// p.TxPtr.(*consts.NewCitizen)
	//p.TxVars[`state_code`] = data.StateId
	//		fmt.Println(data)
	return nil
}

func (p *Parser) NewCitizenFront() error {
	data := p.TxPtr.(*consts.NewCitizen)

	if err := p.generalCheckStruct(fmt.Sprintf(`,%d`, data.CitizenId)); err != nil {
		return err
	}
	/*	err := p.generalCheck()
		if err != nil {
			return p.ErrInfo(err)
		}

		// To not record too small or too big key
		if !utils.CheckInputData(p.TxMap["public_key_hex"], "public_key") {
			return utils.ErrInfo(fmt.Errorf("incorrect public_key %s", p.TxMap["public_key_hex"]))
		}
	*/
	// We get a set of custom fields that need to be in the tx
	/*	statePrefix, err := p.GetStatePrefix(p.TxMaps.Int64["state_id"])
		additionalFields, err := p.Single(`SELECT value FROM ` + statePrefix + `_state_parameters where parameter='citizen_fields'`).Bytes()

		if err != nil {
			return p.ErrInfo(err)
		}

		additionalFieldsMap := []map[string]string{}
		err = json.Unmarshal(additionalFields, &additionalFieldsMap)
		if err != nil {
			return p.ErrInfo(err)
		}

		verifyData := make(map[string]string)
		for _, date := range additionalFieldsMap {
			verifyData[date["name"]] = date["txType"]
		}
		err = p.CheckInputData(verifyData)
		if err != nil {
			return p.ErrInfo(err)
		}
	*/
	// Citizens can only add a citizen of the same country

	// One who adds a citizen must be a valid representative body appointed in ea_state_parameters

	// must be supplemented
	/*	forSign := fmt.Sprintf("%s,%s,%d", p.TxMap["type"], p.TxMap["time"], p.TxCitizenID)
		CheckSignResult, err := utils.CheckSign(p.PublicKeys, forSign, p.TxMap["sign"], false)
		if err != nil {
			return p.ErrInfo(err)
		}
		if !CheckSignResult {
			return p.ErrInfo("incorrect sign")
		}
	*/
	return nil
}

func (p *Parser) NewCitizen() error {

	data := p.TxPtr.(*consts.NewCitizen)

	_, err := p.selectiveLoggingAndUpd([]string{"public_key_0", "block_id"}, []interface{}{data.PublicKey, p.BlockData.BlockId}, "dlt_wallets", nil, nil, true)
	if err != nil {
		return p.ErrInfo(err)
	}

	/*citizenId, err := p.ExecSqlGetLastInsertId(`INSERT INTO `+p.States[data.StateId]+`_citizens ( public_key_0, block_id ) VALUES ( [hex], ? )`,
		p.States[data.StateId]+`_citizens`, hex.EncodeToString(data.PublicKey), p.BlockData.BlockId)
	if err != nil {
		return err
	} else {
		if req, err := p.OneRow(`select * from `+p.States[data.StateId]+`_citizens_requests_private where approved=1 AND public=[hex]`,
			hex.EncodeToString(data.PublicKey)).String(); err == nil {

			if err = p.ExecSql(`update `+p.States[data.StateId]+`_citizens_requests_private set approved=? where id=?`,
				citizenId, req[`id`]); err != nil {
				return err
			}
			if len(req[`binary`]) > 0 {
				if err = p.ExecSql(`insert into `+p.States[data.StateId]+`_citizens_private (citizen_id, fields, binary)
				    values(?, ?, [hex])`, citizenId, req[`fields`], hex.EncodeToString([]byte(req[`binary`]))); err != nil {
					return err
				}
			} else if err = p.ExecSql(`insert into `+p.States[data.StateId]+`_citizens_private (citizen_id, fields)
				    values(?, ?, [hex])`, citizenId, req[`fields`]); err != nil {
				return err
			}
		} else {
			return err
		}
	}*/
	return nil
}

func (p *Parser) NewCitizenRollback() error {
	return p.autoRollback()
}

/*func (p *Parser) NewCitizenRollbackFront() error {

	return nil

}*/
