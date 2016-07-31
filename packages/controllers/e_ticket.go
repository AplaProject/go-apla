package controllers

import (
	"errors"
	"github.com/DayLightProject/go-daylight/packages/utils"
	"strings"
	"html"
)

func (c *Controller) ETicket() (string, error) {

	if c.SessUserId == 0 {
		return c.Lang["sign_up_please"], nil
	}

	c.r.ParseForm()
	userId := c.SessUserId
	subject := html.EscapeString(c.r.FormValue("subject"))
	topic := html.EscapeString(c.r.FormValue("topic"))
	email := html.EscapeString(c.r.FormValue("email"))
	idroot := utils.StrToInt64(c.r.FormValue("idroot"))
	userid := utils.StrToInt64(c.r.FormValue("userid"))
	status := 1   // not read

	if userid > 0 && (!c.NodeAdmin || c.SessRestricted != 0) {
		return ``, utils.ErrInfo(errors.New("Permission denied"))
	}
	if userid > 0 {
		exist, err := c.Single(`select id from e_users where id=?`, userid).Int64()
		if exist == 0 || err != nil {
			return ``, utils.ErrInfo(errors.New("Unknown User Id"))
		}
		if idroot == 0 {
			status |= 2  // From admin
			userId = userid
		} else {
			userId = 0
		}
	} else if idroot == 0 {
		if err := c.ExecSql(`update e_users set email=? where id=?`, email, userId ); err != nil {
			return ``, utils.ErrInfo(err)
		}
	}
	
	now := utils.Time()
	err := c.ExecSql(`insert into e_tickets (user_id, subject, topic, idroot, time, status, uptime) 
	                 values(?,?,?,?,?,?,?)`, userId, subject, topic, idroot, now, status, now )
	if err == nil && idroot>0 {
		err = c.ExecSql(`update e_tickets set uptime=? where id=?`, now, idroot )
	}			
	if err!=nil {
		return ``, utils.ErrInfo(err)
	}
	var cmd uint
	email = ``
	if userid > 0 {  // Сообщение от админа
		email,_ = c.Single(`select email from e_users where id=?`, userid).String()
		cmd = utils.ECMD_EXANSWER
	} else { // Сообщение от юзера
		email,_ = c.Single(`select value from e_config where name=?`, `email`).String()
		cmd = utils.ECMD_EXREQUEST
	}
	if len(email) > 0 {
		catalog,_ := c.Single(`select value from e_config where name=?`, `catalog`).String()
		catalog = strings.Replace(catalog, "/", "", -1)
		utils.ExchangeEmail(email, c.r.Host + "/" + catalog + "/", cmd )
	}	
	return `1`, nil
}
