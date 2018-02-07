// MIT License
//
// Copyright (c) 2016 GenesisKernel
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.
package api

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/GenesisCommunity/go-genesis/packages/consts"
	"github.com/GenesisCommunity/go-genesis/packages/converter"
	"github.com/GenesisCommunity/go-genesis/packages/crypto"
)

var (
	gAuth             string
	gAddress          string
	gPrivate, gPublic string
)

type global struct {
	url   string
	value string
}

// PrivateToPublicHex returns the hex public key for the specified hex private key.
func PrivateToPublicHex(hexkey string) (string, error) {
	key, err := hex.DecodeString(hexkey)
	if err != nil {
		return ``, fmt.Errorf("Decode hex error")
	}
	pubKey, err := crypto.PrivateToPublic(key)
	if err != nil {
		return ``, err
	}
	return hex.EncodeToString(pubKey), nil
}

func sendRequest(rtype, url string, form *url.Values, v interface{}) error {
	client := &http.Client{}
	var ioform io.Reader
	if form != nil {
		ioform = strings.NewReader(form.Encode())
	}
	req, err := http.NewRequest(rtype, `http://localhost:7079`+consts.ApiPath+url, ioform)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	if len(gAuth) > 0 {
		req.Header.Set("Authorization", jwtPrefix+gAuth)
	}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}

	defer resp.Body.Close()
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf(`%d %s`, resp.StatusCode, strings.TrimSpace(string(data)))
	}

	if err = json.Unmarshal(data, v); err != nil {
		return err
	}
	return nil
}

func sendGet(url string, form *url.Values, v interface{}) error {
	return sendRequest("GET", url, form, v)
}

func sendPost(url string, form *url.Values, v interface{}) error {
	return sendRequest("POST", url, form, v)
}

func keyLogin(state int64) (err error) {
	var (
		key, sign []byte
	)

	key, err = ioutil.ReadFile(`key`)
	if err != nil {
		return
	}
	if len(key) > 64 {
		key = key[:64]
	}
	var ret getUIDResult
	err = sendGet(`getuid`, nil, &ret)
	if err != nil {
		return
	}
	gAuth = ret.Token
	if len(ret.UID) == 0 {
		return fmt.Errorf(`getuid has returned empty uid`)
	}

	var pub string

	sign, err = crypto.Sign(string(key), ret.UID)
	if err != nil {
		return
	}
	pub, err = PrivateToPublicHex(string(key))
	if err != nil {
		return
	}
	form := url.Values{"pubkey": {pub}, "signature": {hex.EncodeToString(sign)},
		`ecosystem`: {converter.Int64ToStr(state)}}
	var logret loginResult
	err = sendPost(`login`, &form, &logret)
	if err != nil {
		return
	}
	gAddress = logret.Address
	gPrivate = string(key)
	gPublic, err = PrivateToPublicHex(gPrivate)
	gAuth = logret.Token
	if err != nil {
		return
	}
	return
}

func getSign(forSign string) (string, error) {
	sign, err := crypto.Sign(gPrivate, forSign)
	if err != nil {
		return ``, err
	}
	return hex.EncodeToString(sign), nil
}

func getTestSign(forSign string) (string, error) {
	var ret signTestResult
	err := sendPost(`signtest`, &url.Values{"forsign": {forSign},
		"private": {gPrivate}}, &ret)
	if err != nil {
		return ``, err
	}
	return ret.Signature, nil
}

func appendSign(ret map[string]interface{}, form *url.Values) error {
	forsign := ret[`forsign`].(string)
	if ret[`signs`] != nil {
		for _, item := range ret[`signs`].([]interface{}) {
			v := item.(map[string]interface{})
			vsign, err := getSign(v[`forsign`].(string))
			if err != nil {
				return err
			}
			(*form)[v[`field`].(string)] = []string{vsign}
			forsign += `,` + vsign
		}
	}
	sign, err := getSign(forsign)
	if err != nil {
		return err
	}
	(*form)[`time`] = []string{ret[`time`].(string)}
	(*form)[`signature`] = []string{sign}
	return nil
}

func waitTx(hash string) (int64, error) {
	for i := 0; i < 15; i++ {
		var ret txstatusResult
		err := sendGet(`txstatus/`+hash, nil, &ret)
		if err != nil {
			return 0, err
		}
		if len(ret.BlockID) > 0 {
			return converter.StrToInt64(ret.BlockID), fmt.Errorf(ret.Result)
		}
		if ret.Message != nil {
			errtext, err := json.Marshal(ret.Message)
			if err != nil {
				return 0, err
			}
			return 0, errors.New(string(errtext))
		}
		time.Sleep(time.Second)
	}
	return 0, fmt.Errorf(`TxStatus timeout`)
}

func randName(prefix string) string {
	return fmt.Sprintf(`%s%d`, prefix, time.Now().Unix())
}

func postTxResult(txname string, form *url.Values) (id int64, msg string, err error) {
	ret := make(map[string]interface{})
	err = sendPost(`prepare/`+txname, form, &ret)
	if err != nil {
		return
	}
	if err = appendSign(ret, form); err != nil {
		return
	}
	ret = map[string]interface{}{}
	err = sendPost(`contract/`+txname, form, &ret)
	if err != nil {
		return
	}
	if len((*form)[`vde`]) > 0 {
		if ret[`result`] != nil {
			msg = fmt.Sprint(ret[`result`])
			id = converter.StrToInt64(msg)
		}
		return
	}
	id, err = waitTx(ret[`hash`].(string))
	if id != 0 && err != nil {
		msg = err.Error()
		err = nil
	}
	return
}

func RawToString(input json.RawMessage) string {
	out := strings.Trim(string(input), `"`)
	return strings.Replace(out, `\"`, `"`, -1)
}

func postTx(txname string, form *url.Values) error {
	_, _, err := postTxResult(txname, form)
	return err
}

func cutErr(err error) string {
	out := err.Error()
	if off := strings.IndexByte(out, '('); off != -1 {
		out = out[:off]
	}
	return strings.TrimSpace(out)
}
