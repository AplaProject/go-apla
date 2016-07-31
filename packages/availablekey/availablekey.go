package availablekey

import (
	"errors"
	"github.com/DayLightProject/go-daylight/packages/schema"
	"github.com/DayLightProject/go-daylight/packages/utils"
	"github.com/op/go-logging"
	"math/rand"
	"regexp"
)

var log = logging.MustGetLogger("availablekey")

type AvailablekeyStruct struct {
	*utils.DCDB
	Email string
}

func (a *AvailablekeyStruct) checkAvailableKey(key string) (int64, string, error) {
	publicKeyAsn, err := utils.GetPublicFromPrivate(key)
	if err != nil {
		log.Debug("%v", err)
		return 0, "", utils.ErrInfo(err)
	}
	log.Debug("publicKeyAsn: %s", publicKeyAsn)
	userId, err := a.Single("SELECT user_id FROM users WHERE hex(public_key_0) = ?", publicKeyAsn).Int64()
	if err != nil {
		return 0, "", utils.ErrInfo(err)
	}
	log.Debug("userId: %s", userId)
	if userId == 0 {
		return 0, "", errors.New("null userId")
	}
	allTables, err := a.GetAllTables()
	if err != nil {
		return 0, "", utils.ErrInfo(err)
	}
	// может другой юзер уже начал смену ключа. актуально для пула
	if utils.InSliceString(utils.Int64ToStr(userId)+"_my_table", allTables) {
		return 0, "", errors.New("exists _my_table")
	}
	return userId, string(publicKeyAsn), nil
}

func (a *AvailablekeyStruct) GetAvailableKey() (int64, string, error) {

	community, err := a.GetCommunityUsers()
	if err != nil {
		return 0, "", utils.ErrInfo(err)
	}
	// запрещено менять ключ таким методом если в my_table уже есть статус
	if len(community) == 0 {
		status, err := a.Single("SELECT status FROM my_table").String()
		if err != nil {
			return 0, "", utils.ErrInfo(err)
		}
		if status == "waiting_set_new_key" {
			// Была прервана регистрация - обнуляем таблицы
			if err = a.ExecSql("UPDATE my_table SET user_id = 0, status = ?", "my_pending" ); err != nil {
				return 0, "", utils.ErrInfo(err)
			}
			if err = a.ExecSql("DELETE FROM my_keys");err != nil {
				return 0, "", utils.ErrInfo(err)
			}
			status = "my_pending"
		}
		if status != "my_pending" {
			return 0, "", utils.ErrInfo(errors.New("my_table not null"))
		}
	}

	var keys []string
	for i:=0; i<10; i++ {
		keysStr, err := utils.GetHttpTextAnswer("http://dcoin.club/keys")
		if err != nil {
			if err.Error() == `404` {
				break
			}
			return 0, "", utils.ErrInfo(err)
		}
		//keysStr = strings.Replace(keysStr, "\n", "", -1)
		r, _ := regexp.Compile("(?s)-----BEGIN RSA PRIVATE KEY-----(.*?)-----END RSA PRIVATE KEY-----")
		keys = r.FindAllString(keysStr, -1)
		for i := range keys {
			j := rand.Intn(i + 1)
			keys[i], keys[j] = keys[j], keys[i]
		}
		if len(keys) > 0 {
			break
		} else {
			utils.Sleep(5)
		}
	}

	for _, key := range keys {
		userId, pubKey, err := a.checkAvailableKey(key)
		if err != nil {
			log.Error("%s", utils.ErrInfo(err)) // тут ошибка - это нормально
		}
		log.Debug("checkAvailableKey userId: %v", userId)
		if userId > 0 {
			// запишем приватный ключ в БД, чтобы можно было подписать тр-ию на смену ключа
			myPref := ""

			log.Debug("schema_ 0")
			if len(community) > 0 {
				schema_ := &schema.SchemaStruct{}
				schema_.DCDB = a.DCDB
				schema_.DbType = a.ConfigIni["db_type"]
				schema_.PrefixUserId = int(userId)
				schema_.GetSchema()
				myPref = utils.Int64ToStr(userId) + "_"
				err = a.ExecSql("INSERT INTO "+myPref+"my_table (user_id, status, email) VALUES (?, ?, ?)", userId, "waiting_set_new_key", a.Email)
				if err != nil {
					return 0, "", utils.ErrInfo(err)
				}
				err = a.ExecSql("INSERT INTO community ( user_id ) VALUES ( ? )", userId)
				if err != nil {
					return 0, "", utils.ErrInfo(err)
				}
			} else {
				err = a.ExecSql("UPDATE my_table SET user_id = ?, status = ?", userId, "waiting_set_new_key")
				if err != nil {
					return 0, "", utils.ErrInfo(err)
				}
			}
			log.Debug("schema_ 1")

			// пишем приватный в my_keys т.к. им будем подписывать тр-ию на смену ключа
			err = a.ExecSql("INSERT INTO "+myPref+"my_keys (private_key, public_key, status, block_id) VALUES (?, [hex], ?, ?)", key, pubKey, "approved", 1)
			if err != nil {
				return 0, "", utils.ErrInfo(err)
			}
			log.Debug("GenKeys 0")
			newPrivKey, newPubKey := utils.GenKeys()
			log.Debug("GenKeys 1")
			// сразу генерируем новый ключ и пишем приватный временно в my_keys, чтобы можно было выдавать юзеру для скачивания
			err = a.ExecSql("INSERT INTO "+myPref+"my_keys (private_key, public_key, status) VALUES (?, ?, ?)", newPrivKey, utils.HexToBin([]byte(newPubKey)), "my_pending")
			if err != nil {
				return 0, "", utils.ErrInfo(err)
			}
			log.Debug("return userId %d", userId)
			return userId, pubKey, nil
		}
	}
	return 0, "", nil
}
