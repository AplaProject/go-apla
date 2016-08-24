package controllers

import (
	"github.com/DayLightProject/go-daylight/packages/utils"
	"encoding/hex"
)

const NBlockExplorer = `block_explorer`

type blockExplorerPage struct {
	CommonPage
	List       []map[string]string
	BlockId    int64
	BlockData  map[string]string
}

func init() {
	newPage(NBlockExplorer)
}

func (c *Controller) BlockExplorer() (string, error) {
	var pageData blockExplorerPage
	
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
		}
		pageData.BlockData = blockInfo	
	} else {
		blockExplorer,err := c.GetAll("SELECT hash, cb_id, wallet_id, time, tx, id FROM block_chain order by id desc limit 0, 30",-1)
		if err != nil {
			return "", utils.ErrInfo(err)
		}
		for ind := range blockExplorer {
			blockExplorer[ind][`hash`] = hex.EncodeToString([]byte(blockExplorer[ind][`hash`]))
		}
		pageData.List = blockExplorer
	}
	return proceedTemplate( c, NBlockExplorer, &pageData )
}
