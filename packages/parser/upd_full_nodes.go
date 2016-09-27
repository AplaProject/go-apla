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
	"github.com/DayLightProject/go-daylight/packages/utils"
)

func (p *Parser) UpdFullNodesInit() error {
	err := p.GetTxMaps([]map[string]string{{"sign": "bytes"}})
	if err != nil {
		return p.ErrInfo(err)
	}
	return nil
}

func (p *Parser) UpdFullNodesFront() error {
	return nil
}

func (p *Parser) UpdFullNodes() error {

	err := p.selectiveLoggingAndUpd([]string{"time"}, []interface{}{p.TxTime}, "upd_full_nodes", nil, nil, false)
	if err != nil {
		return p.ErrInfo(err)
	}

	// выбирем ноды, где wallet_id
	data, err := p.GetAll(`SELECT * FROM full_nodes WHERE wallet_id > 0`, -1)
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
	rbId, err := p.ExecSqlGetLastInsertId(`INSERT INTO rb_full_nodes (full_nodes_wallet_json, block_id, prev_rb_id) VALUES (?, ?, ?)`, "rb_full_nodes", string(jsonData), p.BlockData.BlockId, data[0]["rb_id"])
	if err != nil {
		return p.ErrInfo(err)
	}

	// удаляем где wallet_id
	err = p.ExecSql(`DELETE FROM full_nodes WHERE wallet_id > 0`)
	if err != nil {
		return p.ErrInfo(err)
	}
	maxId, err := p.Single(`SELECT max(id) FROM full_nodes`).Int64()
	if err != nil {
		return p.ErrInfo(err)
	}

	// обновляем AI
	if p.ConfigIni["db_type"] == "sqlite" {
		err = p.SetAI("full_nodes", maxId)
	} else {
		err = p.SetAI("full_nodes", maxId+1)
	}
	if err != nil {
		return p.ErrInfo(err)
	}

	// получаем новые данные по wallet-нодам
	all, err := p.GetList(`SELECT address_vote FROM dlt_wallets GROUP BY address_vote ORDER BY sum(amount) DESC LIMIT 10`).String()
	if err != nil {
		return p.ErrInfo(err)
	}
	for _, address_vote := range all {
		dlt_wallets, err := p.OneRow(`SELECT host, wallet_id FROM dlt_wallets WHERE address = [hex]`, utils.BinToHex(address_vote)).String()
		if err != nil {
			return p.ErrInfo(err)
		}
		// вставляем новые данные по wallet-нодам с указанием общего rb_id
		err = p.ExecSql(`INSERT INTO full_nodes (wallet_id, host, rb_id) VALUES (?, ?, ?)`, dlt_wallets["wallet_id"], dlt_wallets["host"], rbId)
		if err != nil {
			return p.ErrInfo(err)
		}
	}

	return nil
}

func (p *Parser) UpdFullNodesRollback() error {
	err := p.selectiveRollback("upd_full_nodes", "", false)
	if err != nil {
		return p.ErrInfo(err)
	}

	// получим rb_id чтобы восстановить оттуда данные
	rbId, err := p.Single(`SELECT rb_id FROM full_nodes WHERE wallet_id > 0`).Int64()
	if err != nil {
		return p.ErrInfo(err)
	}

	full_nodes_wallet_json, err := p.Single(`SELECT full_nodes_wallet_json FROM rb_full_nodes WHERE rb_id = ?`, rbId).Bytes()
	if err != nil {
		return p.ErrInfo(err)
	}
	full_nodes_wallet := []map[string]string{{}}
	err = json.Unmarshal(full_nodes_wallet_json, &full_nodes_wallet)
	if err != nil {
		return p.ErrInfo(err)
	}

	// удаляем новые данные
	err = p.ExecSql(`DELETE FROM full_nodes WHERE wallet_id > 0`)
	if err != nil {
		return p.ErrInfo(err)
	}

	maxId, err := p.Single(`SELECT max(id) FROM full_nodes`).Int64()
	if err != nil {
		return p.ErrInfo(err)
	}

	// обновляем AI
	if p.ConfigIni["db_type"] == "sqlite" {
		err = p.SetAI("full_nodes", maxId)
	} else {
		err = p.SetAI("full_nodes", maxId+1)
	}
	if err != nil {
		return p.ErrInfo(err)
	}

	// удаляем новые данные
	err = p.ExecSql(`DELETE FROM rb_full_nodes WHERE rb_id = ?`, rbId)
	if err != nil {
		return p.ErrInfo(err)
	}
	p.rollbackAI("rb_full_nodes", 1)

	for _, data := range full_nodes_wallet {
		// вставляем новые данные по wallet-нодам с указанием общего rb_id
		err = p.ExecSql(`INSERT INTO full_nodes (id, host, wallet_id, state_id, final_delegate_wallet_id, final_delegate_state_id, rb_id) VALUES (?, ?, ?, ?, ?, ?, ?)`, data["id"], data["host"], data["wallet_id"], data["state_id"], data["final_delegate_wallet_id"], data["final_delegate_state_id"], data["rb_id"])
		if err != nil {
			return p.ErrInfo(err)
		}
	}

	return nil
}
func (p *Parser) UpdFullNodesRollbackFront() error {
	return nil
}
