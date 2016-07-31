package controllers

import (
	"github.com/DayLightProject/go-daylight/packages/availablekey"
	"github.com/DayLightProject/go-daylight/packages/utils"
)

type availableKeysPage struct {
	AutoLogin bool
	Key       string
	LangId    int
}

/*
func checkAvailableKey(key string, db *utils.DCDB) (int64, string, error) {
	publicKeyAsn, err := utils.GetPublicFromPrivate(key)
	if err != nil {
		log.Debug("%v", err)
		return 0, "", utils.ErrInfo(err)
	}
	log.Debug("publicKeyAsn: %s", publicKeyAsn)
	userId, err := db.Single("SELECT user_id FROM users WHERE hex(public_key_0) = ?", publicKeyAsn).Int64()
	if err != nil {
		return 0, "", utils.ErrInfo(err)
	}
	log.Debug("userId: %s", userId)
	if userId == 0 {
		return 0, "", errors.New("null userId")
	}
	allTables, err := db.GetAllTables()
	if err != nil {
		return 0, "", utils.ErrInfo(err)
	}
	// может другой юзер уже начал смену ключа. актуально для пула
	if utils.InSliceString(utils.Int64ToStr(userId)+"_my_table", allTables) {
		return 0, "", errors.New("exists _my_table")
	}
	return userId, string(publicKeyAsn), nil
}*/

func (c *Controller) AvailableKeys() (string, error) {

	var email string
	if c.Community {
		// если это пул, то будет прислан email
		email = c.r.FormValue("email")
		if !utils.ValidateEmail(email) {
			return utils.JsonAnswer("Incorrect email", "error").String(), nil
		}
		// если мест в пуле нет, то просто запишем юзера в очередь
		pool_max_users, err := c.Single("SELECT pool_max_users FROM config").Int()
		if err != nil {
			return "", utils.JsonAnswer(utils.ErrInfo(err), "error").Error()
		}
		if len(c.CommunityUsers) >= pool_max_users {
			err = c.ExecSql("INSERT INTO pool_waiting_list ( email, time, user_id ) VALUES ( ?, ?, ? )", email, utils.Time(), 0)
			if err != nil {
				return "", utils.JsonAnswer(utils.ErrInfo(err), "error").Error()
			}
			return utils.JsonAnswer(c.Lang["pool_is_full"], "error").String(), nil
		}
	}

	availablekey := &availablekey.AvailablekeyStruct{}
	availablekey.DCDB = c.DCDB
	availablekey.Email = email
	userId, publicKey, err := availablekey.GetAvailableKey()
	if err != nil {
		return "", utils.JsonAnswer(utils.ErrInfo(err), "error").Error()
	}
	if userId > 0 {
		c.sess.Set("user_id", userId)
		c.sess.Set("public_key", publicKey)
		log.Debug("user_id: %d", userId)
		log.Debug("public_key: %s", publicKey)
		return utils.JsonAnswer("success", "success").String(), nil
	}
	return utils.JsonAnswer("no_available_keys", "error").String(), nil
}
