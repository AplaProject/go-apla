package controllers

import (
	"strings"

	"github.com/DayLightProject/go-daylight/packages/utils"
)

const ACitizenRequest = `ajax_citizen_request`

type CitizenRequestJson struct {
	Host     string `json:"host"`
	Time     int64  `json:"time"`
	TypeName string `json:"type_name"`
	TypeId   int64  `json:"type_id"`
	Error    string `json:"error"`
}

func init() {
	newPage(ACitizenRequest, `json`)
}

func (c *Controller) AjaxCitizenRequest() interface{} {
	var (
		result CitizenRequestJson
		err    error
		host   string
	)

	stateCode := utils.StrToInt64(c.r.FormValue(`state_id`))
	statePrefix, err := c.GetStatePrefix(stateCode)
	if err == nil {
		request, err := c.Single(`SELECT block_id FROM `+statePrefix+`_citizenship_requests where dlt_wallet_id=?`, c.SessWalletId).Int64()
		if err == nil {
			if request > 0 {
				var state map[string]string
				state, err = c.OneRow(`select * from states where state_id=?`, stateCode).String()
				if len(state[`host`]) == 0 {
					if walletId := utils.StrToInt64(state[`delegate_wallet_id`]); walletId > 0 {
						host, _ = c.Single(`select host from dlt_wallets where wallet_id=?`, walletId).String()
					}
					if len(host) == 0 {
						if stateId := utils.StrToInt64(state[`delegate_state_id`]); stateId > 0 {
							host, err = c.Single(`select host from states where state_id=?`, stateId).String()
						}
					}
				}
				result.Time = utils.Time()
				if len(host) > 0 {
					if !strings.HasPrefix(host, `http`) {
						host = `http://` + host
					}
					if !strings.HasSuffix(host, `/`) {
						host += `/`
					}
					result.TypeName = `NewCitizen`
					result.TypeId = utils.TypeInt(result.TypeName)
				}
				result.Host = host
			}
		} else {
			result.Error = err.Error()
		}
	}
	if err != nil {
		result.Error = err.Error()
	}
	return result
}
