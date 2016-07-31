package controllers

import (
	"errors"
	"github.com/DayLightProject/go-daylight/packages/consts"
	"github.com/DayLightProject/go-daylight/packages/utils"
)

func (c *Controller) UpdateDcoin() (string, error) {
	if community, err := c.DCDB.GetCommunityUsers(); err!=nil || (len(community) > 0 &&  !c.NodeAdmin) {
		return "", utils.ErrInfo(errors.New("Permission denied"))
	}	
/*	if c.SessRestricted != 0 || !c.NodeAdmin {
		return "", utils.ErrInfo(errors.New("Permission denied"))
	}*/

	_, url, err := utils.GetUpdVerAndUrl(consts.UPD_AND_VER_URL)
	if err != nil {
		return "", utils.ErrInfo(err)
	}

	if len(url) > 0 {
		err = utils.DcoinUpd(url)
		if err != nil {
			return "", utils.ErrInfo(err)
		}
		return utils.JsonAnswer("success", "success").String(), nil
	}
	return "", nil
}
