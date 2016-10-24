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
	"fmt"
)

func (p *Parser) ChangeNodeKeyDLTInit() error {

	fields := []map[string]string{{"new_node_public_key": "bytes"}, {"sign": "bytes"}}
	err := p.GetTxMaps(fields)
	if err != nil {
		return p.ErrInfo(err)
	}
	p.TxMaps.Bytes["new_node_public_key"] = utils.BinToHex(p.TxMaps.Bytes["new_node_public_key"])
	p.TxMap["new_node_public_key"] = utils.BinToHex(p.TxMap["new_node_public_key"])
	return nil
}

func (p *Parser) ChangeNodeKeyDLTFront() error {

	err := p.generalCheck()
	if err != nil {
		return p.ErrInfo(err)
	}

	verifyData := map[string]string{"new_node_public_key": "public_key"}
	err = p.CheckInputData(verifyData)
	if err != nil {
		return p.ErrInfo(err)
	}

	txTime := p.TxTime
	if p.BlockData!= nil {
		txTime = p.BlockData.Time
	}
	last_forging_data_upd, err := p.Single(`SELECT last_forging_data_upd FROM dlt_wallets WHERE wallet_id = ?`, p.TxWalletID).Int64()
	if err != nil || txTime - last_forging_data_upd < 600 {
		return p.ErrInfo("txTime - last_forging_data_upd < 600 sec")
	}

	forSign := fmt.Sprintf("%s,%s,%d,%s", p.TxMap["type"], p.TxMap["time"], p.TxWalletID, p.TxMap["new_node_public_key"])
	CheckSignResult, err := utils.CheckSign(p.PublicKeys, forSign, p.TxMap["sign"], false)
	if err != nil || !CheckSignResult {
		return p.ErrInfo("incorrect sign "+forSign)
	}
	return nil
}

func (p *Parser) ChangeNodeKeyDLT() error {

	_, err := p.selectiveLoggingAndUpd([]string{"node_public_key", "last_forging_data_upd"}, []interface{}{utils.HexToBin(p.TxMaps.Bytes["new_node_public_key"]), p.BlockData.Time}, "dlt_wallets", []string{"wallet_id"}, []string{utils.Int64ToStr(p.TxWalletID)}, true)
	if err != nil {
		return p.ErrInfo(err)
	}

	myKey, err := p.Single(`SELECT id FROM my_node_keys WHERE block_id = 0 AND public_key = [hex]`, p.TxMaps.Bytes["new_node_public_key"]).Int64()
	if err != nil {
		return p.ErrInfo(err)
	}
	log.Debug("myKey %d", myKey)
	if myKey > 0 {
		_, err := p.selectiveLoggingAndUpd([]string{"block_id"}, []interface{}{p.BlockData.BlockId}, "my_node_keys", []string{"id"}, []string{utils.Int64ToStr(myKey)}, true)
		if err != nil {
			return p.ErrInfo(err)
		}
	}
	return nil
}

func (p *Parser) ChangeNodeKeyDLTRollback() error {
	return p.autoRollback()
}
