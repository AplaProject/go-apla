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
	"github.com/DayLightProject/go-daylight/packages/utils"
)

func (p *Parser) ChangeNodeKeyInit() error {

	fields := []map[string]string{{"new_node_public_key": "bytes"}, {"sign": "bytes"}}
	err := p.GetTxMaps(fields)
	if err != nil {
		return p.ErrInfo(err)
	}
	p.TxMaps.Bytes["new_node_public_key"] = utils.BinToHex(p.TxMaps.Bytes["new_node_public_key"])
	p.TxMap["new_node_public_key"] = utils.BinToHex(p.TxMap["new_node_public_key"])
	return nil
}

func (p *Parser) ChangeNodeKeyFront() error {

	/*err := p.generalCheck()
	if err != nil {
		return p.ErrInfo(err)
	}


	verifyData := map[string]string{"new_node_public_key": "public_key"}
	err = p.CheckInputData(verifyData)
	if err != nil {
		return p.ErrInfo(err)
	}

	nodePublicKey, err := p.GetPublicKeyWalletOrCitizen(p.TxMaps.Int64["wallet_id"], p.TxMaps.Int64["citizen_id"])
	if err != nil || len(nodePublicKey) == 0 {
		return p.ErrInfo("incorrect user_id")
	}
	*/
	/*forSign := fmt.Sprintf("%s,%s,%s,%s", p.TxMap["type"], p.TxMap["time"], p.TxMap["user_id"], p.TxMap["new_node_public_key"])
	CheckSignResult, err := utils.CheckSign([][]byte{nodePublicKey}, forSign, p.TxMap["sign"], true)
	if err != nil || !CheckSignResult {
		forSign := fmt.Sprintf("%s,%s,%s,%s", p.TxMap["type"], p.TxMap["time"], p.TxMap["user_id"], p.TxMap["new_node_public_key"])
		CheckSignResult, err := utils.CheckSign(p.PublicKeys, forSign, p.TxMap["sign"], false)
		if err != nil || !CheckSignResult {
			return p.ErrInfo("incorrect sign")
		}
	}

	err = p.limitRequest(p.Variables.Int64["limit_node_key"], "node_key", p.Variables.Int64["limit_node_key_period"])
	if err != nil {
		return p.ErrInfo(err)
	}
	*/
	return nil
}

func (p *Parser) ChangeNodeKey() error {

	if p.TxMaps.Int64["wallet_id"] > 0 {
		_, err := p.selectiveLoggingAndUpd([]string{"node_public_key"}, []interface{}{p.TxMaps.Bytes["new_node_public_key"]}, "dlt_wallets", []string{"wallet_id"}, []string{utils.Int64ToStr(p.TxWalletID)}, true)
		if err != nil {
			return p.ErrInfo(err)
		}
	} else {
		_, err := p.selectiveLoggingAndUpd([]string{"node_public_key"}, []interface{}{p.TxMaps.Bytes["new_node_public_key"]}, "central_banks", []string{"head_citizen_id"}, []string{utils.Int64ToStr(p.TxCitizenID)}, true)
		if err != nil {
			return p.ErrInfo(err)
		}
	}
	return nil
}

func (p *Parser) ChangeNodeKeyRollback() error {
	/*
		return p.autoRollback()

		if p.TxMaps.Int64["wallet_id"] > 0 {
			log.Debug("p.TxWalletID %d", p.TxWalletID)
			err := p.selectiveRollback([]string{"node_public_key"}, "dlt_wallets", "wallet_id="+utils.Int64ToStr(p.TxWalletID), false)
			if err != nil {
				return p.ErrInfo(err)
			}
		} else {
			err := p.selectiveRollback([]string{"node_public_key"}, "central_banks", "head_citizen_id="+utils.Int64ToStr(p.TxCitizenID), false)
			if err != nil {
				return p.ErrInfo(err)
			}
		}*/
	return p.autoRollback()
}

/*func (p *Parser) ChangeNodeKeyRollbackFront() error {
	return nil
}
*/
