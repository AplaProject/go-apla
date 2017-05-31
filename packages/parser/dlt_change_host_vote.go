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
	"encoding/hex"
	"fmt"
	"github.com/EGaaS/go-egaas-mvp/packages/lib"
	"github.com/EGaaS/go-egaas-mvp/packages/utils"
)

// DLTChangeHostVoteInit initializes DLTChangeHostVote transaction
func (p *Parser) DLTChangeHostVoteInit() error {

	fields := []map[string]string{{"host": "string"}, {"addressVote": "string"}, {"fuelRate": "int64"}, {"public_key": "bytes"}, {"sign": "bytes"}}
	err := p.GetTxMaps(fields)
	if err != nil {
		return p.ErrInfo(err)
	}

	p.TxMaps.Bytes["public_key"] = utils.BinToHex(p.TxMaps.Bytes["public_key"])
	p.TxMap["public_key"] = utils.BinToHex(p.TxMap["public_key"])
	log.Debug("p.TxMaps.String[addressVote] %s", p.TxMaps.String["addressVote"])
	return nil
}

// DLTChangeHostVoteFront checks conditions of DLTChangeHostVote transaction
func (p *Parser) DLTChangeHostVoteFront() error {

	err := p.generalCheck(`change_host_vote`)
	if err != nil {
		return p.ErrInfo(err)
	}

	verifyData := map[string]string{"host": "host", "addressVote": "walletAddress", "fuelRate": "int64", "public_key": "public_key"}
	err = p.CheckInputData(verifyData)
	if err != nil {
		return p.ErrInfo(err)
	}

	// public key need only when we don't have public_key in the dlt_wallets table
	publicKey, err := p.Single(`SELECT public_key_0 FROM dlt_wallets WHERE wallet_id = ?`, p.TxWalletID).String()
	if err != nil {
		return p.ErrInfo(err)
	}
	if len(publicKey) == 0 {
		bkey, err := hex.DecodeString(string(p.TxMaps.Bytes["public_key"]))
		if err != nil {
			return p.ErrInfo(err)
		}
		if lib.KeyToAddress(bkey) != lib.AddressToString(p.TxWalletID) {
			return p.ErrInfo("incorrect public_key")
		}
	}

	txTime := p.TxTime
	if p.BlockData != nil {
		txTime = p.BlockData.Time
	}
	lastForgingDataUpd, err := p.Single(`SELECT last_forging_data_upd FROM dlt_wallets WHERE wallet_id = ?`, p.TxWalletID).Int64()
	if err != nil || txTime-lastForgingDataUpd < 600 {
		return p.ErrInfo("txTime - lastForgingDataUpd < 600 sec")
	}

	forSign := fmt.Sprintf("%s,%s,%d,%s,%s,%s", p.TxMap["type"], p.TxMap["time"], p.TxWalletID, p.TxMap["host"], p.TxMap["addressVote"], p.TxMap["fuelRate"])
	CheckSignResult, err := utils.CheckSign(p.PublicKeys, forSign, p.TxMap["sign"], false)
	if err != nil {
		return p.ErrInfo(err)
	}
	if !CheckSignResult {
		return p.ErrInfo("incorrect sign")
	}

	return nil
}

// DLTChangeHostVote proceeds DLTChangeHostVote transaction
func (p *Parser) DLTChangeHostVote() error {
	var err error

	log.Debug("p.TxMaps.String[addressVote] %s", p.TxMaps.String["addressVote"])

	pkey, err := p.Single(`SELECT public_key_0 FROM dlt_wallets WHERE public_key_0 = [hex]`, p.TxMaps.Bytes["public_key"]).String()
	if err != nil {
		return p.ErrInfo(err)
	}

	if len(p.TxMaps.Bytes["public_key"]) > 0 && len(pkey) == 0 {
		_, err = p.selectiveLoggingAndUpd([]string{"host", "address_vote", "fuel_rate", "public_key_0", "last_forging_data_upd"}, []interface{}{p.TxMaps.String["host"], string(p.TxMaps.Int64["addressVote"]), string(p.TxMaps.String["fuelRate"]), utils.HexToBin(p.TxMaps.Bytes["public_key"]), p.BlockData.Time}, "dlt_wallets", []string{"wallet_id"}, []string{utils.Int64ToStr(p.TxWalletID)}, true)
	} else {
		_, err = p.selectiveLoggingAndUpd([]string{"host", "address_vote", "fuel_rate", "last_forging_data_upd"}, []interface{}{p.TxMaps.String["host"], p.TxMaps.String["addressVote"], p.TxMaps.Int64["fuelRate"], p.BlockData.Time}, "dlt_wallets", []string{"wallet_id"}, []string{utils.Int64ToStr(p.TxWalletID)}, true)
	}
	if err != nil {
		return p.ErrInfo(err)
	}
	p.UpdateFuel() // uncache fuel
	return nil
}

// DLTChangeHostVoteRollback rollbacks DLTChangeHostVote transaction
func (p *Parser) DLTChangeHostVoteRollback() error {
	return p.autoRollback()
}
