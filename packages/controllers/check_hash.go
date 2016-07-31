// check_hash
package controllers

import (
	"github.com/DayLightProject/go-daylight/packages/utils"
	"encoding/json"
	"fmt"
)

func (c *Controller) CheckHash() (string, error) {
	values := make( map[string]string )
	blockId := utils.StrToInt64( c.r.FormValue(`block_id`))
	curId, err := c.GetBlockId()
	var hashes []map[string]string
	if err == nil {
		if curId >= blockId {
			hashes,err = c.GetAll(`select id,hash from block_chain where id >=? order by id desc`, 
			                        5, blockId-4 )
			if curId > blockId + 4 {
				if last,err := c.GetAll(`select id,hash from block_chain order by id desc`, 5 ); err == nil {
					for key, val := range last {
    					hashes[key] = val
					}	
				}
			}
		} else {
			hashes,err = c.GetAll(`select id,hash from block_chain order by id desc`, 5 )
		}
		if err == nil && hashes != nil {
			for _, val := range hashes {
				values[val[`id`]] = fmt.Sprintf("%x", val[`hash`])
			}	
		}
	}
	res, err := json.Marshal( values )
	return string(res),err
}
