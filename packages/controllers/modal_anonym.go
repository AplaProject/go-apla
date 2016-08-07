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
	MyWallet			  string
}

func (c *Controller) ModalAnonym() (string, error) {

	walletHash, err := c.Single("SELECT hex(hash) FROM wallets WHERE wallet_id = ?", c.SessWalletId).String()
	if err != nil {
		return "", utils.ErrInfo(err)
	}

	TemplateStr, err := makeTemplate("modal_anonym", "modalAnonym", &modalAnonymPage{
		CountSignArr:          c.CountSignArr,
		CountSign:             c.CountSign,
		Lang:                  c.Lang,
		MyWallet:              walletHash,
		Title:                 "modalAnonym",
		ShowSignData:          c.ShowSignData,
		SignData:              ""})
	if err != nil {
		return "", utils.ErrInfo(err)
	}
	return TemplateStr, nil
}
