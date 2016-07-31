// send_key
package controllers

import (
	"github.com/DayLightProject/go-daylight/packages/utils"
	"encoding/json"
	"encoding/base64"
	"io/ioutil"
	"path"
	"path/filepath"
)

func (c *Controller) SendKey() (string, error) {
	resval := false
	result := func(msg, data string, success bool ) (string, error) {
		res, err := json.Marshal( answerJson{Result:resval, Error: msg,
		                          Data: data, Success: success})
		return string(res), err
	}
	
	email := c.r.FormValue(`email`)
	params := map[string]string{ `user_id`: utils.Int64ToStr(c.SessUserId),
		`subject`: c.r.FormValue(`subject`), `text`: c.r.FormValue(`text`),
		`refid`: c.r.FormValue(`refuser`)}
    //PngKey    string    `json:"png_key"`
	
	txtkey, err := ioutil.ReadFile( filepath.Join(*utils.Dir,`public`, path.Base(c.r.FormValue(`keyurl`)))+`.txt`)
	if err != nil {
		return result(err.Error(), ``, false)
	}
	params[`txt_key`] = base64.StdEncoding.EncodeToString(txtkey)
	pngkey, err := ioutil.ReadFile( filepath.Join(*utils.Dir,`public`, path.Base(c.r.FormValue(`keyurl`)))+`.png`)
	if err != nil {
		return result(err.Error(), ``, false)
	}
	params[`png_key`] = base64.StdEncoding.EncodeToString(pngkey)
	err = utils.SendEmail( email, utils.EXCHANGE_USER, utils.ECMD_SENDKEY, &params )
	if err != nil {
		return result(err.Error(), ``, false)
	}
	return result( ``, `The email has been successfully sent.`, true )
}
