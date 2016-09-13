package controllers

import (
	"math/rand"
	"time"

	"github.com/DayLightProject/go-daylight/packages/utils"
)

const AGetUid = `ajax_get_uid`

type GetUidJson struct {
	Uid   string `json:"uid"`
	Error string `json:"error"`
}

func init() {
	newPage(AGetUid, `json`)
}

func (c *Controller) AjaxGetUid() interface{} {
	var result GetUidJson

	r := rand.New(rand.NewSource(time.Now().Unix()))
	result.Uid = utils.Int64ToStr(r.Int63())
	c.sess.Set("uid", result.Uid)
	return result
}
