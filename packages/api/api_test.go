// Copyright 2016 The go-daylight Authors
// This file is part of the go-daylight library.
//
// The go-daylight library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-daylight library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-daylight library. If not, see <http://www.gnu.org/licenses/>.

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

	"github.com/GenesisKernel/go-genesis/packages/consts"
	"github.com/GenesisKernel/go-genesis/packages/converter"
	"github.com/GenesisKernel/go-genesis/packages/crypto"
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
