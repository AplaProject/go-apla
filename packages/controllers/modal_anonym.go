package controllers

import (
	"github.com/DayLightProject/go-daylight/packages/utils"
)


type modalAnonymPage struct {
	Lang                  map[string]string
	Title                 string
	CountSign             int
	CountSignArr          []int
	SignData              string
	ShowSignData          bool
	MyWalletData		  map[string]string
	WalletId int64
	CitizenId int64
	TxType       string
	TxTypeId     int64
	TimeNow      int64
}

func (c *Controller) ModalAnonym() (string, error) {

	txType := "DLTChangeHostVote"
	txTypeId := utils.TypeInt(txType)
	timeNow := utils.Time()

	MyWalletData, err := c.OneRow("SELECT hex(address) as address, host, hex(addressVote) as addressVote  FROM dlt_wallets WHERE wallet_id = ?", c.SessWalletId).String()
	if err != nil {
		return "", utils.ErrInfo(err)
	}
	log.Debug("MyWalletData %v", MyWalletData);


	TemplateStr, err := makeTemplate("modal_anonym", "modalAnonym", &modalAnonymPage{
		Lang:                  c.Lang,
		MyWalletData:          MyWalletData,
		Title:                 "modalAnonym",
		ShowSignData:          c.ShowSignData,
		SignData:              "",
		WalletId: c.SessWalletId,
		CitizenId: c.SessCitizenId,
		CountSignArr:          c.CountSignArr,
		CountSign:             c.CountSign,
		TimeNow:      timeNow,
		TxType:       txType,
		TxTypeId:     txTypeId})
	if err != nil {
		return "", utils.ErrInfo(err)
	}
	return TemplateStr, nil
}
