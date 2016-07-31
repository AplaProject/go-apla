package controllers

import (
	"github.com/DayLightProject/go-daylight/packages/utils"
	"strings"
//	"fmt"
/*	"encoding/base64"
	"sort"
	"time"
	"math"*/
)

type eMySupportPage struct {
	Lang             map[string]string
	UserId           int64
	Email            string
	List             []map[string]string
	Topic            string
	IdRoot           int64
}

func (c *Controller) EMySupport() (string, error) {
	var ( err error
		topic string
		list []map[string]string
		idRoot int64
	)
	if c.SessUserId == 0 {
		return `<script language="javascript"> window.location.href = "` + c.EURL + `"</script>If you are not redirected automatically, follow the <a href="` + c.EURL + `">` + c.EURL + `</a>`, nil
	}
	email,err := c.Single(`select email from e_users where id=?`, c.SessUserId ).String()
	if err != nil {
		return "", utils.ErrInfo(err)
	}
	
	list = make([]map[string]string, 0)
	if idroot, ok := c.Parameters[`idroot`]; ok {
		first, err := c.OneRow( `select * from e_tickets where id=? and user_id=?`, idroot, c.SessUserId ).String()
		if err != nil {
			return "", utils.ErrInfo(err)
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
		user_id := utils.StrToInt64(latest[`user_id`])
		status := utils.StrToInt64(latest[`status`])
		if ( user_id != c.SessUserId || status & 2 != 0 ) && status & 1 != 0 {
			c.ExecSql(`update e_tickets set status = status & ~1 where id=?`, latest[`id`] )
		}
	} else {
		list, err = c.GetAll( `select id, subject, user_id, uptime, status,
		(select count(id) from e_tickets where idroot=e.id) as count
		from e_tickets as e where e.idroot=0 and e.user_id=? order by uptime desc`, 20, c.SessUserId )

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
			if ( user_id != c.SessUserId || status & 2 != 0 ) && status & 1 != 0 {
				list[key][`toread`] = `1`
			} else {
				list[key][`toread`] = ``
			}
		}
		if user_id == c.SessUserId {
			list[key][`author`] = ``
		} else {
			list[key][`author`] = `admin`
		}
	} 
	TemplateStr, err := makeTemplate("e_my_support", "eMySupport", &eMySupportPage{
		Lang:             c.Lang,
		UserId:           c.SessUserId,
		Email:            email,
		List:             list,
		Topic:            topic,
		IdRoot:           idRoot,
//		Collapse:         collapse,
	})
	if err != nil {
		return "", utils.ErrInfo(err)
	}
	return TemplateStr, nil
}
