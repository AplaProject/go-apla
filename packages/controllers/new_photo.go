package controllers

import (
	"encoding/json"
	"github.com/DayLightProject/go-daylight/packages/utils"
)

func (c *Controller) NewPhoto() (string, error) {

	c.r.ParseForm()

	userId := int64(utils.StrToFloat64(c.r.FormValue("user_id")))

	data, err := c.OneRow("SELECT photo_block_id, photo_max_miner_id, miners_keepers FROM miners_data WHERE user_id = ?", userId).String()
	if err != nil {
		return "", utils.ErrInfo(err)
	}

	// получим ID майнеров, у которых лежат фото нужного нам юзера
	minersIds := utils.GetMinersKeepers(data["photo_block_id"], data["photo_max_miner_id"], data["miners_keepers"], true)

	// берем 1 случайный из 10-и ID майнеров
	k := utils.RandInt(0, len(minersIds))
	minerId := minersIds[k]
	host, err := c.Single("SELECT http_host FROM miners_data WHERE miner_id  =  ?", minerId).String()
	if err != nil {
		return "", utils.ErrInfo(err)
	}
	if host == `0` {
		host = `http://pool.dcoin.club/`
	} 
	result, err := json.Marshal(map[string]string{"face": host + "public/face_" + utils.Int64ToStr(userId) + ".jpg", "profile": host + "public/profile_" + utils.Int64ToStr(userId) + ".jpg"})
	if err != nil {
		return "", utils.ErrInfo(err)
	}
	return string(result), nil
}
