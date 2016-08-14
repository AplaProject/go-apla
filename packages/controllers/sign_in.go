package controllers

import (
	"github.com/DayLightProject/go-daylight/packages/utils"
)

func (c *Controller) SignIn() (string, error) {

	c.r.ParseForm()
	n := []byte(c.r.FormValue("n"))
	e := []byte(c.r.FormValue("e"))
	setupPassword := c.r.FormValue("setup_password")

	if !utils.CheckInputData(n, "hex") {
		log.Error("incorrect n %v", n)
		return `{"result":"incorrect n"}`, nil
	}
	if !utils.CheckInputData(e, "hex") {
		log.Error("incorrect e %v", e)
		return `{"result":"incorrect e"}`, nil
	}

	log.Debug("n %s", n)
	log.Debug("e %s", e)
	log.Debug("c.r.RemoteAddr %s", c.r.RemoteAddr)
	log.Debug("c.r.Header.Get(User-Agent) %s", c.r.Header.Get("User-Agent"))

	// проверим, верный ли установочный пароль, если он, конечно, есть
	setupPassword_, err := c.Single("SELECT setup_password FROM config").String()
	if err != nil {
		log.Error("err %v", err)
		return "{\"result\":0}", err
	}
	if len(setupPassword_) > 0 && setupPassword_ != string(utils.DSha256(setupPassword)) {
		log.Error(setupPassword_, string(utils.DSha256(setupPassword)), setupPassword)
		return "{\"result\":0}", nil
	}

	publicKey := utils.MakeAsn1(n, e)
	log.Debug("new key", string(publicKey))
	walletId, err := c.GetWalletIdByPublicKey(publicKey)
	if err != nil {
		log.Error("err %v", err)
		return "{\"result\":0}", err
	}
	if walletId > 0 {
		err = c.ExecSql("UPDATE config SET dlt_wallet_id = ?", walletId)
		if err != nil {
			log.Error("err %v", err)
			return "{\"result\":0}", err
		}
		c.sess.Set("wallet_id", walletId)
	} else {
		citizenId, err := c.GetCitizenIdByPublicKey(publicKey)
		if err != nil {
			log.Error("err %v", err)
			return "{\"result\":0}", err
		}
		err = c.ExecSql("UPDATE config SET citizen_id = ?", citizenId)
		if err != nil {
			log.Error("err %v", err)
			return "{\"result\":0}", err
		}
		c.sess.Set("citizen_id", citizenId)
	}

	return "{\"result\":1}", nil
}
