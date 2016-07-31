package controllers

import (
	"errors"
	"github.com/DayLightProject/go-daylight/packages/utils"
	"io/ioutil"
	"net"
	"os"
	"time"
)

func (c *Controller) SendPromisedAmountToPool() (string, error) {

	if c.SessRestricted != 0 {
		return "", utils.ErrInfo(errors.New("Permission denied"))
	}

	filesSign := c.r.FormValue("filesSign")
	currencyId := utils.StrToInt64(c.r.FormValue("currencyId"))
	miners_data, err := c.OneRow(`SELECT tcp_host, pool_user_id FROM miners_data WHERE user_id = ?`, c.SessUserId).String()
	if err != nil {
		return "", utils.ErrInfo(err)
	}
	tcpHost := miners_data["tcp_host"]
	if miners_data["pool_user_id"] != "0" {
		tcpHost, err = c.Single(`SELECT tcp_host FROM miners_data WHERE user_id = ?`, miners_data["pool_user_id"]).String()
		if err != nil {
			return "", utils.ErrInfo(err)
		}
	}
	conn, err := net.DialTimeout("tcp", tcpHost, 5*time.Second)
	if err != nil {
		return "", utils.ErrInfo(err)
	}
	defer conn.Close()

	conn.SetReadDeadline(time.Now().Add(240 * time.Second))
	conn.SetWriteDeadline(time.Now().Add(240 * time.Second))

	var data []byte
	data = append(data, utils.DecToBin(c.SessUserId, 5)...)
	data = append(data, utils.DecToBin(currencyId, 1)...)
	data = append(data, utils.EncodeLengthPlusData(filesSign)...)

	if _, err := os.Stat(*utils.Dir + "/public/" + utils.Int64ToStr(c.SessUserId) + "_promised_amount_" + utils.Int64ToStr(currencyId) + ".mp4"); err == nil {
		file, err := ioutil.ReadFile(*utils.Dir + "/public/" + utils.Int64ToStr(c.SessUserId) + "_promised_amount_" + utils.Int64ToStr(currencyId) + ".mp4")
		if err != nil {
			return "", utils.ErrInfo(err)
		}
		data = append(data, utils.EncodeLengthPlusData(append(utils.DecToBin(2, 1), file...))...)
	}

	// тип данных
	_, err = conn.Write(utils.DecToBin(12, 2))
	if err != nil {
		return "", utils.ErrInfo(err)
	}

	// в 4-х байтах пишем размер данных, которые пошлем далее
	size := utils.DecToBin(len(data), 4)
	_, err = conn.Write(size)
	if err != nil {
		return "", utils.ErrInfo(err)
	}

	// далее шлем сами данные
	_, err = conn.Write([]byte(data))
	if err != nil {
		return "", utils.ErrInfo(err)
	}

	// в ответ получаем статус
	buf := make([]byte, 1)
	_, err = conn.Read(buf)
	if err != nil {
		return "", utils.ErrInfo(err)
	}
	status := utils.BinToDec(buf)
	result := ""
	if status == 1 {
		result = utils.JsonAnswer("1", "success").String()
	} else {
		result = utils.JsonAnswer("error", "error").String()
	}

	return result, nil
}
