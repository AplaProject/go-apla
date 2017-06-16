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
	if err := msgpack.Unmarshal(p.BinaryData, changeNodeKey); err != nil {
		return p.ErrInfo(err)
	}
	p.ChangeNodeKey = changeNodeKey
	p.ChangeNodeKey.NewNodePublicKey = utils.BinToHex(p.ChangeNodeKey.NewNodePublicKey)
	return nil
}

func (p *ChangeNodeKeyParser) Validate() error {
	nodePublicKey, err := p.GetPublicKeyWalletOrCitizen(p.TxMaps.Int64["wallet_id"], p.ChangeNodeKey.Header.UserID)
	if err != nil || len(nodePublicKey) == 0 {
		return p.ErrInfo("incorrect user_id")
	}

	CheckSignResult, err := utils.CheckSign([][]byte{nodePublicKey}, p.ChangeNodeKey.ForSign(), p.TxMap["sign"], true)
	if err != nil || !CheckSignResult {
		CheckSignResult, err := utils.CheckSign(p.PublicKeys, p.ChangeNodeKey.ForSign(), p.TxMap["sign"], false)
		if err != nil || !CheckSignResult {
			return p.ErrInfo("incorrect sign")
		}
	}
	return nil
}

func (p *ChangeNodeKeyParser) Action() error {

	_, err := p.selectiveLoggingAndUpd([]string{"node_public_key"}, []interface{}{utils.HexToBin(p.ChangeNodeKey.NewNodePublicKey)}, "system_recognized_states", []string{"state_id"}, []string{utils.Int64ToStr(p.ChangeNodeKey.Header.StateID)}, true)
	if err != nil {
		return p.ErrInfo(err)
	}
	myKey, err := p.Single(`SELECT id FROM my_node_keys WHERE block_id = 0 AND public_key = [hex]`, p.ChangeNodeKey.NewNodePublicKey).Int64()
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

func (p *ChangeNodeKeyParser) Rollback() error {
	return p.autoRollback()
}

func (p ChangeNodeKeyParser) Header() *tx.Header {
	return &p.ChangeNodeKey.Header
}
