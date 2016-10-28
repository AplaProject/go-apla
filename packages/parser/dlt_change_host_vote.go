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
	"github.com/EGaaS/go-mvp/packages/utils"
	"encoding/hex"
	"github.com/EGaaS/go-mvp/packages/lib"
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

	err := p.generalCheck()
	if err != nil {
		return p.ErrInfo(err)
	}

	verifyData := map[string]string{"host": "host", "addressVote": "walletAddress", "public_key": "public_key"}
	err = p.CheckInputData(verifyData)
	if err != nil {
		return p.ErrInfo(err)
	}

	// public key need only when we don't have public_key in the dlt_wallets table
	public_key_0, err := p.Single(`SELECT public_key_0 FROM dlt_wallets WHERE wallet_id = ?`, p.TxWalletID).String()
	if err != nil {
		return p.ErrInfo(err)
	}
	if len(public_key_0) == 0 {
		bkey, err := hex.DecodeString(string(p.TxMaps.Bytes["public_key"]))
		if err != nil {
			return p.ErrInfo(err)
		}
		if lib.KeyToAddress(bkey) != lib.AddressToString(uint64(p.TxWalletID)) {
			return p.ErrInfo("incorrect public_key")
		}
	}

	txTime := p.TxTime
	if p.BlockData!= nil {
		txTime = p.BlockData.Time
	}
	last_forging_data_upd, err := p.Single(`SELECT last_forging_data_upd FROM dlt_wallets WHERE wallet_id = ?`, p.TxWalletID).Int64()
	if err != nil || txTime - last_forging_data_upd < 600 {
		return p.ErrInfo("txTime - last_forging_data_upd < 600 sec")
	}

	forSign := fmt.Sprintf("%s,%s,%d,%s,%s", p.TxMap["type"], p.TxMap["time"], p.TxWalletID, p.TxMap["host"], p.TxMap["addressVote"])
	CheckSignResult, err := utils.CheckSign(p.PublicKeys, forSign, p.TxMap["sign"], false)
	if err != nil {
		return p.ErrInfo(err)
	}
	if !CheckSignResult {
		return p.ErrInfo("incorrect sign")
	}

	return nil
}

func (p *Parser) DLTChangeHostVote() error {
	var err error

	log.Debug("p.TxMaps.String[addressVote] %s", p.TxMaps.String["addressVote"])

	pkey, err := p.Single(`SELECT public_key_0 FROM dlt_wallets WHERE public_key_0 = [hex]`, p.TxMaps.Bytes["public_key"]).String()
	if err != nil {
		return p.ErrInfo(err)
	}

	if len(p.TxMaps.Bytes["public_key"]) > 0 && len(pkey) == 0 {
		_, err = p.selectiveLoggingAndUpd([]string{"host", "address_vote", "public_key_0", "last_forging_data_upd"}, []interface{}{p.TxMaps.String["host"], string(p.TxMaps.String["addressVote"]), utils.HexToBin(p.TxMaps.Bytes["public_key"]), p.BlockData.Time}, "dlt_wallets", []string{"wallet_id"}, []string{utils.Int64ToStr(p.TxWalletID)}, true)
	} else {
		_, err = p.selectiveLoggingAndUpd([]string{"host", "address_vote", "last_forging_data_upd"}, []interface{}{p.TxMaps.String["host"], p.TxMaps.String["addressVote"], p.BlockData.Time}, "dlt_wallets", []string{"wallet_id"}, []string{utils.Int64ToStr(p.TxWalletID)}, true)
	}
	if err != nil {
		return p.ErrInfo(err)
	}
	return nil
}

func (p *Parser) DLTChangeHostVoteRollback() error {
	return p.autoRollback()
}