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

	"github.com/EGaaS/go-egaas-mvp/packages/lib"
	"github.com/EGaaS/go-egaas-mvp/packages/utils"
	"github.com/EGaaS/go-egaas-mvp/packages/utils/tx"

	"gopkg.in/vmihailenco/msgpack.v2"
)

type DLTChangeHostVoteParser struct {
	*Parser
	DLTChangeHostVote *tx.DLTChangeHostVote
}

func (p *DLTChangeHostVoteParser) Init() error {
	dltChangeHostVote := &tx.DLTChangeHostVote{}
	if err := msgpack.Unmarshal(p.BinaryData, dltChangeHostVote); err != nil {
		return p.ErrInfo(err)
	}
	p.DLTChangeHostVote = dltChangeHostVote
	p.DLTChangeHostVote.PublicKey = utils.BinToHex(p.DLTChangeHostVote.Header.PublicKey)
	return nil
}

func (p *DLTChangeHostVoteParser) Validate() error {
	err := p.generalCheck(`change_host_vote`, &p.DLTChangeHostVote.Header)
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
		bkey, err := hex.DecodeString(string(p.DLTChangeHostVote.PublicKey))
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

	CheckSignResult, err := utils.CheckSign(p.PublicKeys, p.DLTChangeHostVote.ForSign(), p.TxMap["sign"], false)
	if err != nil {
		return p.ErrInfo(err)
	}
	if !CheckSignResult {
		return p.ErrInfo("incorrect sign")
	}

	return nil
}

func (p *DLTChangeHostVoteParser) Action() error {
	var err error
	log.Debug("p.TxMaps.String[addressVote] %s", p.TxMaps.String["addressVote"])
	pkey, err := p.Single(`SELECT public_key_0 FROM dlt_wallets WHERE public_key_0 = [hex]`, p.DLTChangeHostVote.Header.PublicKey).String()
	if err != nil {
		return p.ErrInfo(err)
	}

	if len(p.DLTChangeHostVote.Header.PublicKey) > 0 && len(pkey) == 0 {
		_, err = p.selectiveLoggingAndUpd([]string{"host", "address_vote", "fuel_rate", "public_key_0", "last_forging_data_upd"}, []interface{}{p.DLTChangeHostVote.Host, string(p.DLTChangeHostVote.AddressVote), string(p.DLTChangeHostVote.FuelRate), utils.HexToBin(p.DLTChangeHostVote.Header.PublicKey), p.BlockData.Time}, "dlt_wallets", []string{"wallet_id"}, []string{utils.Int64ToStr(p.TxWalletID)}, true)
	} else {
		_, err = p.selectiveLoggingAndUpd([]string{"host", "address_vote", "fuel_rate", "last_forging_data_upd"}, []interface{}{p.DLTChangeHostVote.Host, p.DLTChangeHostVote.AddressVote, p.DLTChangeHostVote.FuelRate, p.BlockData.Time}, "dlt_wallets", []string{"wallet_id"}, []string{utils.Int64ToStr(p.TxWalletID)}, true)
	}
	if err != nil {
		return p.ErrInfo(err)
	}
	p.UpdateFuel() // uncache fuel
	return nil
}

func (p *DLTChangeHostVoteParser) Rollback() error {
	return p.autoRollback()
}

func (p DLTChangeHostVoteParser) Header() *tx.Header {
	return &p.DLTChangeHostVote.Header
}
