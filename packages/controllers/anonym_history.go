package controllers

import (
	"github.com/DayLightProject/go-daylight/packages/utils"
)



func (c *Controller) AnonymHistory() (string, error) {
	var str string
	var err error
		if c.SessWalletId > 0 {
		str, err = c.GetJSON(`SELECT id, hex(recipient_wallet_address) as recipient_wallet_address, amount, time, comment, block_id FROM dlt_transactions WHERE recipient_wallet_id= ? OR sender_wallet_id`, c.SessWalletId, c.SessWalletId);
		if err != nil {
			return "", utils.ErrInfo(err)
		}
	}
	return string(str), nil
}