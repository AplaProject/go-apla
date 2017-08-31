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
	"fmt"
	"strconv"
	"strings"

	"github.com/EGaaS/go-egaas-mvp/packages/consts"
	"github.com/EGaaS/go-egaas-mvp/packages/converter"
	logger "github.com/EGaaS/go-egaas-mvp/packages/log"
	"github.com/EGaaS/go-egaas-mvp/packages/model"
	"github.com/EGaaS/go-egaas-mvp/packages/smart"
	"github.com/EGaaS/go-egaas-mvp/packages/utils"
)

const nBlockExplorer = `block_explorer`

type blockExplorerPage struct {
	Data       *CommonPage
	List       []map[string]string
	Latest     int64
	BlockID    int64
	BlockData  map[string]string
	SinglePage int64
	Host       string
}

func init() {
	newPage(nBlockExplorer)
}

// BlockExplorer is a controller for block explorer template page
func (c *Controller) BlockExplorer() (string, error) {
	pageData := blockExplorerPage{Data: c.Data, Host: c.r.Host}

	blockID, err := strconv.ParseInt(c.r.FormValue("blockId"), 10, 64)
	if err != nil {
		logger.LogInfo(consts.StrToIntError, c.r.FormValue("blockId"))
	}

	pageData.SinglePage, err = strconv.ParseInt(c.r.FormValue("singlePage"), 10, 64)
	if err != nil {
		logger.LogInfo(consts.StrToIntError, c.r.FormValue("singlePage"))
	}
	if blockID > 0 {
		pageData.BlockID = blockID
		block := &model.Block{}
		err := block.GetBlock(blockID)
		if err != nil {
			return "", utils.ErrInfo(err)
		}

		blockInfo := block.ToMap()
		if len(blockInfo) > 0 {
			blockInfo[`hash`] = string(converter.BinToHex(block.Hash))
			blockInfo[`size`] = converter.IntToStr(len(block.Data))
			blockInfo[`wallet_address`] = converter.AddressToString(block.WalletID)
			tmp := string(converter.BinToHex(block.Data))
			out := ``
			for i, ch := range tmp {
				out += string(ch)
				if (i & 1) != 0 {
					out += ` `
				}
			}
			if blockID > 1 {
				parent := &model.Block{}
				err = parent.GetBlock(blockID - 1)
				if err == nil {
					blockInfo[`parent`] = string(converter.BinToHex(parent.Hash))
				} else {
					blockInfo[`parent`] = err.Error()
				}
			}
			txlist := make([]string, 0)
			block := block.Data[1:]
			utils.ParseBlockHeader(&block)

			for len(block) > 0 {
				length, err := converter.DecodeLength(&block)
				if err != nil {
					log.Fatal(err)
				}
				size := int(length)
				if size == 0 || len(block) < size {
					break
				}
				var name string
				itype := int(block[0])
				if itype < 128 {
					if stype, ok := consts.TxTypes[itype]; ok {
						name = stype
					} else {
						name = fmt.Sprintf("unknown %d", itype)
					}
				} else {
					itype -= 128
					tmp := make([]byte, 4)
					for i := 0; i < itype; i++ {
						tmp[4-itype+i] = block[i+1]
					}
					idc := int32(binary.BigEndian.Uint32(tmp))
					contract := smart.GetContractByID(idc)
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
		if c.r.FormValue("modal") == `1` {
			return proceedTemplate(c, `modal_block_detail`, &pageData)
		}
	} else {
		latest, err := strconv.ParseInt(c.r.FormValue("latest"), 10, 64)
		if err != nil {
			logger.LogInfo(consts.StrToIntError, c.r.FormValue("latest"))
		}
		block := &model.Block{}
		if latest > 0 {
			block.GetMaxBlock()
			if block.ID <= latest {
				return ``, nil
			}
		}

		blockchain, err := block.GetBlocks(-1, 30)
		if err != nil {
			log.Debugf("can't get last 30 blocks")
			return "", utils.ErrInfo(err)
		}

		blockExplorer := make([]map[string]string, 0)
		for _, block := range blockchain {
			row := block.ToMap()
			row["wallet_address"] = converter.AddressToString(block.WalletID)
			blockExplorer = append(blockExplorer, row)
		}

		pageData.List = blockExplorer
		if blockExplorer != nil && len(blockExplorer) > 0 {
			pageData.Latest, err = strconv.ParseInt(blockExplorer[0][`id`], 10, 64)
			if err != nil {
				logger.LogInfo(consts.StrToIntError, blockExplorer[0][`id`])
			}
		}
	}
	return proceedTemplate(c, nBlockExplorer, &pageData)
}
