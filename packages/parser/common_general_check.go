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
	"github.com/AplaProject/go-apla/packages/converter"
	"github.com/AplaProject/go-apla/packages/crypto"
	"github.com/AplaProject/go-apla/packages/model"
	"github.com/AplaProject/go-apla/packages/smart"
	"github.com/AplaProject/go-apla/packages/utils"
	"github.com/AplaProject/go-apla/packages/utils/tx"
)

// common check for all transactions
func (p *Parser) generalCheck(name string, header *tx.Header, conditionsCheck map[string]string) error {
	txType := int64(header.Type)
	if header.StateID > 0 {
		p.TxStateID = uint32(header.StateID)
		p.TxStateIDStr = converter.Int64ToStr(header.StateID)
		p.TxCitizenID = header.UserID
		p.TxWalletID = 0
	} else {
		p.TxStateID = 0
		p.TxStateIDStr = ""
		p.TxWalletID = header.UserID
		p.TxCitizenID = 0
	}
	if txType == utils.TypeInt("DLTTransfer") || txType == utils.TypeInt("NewState") || txType == utils.TypeInt("DLTChangeHostVote") || txType == utils.TypeInt("ChangeNodeKeyDLT") || txType == utils.TypeInt("CitizenRequest") || txType == utils.TypeInt("UpdFullNodes") {
		dltWallet := &model.DltWallet{}
		_, err := dltWallet.Get(p.DbTransaction, p.TxWalletID)
		if err != nil {
			return utils.ErrInfo(err)
		}

		if len(dltWallet.PublicKey) == 0 {
			if len(header.PublicKey) == 0 {
				return utils.ErrInfoFmt("incorrect public_key")
			}
			walletID, err := crypto.GetWalletIDByPublicKey(header.PublicKey)
			if err != nil {
				return utils.ErrInfo(err)
			}
			if walletID == 0 {
				return utils.ErrInfoFmt("incorrect wallet_id or public_key")
			}
			p.PublicKeys = append(p.PublicKeys, header.PublicKey)
		} else {
			p.PublicKeys = append(p.PublicKeys, []byte(dltWallet.PublicKey))
			log.Debug("data[public_key_0]", dltWallet.PublicKey)
		}
	} else {
		log.Debugf("parser general check, user_id = %d", header.UserID)
		dltWallet := &model.DltWallet{}
		_, err := dltWallet.Get(p.DbTransaction, header.UserID)
		if err != nil {
			return utils.ErrInfo(err)
		}
		if len(dltWallet.PublicKey) == 0 {
			return utils.ErrInfoFmt("incorrect user_id")
		}
		p.PublicKeys = append(p.PublicKeys, []byte(dltWallet.PublicKey))
	}

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
