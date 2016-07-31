package controllers

import (
	"github.com/DayLightProject/go-daylight/packages/utils"
	"time"
)

type arbitrationSellerPage struct {
	SignData        string
	ShowSignData    bool
	TxType          string
	TxTypeId        int64
	TimeNow         int64
	UserId          int64
	Alert           string
	Lang            map[string]string
	CountSignArr    []int
	LastTxFormatted string
	CurrencyList    map[int64]string
	MinerId         int64
	PendingTx       int64
	MyOrders        []map[string]string
	SessRestricted  int64
	ShopData        map[string]string
	HoldBack        map[string]string
}

func (c *Controller) ArbitrationSeller() (string, error) {

	log.Debug("ArbitrationSeller")

	txType := "ChangeSellerHoldBack"
	txTypeId := utils.TypeInt(txType)
	timeNow := time.Now().Unix()

	holdBack, err := c.OneRow("SELECT arbitration_days_refund, seller_hold_back_pct FROM users WHERE user_id  =  ?", c.SessUserId).String()
	if err != nil {
		return "", utils.ErrInfo(err)
	}

	myOrders, err := c.GetAll("SELECT id, time, amount, seller, status, comment_status, comment	FROM orders	WHERE seller = ? ORDER BY time DESC LIMIT 20", 20, c.SessUserId)
	for k, data := range myOrders {
		if data["status"] == "refund" {
			if c.SessRestricted == 0 {
				data_, err := c.OneRow("SELECT comment, comment_status FROM "+c.MyPrefix+"my_comments WHERE id  =  ? AND type  =  'seller'", data["id"]).String()
				if err != nil {
					return "", utils.ErrInfo(err)
				}
				data["status"] = data_["comment"]
				data["comment_status"] = data_["comment_status"]
			}
		}
		myOrders[k] = data
	}

	var shopData map[string]string
	if c.SessRestricted == 0 {
		shopData, err = c.OneRow("SELECT shop_secret_key, shop_callback_url FROM " + c.MyPrefix + "my_table").String()
		if err != nil {
			return "", utils.ErrInfo(err)
		}
	}

	last_tx, err := c.GetLastTx(c.SessUserId, utils.TypesToIds([]string{"ChangeSellerHoldBack", "MoneyBack"}), 3, c.TimeFormat)
	lastTxFormatted := ""
	var pendingTx_ map[int64]int64
	if len(last_tx) > 0 {
		lastTxFormatted, pendingTx_ = utils.MakeLastTx(last_tx, c.Lang)
	}
	pendingTx := pendingTx_[txTypeId]

	TemplateStr, err := makeTemplate("arbitration_seller", "arbitrationSeller", &arbitrationSellerPage{
		Alert:           c.Alert,
		Lang:            c.Lang,
		CountSignArr:    c.CountSignArr,
		ShowSignData:    c.ShowSignData,
		UserId:          c.SessUserId,
		TimeNow:         timeNow,
		TxType:          txType,
		TxTypeId:        txTypeId,
		SignData:        "",
		LastTxFormatted: lastTxFormatted,
		CurrencyList:    c.CurrencyList,
		MinerId:         c.MinerId,
		PendingTx:       pendingTx,
		MyOrders:        myOrders,
		SessRestricted:  c.SessRestricted,
		ShopData:        shopData,
		HoldBack:        holdBack})
	if err != nil {
		return "", utils.ErrInfo(err)
	}
	return TemplateStr, nil
}
