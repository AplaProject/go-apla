package dcparser

import (
	"encoding/json"
	"fmt"
	"github.com/DayLightProject/go-daylight/packages/utils"
)

func (p *Parser) AdminSpotsInit() error {
	fields := []string{"example_spots", "segments", "tolerances", "compatibility", "sign"}
	TxMap := make(map[string][]byte)
	TxMap, err := p.GetTxMap(fields)
	p.TxMap = TxMap
	if err != nil {
		return p.ErrInfo(err)
	}
	return nil
}

type exampleSpotsType struct {
	Face    map[string][]interface{} `json:"face"`
	Profile map[string][]interface{} `json:"profile"`
}
type tolerancesType struct {
	Face    map[string]string `json:"face"`
	Profile map[string]string `json:"profile"`
}

func (p *Parser) AdminSpotsFront() error {
	err := p.generalCheckAdmin()
	if err != nil {
		return p.ErrInfo(err)
	}

	if !utils.CheckInputData(p.TxMap["compatibility"], "compatibility") {
		return p.ErrInfo("incorrect compatibility")
	}
	exampleSpots := new(exampleSpotsType)
	if err := json.Unmarshal([]byte(p.TxMap["example_spots"]), &exampleSpots); err != nil {
		return p.ErrInfo("incorrect example_spots")
	}
	if exampleSpots.Face == nil || exampleSpots.Profile == nil {
		return p.ErrInfo("incorrect example_spots")
	}

	segments := new(exampleSpotsType)
	if err := json.Unmarshal([]byte(p.TxMap["segments"]), &segments); err != nil {
		return p.ErrInfo("incorrect segments")
	}
	if segments.Face == nil || segments.Profile == nil {
		return p.ErrInfo("incorrect segments")
	}
	tolerances := new(tolerancesType)
	if err := json.Unmarshal([]byte(p.TxMap["tolerances"]), &tolerances); err != nil {
		return p.ErrInfo("incorrect tolerances")
	}
	if tolerances.Face == nil || tolerances.Profile == nil {
		return p.ErrInfo("incorrect tolerances")
	}
	forSign := fmt.Sprintf("%s,%s,%s,%s,%s,%s,%s", p.TxMap["type"], p.TxMap["time"], p.TxMap["user_id"], p.TxMap["example_spots"], p.TxMap["segments"], p.TxMap["tolerances"], p.TxMap["compatibility"])
	CheckSignResult, err := utils.CheckSign(p.PublicKeys, forSign, p.TxMap["sign"], false)
	if err != nil {
		return p.ErrInfo(err)
	}
	if !CheckSignResult {
		return p.ErrInfo("incorrect sign")
	}
	return nil
}

func (p *Parser) AdminSpots() error {
	logData, err := p.OneRow("SELECT * FROM spots_compatibility").String()
	if err != nil {
		return p.ErrInfo(err)
	}
	logId, err := p.ExecSqlGetLastInsertId("INSERT INTO log_spots_compatibility ( version, example_spots, compatibility, segments, tolerances, block_id, prev_log_id ) VALUES ( ?, ?, ?, ?, ?, ?, ? )", "log_id", logData["version"], logData["example_spots"], logData["compatibility"], logData["segments"], logData["tolerances"], p.BlockData.BlockId, logData["log_id"])
	if err != nil {
		return p.ErrInfo(err)
	}
	// обновляем данные в рабочих таблицах
	err = p.ExecSql("UPDATE spots_compatibility SET version = version+1, example_spots = ?, compatibility = ?, segments = ?, tolerances = ?, log_id = ?", p.TxMap["example_spots"], p.TxMap["compatibility"], p.TxMap["segments"], p.TxMap["tolerances"], logId)
	if err != nil {
		return p.ErrInfo(err)
	}

	return nil
}

func (p *Parser) AdminSpotsRollback() error {
	logId, err := p.Single("SELECT log_id FROM spots_compatibility LIMIT 1").Int()
	if err != nil {
		return p.ErrInfo(err)
	}
	if logId > 0 {
		// данные, которые восстановим
		logData, err := p.OneRow("SELECT * FROM log_spots_compatibility WHERE log_id = ?", logId).String()
		if err != nil {
			return p.ErrInfo(err)
		}
		err = p.ExecSql("UPDATE spots_compatibility SET version = ?, example_spots = ?, compatibility = ?, segments = ?, tolerances = ?, log_id = ?", logData["version"], logData["example_spots"], logData["compatibility"], logData["segments"], logData["tolerances"], logData["prev_log_id"])
		if err != nil {
			return p.ErrInfo(err)
		}
		// подчищаем _log
		err = p.ExecSql("DELETE FROM log_spots_compatibility WHERE log_id = ?", logId)
		if err != nil {
			return p.ErrInfo(err)
		}
		p.rollbackAI("log_spots_compatibility", 1)
	}
	return nil
}

func (p *Parser) AdminSpotsRollbackFront() error {
	return nil
}
