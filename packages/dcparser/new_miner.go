package dcparser

import (
	"encoding/json"
	"fmt"
	"github.com/DayLightProject/go-daylight/packages/utils"
	"regexp"
)

type exampleSpots struct {
	Face    map[string][]interface{} `json:"face"`
	Profile map[string][]interface{} `json:"profile"`
}

func (p *Parser) NewMinerInit() error {
	var fields []map[string]string
	if p.BlockData != nil && p.BlockData.BlockId < 250900 {
		fields = []map[string]string{{"race": "int64"}, {"country": "int64"}, {"latitude": "float64"}, {"longitude": "float64"}, {"http_host": "string"}, {"face_coords": "string"}, {"profile_coords": "string"}, {"face_hash": "string"}, {"profile_hash": "string"}, {"video_type": "string"}, {"video_url_id": "string"}, {"node_public_key": "bytes"}, {"sign": "bytes"}}
	} else if p.BlockData != nil && p.BlockData.BlockId < 281500 {
		fields = []map[string]string{{"race": "int64"}, {"country": "int64"}, {"latitude": "float64"}, {"longitude": "float64"}, {"http_host": "string"}, {"tcp_host": "string"}, {"face_coords": "string"}, {"profile_coords": "string"}, {"face_hash": "string"}, {"profile_hash": "string"}, {"video_type": "string"}, {"video_url_id": "string"}, {"node_public_key": "bytes"}, {"sign": "bytes"}}
	} else {
		fields = []map[string]string{{"race": "int64"}, {"country": "int64"}, {"latitude": "float64"}, {"longitude": "float64"}, {"http_host": "string"}, {"tcp_host": "string"}, {"face_coords": "string"}, {"profile_coords": "string"}, {"face_hash": "string"}, {"profile_hash": "string"}, {"video_type": "string"}, {"video_url_id": "string"}, {"node_public_key": "bytes"}, {"pool_user_id": "int64"}, {"sign": "bytes"}}
	}
	err := p.GetTxMaps(fields)
	if err != nil {
		return p.ErrInfo(err)
	}
	p.TxMap["node_public_key"] = utils.BinToHex(p.TxMap["node_public_key"])
	p.TxMaps.Bytes["node_public_key"] = utils.BinToHex(p.TxMaps.Bytes["node_public_key"])
	return nil
}

func (p *Parser) NewMinerFront() error {
	err := p.generalCheck()
	if err != nil {
		return p.ErrInfo(err)
	}
	// получим кол-во точек для face и profile
	exampleSpots_, err := p.DCDB.Single("SELECT example_spots FROM spots_compatibility").String()
	if err != nil {
		return p.ErrInfo(err)
	}

	exampleSpots := new(exampleSpots)
	err = json.Unmarshal([]byte(exampleSpots_), &exampleSpots)
	if err != nil {
		return p.ErrInfo(err)
	}

	if !utils.CheckInputData(p.TxMap["race"], "race") {
		return utils.ErrInfoFmt("race")
	}
	if !utils.CheckInputData(p.TxMap["country"], "country") {
		return utils.ErrInfoFmt("country")
	}
	if !utils.CheckInputData(p.TxMap["latitude"], "coordinate") {
		return utils.ErrInfoFmt("latitude")
	}
	if !utils.CheckInputData(p.TxMap["longitude"], "coordinate") {
		return utils.ErrInfoFmt("longitude")
	}
	if p.TxMaps.String["http_host"] != "0" && !utils.CheckInputData(p.TxMap["http_host"], "http_host") {
		return utils.ErrInfoFmt("http_host")
	}
	if p.BlockData == nil || p.BlockData.BlockId > 250900 {
		if p.TxMaps.String["tcp_host"] != "0" && !utils.CheckInputData(p.TxMap["tcp_host"], "tcp_host") {
			return utils.ErrInfoFmt("tcp_host")
		}
	}
	if p.BlockData == nil || p.BlockData.BlockId > 281500 {
		if p.TxMaps.String["pool_user_id"] != "0" && !utils.CheckInputData(p.TxMap["pool_user_id"], "int64") {
			return utils.ErrInfoFmt("pool_user_id")
		}
	}
	if !utils.CheckInputData_(p.TxMap["face_coords"], "coords", utils.IntToStr(len(exampleSpots.Face)-1)) {
		return utils.ErrInfoFmt("face_coords")
	}
	if !utils.CheckInputData_(p.TxMap["profile_coords"], "coords", utils.IntToStr(len(exampleSpots.Profile)-1)) {
		return utils.ErrInfoFmt("profile_coords")
	}
	if !utils.CheckInputData(p.TxMap["face_hash"], "photo_hash") {
		return utils.ErrInfoFmt("face_hash")
	}
	if !utils.CheckInputData(p.TxMap["profile_hash"], "photo_hash") {
		return utils.ErrInfoFmt("profile_hash")
	}
	if !utils.CheckInputData(p.TxMap["video_type"], "video_type") {
		return utils.ErrInfoFmt("video_type")
	}
	if !utils.CheckInputData(p.TxMap["video_url_id"], "video_url_id") {
		return utils.ErrInfoFmt("video_url_id %s", p.TxMap["video_url_id"])
	}
	if !utils.CheckInputData(p.TxMap["node_public_key"], "public_key") {
		return utils.ErrInfoFmt("node_public_key")
	}

	if (p.BlockData == nil || p.BlockData.BlockId > 281500) {
		// проверим, не занял ли кто-то хосты
		if p.TxMaps.String["http_host"]!="0" {
			exists, err := p.Single(`SELECT user_id FROM miners_data WHERE http_host = ?`, p.TxMaps.String["http_host"]).Int64()
			if err != nil {
				return p.ErrInfo(err)
			}
			if exists > 0 {
				return p.ErrInfo("host exists")
			}
		}
		if p.TxMaps.String["tcp_host"]!="0" {
			exists, err := p.Single(`SELECT user_id FROM miners_data WHERE tcp_host = ?`, p.TxMaps.String["tcp_host"]).Int64()
			if err != nil {
				return p.ErrInfo(err)
			}
			if exists > 0 {
				return p.ErrInfo("host exists")
			}
		}
		if (p.TxMaps.String["http_host"]=="0" || p.TxMaps.String["tcp_host"]=="0") && p.TxMaps.Int64["pool_user_id"] == 0 {
			return p.ErrInfo("host and pool empty")
		}
		if p.TxMaps.Int64["pool_user_id"] > 0 {
			// проверим, есть ли такой пул и есть ли места на выбранном пуле
			iAmPool, err := p.Single(`SELECT i_am_pool FROM miners_data WHERE user_id = ? AND pool_count_users < ?`, p.TxMaps.Int64["pool_user_id"], p.Variables.Int64["max_pool_users"]).Int64()
			if err != nil {
				return p.ErrInfo(err)
			}
			if iAmPool == 0 {
				return p.ErrInfo("incorrect pool_user_id")
			}
		}
	}

	forSign := ""
	if p.BlockData != nil && p.BlockData.BlockId < 250900 {
		forSign = fmt.Sprintf("%s,%s,%s,%s,%s,%s,%s,%s,%s,%s,%s,%s,%s,%s,%s", p.TxMap["type"], p.TxMap["time"], p.TxMap["user_id"], p.TxMap["race"], p.TxMap["country"], p.TxMap["latitude"], p.TxMap["longitude"], p.TxMap["http_host"], p.TxMap["face_hash"], p.TxMap["profile_hash"], p.TxMap["face_coords"], p.TxMap["profile_coords"], p.TxMap["video_type"], p.TxMap["video_url_id"], p.TxMap["node_public_key"])
	} else if p.BlockData != nil && p.BlockData.BlockId < 281500 {
		forSign = fmt.Sprintf("%s,%s,%s,%s,%s,%s,%s,%s,%s,%s,%s,%s,%s,%s,%s,%s", p.TxMap["type"], p.TxMap["time"], p.TxMap["user_id"], p.TxMap["race"], p.TxMap["country"], p.TxMap["latitude"], p.TxMap["longitude"], p.TxMap["http_host"], p.TxMap["tcp_host"], p.TxMap["face_hash"], p.TxMap["profile_hash"], p.TxMap["face_coords"], p.TxMap["profile_coords"], p.TxMap["video_type"], p.TxMap["video_url_id"], p.TxMap["node_public_key"])
	} else {
		forSign = fmt.Sprintf("%s,%s,%s,%s,%s,%s,%s,%s,%s,%s,%s,%s,%s,%s,%s,%s,%s", p.TxMap["type"], p.TxMap["time"], p.TxMap["user_id"], p.TxMap["race"], p.TxMap["country"], p.TxMap["latitude"], p.TxMap["longitude"], p.TxMap["http_host"], p.TxMap["tcp_host"], p.TxMap["face_hash"], p.TxMap["profile_hash"], p.TxMap["face_coords"], p.TxMap["profile_coords"], p.TxMap["video_type"], p.TxMap["video_url_id"], p.TxMap["node_public_key"], p.TxMap["pool_user_id"])
	}
	CheckSignResult, err := utils.CheckSign(p.PublicKeys, forSign, p.TxMap["sign"], false)
	if err != nil {
		return p.ErrInfo(err)
	}
	if !CheckSignResult {
		return utils.ErrInfoFmt("incorrect sign")
	}

	// проверим, не кончились ли попытки стать майнером у данного юзера
	count, err := p.CountMinerAttempt(p.TxUserID, "user_voting")
	if count >= p.Variables.Int64["miner_votes_attempt"] {
		return utils.ErrInfoFmt("miner_votes_attempt")
	}
	if err != nil {
		return p.ErrInfo(err)
	}
	//  на всякий случай не даем начать нодовское, если идет юзерское голосование
	userVoting, err := p.DCDB.Single("SELECT id FROM votes_miners WHERE user_id = ? AND type = 'user_voting' AND votes_end = 0", p.TxUserID).String()
	if err != nil {
		return p.ErrInfo(err)
	}
	if len(userVoting) > 0 {
		return utils.ErrInfoFmt("existing $user_voting")
	}

	// проверим, не является ли юзер майнером и  не разжалованный ли это бывший майнер
	minerStatus, err := p.DCDB.Single("SELECT status FROM miners_data WHERE user_id = ? AND status IN ('miner','passive_miner','suspended_miner')", p.TxUserID).String()
	if err != nil {
		return p.ErrInfo(err)
	}
	if len(minerStatus) > 0 {
		return utils.ErrInfoFmt("incorrect miner status")
	}

	// разрешен 1 запрос за сутки
	err = p.limitRequest(p.Variables.Int64["limit_new_miner"], "new_miner", p.Variables.Int64["limit_new_miner_period"])
	if err != nil {
		return p.ErrInfo(err)
	}

	return nil
}

func (p *Parser) NewMiner() error {
	tcpHost := ""
	if p.BlockData != nil && p.BlockData.BlockId < 250900 {
		re := regexp.MustCompile(`^https?:\/\/([0-9a-z\_\.\-:]+)\/`)
		match := re.FindStringSubmatch(p.TxMaps.String["http_host"])
		if len(match) != 0 {
			tcpHost = match[1] + ":8088"
		}
	} else {
		tcpHost = p.TxMaps.String["tcp_host"]
	}
	// получим массив майнеров, которые должны скопировать к себе 2 фото лица юзера
	maxMinerId, err := p.DCDB.Single("SELECT max(miner_id) FROM miners").Int64()
	if err != nil {
		return p.ErrInfo(err)
	}
	// т.к. у юзера это может быть не первая попытка стать майнером, то отменяем голосования по всем предыдущим
	err = p.DCDB.ExecSql("UPDATE votes_miners SET votes_end = 1, end_block_id = ? WHERE user_id = ? AND type = 'node_voting' AND end_block_id = 0 AND votes_end = 0", p.BlockData.BlockId, p.TxUserID)
	if err != nil {
		return p.ErrInfo(err)
	}
	// создаем новое голосование для нодов
	err = p.DCDB.ExecSql("INSERT INTO votes_miners (type,	user_id,	votes_start_time) VALUES ('node_voting', ?, ?)", p.TxUserID, p.BlockData.Time)
	if err != nil {
		return p.ErrInfo(err)
	}

	// переведем все координаты в отрезки.
	var faceCoords [][2]int
	err = json.Unmarshal(p.TxMap["face_coords"], &faceCoords)
	if err != nil {
		return p.ErrInfo(err)
	}
	faceCoords = append([][2]int{{0, 0}}, faceCoords...)

	// получим отрезки
	data, err := p.DCDB.OneRow("SELECT segments, version FROM spots_compatibility").String()
	if err != nil {
		return p.ErrInfo(err)
	}
	spotsVersion := data["version"]

	var segments map[string]map[string][]string
	err = json.Unmarshal([]byte(data["segments"]), &segments)
	if err != nil {
		return p.ErrInfo(err)
	}
	n := len(segments["face"]) + 1
	faceRelations := make([]float64, n, n)
	faceRelations[0] = utils.PpLenght(faceCoords[1], faceCoords[2])

	for num, spots := range segments["face"] {
		// 1. ширина головы
		// 2. глаз-нос
		// 3. нос-губа
		// 4. губа-подбородок
		// 5. ширина челюсти
		faceRelations[utils.StrToInt(num)] = utils.Round((utils.PpLenght(faceCoords[utils.StrToInt(spots[0])], faceCoords[utils.StrToInt(spots[1])]) / faceRelations[0]), 4)
	}
	faceRelations[0] = 1

	// переведем все координаты в отрезки.
	var profileCoords [][2]int
	err = json.Unmarshal(p.TxMap["profile_coords"], &profileCoords)
	if err != nil {
		return p.ErrInfo(err)
	}
	profileCoords = append([][2]int{{0, 0}}, profileCoords...)

	n = len(segments["profile"]) + 1
	profileRelations := make([]float64, n, n)
	profileRelations[0] = utils.PpLenght(profileCoords[1], profileCoords[2])

	for num, spots := range segments["profile"] {
		// 1. край уха - край носа
		// 2. глаз - край носа
		// 3. подбородок - низ уха
		// 4. верх уха - низ уха
		profileRelations[utils.StrToInt(num)] = utils.Round((utils.PpLenght(profileCoords[utils.StrToInt(spots[0])], profileCoords[utils.StrToInt(spots[1])]) / profileRelations[0]), 4)
	}
	profileRelations[0] = 1

	addSql := make(map[string]string)
	addSql["names"] = ""
	addSql["values"] = ""
	addSql["upd"] = ""
	for j := 1; j < len(faceRelations); j++ {
		addSql["names"] += fmt.Sprintf("f%v,\n", j)
		addSql["values"] += fmt.Sprintf("'%v',\n", faceRelations[j])
		addSql["upd"] += fmt.Sprintf("f%v='%v',\n", j, faceRelations[j])
	}
	for j := 1; j < len(profileRelations); j++ {
		addSql["names"] += fmt.Sprintf("p%v,\n", j)
		addSql["values"] += fmt.Sprintf("'%v',\n", profileRelations[j])
		addSql["upd"] += fmt.Sprintf("p%v='%v',\n", j, profileRelations[j])
	}
	addSql["names"] = addSql["names"][0 : len(addSql["names"])-2]
	addSql["values"] = addSql["values"][0 : len(addSql["values"])-2]
	addSql["upd"] = addSql["upd"][0 : len(addSql["upd"])-2]

	// Для откатов
	// проверим, есть ли в БД запись, которую нужно залогировать
	logData, err := p.DCDB.OneRow("SELECT * FROM faces WHERE user_id = ?", p.TxUserID).String()
	if err != nil {
		return p.ErrInfo(err)
	}
	if len(logData) > 0 {
		addSql1 := ""
		addSql2 := ""
		for i := 1; i <= 20; i++ {
			addSql1 += fmt.Sprintf("f%v, ", i)
			addSql2 += fmt.Sprintf("%v,", logData[fmt.Sprintf("f%v", i)])
		}
		for i := 1; i <= 20; i++ {
			addSql1 += fmt.Sprintf("p%v, ", i)
			addSql2 += fmt.Sprintf("%v,", logData[fmt.Sprintf("p%v", i)])
		}
		// для откатов
		logId, err := p.ExecSqlGetLastInsertId(`
			INSERT INTO log_faces (
					user_id,
					version,
					status,
					race,
					country,
					`+addSql1+`
					prev_log_id,
					block_id
				) VALUES (
					`+logData["user_id"]+`,
					`+logData["version"]+`,
					'`+logData["status"]+`',
					`+logData["race"]+`,
					`+logData["country"]+`,
					`+addSql2+`
					`+logData["log_id"]+`,
					`+utils.Int64ToStr(p.BlockData.BlockId)+`
				)`, "log_id")
		if err != nil {
			return p.ErrInfo(err)
		}
		err = p.ExecSql("UPDATE faces SET "+addSql["upd"]+", version = ?, race = ?, country = ?, log_id = ? WHERE user_id = ?", spotsVersion, p.TxMaps.Int64["race"], p.TxMaps.Int64["country"], logId, p.TxUserID)
		if err != nil {
			return p.ErrInfo(err)
		}
	} else {
		// это первая запись в таблицу, и лог писать не с чего
		err = p.ExecSql(`
					INSERT INTO faces (
						user_id,
						version,
						race,
						country,
						` + addSql["names"] + `
					) VALUES (
						` + string(p.TxMap["user_id"]) + `,
						'` + spotsVersion + `',
						` + string(p.TxMap["race"]) + `,
						` + string(p.TxMap["country"]) + `,
						` + addSql["values"] + `
					)`)
		if err != nil {
			return p.ErrInfo(err)
		}
	}

	// проверим, есть ли в БД запись, которую надо залогировать
	logData, err = p.OneRow("SELECT * FROM miners_data WHERE user_id = ?", p.TxUserID).String()
	if err != nil {
		return p.ErrInfo(err)
	}
	if len(logData) > 0 {
		logData["node_public_key"] = string(utils.BinToHex([]byte(logData["node_public_key"])))
		// для откатов
		logId, err := p.ExecSqlGetLastInsertId(`
				INSERT INTO log_miners_data (
					user_id,
					miner_id,
					status,
					node_public_key,
					face_hash,
					profile_hash,
					photo_block_id,
					photo_max_miner_id,
					miners_keepers,
					face_coords,
					profile_coords,
					video_type,
					video_url_id,
					http_host,
					tcp_host,
					latitude,
					longitude,
					country,
					pool_user_id,
					block_id,
					prev_log_id
				) VALUES (
					?, ?, ?, [hex], ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?
				) `, "log_id", logData["user_id"], logData["miner_id"], logData["status"], logData["node_public_key"], logData["face_hash"], logData["profile_hash"], logData["photo_block_id"], logData["photo_max_miner_id"], logData["miners_keepers"], logData["face_coords"], logData["profile_coords"], logData["video_type"], logData["video_url_id"], logData["http_host"], logData["tcp_host"], logData["latitude"], logData["longitude"], logData["country"], p.TxMaps.Int64["pool_user_id"], p.BlockData.BlockId, logData["log_id"])
		if err != nil {
			return p.ErrInfo(err)
		}
		// обновляем таблу
		err = p.ExecSql(`
				UPDATE miners_data
				SET
					node_public_key = [hex],
					face_hash = ?,
					profile_hash = ?,
					photo_block_id = ?,
					photo_max_miner_id = ?,
					miners_keepers = ?,
					face_coords = ?,
					profile_coords = ?,
					video_type = ?,
					video_url_id = ?,
					latitude = ?,
					longitude = ?,
					country = ?,
					http_host = ?,
					tcp_host = ?,
					pool_user_id = ?,
					log_id = ?
				WHERE user_id = ?`, p.TxMap["node_public_key"], p.TxMaps.String["face_hash"], p.TxMaps.String["profile_hash"], p.BlockData.BlockId, maxMinerId, p.Variables.Int64["miners_keepers"], p.TxMaps.String["face_coords"], p.TxMaps.String["profile_coords"], p.TxMaps.String["video_type"], p.TxMaps.String["video_url_id"], p.TxMaps.Float64["latitude"], p.TxMaps.Float64["longitude"], p.TxMaps.Int64["country"], p.TxMaps.String["http_host"], tcpHost, p.TxMaps.Int64["pool_user_id"], logId, p.TxUserID)
		if err != nil {
			return p.ErrInfo(err)
		}
	} else {
		err = p.ExecSql(`
				INSERT INTO miners_data (
					user_id,
					node_public_key,
					face_hash,
					profile_hash,
					photo_block_id,
					photo_max_miner_id,
					miners_keepers,
					face_coords,
					profile_coords,
					video_type,
					video_url_id,
					latitude,
					longitude,
					country,
					http_host,
					tcp_host,
					pool_user_id
			) VALUES (?, [hex], ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
			p.TxUserID, p.TxMap["node_public_key"], p.TxMaps.String["face_hash"], p.TxMaps.String["profile_hash"], p.BlockData.BlockId, maxMinerId, p.Variables.Int64["miners_keepers"], p.TxMaps.String["face_coords"], p.TxMaps.String["profile_coords"], p.TxMaps.String["video_type"], p.TxMaps.String["video_url_id"], p.TxMaps.Float64["latitude"], p.TxMaps.Float64["longitude"], p.TxMaps.Int64["country"], p.TxMaps.String["http_host"], tcpHost, p.TxMaps.Int64["pool_user_id"])
		if err != nil {
			return p.ErrInfo(err)
		}
	}
	// проверим, не наш ли это user_id
	myUserId, myBlockId, myPrefix, _, err := p.GetMyUserId(p.TxUserID)
	if err != nil {
		return err
	}
	if p.TxUserID == myUserId && myBlockId <= p.BlockData.BlockId {
		err = p.DCDB.ExecSql("UPDATE "+myPrefix+"my_node_keys SET block_id = ? WHERE hex(public_key) = ?", p.BlockData.BlockId, p.TxMap["node_public_key"])
		if err != nil {
			return p.ErrInfo(err)
		}
	}
	if p.TxMaps.Int64["pool_user_id"] > 0 {
		err = p.ExecSql(`UPDATE miners_data SET pool_count_users = pool_count_users + 1 WHERE user_id = ?`, p.TxMaps.Int64["pool_user_id"])
		if err != nil {
			return p.ErrInfo(err)
		}
	}
	return nil
}

func (p *Parser) NewMinerRollback() error {
	err := p.generalRollback("faces", p.TxUserID, "", false)
	if err != nil {
		return p.ErrInfo(err)
	}
	err = p.generalRollback("miners_data", p.TxUserID, "", false)
	if err != nil {
		return p.ErrInfo(err)
	}
	// votes_miners
	p.ExecSql(`UPDATE votes_miners
					SET votes_end = 0, end_block_id = 0
					WHERE user_id = ? AND type = 'node_voting' AND end_block_id = ? AND votes_end > 0`,
		p.TxUserID, p.BlockData.BlockId)
	p.ExecSql(`DELETE FROM votes_miners
					WHERE type = 'node_voting' AND user_id = ? AND votes_start_time = ?`, p.TxUserID, p.BlockData.Time)
	p.rollbackAI("votes_miners", 1)
	// проверим, не наш ли это user_id
	myUserId, _, myPrefix, _, err := p.GetMyUserId(p.TxUserID)
	if err != nil {
		return err
	}
	if p.TxUserID == myUserId {
		pub, err := p.Single("SELECT public_key FROM "+myPrefix+"my_node_keys WHERE block_id=?", p.BlockData.BlockId).String()
		if err != nil {
			return err
		}
		if len(pub) > 0 {
			err = p.ExecSql("UPDATE "+myPrefix+"my_node_keys SET block_id = 0 WHERE block_id = ?", p.BlockData.BlockId)
			if err != nil {
				return err
			}
		}
	}
	if p.TxMaps.Int64["pool_user_id"] > 0 {
		err = p.ExecSql(`UPDATE miners_data SET pool_count_users = pool_count_users - 1 WHERE user_id = ?`, p.TxMaps.Int64["pool_user_id"])
		if err != nil {
			return p.ErrInfo(err)
		}
	}
	return nil
}

func (p *Parser) NewMinerRollbackFront() error {
	return p.limitRequestsRollback("new_miner")
}
