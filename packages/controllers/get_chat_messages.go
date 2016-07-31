package controllers

import (
	"crypto/rand"
	"crypto/rsa"
	"encoding/json"
	"github.com/DayLightProject/go-daylight/packages/consts"
	"github.com/DayLightProject/go-daylight/packages/utils"
	"strings"
	"text/template"
	//"fmt"
)

var chatIds = make(map[int64][]int)

type Message struct {

}

func (c *Controller) GetChatMessages() (string, error) {
	c.r.ParseForm()
	first := c.r.FormValue("first")

	if first == "1" {
		chatIds[c.SessUserId] = []int{}
	}

	if err := removeOld(c); err != nil {
		return "", err
	}

	chatData, err := getChatData(c)
	if err != nil {
		return "", err
	}

	var result string
	for i := len(chatData) - 1; i >= 0; i-- {
		data := chatData[i]
		status := data["status"]
		message := data["message"]
		receiver := utils.StrToInt64(data["receiver"])
		sender := utils.StrToInt64(data["sender"])
		if status == "1" {
			// Если юзер хранит приватый ключ в БД, то сможем расшифровать прямо тут
			if receiver == c.SessUserId {
				privateKey, err := c.GetMyPrivateKey(c.MyPrefix)
				if err != nil {
					log.Error("%v", utils.ErrInfo(err))
					continue
				}

				if len(privateKey) > 0 {
					decrypted, err := decrypt(privateKey, data["message"], data["id"], c)
					if err != nil || len(decrypted) < 1 {
						log.Error("%v", err)
						continue
					}
					message = string(decrypted)
					status = "2"
				}
			}
		}

		name := data["sender"]
		ava := "/static/img/noavatar.png"
		// возможно у отпарвителя есть ник
		nameAvaBan, err := c.OneRow(`SELECT name, avatar, chat_ban FROM users WHERE user_id = ?`, sender).String()
		if err != nil {
			return "", utils.ErrInfo(err)
		}
		// возможно юзер забанен
		if nameAvaBan["chat_ban"] == "1" {
			continue
		}
		if len(nameAvaBan["name"]) > 0 {
			name = nameAvaBan["name"]
		}

		minerStatus, err := c.Single(`SELECT status FROM miners_data WHERE user_id = ?`, sender).String()
		if err != nil {
			return "", utils.ErrInfo(err)
		}
		if minerStatus == "miner" && len(nameAvaBan["avatar"]) > 0 {
			ava = nameAvaBan["avatar"]
		}

		row := ""
		message = template.HTMLEscapeString(message)
		avaHtml := `<img src="` + ava + `" onclick='setReceiver("` + name + `", "` + data["sender"] + `")'>`
		nameHtml := `<strong><a class="chatNick" onclick='setReceiver("` + name + `", "` + data["sender"] + `")'>` + name + `</a></strong>`
		if status == "2" { // успешно расшифровали
			row = `<tr><td>` + avaHtml + `` + nameHtml + `: <i class="fa fa-lock"></i> ` + message + `</td></tr>`
		} else if status == "1" && receiver == c.SessUserId { // либо нет ключа, либо какая-то ошибка
			row = `<tr><td>` + avaHtml + `` + nameHtml + `: <div id="comment_` + data["id"] + `" style="display: inline-block;"><input type="hidden" value="` + message + `" id="encrypt_comment_` + data["id"] + `"><a class="btn btn-default btn-lg" onclick="decrypt_comment(` + data["id"] + `, 'chat')"> <i class="fa fa-lock"></i> Decrypt</a></div></td></tr>`
		} else if status == "0" {
			row = `<tr><td>` + avaHtml + `` + nameHtml + `: ` + message + `</td></tr>`
		}
		result += row
		chatIds[c.SessUserId] = append(chatIds[c.SessUserId], utils.StrToInt(data["id"]))
		if first == "1" {
			if utils.StrToInt64(data["sign_time"]) < utils.ChatMinSignTime || utils.ChatMinSignTime == 0 {
				utils.ChatMinSignTime = utils.StrToInt64(data["sign_time"])
				log.Debug("utils.ChatMinSignTime", utils.ChatMinSignTime)
			}
		}
	}

	log.Debug("chat data: %v", result)
	//fmt.Println("Result",result)
	//fmt.Println("chatIds",chatIds)
	chatStatus := "ok"
	if len(utils.ChatInConnections) == 0 || len(utils.ChatOutConnections) == 0 {
		chatStatus = "bad"
	}

	//fmt.Println("result",result)

	resultJson, _ := json.Marshal(map[string]string{"messages": result, "chatStatus": chatStatus})

	return string(resultJson), nil
}


func getChatData(c *Controller) ([]map[string]string, error) {
	room := utils.StrToInt64(c.r.FormValue("room"))
	lang := utils.StrToInt64(c.r.FormValue("lang"))
	var ids string
	if len(chatIds[c.SessUserId]) > 0 {
		ids = `AND id NOT IN(` + strings.Join(utils.IntSliceToStr(chatIds[c.SessUserId]), ",") + `)`
	}
	//fmt.Println("utils.ChatMinSignTime", utils.ChatMinSignTime)
	chatData, err := c.GetAll(`SELECT * FROM chat WHERE sign_time > ? AND room = ? AND lang = ?  ` +
			ids +
			` ORDER BY sign_time DESC LIMIT `+
			utils.Int64ToStr(consts.CHAT_COUNT_MESSAGES),
			consts.CHAT_COUNT_MESSAGES,
			utils.ChatMinSignTime,
			room,
			lang)

	return chatData, utils.ErrInfo(err)
}

func decrypt(privateKey, message, id string, c *Controller) (string, error) {
	rsaPrivateKey, err := utils.MakePrivateKey(privateKey)
	if err != nil {
		return "", utils.ErrInfo(err)
	}

	decrypted, err := rsa.DecryptPKCS1v15(rand.Reader, rsaPrivateKey, utils.HexToBin([]byte(message)))
	if err != nil {
		return "", utils.ErrInfo(err)
	}
	if len(decrypted) > 0 {
		err = c.ExecSql(`UPDATE chat SET enc_message = message, message = ?, status = ? WHERE id = ?`, decrypted, 2, id)
		if err != nil {
			return "", utils.ErrInfo(err)
		}

		return string(decrypted), nil
	}

	return "", nil
}


func removeOld(c *Controller) error {
	maxId, err := c.Single(`SELECT max(id) FROM chat`).Int64()
	if err != nil {
		return utils.ErrInfo(err)
	}

	// удалим старое
	err = c.ExecSql(`DELETE FROM chat WHERE id < ?`, maxId - consts.CHAT_MAX_MESSAGES)
	return utils.ErrInfo(err)
}