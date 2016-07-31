package controllers

import (
	"encoding/base64"
	"errors"
	"fmt"
	"github.com/DayLightProject/go-daylight/packages/utils"
	"strings"
)

func (c *Controller) EGatePayeer() (string, error) {

	c.r.ParseForm()

	fmt.Println(c.r.Form)

	if utils.IPwoPort(c.r.RemoteAddr) != "37.59.221.23" {
		return "", errors.New("Incorrect RemoteAddr " + utils.IPwoPort(c.r.RemoteAddr))
	}

	if len(c.r.FormValue("m_operation_id")) > 0 && len(c.r.FormValue("m_sign")) > 0 {
		sign := strings.ToUpper(string(utils.Sha256(c.r.FormValue("m_operation_id") + ":" + c.r.FormValue("m_operation_ps") + ":" + c.r.FormValue("m_operation_date") + ":" + c.r.FormValue("m_operation_pay_date") + ":" + c.r.FormValue("m_shop") + ":" + c.r.FormValue("m_orderid") + ":" + c.r.FormValue("m_amount") + ":" + c.r.FormValue("m_curr") + ":" + base64.StdEncoding.EncodeToString([]byte(c.r.FormValue("m_desc"))) + ":" + c.r.FormValue("m_status") + ":" + c.EConfig["payeer_s_key"])))
		if c.r.FormValue("m_sign") == sign && c.r.FormValue("m_status") == "success" {

			txTime := utils.Time()

			currencyId := int64(0)

			if c.r.FormValue("m_curr") == "USD" {
				currencyId = 1001
			}
			if currencyId == 0 {
				return c.r.FormValue("m_orderid") + "|success", nil
			}

			amount := utils.StrToFloat64(c.r.FormValue("m_amount"))
			pmId := utils.StrToInt64(c.r.FormValue("m_operation_id"))
			// проверим, не зачисляли ли мы уже это платеж
			existsId, err := c.Single(`SELECT id FROM e_adding_funds_payeer WHERE id = ?`, pmId).Int64()
			if err != nil {
				return c.r.FormValue("m_orderid") + "|success", nil
			}
			if existsId != 0 {
				return c.r.FormValue("m_orderid") + "|success", nil
			}
			paymentInfo := c.r.FormValue("m_desc")

			EPayment(paymentInfo, currencyId, txTime, amount, pmId, "payeer", c.ECommission)
			return c.r.FormValue("m_orderid") + "|success", nil
		}
	}
	return c.r.FormValue("m_orderid") + "|error", nil

}
