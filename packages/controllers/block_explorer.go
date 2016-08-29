package controllers

import (
	"github.com/DayLightProject/go-daylight/packages/utils"
	"encoding/hex"
)

const NBlockExplorer = `block_explorer`

type blockExplorerPage struct {
	Data       *CommonPage
	List       []map[string]string
	Latest     int64
	BlockId    int64
	BlockData  map[string]string
}

func init() {
	newPage(NBlockExplorer)
}

func (c *Controller) BlockExplorer() (string, error) {
	pageData := blockExplorerPage{Data:c.Data}
	
	blockId := utils.StrToInt64( c.r.FormValue("blockId"))
	
	if blockId > 0 {
		pageData.BlockId = blockId
		blockInfo,err := c.OneRow("SELECT * FROM block_chain where id=?", blockId).String()
		if err != nil {
			return "", utils.ErrInfo(err)
		}
		if len(blockInfo) > 0 {
			blockInfo[`hash`] = hex.EncodeToString([]byte(blockInfo[`hash`]))
			blockInfo[`size`] = utils.IntToStr(len(blockInfo[`data`]))
			tmp := hex.EncodeToString([]byte(blockInfo[`data`]))
			out := ``
			for i, ch := range tmp {
				out += string(ch) 
				if (i & 1) != 0 {
					out += ` `
				}
			}
			blockInfo[`data`] = out
			if blockId > 1 {
				parent,err := c.Single("SELECT hash FROM block_chain where id=?", blockId-1).String()
				if err == nil {
					blockInfo[`parent`] = hex.EncodeToString([]byte(parent))
				} else {
					blockInfo[`parent`] = err.Error()
				}
			}
		}
		pageData.BlockData = blockInfo	
	} else {
		latest := utils.StrToInt64( c.r.FormValue("latest"))
		if latest > 0 {
			curid,_ := c.Single("select max(id) from block_chain").Int64()
			if curid <= latest {
				return ``, nil
			}
		}	
		blockExplorer,err := c.GetAll("SELECT hash, cb_id, wallet_id, time, tx, id FROM block_chain order by id desc limit 0, 30", -1 )
		if err != nil {
			return "", utils.ErrInfo(err)
		}
		for ind := range blockExplorer {
			blockExplorer[ind][`hash`] = hex.EncodeToString([]byte(blockExplorer[ind][`hash`]))
		}
		pageData.List = blockExplorer
		if blockExplorer != nil && len(blockExplorer) > 0 {
			pageData.Latest = utils.StrToInt64(blockExplorer[0][`id`])
		}
	}
	return proceedTemplate( c, NBlockExplorer, &pageData )
}
