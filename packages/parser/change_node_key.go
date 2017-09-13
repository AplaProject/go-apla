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
	"github.com/EGaaS/go-egaas-mvp/packages/converter"
	"github.com/EGaaS/go-egaas-mvp/packages/model"
	"github.com/EGaaS/go-egaas-mvp/packages/utils"
	"github.com/EGaaS/go-egaas-mvp/packages/utils/tx"
	"gopkg.in/vmihailenco/msgpack.v2"
)

type ChangeNodeKeyParser struct {
	*Parser
	ChangeNodeKey *tx.ChangeNodeKey
}

func (p *ChangeNodeKeyParser) Init() error {
	changeNodeKey := &tx.ChangeNodeKey{}
	if err := msgpack.Unmarshal(p.TxBinaryData, changeNodeKey); err != nil {
		return p.ErrInfo(err)
	}
	p.ChangeNodeKey = changeNodeKey
	p.ChangeNodeKey.NewNodePublicKey = converter.BinToHex(p.ChangeNodeKey.NewNodePublicKey)
	return nil
}

func (p *ChangeNodeKeyParser) Validate() error {
	wallet := &model.DltWallet{}
	err := wallet.GetWalletTransaction(p.DbTransaction, p.ChangeNodeKey.Header.UserID)
	if err != nil || len(wallet.PublicKey) == 0 {
		return p.ErrInfo("incorrect user_id")
	}
	CheckSignResult, err := utils.CheckSign([][]byte{[]byte(wallet.PublicKey)}, p.ChangeNodeKey.ForSign(), p.ChangeNodeKey.Header.BinSignatures, true)
	if err != nil || !CheckSignResult {
		CheckSignResult, err := utils.CheckSign(p.PublicKeys, p.ChangeNodeKey.ForSign(), p.ChangeNodeKey.Header.BinSignatures, false)
		if err != nil || !CheckSignResult {
			return p.ErrInfo("incorrect sign")
		}
	}
	return nil
}

func (p *ChangeNodeKeyParser) Action() error {
	_, _, err := p.selectiveLoggingAndUpd([]string{"node_public_key"}, []interface{}{converter.HexToBin(p.ChangeNodeKey.NewNodePublicKey)}, "system_recognized_states", []string{"state_id"}, []string{converter.Int64ToStr(p.ChangeNodeKey.Header.StateID)}, true)
	if err != nil {
		return p.ErrInfo(err)
	}
	key := &model.MyNodeKey{}
	found, err := key.GetZeroBlock(converter.HexToBin(p.ChangeNodeKey.NewNodePublicKey))
	if err != nil {
		return p.ErrInfo(err)
	}
	if found {
		_, _, err := p.selectiveLoggingAndUpd([]string{"block_id"}, []interface{}{p.BlockData.BlockID}, "my_node_keys", []string{"id"}, []string{converter.Int64ToStr(int64(key.ID))}, true)
		if err != nil {
			return p.ErrInfo(err)
		}
	}
	return nil
}

func (p *ChangeNodeKeyParser) Rollback() error {
	return p.autoRollback()
}

func (p ChangeNodeKeyParser) Header() *tx.Header {
	return &p.ChangeNodeKey.Header
}
