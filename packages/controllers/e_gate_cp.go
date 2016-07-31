package controllers

import (
	"fmt"
	"strings"
	b64 "encoding/base64"
	"errors"
	"github.com/DayLightProject/go-daylight/packages/utils"
)

func (c *Controller) EGateCP() (string, error) {

	c.r.ParseForm()

	fmt.Println(c.r.Form)
	log.Error("EGateCP %v", c.r.Form)

	fmt.Println(c.r.Header.Get("HTTP_HMAC"))
	log.Error("HTTP_HMAC %v", c.r.Header.Get("HTTP_HMAC"))

	fmt.Println(c.r.Header.Get("Authorization"))
	log.Error("Authorization %v", c.r.Header.Get("Authorization"))

	sEnc := strings.Split(c.r.Header.Get("Authorization"), " ")
	log.Error("sEnc %v", sEnc[0])

	if len(sEnc) > 1 {
		sDec, _ := b64.StdEncoding.DecodeString(sEnc[1])
		sEnc0 := strings.Split(string(sDec), ":")
		if len(sEnc0) > 1 {
			if sEnc0[0] != c.EConfig["cp_id"] || sEnc0[1]!= c.EConfig["cp_s_key"] {
				log.Error("incorrect cp_id cp_s_key %v %v %v %v ", sEnc0[0], c.EConfig["cp_id"], sEnc0[1], c.EConfig["cp_s_key"])
				return "", errors.New("cp_id cp_s_key")
			}
		} else {
			log.Error("incorrect cp_id cp_s_key %v", sEnc0)
			return "", errors.New("cp_id cp_s_key")
		}
	} else {
		log.Error("incorrect cp_id cp_s_key")
		return "", errors.New("cp_id cp_s_key")
	}

	var currencyId int64
	if c.r.FormValue("currency1") == "BTC" {
		currencyId = 1002
	}
	if currencyId == 0 {
		log.Error("Incorrect currencyId")
		return "", errors.New("Incorrect currencyId")
	}

	amount := utils.StrToFloat64(c.r.FormValue("amount1"))
	log.Error("amount %v", amount)
	log.Error("FormValue %v", c.r.FormValue("amount1"))
	pmId := utils.StrToInt64(c.r.FormValue("txn_id"))
	// проверим, не зачисляли ли мы уже это платеж
	existsId, err := c.Single(`SELECT id FROM e_adding_funds_cp WHERE id = ?`, pmId).Int64()
	if err != nil {
		return "", utils.ErrInfo(err)
	}
	if existsId != 0 {
		log.Error("Incorrect txn_id")
		return "", errors.New("Incorrect txn_id")
	}
	paymentInfo := c.r.FormValue("item_name")

	status:=utils.StrToInt64(c.r.FormValue("status"))
	if status >= 100 || status == 2 {
		txTime := utils.Time()
		err = EPayment(paymentInfo, currencyId, txTime, amount, pmId, "cp", c.ECommission)
		if err != nil {
			return "", utils.ErrInfo(err)
		}
	} else {
		log.Error("status %v", status)
		return "", errors.New("Incorrect txn_id")
	}

	return ``, nil
}
