package controllers

import (
	"github.com/DayLightProject/go-daylight/packages/utils"
	"fmt"
)

func (c *Controller) SignIn() (string, error) {
	
	ret := `{"result":0}`
	c.r.ParseForm()
	key := c.r.FormValue("key")
	//msg := c.r.FormValue("msg")
	//sign := utils.HexToBin([]byte(c.r.FormValue("sign")))
/*	n := []byte(c.r.FormValue("n"))
	e := []byte(c.r.FormValue("e"))

	if !utils.CheckInputData(n, "hex") {
		log.Error("incorrect n %v", n)
		return `{"result":"incorrect n"}`, nil
	}
	if !utils.CheckInputData(e, "hex") {
		log.Error("incorrect e %v", e)
		return `{"result":"incorrect e"}`, nil
	}

	log.Debug("n %s", n)
	log.Debug("e %s", e)*/
//	fmt.Printf("Signature %d %s\r\n", len(sign), sign)
	/*if verify,_ := utils.CheckECDSA([][]byte{key}, msg, sign, true); !verify {
		return ret, fmt.Errorf("incorrect signature")
	} */
	address := utils.KeyToAddress(key)
	c.sess.Set("address", address)
	log.Debug("c.r.RemoteAddr %s", c.r.RemoteAddr)
	log.Debug("c.r.Header.Get(User-Agent) %s", c.r.Header.Get("User-Agent"))

//	publicKey := utils.MakeAsn1(n, e)
//	publicKey := []byte(utils.HexToBin(key))
//	log.Debug("new key", string(publicKey))
	publicKey := []byte(key)
	walletId, err := c.GetWalletIdByPublicKey(publicKey)
	fmt.Printf("Sign in wallet=%d address=%s\r\n", walletId, address)
	if err != nil {
		log.Error("err %v", err)
		return ret, err
	}
	if walletId > 0 {
		err = c.ExecSql("UPDATE config SET dlt_wallet_id = ?", walletId)
		if err != nil {
			log.Error("err %v", err)
			return ret, err
		}
		c.sess.Set("wallet_id", walletId)
	} else {
		citizenId, err := c.GetCitizenIdByPublicKey(publicKey)
		if err != nil {
			log.Error("err %v", err)
			return ret, err
		}
		err = c.ExecSql("UPDATE config SET citizen_id = ?", citizenId)
		if err != nil {
			log.Error("err %v", err)
			return ret, err
		}
		c.sess.Set("citizen_id", citizenId)
	}
	return `{"result":1}`, nil
}
