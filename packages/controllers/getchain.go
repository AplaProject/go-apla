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

	"github.com/EGaaS/go-egaas-mvp/packages/converter"
	"github.com/EGaaS/go-egaas-mvp/packages/model"
)

// ChainInfo contains infromation about transaction
type ChainInfo struct {
	ID int64 `json:"id"`
	//	Hash    string `json:"hash"`
	//	Wallet  int64  `json:"wallet_id"`
	Address string `json:"wallet_address"`
	//	State   int64  `json:"state_id"`
	Time string `json:"time"`
	Tx   string `json:"tx"`
}

// ChainMsg contains latest transactions
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

// UpdateChain returns the latest transactions
func UpdateChain(latest int64) (answer ChainMsg) {
	answer.Data = make([]ChainInfo, 0)
	for i := chainOff - 1; i >= 0 && len(answer.Data) < 10 && chainList[i].ID > latest; i-- {
		answer.Data = append(answer.Data, chainList[i])
		if i == chainOff-1 {
			answer.Latest = chainList[i].ID
		}
	}
	return
}

// GetChain updates information about transactions
func GetChain() {
	for {
		if model.DBConn != nil {
			// b.hash, b.state_id,
			block := &model.Block{}
			blockchain, err := block.GetBlocks(chainLatest, 30)
			if err == nil && len(blockchain) > 0 {
				chainLatest = blockchain[0].ID
				if chainOff+len(blockchain) > chainLimit {
					for i := 0; i < 50; i++ {
						chainList[i] = chainList[chainOff-50+i]
					}
					chainOff = 50
				}
				for i := len(blockchain); i > 0; i-- {
					item := blockchain[i-1]
					address := ``
					if item.WalletID != 0 {
						address = converter.AddressToString(item.WalletID)
					}

					chainList[chainOff] = ChainInfo{ID: item.WalletID,
						Address: address, Time: string(item.Time), Tx: string(item.Tx)}
					chainOff++
				}
			}
		}
		time.Sleep(1 * time.Second)
	}
}
