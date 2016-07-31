package dcparser

import (
	"fmt"
	"github.com/DayLightProject/go-daylight/packages/utils"
	//"encoding/json"
	"regexp"
)

func (p *Parser) NewUserInit() error {
	err := p.GetTxMaps([]map[string]string{{"public_key": "bytes"}, {"sign": "bytes"}})
	if err != nil {
		return p.ErrInfo(err)
	}
	p.TxMap["public_key_hex"] = utils.BinToHex(p.TxMap["public_key"])
	p.TxMaps.Bytes["public_key_hex"] = utils.BinToHex(p.TxMaps.Bytes["public_key"])
	return nil
}

func (p *Parser) NewUserFront() error {

	log.Debug("NewUserFront")

	err := p.generalCheck()
	if err != nil {
		return p.ErrInfo(err)
	}

	// является ли данный юзер майнером
	err = p.checkMiner(p.TxUserID)
	if err != nil {
		return p.ErrInfo(err)
	}

	// прошло ли 30 дней с момента регистрации майнера
	err = p.checkMinerNewbie()
	if err != nil {
		return p.ErrInfo(err)
	}

	// чтобы не записали слишком мелкий или слишком крупный ключ
	if !utils.CheckInputData(p.TxMap["public_key_hex"], "public_key") {
		return utils.ErrInfo(fmt.Errorf("incorrect public_key %s", p.TxMap["public_key_hex"]))
	}

	// публичный ключ должен быть без паролей
	if ok, _ := regexp.MatchString("DEK-Info: (.+),(.+)", string(p.TxMap["public_key"])); ok {
		return p.ErrInfo("incorrect public_key")
	}

	forSign := fmt.Sprintf("%s,%s,%s,%s", p.TxMap["type"], p.TxMap["time"], p.TxMap["user_id"], p.TxMap["public_key_hex"])
	CheckSignResult, err := utils.CheckSign(p.PublicKeys, forSign, p.TxMap["sign"], false)
	if err != nil {
		return p.ErrInfo(err)
	}
	if !CheckSignResult {
		return p.ErrInfo("incorrect sign")
	}

	// один ключ не может быть у двух юзеров
	num, err := p.DCDB.Single("SELECT count(user_id) FROM users WHERE hex(public_key_0) = ? OR hex(public_key_1) = ? OR hex(public_key_2) = ?",
		p.TxMap["public_key_hex"], p.TxMap["public_key_hex"], p.TxMap["public_key_hex"]).Int()
	if num > 0 {
		return p.ErrInfo("exists public_key")
	}
	err = p.getAdminUserId()
	if err != nil {
		return p.ErrInfo(err)
	}
	if utils.BytesToInt64(p.TxMap["user_id"]) == p.AdminUserId {
		err = p.limitRequest(1000, "new_user", 86400)
	} else {
		err = p.limitRequest(p.Variables.Int64["limit_new_user"], "new_user", p.Variables.Int64["limit_new_user_period"])
	}
	if err != nil {
		return p.ErrInfo(err)
	}
	return nil
}

func (p *Parser) NewUser() error {
	log.Debug("NewUser")
	// пишем в БД нового юзера
	newUserId, err := p.ExecSqlGetLastInsertId("INSERT INTO users (public_key_0, referral) VALUES ([hex], ?)", "user_id", p.TxMap["public_key_hex"], p.TxUserID)
	if err != nil {
		return p.ErrInfo(err)
	}
	log.Debug("newUserId %v", newUserId)

	// если работаем в режиме пула, то ищем тех, у кого еще нет user_id
	community, err := p.DCDB.GetCommunityUsers()
	if err != nil {
		return p.ErrInfo(err)
	}
	if len(community) > 0 {
		for _, userId := range community {
			myPrefix := utils.Int64ToStr(userId) + "_"
			myUserId, err := p.DCDB.Single("SELECT user_id FROM " + myPrefix + "my_table").Int64()
			if err != nil {
				return p.ErrInfo(err)
			}
			if myUserId == 0 {
				// проверим, не наш ли это public_key, чтобы записать полученный user_id в my_table
				myPublicKey, err := p.DCDB.Single("SELECT public_key FROM "+myPrefix+"my_keys WHERE hex(public_key) = ?", p.TxMap["public_key_hex"]).String()
				if err != nil {
					return p.ErrInfo(err)
				}
				if myPublicKey != "" {
					// теперь у нас полноценный юзерский акк, и его можно апргрейдить до майнерского
					err = p.DCDB.ExecSql("UPDATE "+myPrefix+"my_table SET user_id = ?, status = 'user', notification_status = 0", newUserId)
					if err != nil {
						return p.ErrInfo(err)
					}
					err = p.DCDB.ExecSql("UPDATE "+myPrefix+"my_keys SET block_id = ? WHERE hex(public_key) = ?", p.BlockData.BlockId, p.TxMap["public_key_hex"])
					if err != nil {
						return p.ErrInfo(err)
					}
				}
			}
		}
	} else {
		myUserId, err := p.DCDB.Single("SELECT user_id FROM my_table").Int64()
		if err != nil {
			return p.ErrInfo(err)
		}
		if myUserId == 0 {

			// проверим, не наш ли это public_key, чтобы записать полученный user_id в my_table
			myPublicKey, err := p.DCDB.Single("SELECT public_key FROM my_keys WHERE hex(public_key) = ?", p.TxMap["public_key_hex"]).String()
			if err != nil {
				return p.ErrInfo(err)
			}
			if myPublicKey != "" {
				// теперь у нас полноценный юзерский акк, и его можно апргрейдить до майнерского
				err = p.DCDB.ExecSql("UPDATE my_table SET user_id = ?, status = 'user', notification_status = 0", newUserId)
				if err != nil {
					return p.ErrInfo(err)
				}
				err = p.DCDB.ExecSql("UPDATE my_keys SET block_id = ? WHERE hex(public_key) = ?", p.BlockData.BlockId, p.TxMap["public_key_hex"])
				if err != nil {
					return p.ErrInfo(err)
				}
			}
		}
	}
	// проверим, не наш ли это user_id
	myUserId, myBlockId, myPrefix, _, err := p.GetMyUserId(p.TxUserID)
	if err != nil {
		return p.ErrInfo(err)
	}
	if p.TxUserID == myUserId && myBlockId <= p.BlockData.BlockId {
		p.DCDB.ExecSql("UPDATE "+myPrefix+"my_new_users SET status ='approved', user_id = ? WHERE hex(public_key) = ?", newUserId, p.TxMap["public_key_hex"])
	}
	if p.TxUserID > 1 && myBlockId <= p.BlockData.BlockId {
		p.nfyRefReady(p.TxUserID, newUserId)
	}

	p.nfyStatus(newUserId, `user`)
	return nil
}

func (p *Parser) NewUserRollback() error {
	// если работаем в режиме пула, то ищем тех, у кого записан такой ключ
	community, err := p.DCDB.GetCommunityUsers()
	if err != nil {
		return p.ErrInfo(err)
	}
	if len(community) > 0 {
		for _, userId := range community {
			myPrefix := utils.Int64ToStr(userId) + "_"
			// проверим, не наш ли это public_key, чтобы записать полученный user_id в my_table
			myPublicKey, err := p.DCDB.Single("SELECT public_key FROM "+myPrefix+"my_keys WHERE hex(public_key) = ?", p.TxMap["public_key_hex"]).String()
			if err != nil {
				return p.ErrInfo(err)
			}
			if myPublicKey != "" {
				// теперь у нас полноценный юзерский акк, и его можно апргрейдить до майнерского
				err = p.DCDB.ExecSql("UPDATE " + myPrefix + "my_table SET user_id = 0, status = 'my_pending', notification_status = 0")
				if err != nil {
					return p.ErrInfo(err)
				}
				err = p.DCDB.ExecSql("UPDATE "+myPrefix+"my_keys SET block_id = 0 WHERE block_id = ?", p.BlockData.BlockId)
				if err != nil {
					return p.ErrInfo(err)
				}
			}
		}
	} else {
		// проверим, не наш ли это public_key
		myPublicKey, err := p.DCDB.Single("SELECT public_key FROM my_keys WHERE hex(public_key) = ?", p.TxMap["public_key_hex"]).String()
		if err != nil {
			return p.ErrInfo(err)
		}
		if myPublicKey != "" {
			err = p.DCDB.ExecSql("UPDATE my_table SET user_id = 0, status = 'my_pending', notification_status = 0")
			if err != nil {
				return p.ErrInfo(err)
			}
			err = p.DCDB.ExecSql("UPDATE my_keys SET block_id = 0 WHERE block_id = ?", p.BlockData.BlockId)
			if err != nil {
				return p.ErrInfo(err)
			}
		}
	}
	err = p.DCDB.ExecSql("DELETE FROM users WHERE hex(public_key_0) = ?", p.TxMap["public_key_hex"])
	if err != nil {
		return p.ErrInfo(err)
	}
	err = p.rollbackAI("users", 1)
	if err != nil {
		return p.ErrInfo(err)
	}
	return nil
}

func (p *Parser) NewUserRollbackFront() error {
	return p.limitRequestsRollback("new_user")
}
