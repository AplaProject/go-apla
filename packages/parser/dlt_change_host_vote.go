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

func (p *Parser) DLTChangeHostVoteInit() error {

	fields := []map[string]string{{"host": "string"}, {"addressVote": "string"}, {"public_key": "bytes"}, {"sign": "bytes"}}
	err := p.GetTxMaps(fields)
	if err != nil {
		return p.ErrInfo(err)
	}

	p.TxMaps.Bytes["public_key"] = utils.BinToHex(p.TxMaps.Bytes["public_key"])
	p.TxMap["public_key"] = utils.BinToHex(p.TxMap["public_key"])
	log.Debug("p.TxMaps.String[addressVote] %s", p.TxMaps.String["addressVote"])
	return nil
}

func (p *Parser) DLTChangeHostVoteFront() error {

	/*err := p.generalCheck()
	if err != nil {
		return p.ErrInfo(err)
	}

	verifyData := map[string]string{"host": "host", "addressVote": "sha1"}
	err = p.CheckInputData(verifyData)
	if err != nil {
		return p.ErrInfo(err)
	}*/

/*
	forSign := fmt.Sprintf("%s,%s,%s,%s,%s,%s,%s,%s", p.TxMap["type"], p.TxMap["time"], p.TxMap["walletAddress"], p.TxMap["sell_currency_id"], p.TxMap["sell_rate"], p.TxMap["amount"], p.TxMap["buy_currency_id"], p.TxMap["commission"])
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

func (p *Parser) DLTChangeHostVote() error {
	var err error

	log.Debug("p.TxMaps.String[addressVote] %s", p.TxMaps.String["addressVote"])

	if len(p.TxMaps.Bytes["public_key"]) > 0 {
		err = p.selectiveLoggingAndUpd([]string{"host", "addressVote", "public_key_0"}, []interface{}{p.TxMaps.String["host"], string(p.TxMaps.String["addressVote"]), p.TxMaps.Bytes["public_key"]}, "dlt_wallets", []string{"wallet_id"}, []string{utils.Int64ToStr(p.TxWalletID)}, true)
	} else {
		err = p.selectiveLoggingAndUpd([]string{"host", "addressVote"}, []interface{}{p.TxMaps.String["host"], p.TxMaps.String["addressVote"]}, "dlt_wallets", []string{"wallet_id"}, []string{utils.Int64ToStr(p.TxWalletID)}, true)
	}
	if err != nil {
		return p.ErrInfo(err)
	}
	return nil
}

func (p *Parser) DLTChangeHostVoteRollback() error {

	/*var err error
	if len(p.TxMaps.Bytes["public_key"]) > 0 {
		err = p.selectiveRollback([]string{"host", "addressVote", "public_key_0"}, "dlt_wallets", "", false)
	} else {
		err = p.selectiveRollback([]string{"host", "addressVote"}, "dlt_wallets", "", false)
	}
	if err != nil {
		return p.ErrInfo(err)
	}
	return nil*/
	return p.autoRollback()
}

func (p *Parser) DLTChangeHostVoteRollbackFront() error {
	return nil
}
