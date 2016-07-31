// exchange_user
package controllers

import (
	"errors"
//	"strings"
	"github.com/DayLightProject/go-daylight/packages/utils"
//	"fmt"
)

type ExchangeUserPage struct {
	UserId     int64
	Lang       map[string]string
	List       map[string][]map[string]string
	Headers    map[string][]string
	UserInfo   int64
}

func (c *Controller) ExchangeUser() (string, error) {

	var ( err error
		list map[string][]map[string]string
		userInfo int64
	)
	if !c.NodeAdmin || c.SessRestricted != 0 {
		return "", utils.ErrInfo(errors.New("Permission denied"))
	}
	list = make(map[string][]map[string]string)
	headers := make(map[string][]string)
	if iduser, ok := c.Parameters[`iduser`]; ok {
		userInfo = utils.StrToInt64(iduser)
		tables := []string{`e_adding_funds`, `e_adding_funds_cp`, `e_adding_funds_payeer`, `e_adding_funds_pm`,
		             `e_orders`, `e_tokens`, `e_trade`, `e_wallets`, `e_withdraw`}
		for _,tbl := range tables {
			sort := `id`
			if tbl == `e_wallets` {
				sort = `user_id`
			}
			tableList, err := c.GetAll(`select * from `+tbl+` where user_id=? order by `+sort +` desc`, 20, userInfo )
			if err != nil {
				return "", utils.ErrInfo(err)
			}
			if len(tableList) > 0 {
				showcol := `show columns from ` + tbl
				if c.ConfigIni[`db_type`] == `sqlite` {
					showcol = `pragma table_info('`+tbl+`')`
				}
				headList, err := c.GetAll(showcol,-1)
				if err != nil {
					return "", utils.ErrInfo(err)
				}
				head := make([]string, 0)
				for _,hval := range headList {
					if c.ConfigIni[`db_type`] == `sqlite` {
						head = append(head, hval[`name`])
					} else if field, ok := hval[`Field`]; ok {
						head = append(head, field)
					} else {
						head = append(head, hval[`field`])
					}
				}
				headers[tbl] = head
				list[tbl] = tableList
			}
		}
	}
	TemplateStr, err := makeTemplate("exchange_user", "exchangeUser", &ExchangeUserPage{
		Lang:     c.Lang,
		List:     list,
		Headers:  headers,
		UserInfo: userInfo,
		UserId:   c.SessUserId})
	if err != nil {
		return "", utils.ErrInfo(err)
	}
	return TemplateStr, nil
}
