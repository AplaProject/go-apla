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
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/DayLightProject/go-daylight/packages/lib"
	"github.com/DayLightProject/go-daylight/packages/smart"
	"github.com/DayLightProject/go-daylight/packages/utils"
)

const NBlockExplorer = `block_explorer`

type blockExplorerPage struct {
	Data       *CommonPage
	List       []map[string]string
	Latest     int64
	BlockId    int64
	BlockData  map[string]string
	SinglePage int64
}

func init() {
	newPage(NBlockExplorer)
}

func (c *Controller) BlockExplorer() (string, error) {
	pageData := blockExplorerPage{Data: c.Data}

	blockId := utils.StrToInt64(c.r.FormValue("blockId"))
	pageData.SinglePage = utils.StrToInt64(c.r.FormValue("singlePage"))

	if blockId > 0 {
		pageData.BlockId = blockId
		blockInfo, err := c.OneRow(`SELECT b.*, w.address FROM block_chain as b
		left join dlt_wallets as w on b.wallet_id=w.wallet_id
		where b.id=?`, blockId).String()
		if err != nil {
			return "", utils.ErrInfo(err)
		}
		if len(blockInfo) > 0 {
			blockInfo[`hash`] = hex.EncodeToString([]byte(blockInfo[`hash`]))
			blockInfo[`size`] = utils.IntToStr(len(blockInfo[`data`]))
			if len(blockInfo[`address`]) > 0 && blockInfo[`address`] != `NULL` {
				blockInfo[`wallet_address`] = lib.KeyToAddress([]byte(blockInfo[`address`]))
			} else {
				blockInfo[`wallet_address`] = ``
			}
			tmp := hex.EncodeToString([]byte(blockInfo[`data`]))
			out := ``
			for i, ch := range tmp {
				out += string(ch)
				if (i & 1) != 0 {
					out += ` `
				}
			}
			if blockId > 1 {
				parent, err := c.Single("SELECT hash FROM block_chain where id=?", blockId-1).String()
				if err == nil {
					blockInfo[`parent`] = hex.EncodeToString([]byte(parent))
				} else {
					blockInfo[`parent`] = err.Error()
				}
			}
			txlist := make([]string, 0)
			block := ([]byte(blockInfo[`data`]))[1:]
			utils.ParseBlockHeader(&block)
			//			fmt.Printf("Block OK %v sign=%d %d %x", *pblock, len((*pblock).Sign), len(block), block)
			for len(block) > 0 {
				size := int(utils.DecodeLength(&block))
				if size == 0 || len(block) < size {
					break
				}
				var name string
				itype := int(block[0])
				if itype < 128 {
					name = fmt.Sprintf("%d", itype)
				} else {
					itype -= 128
					tmp := make([]byte, 4)
					for i := 0; i < itype; i++ {
						tmp[4-itype+i] = block[i+1]
					}
					idc := int32(binary.BigEndian.Uint32(tmp))
					contract := smart.GetContractById(idc)
					if contract != nil {
						name = contract.Name
					} else {
						name = fmt.Sprintf(`Unknown=%d`, idc)
					}
				}
				txlist = append(txlist, name)
				block = block[size:]
			}
			blockInfo[`data`] = out
			blockInfo[`tx_list`] = strings.Join(txlist, `, `)
		}
		pageData.BlockData = blockInfo
	} else {
		latest := utils.StrToInt64(c.r.FormValue("latest"))
		if latest > 0 {
			curid, _ := c.Single("select max(id) from block_chain").Int64()
			if curid <= latest {
				return ``, nil
			}
		}
		blockExplorer, err := c.GetAll(`SELECT  w.address, b.hash, b.state_id, b.wallet_id, b.time, b.tx, b.id FROM block_chain as b
		left join dlt_wallets as w on b.wallet_id=w.wallet_id
		order by b.id desc limit 30 offset 0`, -1)
		if err != nil {
			return "", utils.ErrInfo(err)
		}
		for ind := range blockExplorer {
			blockExplorer[ind][`hash`] = hex.EncodeToString([]byte(blockExplorer[ind][`hash`]))
			if len(blockExplorer[ind][`address`]) > 0 && blockExplorer[ind][`address`] != `NULL` {
				blockExplorer[ind][`wallet_address`] = blockExplorer[ind][`address`]
			} else {
				blockExplorer[ind][`wallet_address`] = ``
			}
			/*			if blockExplorer[ind][`tx`] == `[]` {
							blockExplorer[ind][`tx_count`] = `0`
						} else {
							var tx []string
							json.Unmarshal([]byte(blockExplorer[ind][`tx`]), &tx)
							if tx != nil && len(tx) > 0 {
								blockExplorer[ind][`tx_count`] = utils.IntToStr(len(tx))
							}
						}*/
		}
		pageData.List = blockExplorer
		if blockExplorer != nil && len(blockExplorer) > 0 {
			pageData.Latest = utils.StrToInt64(blockExplorer[0][`id`])
		}
	}
	return proceedTemplate(c, NBlockExplorer, &pageData)
}
