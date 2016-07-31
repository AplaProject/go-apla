package controllers

import (
	"crypto/rand"
	"crypto/rsa"
	"errors"
	"github.com/DayLightProject/go-daylight/packages/utils"
	"text/template"
)

func (c *Controller) SaveDecryptComment() (string, error) {

	if c.SessRestricted != 0 {
		return "", utils.ErrInfo(errors.New("Permission denied"))
	}

	c.r.ParseForm()
	commentType := c.r.FormValue("type")
	id := utils.StrToInt64(c.r.FormValue("id"))
	comment := c.r.FormValue("comment")
	if !utils.InSliceString(commentType, []string{"chat", "dc_transactions", "arbitrator", "seller", "cash_requests", "comments"}) {
		return "", utils.ErrInfo(errors.New("incorrect type"))
	}

	// == если мы майнер и это dc_transactions, то сюда прислан зашифрованный коммент, который можно расшифровать только нод-кдючем
	minerId, err := c.GetMinerId(c.SessUserId)
	if err != nil {
		return "", utils.ErrInfo(err)
	}
	if minerId > 0 && utils.InSliceString(commentType, []string{"dc_transactions", "arbitrator", "seller"}) {
		nodePrivateKey, err := c.GetNodePrivateKey(c.MyPrefix)
		if len(nodePrivateKey) == 0 {
			return `***`, nil
		}
/*		if commentType == `dc_transactions` {
			idBlock,err := c.Single(`select block_id from `+c.MyPrefix+`my_dc_transactions where id=?`, id).Int64()
			if err != nil {
				return "", utils.ErrInfo(err)
			}
			nodeBlock,err := c.Single(`select block_id from `+c.MyPrefix+`my_node_keys order by id desc`).Int64()
			if err != nil {
				return "", utils.ErrInfo(err)
			}
			if idBlock <= nodeBlock {
				return `****`, nil
			}
			nodePublic, err := c.GetNodePublicKey(c.SessUserId)
			if err != nil {
				return "", utils.ErrInfo(err)
			}
			myPublic, err := c.GetMyNodePublicKey(c.MyPrefix)
			if err != nil {
				return "", utils.ErrInfo(err)
			}
			if myPublic != string(nodePublic) {
				return `*****`, nil
			}
		}*/
		// расшифруем коммент
		rsaPrivateKey, err := utils.MakePrivateKey(nodePrivateKey)
		if err != nil {
			return "", utils.ErrInfo(err)
		}
		comment_, err := rsa.DecryptPKCS1v15(rand.Reader, rsaPrivateKey, utils.HexToBin([]byte(comment)))
		if err != nil {
			comment = c.Lang[`decrypt_impossible`]
		} else {
			comment = string(comment_)
		}
	}
	comment = template.HTMLEscapeString(comment)
	if len(comment) > 0 {
		if utils.InSliceString(commentType, []string{"arbitrator", "seller"}) {
			err = c.ExecSql(`
				UPDATE `+c.MyPrefix+`my_comments
				SET comment = ?,
					comment_status = ?
				WHERE id = ? AND type = ?`, comment, "decrypted", id, commentType)
			if err != nil {
				return "", utils.ErrInfo(err)
			}
		} else if commentType == "chat" {
			err = c.ExecSql(`
				UPDATE chat
				SET enc_message = message,
					message = ?,
					status = ?
				WHERE id = ? AND receiver = ?`, comment, 2, id, c.SessUserId)
			if err != nil {
				return "", utils.ErrInfo(err)
			}
		} else {
			err = c.ExecSql(`
				UPDATE `+c.MyPrefix+`my_`+commentType+`
				SET comment = ?,
					comment_status = 'decrypted'
				WHERE id = ?`, comment, id)
			if err != nil {
				return "", utils.ErrInfo(err)
			}
		}
	} else {
		comment = "NULL"
	}
	return comment, nil
}
