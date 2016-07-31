package controllers

import (
	"errors"
	"github.com/DayLightProject/go-daylight/packages/utils"
	"io/ioutil"
	"net"
	"os"
	"time"
)

func (c *Controller) SendToPool() (string, error) {

	if c.SessRestricted != 0 {
		return "", utils.ErrInfo(errors.New("Permission denied"))
	}

	filesSign := c.r.FormValue("filesSign")

	poolUid := utils.StrToInt64(c.r.FormValue("poolUid"))
	data_, err := c.OneRow(`SELECT CASE WHEN m.pool_user_id > 0 then (SELECT tcp_host FROM miners_data WHERE user_id = m.pool_user_id) ELSE tcp_host end as tcp_host, CASE WHEN m.pool_user_id > 0 then (SELECT http_host FROM miners_data WHERE user_id = m.pool_user_id) ELSE http_host end as http_host FROM miners_data as m WHERE m.user_id = ?`, poolUid).String()
	tcpHost := data_["tcp_host"]
	httpHost := data_["http_host"]
	if err != nil {
		return "", utils.ErrInfo(err)
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
	data = append(data, utils.EncodeLengthPlusData(filesSign)...)

	if _, err := os.Stat(*utils.Dir + "/public/" + utils.Int64ToStr(c.SessUserId) + "_user_face.jpg"); err == nil {
		file, err := ioutil.ReadFile(*utils.Dir + "/public/" + utils.Int64ToStr(c.SessUserId) + "_user_face.jpg")
		if err != nil {
			return "", utils.ErrInfo(err)
		}
		data = append(data, utils.EncodeLengthPlusData(append(utils.DecToBin(0, 1), file...))...)
	}
	if _, err := os.Stat(*utils.Dir + "/public/" + utils.Int64ToStr(c.SessUserId) + "_user_profile.jpg"); err == nil {
		file, err := ioutil.ReadFile(*utils.Dir + "/public/" + utils.Int64ToStr(c.SessUserId) + "_user_profile.jpg")
		if err != nil {
			return "", utils.ErrInfo(err)
		}
		data = append(data, utils.EncodeLengthPlusData(append(utils.DecToBin(1, 1), file...))...)
	}
	if _, err := os.Stat(*utils.Dir + "/public/" + utils.Int64ToStr(c.SessUserId) + "_user_video.mp4"); err == nil {
		file, err := ioutil.ReadFile(*utils.Dir + "/public/" + utils.Int64ToStr(c.SessUserId) + "_user_video.mp4")
		if err != nil {
			return "", utils.ErrInfo(err)
		}
		data = append(data, utils.EncodeLengthPlusData(append(utils.DecToBin(2, 1), file...))...)
	}

	// тип данных
	_, err = conn.Write(utils.DecToBin(11, 2))
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
		err = c.ExecSql(`UPDATE `+c.MyPrefix+`my_table SET pool_user_id = ?, tcp_host = ?, http_host = ?`, poolUid, tcpHost, httpHost)
		if err != nil {
			return "", utils.ErrInfo(err)
		}
	} else {
		result = utils.JsonAnswer("error", "error").String()
	}

	return result, nil
}
