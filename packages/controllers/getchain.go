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

package controllers

import (
	"time"

	"github.com/EGaaS/go-egaas-mvp/packages/lib"
	"github.com/EGaaS/go-egaas-mvp/packages/utils"
)

type ChainInfo struct {
	Id int64 `json:"id"`
	//	Hash    string `json:"hash"`
	//	Wallet  int64  `json:"wallet_id"`
	Address string `json:"wallet_address"`
	//	State   int64  `json:"state_id"`
	Time string `json:"time"`
	Tx   string `json:"tx"`
}

type ChainMsg struct {
	Data   []ChainInfo `json:"data"`
	Latest int64       `json:"latest"`
}

const (
	chainLimit = 100
)

var (
	chainLatest int64
	chainOff    int
	chainList   = make([]ChainInfo, chainLimit)
)

func UpdateChain(latest int64) (answer ChainMsg) {
	answer.Data = make([]ChainInfo, 0)
	for i := chainOff - 1; i >= 0 && len(answer.Data) < 10 && chainList[i].Id > latest; i-- {
		answer.Data = append(answer.Data, chainList[i])
		if i == chainOff-1 {
			answer.Latest = chainList[i].Id
		}
	}
	return
}

func GetChain() {
	for {
		// b.hash, b.state_id,
		explorer, err := utils.DB.GetAll(`SELECT   b.wallet_id, b.time, b.tx, b.id FROM block_chain as b
		where b.id > $1	order by b.id desc limit 30 offset 0`, -1, chainLatest)
		if err == nil && len(explorer) > 0 {
			chainLatest = utils.StrToInt64(explorer[0][`id`])
			if chainOff+len(explorer) > chainLimit {
				for i := 0; i < 50; i++ {
					chainList[i] = chainList[chainOff-50+i]
				}
				chainOff = 50
			}
			for i := len(explorer); i > 0; i-- {
				item := explorer[i-1]
				wallet_id := utils.StrToInt64(item[`wallet_id`])
				address := ``
				if wallet_id != 0 {
					address = lib.AddressToString(uint64(wallet_id))
				}

				chainList[chainOff] = ChainInfo{Id: utils.StrToInt64(item[`id`]),
					//Hash: hex.EncodeToString([]byte(item[`hash`])), Wallet: wallet_id,State: utils.StrToInt64(item[`state_id`]),
					Address: address, Time: item[`time`], Tx: item[`tx`]}
				chainOff++
			}
		}
		time.Sleep(1 * time.Second)
	}
}
