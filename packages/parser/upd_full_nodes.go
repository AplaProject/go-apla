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

	"github.com/EGaaS/go-egaas-mvp/packages/config"
	"github.com/EGaaS/go-egaas-mvp/packages/consts"
	"github.com/EGaaS/go-egaas-mvp/packages/converter"
	"github.com/EGaaS/go-egaas-mvp/packages/utils"
	"github.com/EGaaS/go-egaas-mvp/packages/utils/tx"

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
	updFullNodes, err := p.Single("SELECT time FROM upd_full_nodes").Int64()
	if err != nil {
		return p.ErrInfo(err)
	}
	txTime := p.UpdFullNodes.Header.Time
	if p.BlockData != nil {
		txTime = p.BlockData.Time
	}
	if txTime-updFullNodes <= consts.UPD_FULL_NODES_PERIOD {
		return utils.ErrInfoFmt("txTime - upd_full_nodes <= consts.UPD_FULL_NODES_PERIOD")
	}

	p.nodePublicKey, err = p.GetNodePublicKey(p.UpdFullNodes.UserID)
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
	data, err := p.GetAll(`SELECT * FROM full_nodes WHERE wallet_id != 0`, -1)
	if err != nil {
		return p.ErrInfo(err)
	}
	jsonData, err := json.Marshal(data)
	if err != nil {
		return p.ErrInfo(err)
	}

	log.Debug("data %v", data)
	log.Debug("data %v", data[0])
	log.Debug("data %v", data[0]["rb_id"])
	// логируем их в одну запись JSON
	// log them into the one record JSON
	rbID, err := p.ExecSQLGetLastInsertID(`INSERT INTO rb_full_nodes (full_nodes_wallet_json, block_id, prev_rb_id) VALUES (?, ?, ?)`, "rb_full_nodes", string(jsonData), p.BlockData.BlockID, data[0]["rb_id"])
	if err != nil {
		return p.ErrInfo(err)
	}

	// удаляем где wallet_id
	// delete where the wallet_id is
	err = p.ExecSQL(`DELETE FROM full_nodes WHERE wallet_id != 0`)
	if err != nil {
		return p.ErrInfo(err)
	}
	maxID, err := p.Single(`SELECT max(id) FROM full_nodes`).Int64()
	if err != nil {
		return p.ErrInfo(err)
	}

	// обновляем AI
	// update the AI
	err = p.SetAI("full_nodes", maxID+1)
	if err != nil {
		return p.ErrInfo(err)
	}

	// получаем новые данные по wallet-нодам
	// obtain new data on wallet-nodes
	all, err := p.GetList(`SELECT address_vote FROM dlt_wallets WHERE address_vote !='' AND amount > 10000000000000000000000 GROUP BY address_vote ORDER BY sum(amount) DESC LIMIT 100`).String()
	if err != nil {
		return p.ErrInfo(err)
	}
	for _, addressVote := range all {
		dltWallets, err := p.OneRow(`SELECT host, wallet_id FROM dlt_wallets WHERE wallet_id = ?`, int64(converter.StringToAddress(addressVote))).String()
		if err != nil {
			return p.ErrInfo(err)
		}
		// вставляем новые данные по wallet-нодам с указанием общего rb_id
		// insert new data on wallet-nodes with the indication of the common rb_id
		err = p.ExecSQL(`INSERT INTO full_nodes (wallet_id, host, rb_id) VALUES (?, ?, ?)`, dltWallets["wallet_id"], dltWallets["host"], rbID)
		if err != nil {
			return p.ErrInfo(err)
		}
	}

	newRate, err := p.Single(`SELECT fuel_rate FROM dlt_wallets WHERE fuel_rate !=0 GROUP BY fuel_rate ORDER BY sum(amount) DESC LIMIT 1`).String()
	if err != nil {
		return p.ErrInfo(err)
	}
	if len(newRate) > 0 {
		_, _, err = p.selectiveLoggingAndUpd([]string{"value"}, []interface{}{newRate}, "system_parameters", []string{"name"}, []string{"fuel_rate"}, true)
		if err != nil {
			return p.ErrInfo(err)
		}
		p.UpdateFuel() // update fuel
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
	rbID, err := p.Single(`SELECT rb_id FROM full_nodes WHERE wallet_id != 0`).Int64()
	if err != nil {
		return p.ErrInfo(err)
	}

	fullNodesWalletJSON, err := p.Single(`SELECT full_nodes_wallet_json FROM rb_full_nodes WHERE rb_id = ?`, rbID).Bytes()
	if err != nil {
		return p.ErrInfo(err)
	}
	fullNodesWallet := []map[string]string{{}}
	err = json.Unmarshal(fullNodesWalletJSON, &fullNodesWallet)
	if err != nil {
		return p.ErrInfo(err)
	}

	// удаляем новые данные
	// delete new data
	err = p.ExecSQL(`DELETE FROM full_nodes WHERE wallet_id != 0`)
	if err != nil {
		return p.ErrInfo(err)
	}

	maxID, err := p.Single(`SELECT max(id) FROM full_nodes`).Int64()
	if err != nil {
		return p.ErrInfo(err)
	}

	// обновляем AI
	// update the AI
	if config.ConfigIni["db_type"] == "sqlite" {
		err = p.SetAI("full_nodes", maxID)
	} else {
		err = p.SetAI("full_nodes", maxID+1)
	}
	if err != nil {
		return p.ErrInfo(err)
	}

	// удаляем новые данные
	// delete new data
	err = p.ExecSQL(`DELETE FROM rb_full_nodes WHERE rb_id = ?`, rbID)
	if err != nil {
		return p.ErrInfo(err)
	}
	p.rollbackAI("rb_full_nodes", 1)

	for _, data := range fullNodesWallet {
		// вставляем новые данные по wallet-нодам с указанием общего rb_id
		// insert new data on wallet-nodes with the indication of the common rb_id
		err = p.ExecSQL(`INSERT INTO full_nodes (id, host, wallet_id, state_id, final_delegate_wallet_id, final_delegate_state_id, rb_id) VALUES (?, ?, ?, ?, ?, ?, ?)`,
			data["id"], data["host"], data["wallet_id"], data["state_id"], data["final_delegate_wallet_id"], data["final_delegate_state_id"], data["rb_id"])
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
