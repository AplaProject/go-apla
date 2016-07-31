package controllers

import (

)

func (c *Controller) EGateIk() (string, error) {
/*
	c.r.ParseForm()
	fmt.Println(c.r.Form)
	var ikNames []string
	for name, _ := range c.r.Form {
		if name[:2] == "ik" && name != "ik_sign" {
			ikNames = append(ikNames, name)
		}
	}
	sort.Strings(ikNames)
	fmt.Println(ikNames)

	var ikValues []string
	for _, names := range ikNames {
		ikValues = append(ikValues, c.r.FormValue(names))
	}
	ikValues = append(ikValues, c.EConfig["ik_s_key"])
	fmt.Println(ikValues)
	sign := strings.Join(ikValues, ":")
	fmt.Println(sign)
	sign = base64.StdEncoding.EncodeToString(utils.HexToBin(utils.Md5(sign)))
	fmt.Println(sign)
	if sign != c.r.FormValue("ik_sign") {
		return "", errors.New("Incorrect signature")
	}
	currencyId := int64(0)

	if c.r.FormValue("ik_cur") == "USD" {
		currencyId = 1001
	}
	if currencyId == 0 {
		return "", errors.New("Incorrect currencyId")
	}

	amount := utils.StrToFloat64(c.r.FormValue("ik_am"))
	pmId := utils.StrToInt64(c.r.FormValue("ik_inv_id"))
	// проверим, не зачисляли ли мы уже это платеж
	existsId, err := c.Single(`SELECT id FROM e_adding_funds_ik WHERE id = ?`, pmId).Int64()
	if err != nil {
		return "", utils.ErrInfo(err)
	}
	if existsId != 0 {
		return "", errors.New("Incorrect ik_inv_id")
	}
	paymentInfo := c.r.FormValue("ik_desc")

	txTime := utils.Time()
	err = EPayment(paymentInfo, currencyId, txTime, amount, pmId, "ik", c.ECommission)
	if err != nil {
		return "", utils.ErrInfo(err)
	}
*/
	return ``, nil
}
