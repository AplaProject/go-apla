package dcparser

import (
	"fmt"
	"github.com/DayLightProject/go-daylight/packages/utils"
	"time"
)

func (p *Parser) NewHolidaysInit() error {
	fields := []map[string]string{{"start_time": "int64"}, {"end_time": "int64"}, {"sign": "bytes"}}
	err := p.GetTxMaps(fields)
	if err != nil {
		return p.ErrInfo(err)
	}
	return nil
}

func (p *Parser) NewHolidaysFront() error {

	err := p.generalCheck()
	if err != nil {
		return p.ErrInfo(err)
	}

	verifyData := map[string]string{"start_time": "bigint", "end_time": "bigint"}
	err = p.CheckInputData(verifyData)
	if err != nil {
		return p.ErrInfo(err)
	}

	// является ли данный юзер майнером
	err = p.checkMiner(p.TxUserID)
	if err != nil {
		return p.ErrInfo(err)
	}

	forSign := fmt.Sprintf("%s,%s,%s,%s,%s", p.TxMap["type"], p.TxMap["time"], p.TxMap["user_id"], p.TxMap["start_time"], p.TxMap["end_time"])
	CheckSignResult, err := utils.CheckSign(p.PublicKeys, forSign, p.TxMap["sign"], false)
	if err != nil {
		return p.ErrInfo(err)
	}
	if !CheckSignResult {
		return p.ErrInfo("incorrect sign")
	}
	if p.TxMaps.Int64["start_time"] >= p.TxMaps.Int64["end_time"] {
		return p.ErrInfo("start_time >= end_time")
	}

	var txTime int64
	if p.BlockData != nil {
		txTime = p.BlockData.Time
	} else {
		// если каникулы попадут в один блок с cash_requet_out и у каникул будет время начала равно времени блока, то будет ошибка. Делаем запас 1 час
		//у голой тр-ии проверка идет жестче
		txTime = time.Now().Unix() + 3600
	}
	if p.TxMaps.Int64["start_time"] <= txTime {
		return p.ErrInfo("start_time <= txTime")
	}

	// допустим отпуск не более чем на X дней.
	if p.TxMaps.Int64["end_time"]-p.TxMaps.Int64["start_time"] > p.Variables.Int64["holidays_max"] {
		return p.ErrInfo("end_time - start_time > holidays_max")
	}

	// проверяем, чтобы не было перекрывания
	num, err := p.Single("SELECT id FROM holidays WHERE user_id  =  ? AND del  =  0 AND ( start_time < ? AND end_time > ? )", p.TxUserID, p.TxMaps.Int64["end_time"], p.TxMaps.Int64["start_time"]).Int64()
	if err != nil {
		return p.ErrInfo(err)
	}
	if num > 0 {
		return p.ErrInfo("cross time")
	}

	// У юзера должно либо вообще не быть cash_requests, либо должен быть последний со статусом approved. Иначе у него заморожен весь майнинг
	err = p.CheckCashRequests(p.TxUserID)
	if err != nil {
		return p.ErrInfo(err)
	}

	// добавлять можно не более X запросов на добавление и удаление выходных за неделю
	err = p.limitRequest(p.Variables.Int64["limit_holidays"], "holidays", p.Variables.Int64["limit_holidays_period"])
	if err != nil {
		return p.ErrInfo(err)
	}

	return nil
}
func (p *Parser) NewHolidays() error {
	//fmt.Println("TxMap", p.TxMap)
	//var myUserIds []int64;
	err := p.ExecSql(`INSERT INTO holidays (user_id, start_time,end_time) VALUES (?, ?, ?)`,
		p.TxUserID, p.TxMaps.Int64["start_time"], p.TxMaps.Int64["end_time"])
	if err != nil {
		return err
	}
	// проверим, не наш ли это user_id
	myUserId, myBlockId, myPrefix, _, err := p.GetMyUserId(p.TxUserID)
	if err != nil {
		return err
	}
	//fmt.Println(myUserIds)
	if p.TxUserID == myUserId && myBlockId <= p.BlockData.BlockId {
		// обновим статус в нашей локальной табле
		err := p.ExecSql("DELETE FROM "+myPrefix+"my_holidays WHERE start_time=? AND end_time=?", p.TxMaps.Int64["start_time"], p.TxMaps.Int64["end_time"])
		if err != nil {
			return err
		}
	}
	return nil
}
func (p *Parser) NewHolidaysRollback() error {
	//fmt.Println(p.TxMap)
	err := p.ExecSql("DELETE FROM holidays WHERE user_id=? AND start_time=? AND end_time=?", p.TxUserID, p.TxMaps.Int64["start_time"], p.TxMaps.Int64["end_time"])

	if err != nil {
		return utils.ErrInfo(err)
	}

	err = p.rollbackAI("holidays", 1)
	if err != nil {
		return utils.ErrInfo(err)
	}
	return err
}

func (p *Parser) NewHolidaysRollbackFront() error {
	return p.limitRequestsRollback("holidays")
}
