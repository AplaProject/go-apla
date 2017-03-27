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

package main

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/EGaaS/go-egaas-mvp/packages/consts"
	"github.com/EGaaS/go-egaas-mvp/packages/lib"
	"github.com/EGaaS/go-egaas-mvp/packages/utils"
)

type GetCntJson struct {
	Name  string `json:"name"`
	Error string `json:"error"`
}

type TxInfo struct {
	Id      int64  `json:"id"`
	BlockId int64  `json:"block"`
	Address string `json:"wallet"`
	State   string `json:"state"`
	Time    string `json:"time"`
	TxName  string `json:"tx"`
	Comment string `json:"comment"`
	next    *TxInfo
	prev    *TxInfo
}

const (
	txLimit = 300
)

var (
	txId          int64
	txLatest      int64
	txOff         int
	txList, txTop *TxInfo
	txStates      = make(map[int64]string)
	txContracts   = make(map[int32]string)
)

func DataToInfo(data []byte) {

}

func GetTx() {
	txList = &TxInfo{}
	txTop = txList
	prev := txList
	for i := 0; i < txLimit; i++ {
		txTop.next = &TxInfo{}
		txTop = txTop.next
		txTop.prev = prev
		prev = txTop
	}
	txTop.next = txList
	txList.prev = prev
	txTop = txList

	for {
		// b.hash, b.state_id,
		explorer, err := utils.DB.GetAll(`SELECT b.data, b.time, b.tx, b.id FROM block_chain as b
		where b.id > $1	order by b.id desc limit 30 offset 0`, -1, txLatest)
		if err == nil && len(explorer) > 0 {
			txLatest = utils.StrToInt64(explorer[0][`id`])
			for i := len(explorer); i > 0; i-- {
				item := explorer[i-1]
				if utils.StrToInt64(item[`tx`]) == 0 {
					continue
				}
				block := ([]byte(item[`data`]))[1:]
				utils.ParseBlockHeader(&block)
				for len(block) > 0 {
					size := int(utils.DecodeLength(&block))
					if size == 0 || len(block) < size {
						break
					}
					var (
						name, comment string
						txtime        int64
						wallet, state int64
					)
					itype := int(block[0])
					if itype < 128 {
						if stype, ok := consts.TxTypes[itype]; ok {
							name = stype
						} else {
							name = fmt.Sprintf("unknown %d", itype)
						}
						input := block[1:]
						txtime = utils.BinToDecBytesShift(&input, 4)
						length := utils.DecodeLength(&input)
						if length > 0 && length < 30 {
							wallet = utils.BytesToInt64(utils.BytesShift(&input, length))
							length = utils.DecodeLength(&input)
							if length > 0 && length < 20 {
								state = utils.BytesToInt64(utils.BytesShift(&input, length))
							}
						} else {
							break
						}
						//wallet, _ = utils.DecodeLenInt64(&input)
						//state, _ = utils.DecodeLenInt64(&input)
					} else {
						itype -= 128
						tmp := make([]byte, 4)
						for i := 0; i < itype; i++ {
							tmp[4-itype+i] = block[i+1]
						}
						idc := int32(binary.BigEndian.Uint32(tmp))

						if val, ok := txContracts[idc]; ok {
							name = val
						} else {
							resp, err := http.Get(strings.TrimRight(GSettings.Node, `/`) +
								fmt.Sprintf(`/ajax?json=ajax_get_cnt&id=%d`, idc))
							if err != nil {
								break
							}
							if answer, err := ioutil.ReadAll(resp.Body); err != nil {
								resp.Body.Close()
								break
							} else {
								var answerJson GetCntJson
								resp.Body.Close()
								if err = json.Unmarshal(answer, &answerJson); err != nil {
									break
								}
								var off int
								for off < len(answerJson.Name) && answerJson.Name[off] < 'A' {
									off++
								}
								name = answerJson.Name[off:]
								txContracts[idc] = name
							}
						}
						input := block[:]
						header := consts.TXHeader{}
						if err = lib.BinUnmarshal(&input, &header); err != nil {
							break
						}
						txtime = int64(header.Time)
						wallet = int64(header.WalletId)
						state = int64(header.StateId)
					}
					switch name {
					case `GenCitizen`:
						comment = `1`
					}
					if name == `GenCitizen` && txTop.TxName == name && txTop.BlockId == utils.StrToInt64(item[`id`]) &&
						txTop.Address == lib.AddressToString(uint64(wallet)) {
						txTop.Comment = fmt.Sprintf(`%d`, utils.StrToInt64(txTop.Comment)+1)
					} else {
						txTop = txTop.next
						txId++
						txTop.Id = txId
						txTop.BlockId = utils.StrToInt64(item[`id`])
						txTop.Address = lib.AddressToString(uint64(wallet))
						txTop.Comment = comment
						txTop.TxName = name
						txTop.Time = time.Unix(txtime, 0).String()[:19]
						if state > 0 {
							if val, ok := txStates[state]; ok {
								txTop.State = val
							} else {
								stateName, _ := utils.DB.Single(`select state_name from global_states_list where gstate_id=?`, state).String()
								if len(stateName) > 0 {
									txStates[state] = stateName
									txTop.State = stateName
								}
							}
							//					txTop.State = utils.Int64ToStr(state)
						} else {
							txTop.State = ``
						}
					}
					//					fmt.Println(`NAME`, *txTop)
					block = block[size:]
				}
			}
		}
		time.Sleep(1 * time.Second)
	}
}
