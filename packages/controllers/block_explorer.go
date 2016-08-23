package controllers

import (
	"github.com/DayLightProject/go-daylight/packages/utils"
	"encoding/hex"
)

const NBlockExplorer = `block_explorer`

type blockExplorerPage struct {
	CommonPage
	List   []map[string]string
}

func init() {
	newPage(NBlockExplorer)
}

func (c *Controller) BlockExplorer() (string, error) {

	blockExplorer,err := c.GetAll("SELECT hash, cb_id, wallet_id, time, tx, id FROM block_chain order by id desc limit 0, 30",-1)
	if err != nil {
		return "", utils.ErrInfo(err)
	}
	for ind := range blockExplorer {
		blockExplorer[ind][`hash`] = hex.EncodeToString([]byte(blockExplorer[ind][`hash`]))
	}
		
	return proceedTemplate( c, NBlockExplorer, &blockExplorerPage{
//		CommonPage{`Test`},
		List:    blockExplorer,
	})
}
