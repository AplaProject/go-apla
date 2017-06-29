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
	"github.com/EGaaS/go-egaas-mvp/packages/utils/tx"
)

// общая проверка для всех _front
func (p *Parser) generalCheck(name string, header *tx.Header, conditionsCheck map[string]string) error {
	// проверим, есть ли такой юзер и заодно получим public_key
	txType := int64(header.Type)
	if header.StateID > 0 {
		p.TxStateID = uint32(header.StateID)
		p.TxStateIDStr = utils.Int64ToStr(header.StateID)
		p.TxCitizenID = header.UserID
		p.TxWalletID = 0
	} else {
		p.TxStateID = 0
		p.TxStateIDStr = ""
		p.TxWalletID = header.UserID
		p.TxCitizenID = 0
	}
	if txType == utils.TypeInt("DLTTransfer") || txType == utils.TypeInt("NewState") || txType == utils.TypeInt("DLTChangeHostVote") || txType == utils.TypeInt("ChangeNodeKeyDLT") || txType == utils.TypeInt("CitizenRequest") || txType == utils.TypeInt("UpdFullNodes") {
		data, err := p.OneRow("SELECT public_key_0 FROM dlt_wallets WHERE wallet_id = ?", p.TxWalletID).String()
		if err != nil {
			return utils.ErrInfo(err)
		}
		log.Debug("datausers", data)
		if len(data["public_key_0"]) == 0 {
			if len(header.PublicKey) == 0 {
				return utils.ErrInfoFmt("incorrect public_key")
			}
			// возможно юзер послал ключ с тр-ией
			log.Debug("pubkey %x", header.PublicKey)
			walletID, err := p.GetWalletIdByPublicKey(header.PublicKey)
			if err != nil {
				return utils.ErrInfo(err)
			}
			log.Debug("walletId %d", walletID)
			if walletID == 0 {
				return utils.ErrInfoFmt("incorrect wallet_id or public_key")
			}
			p.PublicKeys = append(p.PublicKeys, header.PublicKey)
		} else {
			p.PublicKeys = append(p.PublicKeys, []byte(data["public_key_0"]))
			log.Debug("data[public_key_0]", data["public_key_0"])
		}
	} else {
		log.Debug(`SELECT * FROM "`+utils.Int64ToStr(header.StateID)+`_citizens" WHERE id = %d`, header.UserID)
		data, err := p.OneRow("SELECT public_key_0 FROM dlt_wallets WHERE wallet_id = ?", header.UserID).String()
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
	// 128 - это нод-ключ
	if len(header.BinSignatures) < 64 || len(header.BinSignatures) > 5120 {
		return utils.ErrInfoFmt("incorrect sign size %d", len(header.BinSignatures))
	}
	for _, cond := range []string{`conditions`, `conditions_change`, `permissions`} {
		if val, ok := conditionsCheck[cond]; ok && len(val) == 0 {
			return utils.ErrInfoFmt("Conditions cannot be empty")
		}
		if err := smart.CompileEval(string(conditionsCheck[cond]), uint32(p.TxStateID)); err != nil {
			return utils.ErrInfo(err)
		}
	}

	return p.checkPrice(name)
}
