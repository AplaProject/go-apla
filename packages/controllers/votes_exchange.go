package controllers

import (
	"fmt"
	"github.com/DayLightProject/go-daylight/packages/utils"
	"time"
)

type VotesExchangePage struct {
	Alert        string
	SignData     string
	ShowSignData bool
	TxType       string
	TxTypeId     int64
	TimeNow      int64
	UserId       int64
	Lang         map[string]string
	CountSignArr []int
	EOwner       int64
	Result       int64
}

func (c *Controller) VotesExchange() (string, error) {

	txType := "VotesExchange"
	txTypeId := utils.TypeInt(txType)
	timeNow := time.Now().Unix()

	eOwner := utils.StrToInt64(c.Parameters["e_owner_id"])
	result := utils.StrToInt64(c.Parameters["result"])

	signData := fmt.Sprintf("%d,%d,%d,%d,%d", txTypeId, timeNow, c.SessUserId, eOwner, result)
	if eOwner == 0 {
		signData = ""
	}
	TemplateStr, err := makeTemplate("votes_exchange", "votesExchange", &VotesExchangePage{
		Alert:        c.Alert,
		Lang:         c.Lang,
		CountSignArr: c.CountSignArr,
		ShowSignData: c.ShowSignData,
		UserId:       c.SessUserId,
		TimeNow:      timeNow,
		TxType:       txType,
		TxTypeId:     txTypeId,
		EOwner:       eOwner,
		Result:       result,
		SignData:     signData})
	if err != nil {
		return "", utils.ErrInfo(err)
	}
	return TemplateStr, nil
}
