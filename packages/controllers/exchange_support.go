package controllers

import (
	"errors"
	"strings"
	"github.com/DayLightProject/go-daylight/packages/utils"
	"fmt"
)

type ExchangeSupportPage struct {
	UserId   int64
	ToUserId   int64
	Lang     map[string]string
	List             []map[string]string
	Topic            string
	IdRoot           int64
}

func (c *Controller) ExchangeSupport() (string, error) {

	var ( err error
		topic string
		list []map[string]string
		idRoot, toUserId int64
	)
	if !c.NodeAdmin || c.SessRestricted != 0 {
		return "", utils.ErrInfo(errors.New("Permission denied"))
	}

	list = make([]map[string]string, 0)
	if idroot, ok := c.Parameters[`idroot`]; ok {
		first, err := c.OneRow( `select * from e_tickets where id=?`, idroot).String()
		if err != nil {
			return "", utils.ErrInfo(first)
		}
		if len(first) > 0 {
			topic = first[`subject`]
			idRoot = utils.StrToInt64( idroot )
			answers, err := c.GetAll( `select * from e_tickets where idroot=? order by id`, -1, idroot)
			if err != nil {
				return "", utils.ErrInfo(err)
			}
			list = append(list, first)
			list = append(list, answers... )
		}
		latest := list[len(list)-1]
		toUserId = utils.StrToInt64(list[0][`user_id`])
		user_id := utils.StrToInt64(latest[`user_id`])
		status := utils.StrToInt64(latest[`status`])
		if user_id > 0 && status & 2 == 0 && status & 1 == 1 {
			c.ExecSql(`update e_tickets set status = status & ~1 where id=?`, latest[`id`] )
		}
	} else {
		list, err = c.GetAll( `select id, subject, user_id, uptime, status,
		(select count(id) from e_tickets where idroot=e.id) as count
		from e_tickets as e where e.idroot=0 order by uptime desc`, 50 )

		if err != nil {
			return "", utils.ErrInfo(err)
		}
	}
	for key, val := range list {
		user_id := utils.StrToInt64( val[`user_id`] )
		if idRoot > 0 {
			list[key][`topic`] = strings.Replace(val[`topic`], "\n", "<br>", -1)
		} else {
			status := utils.StrToInt64( val[`status`])
			if list[key][`count`] != `0` {
				latest,_ := c.OneRow(`select user_id, status from e_tickets where idroot=? order by id desc`, val[`id`]).Int64()
				user_id = latest[`user_id`]
				status = latest[`status`]
			}
			if user_id > 0 && status & 2 == 0 && status & 1 == 1 {
				list[key][`toread`] = `1`
			} else {
				list[key][`toread`] = ``
			}

		}
		if user_id == 0 {
			list[key][`author`] = ``
		} else {
			list[key][`author`] = fmt.Sprintf( `ID %d`, user_id )
		}
	} 

	TemplateStr, err := makeTemplate("exchange_support", "exchangeSupport", &ExchangeSupportPage{
		Lang:     c.Lang,
		List:             list,
		Topic:            topic,
		IdRoot:           idRoot,
		ToUserId:         toUserId,
		UserId:   c.SessUserId})
	if err != nil {
		return "", utils.ErrInfo(err)
	}
	return TemplateStr, nil
}
