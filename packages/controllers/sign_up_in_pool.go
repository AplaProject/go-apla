package controllers

import (
	"github.com/DayLightProject/go-daylight/packages/schema"
	"github.com/DayLightProject/go-daylight/packages/utils"
	"regexp"
)

func (c *Controller) SignUpInPool() (string, error) {

	c.w.Header().Set("Access-Control-Allow-Origin", "*")

	if !c.Community {
		return "", utils.JsonAnswer("Not pool", "error").Error()
	}

	c.r.ParseForm()

	var userId int64
	var codeSign string
	if c.SessUserId <= 0 {
		// запрос пришел с десктопного кошелька юзера
		codeSign = c.r.FormValue("code_sign")
		if !utils.CheckInputData(codeSign, "hex_sign") {
			return "", utils.JsonAnswer("Incorrect code_sign", "error").Error()
		}
		userId = utils.StrToInt64(c.r.FormValue("user_id"))
		if !utils.CheckInputData(userId, "int64") {
			return "", utils.JsonAnswer("Incorrect userId", "error").Error()
		}
		// получим данные для подписи
		var hash []byte

		RemoteAddr := utils.RemoteAddrFix(c.r.RemoteAddr)
		re := regexp.MustCompile(`(.*?):[0-9]+$`)
		match := re.FindStringSubmatch(RemoteAddr)
		if len(match) != 0 {
			RemoteAddr = match[1]
		}
		log.Debug("RemoteAddr %s", RemoteAddr)
		hash = utils.Md5(c.r.Header.Get("User-Agent") + RemoteAddr)
		log.Debug("hash %s", hash)

		forSign, err := c.GetDataAuthorization(hash)
		log.Debug("forSign: %v", forSign)
		publicKey, err := c.GetUserPublicKey(userId)
		log.Debug("publicKey: %x", publicKey)
		if err != nil {
			return "", utils.JsonAnswer(utils.ErrInfo(err), "error").Error()
		}
		// проверим подпись
		resultCheckSign, err := utils.CheckSign([][]byte{[]byte(publicKey)}, forSign, utils.HexToBin([]byte(codeSign)), true)
		if err != nil {
			return "", utils.JsonAnswer(utils.ErrInfo(err), "error").Error()
		}
		if !resultCheckSign {
			return "", utils.JsonAnswer("Incorrect codeSign", "error").Error()
		}
	} else {
		// запрос внутри пула
		userId = c.SessUserId
	}
	/*e:=c.r.FormValue("e")
	n:=c.r.FormValue("n")
	if len(e) == 0 || len(n) == 0 {
		result, _ := json.Marshal(map[string]string{"error": c.Lang["pool_error"]})
		return "", errors.New(string(result))
	}*/
	email := c.r.FormValue("email")
	if !utils.ValidateEmail(email) {
		return "", utils.JsonAnswer("Incorrect email", "error").Error()
	}
	nodePrivateKey := c.r.FormValue("node_private_key")
	if !utils.CheckInputData(nodePrivateKey, "private_key") {
		return "", utils.JsonAnswer("Incorrect private_key", "error").Error()
	}
	//publicKey := utils.MakeAsn1([]byte(n), []byte(e))
	log.Debug("3")

	// если мест в пуле нет, то просто запишем юзера в очередь
	pool_max_users, err := c.Single("SELECT pool_max_users FROM config").Int()
	if err != nil {
		return "", utils.JsonAnswer(utils.ErrInfo(err), "error").Error()
	}
	if len(c.CommunityUsers) >= pool_max_users {
		if existId,err := c.Single(`SELECT user_id FROM pool_waiting_list WHERE email=?`, email ).Int64(); existId == 0 {
			if err == nil {
				err = c.ExecSql("INSERT INTO pool_waiting_list ( email, time, user_id ) VALUES ( ?, ?, ? )", email, utils.Time(), userId)
			}
			if err != nil {
				return "", utils.JsonAnswer(utils.ErrInfo(err), "error").Error()
			}
		}
		return "", utils.JsonAnswer(c.Lang["pool_is_full"], "error").Error()
	}

	// регистрируем юзера в пуле
	// вначале убедимся, что такой user_id у нас уже не зареган
	community, err := c.Single("SELECT user_id FROM community WHERE user_id  =  ?", userId).Int64()
	if err != nil {
		return "", utils.JsonAnswer(utils.ErrInfo(err), "error").Error()
	}
	if community != 0 {
		return utils.JsonAnswer(c.Lang["pool_user_id_is_busy"], "success").String(), nil
	}
	err = c.ExecSql("INSERT INTO community ( user_id ) VALUES ( ? )", userId)
	if err != nil {
		return "", utils.JsonAnswer(utils.ErrInfo(err), "error").Error()
	}

	schema_ := &schema.SchemaStruct{}
	schema_.DCDB = c.DCDB
	schema_.DbType = c.ConfigIni["db_type"]
	schema_.PrefixUserId = int(userId)
	schema_.GetSchema()

	prefix := utils.Int64ToStr(userId) + "_"
	minerId, err := c.GetMinerId(userId)
	if err != nil {
		return "", utils.JsonAnswer(utils.ErrInfo(err), "error").Error()
	}
	if minerId > 0 {
		err = c.ExecSql("INSERT INTO "+prefix+"my_table ( user_id, miner_id, status, email ) VALUES  (?, ?, ?, ?)", userId, minerId, "miner", email)
		if err != nil {
			return "", utils.JsonAnswer(utils.ErrInfo(err), "error").Error()
		}
	} else {
		err = c.ExecSql("INSERT INTO "+prefix+"my_table ( user_id, status, email ) VALUES (?, ?, ?)", userId, "miner", email)
		if err != nil {
			return "", utils.JsonAnswer(utils.ErrInfo(err), "error").Error()
		}
	}

	publicKey, err := c.GetUserPublicKey(userId)
	err = c.ExecSql("INSERT INTO "+prefix+"my_keys ( public_key, status ) VALUES ( [hex], 'approved' )", utils.BinToHex(publicKey))
	if err != nil {
		return "", utils.JsonAnswer(utils.ErrInfo(err), "error").Error()
	}
	nodePublicKey, err := utils.GetPublicFromPrivate(nodePrivateKey)
	if err != nil {
		return "", utils.JsonAnswer(utils.ErrInfo(err), "error").Error()
	}
	err = c.ExecSql("INSERT INTO "+prefix+"my_node_keys ( private_key, public_key ) VALUES ( ?, [hex] )", nodePrivateKey, nodePublicKey)
	if err != nil {
		return "", utils.JsonAnswer(utils.ErrInfo(err), "error").Error()
	}
	// Ничего не отправляем, просто добавляем email на сервер
	utils.SendEmail( email, userId, utils.ECMD_SIGNUP, nil )
										
	c.sess.Delete("restricted")
	return utils.JsonAnswer(c.Lang["pool_sign_up_success"], "success").String(), nil
}
