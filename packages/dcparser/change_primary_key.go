package dcparser

import (
	"bytes"
	"database/sql"
	"fmt"
	"github.com/DayLightProject/go-daylight/packages/utils"
)

func (p *Parser) ChangePrimaryKeyInit() error {
	var err error
	fields := []string{"bin_public_keys", "sign"}
	p.TxMap, err = p.GetTxMap(fields)
	if err != nil {
		return p.ErrInfo(err)
	}

	p.newPublicKeysHex = [3][]byte{}

	// в 1 new_public_keys может быть от 1 до 3-х ключей
	i := 0
	bin_public_keys := p.TxMap["bin_public_keys"]
	for {
		length := utils.DecodeLength(&bin_public_keys)
		pKey := utils.BytesShift(&bin_public_keys, length)
		p.newPublicKeysHex[i] = utils.BinToHex(pKey)
		if len(bin_public_keys) == 0 || i > 1 {
			break
		}
		i++
	}
	return nil
}

func (p *Parser) ChangePrimaryKeyFront() error {

	err := p.generalCheck()
	if err != nil {
		return p.ErrInfo(err)
	}

	if !utils.CheckInputData(p.newPublicKeysHex[0], "public_key") {
		return p.ErrInfo("public_key")
	}
	if len(p.newPublicKeysHex[1]) > 0 && !utils.CheckInputData(p.newPublicKeysHex[1], "public_key") {
		return p.ErrInfo("public_key 1")
	}
	if len(p.newPublicKeysHex[2]) > 0 && !utils.CheckInputData(p.newPublicKeysHex[2], "public_key") {
		return p.ErrInfo("public_key 2")
	}

	forSign := fmt.Sprintf("%s,%s,%s,%s,%s,%s", p.TxMap["type"], p.TxMap["time"], p.TxMap["user_id"], p.newPublicKeysHex[0], p.newPublicKeysHex[1], p.newPublicKeysHex[2])
	CheckSignResult, err := utils.CheckSign(p.PublicKeys, forSign, p.TxMap["sign"], false)
	if err != nil {
		return p.ErrInfo(err)
	}
	if !CheckSignResult {
		return p.ErrInfo("incorrect sign")
	}

	err = p.limitRequest(p.Variables.Int64["limit_primary_key"], "primary_key", p.Variables.Int64["limit_primary_key_period"])
	if err != nil {
		return p.ErrInfo(err)
	}
	return nil
}

func (p *Parser) ChangePrimaryKey() error {

	// Всегда есть, что логировать, т.к. это обновление ключа
	var public_key_0, public_key_1, public_key_2 []byte
	var log_id int64
	err := p.QueryRow(p.FormatQuery("SELECT hex(public_key_0), hex(public_key_1), hex(public_key_2), log_id FROM users WHERE user_id  =  ?"), p.TxUserID).Scan(&public_key_0, &public_key_1, &public_key_2, &log_id)
	if err != nil && err != sql.ErrNoRows {
		return p.ErrInfo(err)
	}

	logId, err := p.ExecSqlGetLastInsertId("INSERT INTO log_users ( public_key_0, public_key_1, public_key_2, block_id, prev_log_id ) VALUES ( [hex], [hex], [hex], ?, ? )", "log_id", public_key_0, public_key_1, public_key_2, p.BlockData.BlockId, log_id)
	if err != nil {
		return p.ErrInfo(err)
	}
	err = p.ExecSql("UPDATE users SET public_key_0 = [hex], public_key_1 = [hex], public_key_2 = [hex], log_id = ? WHERE user_id = ?", p.newPublicKeysHex[0], p.newPublicKeysHex[1], p.newPublicKeysHex[2], logId, p.TxUserID)
	if err != nil {
		return p.ErrInfo(err)
	}

	// проверим, не наш ли это user_id или не наш ли это паблик-ключ
	myUserId, myBlockId, myPrefix, _, err := p.GetMyUserId(p.TxUserID)
	if err != nil {
		return err
	}
	community, err := p.GetCommunityUsers()
	if err != nil {
		return err
	}
	var myPublicKey []byte
	if myUserId > 0 || len(community) == 0 {
		var err error
		// проверим, не наш ли это public_key, чтобы записать полученный user_id в my_table
		myPublicKey, err = p.Single("SELECT public_key FROM " + myPrefix + "my_keys WHERE id  =  (SELECT max(id) FROM " + myPrefix + "my_keys )").Bytes()
		myPublicKey = utils.BinToHex(myPublicKey)
		if err != nil {
			return p.ErrInfo(err)
		}
	}

	log.Debug("p.TxUserID: %d myUserId: %d myPublicKey: %s p.newPublicKeysHex[0]: %s myBlockId: %d p.BlockData.BlockId: %d", p.TxUserID, myUserId, myPublicKey, p.newPublicKeysHex[0], myBlockId, p.BlockData.BlockId)
	// возможна ситуация, когда юзер зарегался по уже занятому ключу. В этом случае тут будет новый ключ, а в my_keys не будет
	// my_user_id он уже успел заполучить в предыдущих блоках
	if p.TxUserID == myUserId && !bytes.Equal(myPublicKey, p.newPublicKeysHex[0]) && myBlockId <= p.BlockData.BlockId {
		err = p.ExecSql("UPDATE " + myPrefix + "my_table SET status = 'bad_key'")
		if err != nil {
			return p.ErrInfo(err)
		}
	} else if (p.TxUserID == myUserId || bytes.Equal(myPublicKey, p.newPublicKeysHex[0])) && myBlockId <= p.BlockData.BlockId {
		// если есть user_id, значит уже точно нету bad_key и в прошлых блоках уже было соотвествие my_key с ключем в new_public_keys_hex

		log.Debug("UPDATE " + myPrefix + "my_keys SET status = 'approved', block_id = ?, time = ? WHERE hex(public_key) = ? AND status = 'my_pending'")
		// обновим статус в нашей локальной табле.
		err = p.ExecSql("UPDATE "+myPrefix+"my_keys SET status = 'approved', block_id = ?, time = ? WHERE hex(public_key) = ? AND status = 'my_pending'", p.BlockData.BlockId, p.BlockData.Time, p.newPublicKeysHex[0])
		if err != nil {
			return p.ErrInfo(err)
		}

		// и если у нас в таблицах my_ ничего нет, т.к. мы только нашли соотвествие нашего ключа, то заносим все данные
		if len(myPublicKey) > 0 && myUserId == 0 {
			myUserId, err = p.Single("SELECT user_id FROM users WHERE hex(public_key_0) = ?", myPublicKey).Int64()
			if err != nil {
				return p.ErrInfo(err)
			}
			minersData, err := p.OneRow("SELECT * FROM miners_data WHERE user_id  =  ?", myUserId).String()
			if err != nil {
				return p.ErrInfo(err)
			}
			if len(minersData) > 0 {
				err = p.ExecSql("UPDATE "+myPrefix+"my_table SET user_id = ?, miner_id = ?, status = ?, face_coords = ?, profile_coords = ?, video_type = ?, video_url_id = ?, http_host = ?, tcp_host = ?, geolocation = ?, geolocation_status = 'approved' WHERE status != 'bad_key'",
					minersData["user_id"], minersData["miner_id"], minersData["status"], minersData["face_coords"], minersData["profile_coords"], minersData["video_type"], minersData["video_url_id"], minersData["http_host"], minersData["tcp_host"], minersData["latitude"]+", "+minersData["longitude"])
				if err != nil {
					return p.ErrInfo(err)
				}
			} else {
				err = p.ExecSql("UPDATE "+myPrefix+"my_table SET user_id = ?, status = 'user' WHERE status != 'bad_key'", myUserId)
				if err != nil {
					return p.ErrInfo(err)
				}
			}

			// cash_requests
			rows, err := p.Query(p.FormatQuery("SELECT to_user_id, currency_id, amount,  hash_code, status, id FROM cash_requests WHERE to_user_id = ? OR from_user_id = ?"), myUserId, myUserId)
			if err != nil {
				return p.ErrInfo(err)
			}
			defer rows.Close()
			for rows.Next() {
				var to_user_id, currency_id, amount, status, id string
				var hash_code []byte
				err = rows.Scan(&to_user_id, &currency_id, &amount, &hash_code, &status, &id)
				if err != nil {
					return p.ErrInfo(err)
				}
				err = p.ExecSql("INSERT INTO "+myPrefix+"cash_requests ( to_user_id, currency_id, amount, hash_code, status, cash_request_id ) VALUES ( ?, ?, ?, [hex], ?, ? )", to_user_id, currency_id, amount, utils.BinToHex(hash_code), status, id)
				if err != nil {
					return p.ErrInfo(err)
				}
			}

			//holidays
			rows, err = p.Query(p.FormatQuery("SELECT start_time, end_time, id FROM holidays WHERE user_id = ?"), myUserId)
			if err != nil {
				return p.ErrInfo(err)
			}
			defer rows.Close()
			for rows.Next() {
				var start_time, end_time, holidays_id string
				err = rows.Scan(&start_time, &end_time, &holidays_id)
				if err != nil {
					return p.ErrInfo(err)
				}
				err = p.ExecSql("INSERT INTO "+myPrefix+"my_holidays ( start_time, end_time, holidays_id) VALUES ( ?, ?, ? )", start_time, end_time, holidays_id)
				if err != nil {
					return p.ErrInfo(err)
				}
			}

		}

	}

	return nil
}

func (p *Parser) ChangePrimaryKeyRollback() error {

	// получим log_id, по которому можно найти данные, которые были до этого
	// $log_id всегда больше нуля, т.к. это откат обновления ключа
	logId, err := p.Single("SELECT log_id FROM users WHERE user_id  =  ?", p.TxUserID).Int64()
	if err != nil {
		return p.ErrInfo(err)
	}

	// данные, которые восстановим
	var public_key_0, public_key_1, public_key_2 []byte
	var prev_log_id int64
	err = p.QueryRow(p.FormatQuery("SELECT public_key_0, public_key_1, public_key_2, prev_log_id FROM log_users WHERE log_id  =  ?"), logId).Scan(&public_key_0, &public_key_1, &public_key_2, &prev_log_id)
	if err != nil && err != sql.ErrNoRows {
		return p.ErrInfo(err)
	}
	if len(public_key_0) > 0 {
		public_key_0 = utils.BinToHex(public_key_0)
	}
	if len(public_key_1) > 0 {
		public_key_1 = utils.BinToHex(public_key_1)
	}
	if len(public_key_2) > 0 {
		public_key_2 = utils.BinToHex(public_key_2)
	}

	err = p.ExecSql("UPDATE users SET public_key_0 = [hex], public_key_1 = [hex], public_key_2 = [hex], log_id = ? WHERE user_id = ?", public_key_0, public_key_1, public_key_2, prev_log_id, p.TxUserID)
	if err != nil {
		return p.ErrInfo(err)
	}
	// подчищаем _log
	err = p.ExecSql("DELETE FROM log_users WHERE log_id = ?", logId)
	if err != nil {
		return p.ErrInfo(err)
	}
	p.rollbackAI("log_users", 1)

	// проверим, не наш ли это user_id
	myUserId, _, myPrefix, _, err := p.GetMyUserId(p.TxUserID)
	if err != nil {
		return err
	}
	if p.TxUserID == myUserId {
		// обновим статус в нашей локальной табле.
		err = p.ExecSql("UPDATE "+myPrefix+"my_keys SET status = 'my_pending', block_id = 0, time = 0 WHERE hex(public_key) = ? AND status = 'approved' AND block_id = ?", p.newPublicKeysHex[0], p.BlockData.BlockId)
		if err != nil {
			return p.ErrInfo(err)
		}
	}

	return nil
}

func (p *Parser) ChangePrimaryKeyRollbackFront() error {
	return p.limitRequestsRollback("primary_key")
}
