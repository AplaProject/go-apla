package controllers

import (
	"errors"
	"github.com/DayLightProject/go-daylight/packages/utils"
)

type ExchangeAdminPage struct {
	EPages   map[string]map[string]string
	Alert    string
	UserId   int64
	Lang     map[string]string
	Withdraw []map[string]string
	Lock     int64
	Users    []map[string]string
}

func (c *Controller) ExchangeAdmin() (string, error) {

	if !c.NodeAdmin || c.SessRestricted != 0 {
		return "", utils.ErrInfo(errors.New("Permission denied"))
	}

	log.Debug("c.Parameters", c.Parameters)

	if _, ok := c.Parameters["e_pages_about_title"]; ok {
		err := c.ExecSql("DELETE FROM e_pages WHERE lang = ?", c.LangInt)
		if err != nil {
			return "", utils.ErrInfo(err)
		}

		params := [][]string{{"about_title", "about"}, {"rules_title", "rules"}, {"faq_title", "faq"}, {"contacts_title", "contacts"}}
		for _, data := range params {
			err = c.ExecSql(`INSERT INTO e_pages (lang, name, title, text) VALUES (?, ?, ?, ?)`, c.LangInt, data[1], c.Parameters["e_pages_"+data[0]], c.Parameters["e_pages_"+data[1]])
			if err != nil {
				return "", utils.ErrInfo(err)
			}
		}
	}

	withdrawId := utils.StrToInt64(c.Parameters["withdraw_id"])
	if withdrawId > 0 {
		err := c.ExecSql(`UPDATE e_withdraw SET close_time = ? WHERE id = ?`, utils.Time(), withdrawId)
		if err != nil {
			return "", utils.ErrInfo(err)
		}
	}

	lock, err := c.Single(`SELECT time FROM e_reduction_lock`).Int64()
	if len(c.Parameters["change_reduction_lock"]) > 0 {
		if lock > 0 {
			err := c.ExecSql(`DELETE FROM e_reduction_lock`)
			if err != nil {
				return "", utils.ErrInfo(err)
			}
			lock = 0
		} else {
			err := c.ExecSql(`INSERT INTO e_reduction_lock (time) VALUES (?)`, utils.Time())
			if err != nil {
				return "", utils.ErrInfo(err)
			}
			lock = utils.Time()
		}

	}

	withdraw, err := c.GetAll(`SELECT e_withdraw.id as e_withdraw_id, open_time, close_time, e_users.user_id, currency_id, account, amount,  wd_amount, method, email
    		FROM e_withdraw
    		LEFT JOIN e_users on e_users.id = e_withdraw.user_id
   			ORDER BY e_withdraw.id DESC`, 100)
	if err != nil {
		return "", utils.ErrInfo(err)
	}

	users, err := c.GetAll(`SELECT id, email, ip, lock, user_id
    		FROM e_users
   			ORDER BY id DESC`, 100)
	if err != nil {
		return "", utils.ErrInfo(err)
	}

	ePages := make(map[string]map[string]string)
	rows, err := c.Query(c.FormatQuery("SELECT name, title, text FROM e_pages WHERE lang = ?"), c.LangInt)
	if err != nil {
		return "", utils.ErrInfo(err)
	}
	defer rows.Close()
	for rows.Next() {
		var name, title, text string
		err = rows.Scan(&name, &title, &text)
		if err != nil {
			return "", utils.ErrInfo(err)
		}
		ePages[name] = map[string]string{"title": title, "text": text}
	}

	TemplateStr, err := makeTemplate("exchange_admin", "exchangeAdmin", &ExchangeAdminPage{
		EPages:   ePages,
		Alert:    c.Alert,
		Lang:     c.Lang,
		Withdraw: withdraw,
		Lock:     lock,
		Users:    users,
		UserId:   c.SessUserId})
	if err != nil {
		return "", utils.ErrInfo(err)
	}
	return TemplateStr, nil
}
