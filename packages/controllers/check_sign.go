package controllers

import (
	"github.com/DayLightProject/go-daylight/packages/utils"
	"regexp"
	"time"
)

func (c *Controller) Check_sign() (string, error) {

	var checkError bool

	c.r.ParseForm()
	forsignature := []byte(c.r.FormValue("forsignature")) // для дебага
	n := []byte(c.r.FormValue("n"))
	e := []byte(c.r.FormValue("e"))
	sign := []byte(c.r.FormValue("sign"))
	setupPassword := c.r.FormValue("setup_password")
	private_key := c.r.FormValue("private_key")
	if !utils.CheckInputData(n, "hex") {
		log.Error("incorrect n %v", n)
		return `{"result":"incorrect n"}`, nil
	}
	if !utils.CheckInputData(e, "hex") {
		log.Error("incorrect e %v", e)
		return `{"result":"incorrect e"}`, nil
	}
	if !utils.CheckInputData(string(sign), "hex_sign") {
		log.Error("incorrect sign %v", string(sign))
		return `{"result":"incorrect sign"}`, nil
	}

	allTables, err := c.DCDB.GetAllTables()
	if err != nil {
		log.Error("err %v", err)
		return "{\"result\":0}", err
	}

	var hash []byte
	log.Debug("forsignature %s", forsignature)
	log.Debug("string(sign) %s", string(sign))
	log.Debug("n %s", n)
	log.Debug("e %s", e)
	log.Debug("c.r.RemoteAddr %s", c.r.RemoteAddr)
	log.Debug("c.r.Header.Get(User-Agent) %s", c.r.Header.Get("User-Agent"))
	RemoteAddr := utils.RemoteAddrFix(c.r.RemoteAddr)
	re := regexp.MustCompile(`(.*?):[0-9]+$`)
	match := re.FindStringSubmatch(RemoteAddr)
	if len(match) != 0 {
		RemoteAddr = match[1]
	}
	log.Debug("RemoteAddr %s", RemoteAddr)
	hash = utils.Md5(c.r.Header.Get("User-Agent") + RemoteAddr)
	log.Debug("hash %s", hash)

	if len(c.CommunityUsers) > 0 {

		// в цикле проверяем, кому подойдет присланная подпись
		for _, userId := range c.CommunityUsers {

			myPrefix := utils.Int64ToStr(userId) + "_"
			if !utils.InSliceString(myPrefix+"my_keys", allTables) {
				continue
			}

			// получим открытый ключ юзера
			publicKey, err := c.DCDB.GetMyPublicKey(myPrefix)
			if err != nil {
				log.Error("err %v", err)
				return "{\"result\":0}", err
			}

			// получим данные для подписи
			forSign, err := c.DCDB.GetDataAuthorization(hash)

			log.Debug("publicKey: %x\n", publicKey)
			log.Debug("myPrefix: ", myPrefix)
			log.Debug("sign:  %s\n", sign)
			log.Debug("hash: %s\n", hash)
			log.Debug("forSign: ", forSign)
			// проверим подпись
			resultCheckSign, err := utils.CheckSign([][]byte{publicKey}, forSign, utils.HexToBin(sign), true)
			if err != nil {
				continue
			}
			// если подпись верная, значит мы нашли юзера, который эту подпись смог сделать
			if resultCheckSign {
				myUserId := userId
				// убираем ограниченный режим
				c.sess.Delete("restricted")
				c.sess.Set("user_id", myUserId)
				log.Debug("c.sess.Set(user_id) %d", myUserId)
				public_key, err := c.DCDB.GetUserPublicKey(myUserId)
				if err != nil {
					log.Error("err %v", err)
					return "{\"result\":0}", err
				}
				// паблик кей в сессии нужен чтобы выбрасывать юзера, если ключ изменился
				c.sess.Set("public_key", string(utils.BinToHex([]byte(public_key))))
				log.Debug("string(utils.BinToHex([]byte(public_key))) %s", string(utils.BinToHex([]byte(public_key))))

				adminUSerID, err := c.DCDB.GetAdminUserId()
				if err != nil {
					log.Error("err %v", err)
					return "{\"result\":0}", err
				}
				if adminUSerID == myUserId {
					c.sess.Set("admin", 1)
				}
				return "{\"result\":1}", nil
			}
		}
		log.Debug("restricted test")
		// если дошли досюда, значит ни один ключ не подошел и даем возможность войти в ограниченном режиме
		publicKey := utils.MakeAsn1(n, e)
		userId_, err := c.DCDB.GetUserIdByPublicKey(publicKey)
		userId := utils.StrToInt64(userId_)
		if err != nil {
			log.Error("err %v", err)
			return "{\"result\":0}", err
		}

		log.Debug("userId: %v / publicKey %s", userId, string(publicKey))
		// юзер с таким ключем есть в БД
		if userId > 0 {

			// получим данные для подписи
			forSign, err := c.DCDB.GetDataAuthorization(hash)
			log.Debug("forSign", forSign)
			log.Debug("publicKey %x\n", utils.HexToBin(publicKey))
			log.Debug("sign_", string(sign))
			// проверим подпись
			resultCheckSign, err := utils.CheckSign([][]byte{utils.HexToBin(publicKey)}, forSign, utils.HexToBin(sign), true)
			if err != nil {
				log.Error("err %v", err)
				return "{\"result\":0}", err
			}
			if resultCheckSign {

				// если юзер смог подписать наш хэш, значит у него актуальный праймари ключ
				c.sess.Set("user_id", userId)
				log.Debug("c.sess.Set(user_id) %d", userId)

				// паблик кей в сессии нужен чтобы выбрасывать юзера, если ключ изменился
				c.sess.Set("public_key", string(publicKey))

				// возможно в табле my_keys старые данные, но если эта табла есть, то нужно добавить туда ключ
				if utils.InSliceString(utils.Int64ToStr(userId)+"_my_keys", allTables) {
					curBlockId, err := c.DCDB.GetBlockId()
					if err != nil {
						log.Error("err %v", err)
						return "{\"result\":0}", err
					}
					err = c.DCDB.InsertIntoMyKey(utils.Int64ToStr(userId)+"_", publicKey, utils.Int64ToStr(curBlockId))
					if err != nil {
						log.Error("err %v", err)
						return "{\"result\":0}", err
					}
					c.sess.Delete("restricted")
				} else {
					c.sess.Set("restricted", int64(1))
					log.Debug("c.sess.Set(restricted) 1")
				}
				return "{\"result\":1}", nil
			} else {
				log.Error("!resultCheckSign")
				return "{\"result\":0}", nil
			}
		}
	} else {
		if userId, err := c.DCDB.GetMyUserId(""); err != nil {
				return "{\"result\":0}", err
		} else if userId == 0 {
			// Пока не ясно, но иногда в my_keys имеется public key при user_id==0 в my_table
			// Что не дает залогиниться. Как вариант решения.
			c.DCDB.ExecSql(`delete from my_keys`)
		}
		// получим открытый ключ юзера
		publicKey, err := c.DCDB.GetMyPublicKey("")
		if err != nil {
			log.Error("err %v", err)
			return "{\"result\":0}", err
		}

		// Если ключ еще не успели установить
		if len(publicKey) == 0 {

			// пока не собрана цепочка блоков не даем ввести ключ
			infoBlock, err := c.DCDB.GetInfoBlock()
			if err != nil {
				log.Error("err %v", err)
				return "{\"result\":0}", err
			}
			// если последний блок не старше 2-х часов
			wTime := int64(2)
			if c.ConfigIni["test_mode"] == "1" {
				wTime = 2 * 365 * 24
			}
			log.Debug("%v/%v/%v", time.Now().Unix(), utils.StrToInt64(infoBlock["time"]), wTime)
			if (time.Now().Unix() - utils.StrToInt64(infoBlock["time"])) < 3600*wTime {

				// проверим, верный ли установочный пароль, если он, конечно, есть
				setupPassword_, err := c.Single("SELECT setup_password FROM config").String()
				if err != nil {
					log.Error("err %v", err)
					return "{\"result\":0}", err
				}
				if len(setupPassword_) > 0 && setupPassword_ != string(utils.DSha256(setupPassword)) {
					log.Error(setupPassword_, string(utils.DSha256(setupPassword)), setupPassword)
					return "{\"result\":0}", nil
				}

				publicKey := utils.MakeAsn1(n, e)
				log.Debug("new key", string(publicKey))
				userId, err := c.GetUserIdByPublicKey(publicKey)
				if err != nil {
					log.Error("err %v", err)
					return "{\"result\":0}", err
				}

				// получим данные для подписи
				forSign, err := c.DCDB.GetDataAuthorization(hash)
				if err != nil {
					log.Error("err %v", err)
					return "{\"result\":0}", err
				}
				log.Debug("forSign", forSign)
				log.Debug("publicKey %x\n", utils.HexToBin(publicKey))
				log.Debug("sign_", string(sign))
				// проверим подпись
				resultCheckSign, err := utils.CheckSign([][]byte{utils.HexToBin(publicKey)}, forSign, utils.HexToBin(sign), true)
				if err != nil {
					log.Error("err %v", err)
					return "{\"result\":0}", err
				}
				if !resultCheckSign {
					log.Error("!resultCheckSign")
					return "{\"result\":0}", nil
				}

				if len(userId) > 0 {
					err := c.InsertIntoMyKey("", publicKey, "0")
					if err != nil {
						log.Error("err %v", err)
						return "{\"result\":0}", err
					}
					minerId, err := c.GetMinerId(utils.StrToInt64(userId))
					if err != nil {
						log.Error("err %v", err)
						return "{\"result\":0}", err
					}
					//myUserId, err := c.GetMyUserId("")
					//if myUserId > 0 {
					if minerId > 0 {
						err = c.ExecSql("UPDATE my_table SET user_id = ?, miner_id = ?, status = 'miner'", userId, minerId)
						if err != nil {
							log.Error("err %v", err)
							return "{\"result\":0}", err
						}
					} else {
						err = c.ExecSql("UPDATE my_table SET user_id = ?, status = 'user'", userId)
						if err != nil {
							log.Error("err %v", err)
							return "{\"result\":0}", err
						}
					}
					//} else {
					//	c.ExecSql("INSERT INTO my_table (user_id, status) VALUES (?, 'user')", userId)
					//}
					// возможно юзер хочет сохранить свой ключ
					if len(private_key) > 0 {
						c.ExecSql("UPDATE my_keys SET private_key = ? WHERE block_id = (SELECT max(block_id) FROM my_keys)", private_key)
					}

				} else {
					checkError = true
				}
			} else {
				checkError = true
			}
		} else {

			log.Debug("RemoteAddr %s", RemoteAddr)
			hash = utils.Md5(c.r.Header.Get("User-Agent") + RemoteAddr)
			log.Debug("hash %s", hash)

			// получим данные для подписи
			forSign, err := c.DCDB.GetDataAuthorization(hash)
			log.Debug("forSign", forSign)
			log.Debug("publicKey %x\n", string(publicKey))
			log.Debug("sign_", string(sign))
			// проверим подпись
			resultCheckSign, err := utils.CheckSign([][]byte{publicKey}, forSign, utils.HexToBin(sign), true)
			if err != nil {
				log.Error("err %v", err)
				return "{\"result\":0}", err
			}
			if !resultCheckSign {
				log.Error("!resultCheckSign")
				return "{\"result\":0}", nil
			}

		}

		if checkError {
			log.Error("checkError")
			return "{\"result\":0}", nil
		} else {
			myUserId, err := c.DCDB.GetMyUserId("")
			if myUserId == 0 {
				myUserId = -1
			}
			if err != nil {
				log.Error("err %v", err)
				return "{\"result\":0}", err
			}
			c.sess.Delete("restricted")
			c.sess.Set("user_id", myUserId)

			// если уже пришел блок, в котором зареган ключ юзера
			if myUserId != -1 {

				public_key, err := c.DCDB.GetUserPublicKey(myUserId)
				if err != nil {
					log.Error("err %v", err)
					return "{\"result\":0}", err
				}
				// паблик кей в сессии нужен чтобы выбрасывать юзера, если ключ изменился
				c.sess.Set("public_key", string(utils.BinToHex(public_key)))

				// возможно юзер хочет сохранить свой ключ
				if len(private_key) > 0 {
					c.ExecSql("UPDATE my_keys SET private_key = ? WHERE block_id = (SELECT max(block_id) FROM my_keys)", private_key)
				}

				AdminUserId, err := c.DCDB.GetAdminUserId()
				if err != nil {
					log.Error("err %v", err)
					return "{\"result\":0}", err
				}
				if AdminUserId == myUserId {
					c.sess.Set("admin", int64(1))
				}
				return "{\"result\":1}", nil
			}
		}
	}
	log.Error("result 0 ")
	return "{\"result\":0}", nil
}
