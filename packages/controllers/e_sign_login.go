package controllers

import (
	"github.com/DayLightProject/go-daylight/packages/utils"
	"regexp"
)

func (c *Controller) ESignLogin() (string, error) {

	c.w.Header().Set("Access-Control-Allow-Origin", "*")

	var hash []byte
	loginCode := utils.RandSeq(20)

	RemoteAddr := utils.RemoteAddrFix(c.r.RemoteAddr)
	re := regexp.MustCompile(`(.*?):[0-9]+$`)
	match := re.FindStringSubmatch(RemoteAddr)
	if len(match) != 0 {
		RemoteAddr = match[1]
	}
	log.Debug("RemoteAddr %s", RemoteAddr)
	hash = utils.Md5(c.r.Header.Get("User-Agent") + RemoteAddr)
	log.Debug("hash %s", hash)

	err := c.DCDB.ExecSql(`DELETE FROM e_authorization WHERE hex(hash) = ?`, hash)
	if err != nil {
		return "", err
	}
	err = c.DCDB.ExecSql(`INSERT INTO e_authorization (hash, data) VALUES ([hex], ?)`, hash, loginCode)
	if err != nil {
		return "", err
	}
	log.Debug("loginCode %v", loginCode)
	return "\"" + loginCode + "\"", nil
}
