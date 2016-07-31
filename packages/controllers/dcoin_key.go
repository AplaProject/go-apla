package controllers

import (
	"encoding/base64"
	"errors"
	"github.com/DayLightProject/go-daylight/packages/utils"
	"regexp"
	"strings"
	"fmt"
)

func (c *Controller) DcoinKey() (string, error) {

	fmt.Println("DcoinKey")
	var err error
	c.r.ParseForm()
	// на IOS/Android запрос ключа идет без сессии из objective C (UIImage *image = [UIImage imageWithData:[NSData dataWithContentsOfURL:[NSURL URLWithString:@"http://127.0.0.1:8089/ajax?controllerName=dcoinKey&ios=1"]]];)
	local := false
	// чтобы по локалке никто не украл приватный ключ
	if ok, _ := regexp.MatchString(`^(\:\:)|(127\.0\.0\.1)(:[0-9]+)?$`, c.r.RemoteAddr); ok {
		local = true
		fmt.Println("local = true")
	}
	if utils.Mobile() && c.SessUserId == 0 && !local {
		return "", utils.ErrInfo(errors.New("Not local request from " + c.r.RemoteAddr))
	}
	privKey := ""
	if len(c.r.FormValue("first")) > 0 {
		privKey, err = c.Single(`SELECT private_key FROM ` + c.MyPrefix + `my_keys WHERE status='my_pending'`).String()
		if err != nil {
			return "", utils.ErrInfo(err)
		}
	} else {
		privKey, _ = utils.GenKeys()
	}

	paramNoPass := utils.ParamType{X: 176, Y: 100, Width: 100, Bg_path: "static/img/k_bg.png"}
	paramPass := utils.ParamType{X: 167, Y: 93, Width: 118, Bg_path: "static/img/k_bg_pass.png"}

	var param utils.ParamType
	var privateKey string
	if len(c.r.FormValue("password")) > 0 {
		privateKey_, err := utils.Encrypt(utils.Md5(c.r.FormValue("password")), []byte(privKey))
		privateKey = base64.StdEncoding.EncodeToString(privateKey_)
		if err != nil {
			return "", utils.ErrInfo(err)
		}
		param = paramPass
	} else {
		privateKey = strings.Replace(privKey, "-----BEGIN RSA PRIVATE KEY-----", "", -1)
		privateKey = strings.Replace(privateKey, "-----END RSA PRIVATE KEY-----", "", -1)
		param = paramNoPass
	}

	ios := false
	if ok, _ := regexp.MatchString("(iPod|iPhone|iPad)", c.r.UserAgent()); ok {
		ios = true
	}
	if len(c.r.FormValue("ios")) > 0 {
		ios = true
	}

	if ios || utils.Android() {

		fmt.Println("DcoinKey image/png")
		buffer, err := utils.KeyToImg(privateKey, "", c.SessUserId, c.TimeFormat, param)
		if err != nil {
			return "", utils.ErrInfo(err)
		}
		c.w.Header().Set("Content-Type", "image/png")
		c.w.Header().Set("Content-Length", utils.IntToStr(len(buffer.Bytes())))
		c.w.Header().Set("Content-Disposition", `attachment; filename="Dcoin-private-key-`+utils.Int64ToStr(c.SessUserId)+`.png"`)
		if _, err := c.w.Write(buffer.Bytes()); err != nil {
			return "", utils.ErrInfo(errors.New("unable to write image"))
		}
	} else {
		c.w.Header().Set("Content-Type", "text/plain")
		c.w.Header().Set("Content-Length", utils.IntToStr(len(privateKey)))
		c.w.Header().Set("Content-Disposition", `attachment; filename="Dcoin-private-key-`+utils.Int64ToStr(c.SessUserId)+`.txt"`)
		if _, err := c.w.Write([]byte(privateKey)); err != nil {
			return "", utils.ErrInfo(errors.New("unable to write text"))
		}
	}
	fmt.Println("DcoinKey ok")

	return "", nil
}
