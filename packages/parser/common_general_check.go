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
	"github.com/EGaaS/go-egaas-mvp/packages/smart"
	"github.com/EGaaS/go-egaas-mvp/packages/utils"
)

// общая проверка для всех _front
// general check for all _front
func (p *Parser) generalCheck(name string) error {
	log.Debug("%s", p.TxMap)
	if !utils.CheckInputData(p.TxMap["wallet_id"], "int64") {
		return utils.ErrInfoFmt("incorrect wallet_id")
	}
	if !utils.CheckInputData(p.TxMap["citizen_id"], "int64") {
		return utils.ErrInfoFmt("incorrect citizen_id")
	}
	if !utils.CheckInputData(p.TxMap["time"], "int") {
		return utils.ErrInfoFmt("incorrect time")
	}

	// проверим, есть ли такой юзер и заодно получим public_key
	// check if such a user exists and at the same time we will get the public_key
	if p.TxMaps.Int64["type"] == utils.TypeInt("DLTTransfer") || p.TxMaps.Int64["type"] == utils.TypeInt("NewState") || p.TxMaps.Int64["type"] == utils.TypeInt("DLTChangeHostVote") || p.TxMaps.Int64["type"] == utils.TypeInt("ChangeNodeKeyDLT") || p.TxMaps.Int64["type"] == utils.TypeInt("CitizenRequest") || p.TxMaps.Int64["type"] == utils.TypeInt("UpdFullNodes") {
		data, err := p.OneRow("SELECT public_key_0 FROM dlt_wallets WHERE wallet_id = ?", utils.BytesToInt64(p.TxMap["wallet_id"])).String()
		if err != nil {
			return utils.ErrInfo(err)
		}
		log.Debug("datausers", data)
		if len(data["public_key_0"]) == 0 {
			if len(p.TxMap["public_key"]) == 0 {
				return utils.ErrInfoFmt("incorrect public_key")
			}
			// возможно юзер послал ключ с тр-ией
			// probably user sent the key with territory
			log.Debug("pubkey %s", p.TxMap["public_key"])
			log.Debug("pubkey %x", p.TxMap["public_key"])
			walletID, err := p.GetWalletIDByPublicKey(p.TxMap["public_key"])
			if err != nil {
				return utils.ErrInfo(err)
			}
			log.Debug("walletId %d", walletID)
			if walletID == 0 {
				return utils.ErrInfoFmt("incorrect wallet_id or public_key")
			}
			p.PublicKeys = append(p.PublicKeys, utils.HexToBin(p.TxMap["public_key"]))
		} else {
			p.PublicKeys = append(p.PublicKeys, []byte(data["public_key_0"]))
			log.Debug("data[public_key_0]", data["public_key_0"])
		}
	} else {
		log.Debug(`SELECT * FROM "`+utils.UInt32ToStr(p.TxStateID)+`_citizens" WHERE id = %d`, p.TxCitizenID)
		data, err := p.OneRow("SELECT public_key_0 FROM dlt_wallets WHERE wallet_id = ?", utils.Int64ToStr(p.TxCitizenID)).String()
		if err != nil {
			return utils.ErrInfo(err)
		}
		log.Debug("datausers", data)
		if len(data["public_key_0"]) == 0 {
			return utils.ErrInfoFmt("incorrect user_id")
		}
		p.PublicKeys = append(p.PublicKeys, []byte(data["public_key_0"]))
	}
	// чтобы не записали слишком длинную подпись
	// for not to record too long signature 
	// 128 - это нод-ключ
	// 128 is the node-key
	if len(p.TxMap["sign"]) < 64 || len(p.TxMap["sign"]) > 5120 {
		return utils.ErrInfoFmt("incorrect sign size %d", len(p.TxMap["sign"]))
	}
	for _, cond := range []string{`conditions`, `conditions_change`, `permissions`} {
		if val, ok := p.TxMap[cond]; ok && len(val) == 0 {
			return utils.ErrInfoFmt("Conditions cannot be empty")
		}
		if err := smart.CompileEval(string(p.TxMap[cond]), uint32(p.TxStateID)); err != nil {
			return utils.ErrInfo(err)
		}
	}

	return p.checkPrice(name)
}
