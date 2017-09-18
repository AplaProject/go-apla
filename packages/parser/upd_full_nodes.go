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
	"encoding/json"

	"github.com/EGaaS/go-egaas-mvp/packages/converter"
	"github.com/EGaaS/go-egaas-mvp/packages/model"
	"github.com/EGaaS/go-egaas-mvp/packages/utils"
	"github.com/EGaaS/go-egaas-mvp/packages/utils/tx"

	"strconv"

	"github.com/EGaaS/go-egaas-mvp/packages/config/syspar"
	"github.com/jinzhu/gorm"
	"gopkg.in/vmihailenco/msgpack.v2"
)

type UpdFullNodesParser struct {
	*Parser
	UpdFullNodes *tx.UpdFullNodes
}

func (p *UpdFullNodesParser) Init() error {
	updFullNodes := &tx.UpdFullNodes{}
	if err := msgpack.Unmarshal(p.TxBinaryData, updFullNodes); err != nil {
		return p.ErrInfo(err)
	}
	p.UpdFullNodes = updFullNodes
	return nil
}

func (p *UpdFullNodesParser) Validate() error {
	err := p.generalCheck(`upd_full_nodes`, &p.UpdFullNodes.Header, map[string]string{}) // undefined, cost=0
	if err != nil {
		return p.ErrInfo(err)
	}

	// We check to see if the time elapsed since the last update
	ufn := &model.UpdFullNode{}
	err = ufn.Read()
	if err != nil {
		return p.ErrInfo(err)
	}
	txTime := p.UpdFullNodes.Header.Time
	if p.BlockData != nil {
		txTime = p.BlockData.Time
	}
	if txTime-ufn.Time <= syspar.GetUpdFullNodesPeriod() {
		return utils.ErrInfoFmt("txTime - upd_full_nodes <= consts.UPD_FULL_NODES_PERIOD")
	}

	wallet := &model.DltWallet{}
	err = wallet.GetWallet(p.UpdFullNodes.UserID)
	if err != nil {
		return p.ErrInfo(err)
	}
	p.nodePublicKey = []byte(wallet.NodePublicKey)
	if len(p.nodePublicKey) == 0 {
		return utils.ErrInfoFmt("len(nodePublicKey) = 0")
	}

	CheckSignResult, err := utils.CheckSign([][]byte{p.nodePublicKey}, p.UpdFullNodes.ForSign(), p.UpdFullNodes.BinSignatures, true)
	if err != nil {
		return p.ErrInfo(err)
	}
	if !CheckSignResult {
		return p.ErrInfo("incorrect sign")
	}
	return nil
}

// UpdFullNodes proceeds UpdFullNodes transaction
func (p *UpdFullNodesParser) Action() error {
	_, _, err := p.selectiveLoggingAndUpd([]string{"time"}, []interface{}{p.BlockData.Time}, "upd_full_nodes", []string{`update`}, nil, false)
	if err != nil {
		return p.ErrInfo(err)
	}

	// выбирем ноды, где wallet_id
	// choose nodes where wallet_id is
	fns := &model.FullNode{}
	nodes, err := fns.GetAllFullNodesHasWalletID()
	if err != nil {
		return p.ErrInfo(err)
	}

	data := make([]map[string]string, 0)
	for _, node := range nodes {
		data = append(data, node.ToMap())
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		return p.ErrInfo(err)
	}

	// log them into the one record JSON
	rbFN := &model.RbFullNode{
		FullNodesWalletJson: jsonData,
		BlockID:             p.BlockData.BlockID,
		PrevRbID:            converter.StrToInt64(data[0]["rb_id"]),
	}
	err = rbFN.Create()
	if err != nil {
		return p.ErrInfo(err)
	}

	// удаляем где wallet_id
	// delete where the wallet_id is
	fn := &model.FullNode{}
	err = fn.DeleteNodesWithWallets()
	if err != nil {
		return p.ErrInfo(err)
	}
	maxID, err := fn.GetMaxID()
	if err != nil {
		return p.ErrInfo(err)
	}

	// update the AI
	err = model.SetAI("full_nodes", int64(maxID+1))
	if err != nil {
		return p.ErrInfo(err)
	}

	// получаем новые данные по wallet-нодам
	// obtain new data on wallet-nodes
	dw := &model.DltWallet{}
	all, err := dw.GetAddressVotes()
	if err != nil {
		return p.ErrInfo(err)
	}
	for _, addressVote := range all {
		wallet := &model.DltWallet{}
		err := wallet.GetWallet(int64(converter.StringToAddress(addressVote)))
		if err != nil {
			return p.ErrInfo(err)
		}
		// вставляем новые данные по wallet-нодам с указанием общего rb_id
		// insert new data on wallet-nodes with the indication of the common rb_id
		fn := &model.FullNode{WalletID: wallet.WalletID, Host: wallet.Host, RbID: rbFN.RbID}
		err = fn.Create()
		if err != nil {
			return p.ErrInfo(err)
		}
	}

	w := &model.DltWallet{}
	if err := w.GetNewFuelRate(); err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil
		}
		return p.ErrInfo(err)
	}
	newRate := strconv.FormatInt(w.FuelRate, 10)
	if len(newRate) > 0 {
		_, _, err = p.selectiveLoggingAndUpd([]string{"value"}, []interface{}{newRate}, "system_parameters", []string{"name"}, []string{"fuel_rate"}, true)
		if err != nil {
			return p.ErrInfo(err)
		}
	}
	return nil
}

func (p *UpdFullNodesParser) Rollback() error {
	err := p.selectiveRollback("upd_full_nodes", "", false)
	if err != nil {
		return p.ErrInfo(err)
	}

	// получим rb_id чтобы восстановить оттуда данные
	// get rb_id to restore the data from there
	fnRB := &model.FullNode{}
	if err := fnRB.GetRbIDFullNodesWithWallet(); err != nil {
		return p.ErrInfo(err)
	}

	rbFN := &model.RbFullNode{}
	err = rbFN.GetByRbID(int64(fnRB.ID)) // TODO: change rb_id type
	if err != nil {
		return p.ErrInfo(err)
	}
	fullNodesWallet := []map[string]string{{}}
	err = json.Unmarshal(rbFN.FullNodesWalletJson, &fullNodesWallet)
	if err != nil {
		return p.ErrInfo(err)
	}

	// удаляем новые данные
	// delete new data
	fn := &model.FullNode{}
	err = fn.DeleteNodesWithWallets()
	if err != nil {
		return p.ErrInfo(err)
	}

	maxID, err := fn.GetMaxID()
	if err != nil {
		return p.ErrInfo(err)
	}

	// update the AI
	err = model.SetAI("full_nodes", int64(maxID+1))
	if err != nil {
		return p.ErrInfo(err)
	}

	p.rollbackAI("rb_full_nodes", 1)

	for _, data := range fullNodesWallet {
		// вставляем новые данные по wallet-нодам с указанием общего rb_id
		// insert new data on wallet-nodes with the indication of the common rb_id
		fn := &model.FullNode{
			ID:                    int32(converter.StrToInt64(data["id"])),
			Host:                  data["host"],
			WalletID:              converter.StrToInt64(data["wallet_id"]),
			StateID:               converter.StrToInt64(data["state_id"]),
			FinalDelegateWalletID: converter.StrToInt64(data["final_delegate_wallet_id"]),
			FinalDelegateStateID:  converter.StrToInt64(data["final_delegate_state_id"]),
			RbID:                  converter.StrToInt64(data["rb_id"]),
		}
		err = fn.Create()
		if err != nil {
			return p.ErrInfo(err)
		}
	}
	err = p.autoRollback()
	if err != nil {
		return err
	}
	return nil
}

func (p *UpdFullNodesParser) Header() *tx.Header {
	return &p.UpdFullNodes.Header
}
