// Copyright (C) 2017, 2018, 2019 EGAAS S.A.
//
// This program is free software; you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation; either version 2 of the License, or (at
// your option) any later version.
//
// This program is distributed in the hope that it will be useful, but
// WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the GNU
// General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program; if not, write to the Free Software
// Foundation, Inc., 51 Franklin Street, Fifth Floor, Boston, MA 02110-1301, USA.

package api

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/AplaProject/go-apla/packages/conf"
	"github.com/AplaProject/go-apla/packages/converter"
	"github.com/AplaProject/go-apla/packages/crypto"
	"github.com/AplaProject/go-apla/packages/utils/tx"
)

type Connect struct {
	Auth       string
	Root       string
	PrivateKey []byte
	PublicKey  string
}

type WaitResult struct {
	BlockID int64
	Msg     string
}

func SendRawRequest(rtype, url, auth string, form *url.Values) ([]byte, error) {
	client := &http.Client{}
	var ioform io.Reader
	if form != nil {
		ioform = strings.NewReader(form.Encode())
	}
	req, err := http.NewRequest(rtype, url, ioform)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	if len(auth) > 0 {
		req.Header.Set("Authorization", jwtPrefix+auth)
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf(`%d %s`, resp.StatusCode, strings.TrimSpace(string(data)))
	}

	return data, nil
}

func SendRequest(rtype, url, auth string, form *url.Values, v interface{}) error {
	data, err := SendRawRequest(rtype, url, auth, form)
	if err != nil {
		return err
	}

	return json.Unmarshal(data, v)
}

func (connect *Connect) SendGet(url string, form *url.Values, v interface{}) error {
	return SendRequest("GET", connect.Root+url, connect.Auth, form, v)
}

func (connect *Connect) SendPost(url string, form *url.Values, v interface{}) error {
	return SendRequest("POST", connect.Root+url, connect.Auth, form, v)
}

func (connect *Connect) SendMultipart(url string, files map[string][]byte, v interface{}) error {
	body := new(bytes.Buffer)
	writer := multipart.NewWriter(body)

	for key, data := range files {
		part, err := writer.CreateFormFile(key, key)
		if err != nil {
			return err
		}
		if _, err := part.Write(data); err != nil {
			return err
		}
	}

	if err := writer.Close(); err != nil {
		return err
	}

	req, err := http.NewRequest("POST", connect.Root+url, body)
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", writer.FormDataContentType())

	if len(connect.Auth) > 0 {
		req.Header.Set("Authorization", jwtPrefix+connect.Auth)
	}

	client := &http.Client{}
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

	return json.Unmarshal(data, &v)
}

func (connect *Connect) WaitTx(hash string) (int64, error) {
	data, err := json.Marshal(&txstatusRequest{
		Hashes: []string{hash},
	})
	if err != nil {
		return 0, err
	}

	for i := 0; i < 15; i++ {
		var multiRet multiTxStatusResult
		err := connect.SendPost(`txstatus`, &url.Values{
			"data": {string(data)},
		}, &multiRet)
		if err != nil {
			return 0, err
		}

		ret := multiRet.Results[hash]

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

func (connect *Connect) WaitTxList(hashes []string) (map[string]WaitResult, error) {
	data, err := json.Marshal(&txstatusRequest{
		Hashes: hashes,
	})
	if err != nil {
		return nil, err
	}
	var multiRet multiTxStatusResult
	err = connect.SendPost(`txstatus`, &url.Values{
		"data": {string(data)},
	}, &multiRet)
	if err != nil {
		return nil, err
	}
	waitResults := map[string]WaitResult{}
	for key, ret := range multiRet.Results {
		if len(ret.BlockID) > 0 {
			waitResults[key] = WaitResult{
				BlockID: converter.StrToInt64(ret.BlockID),
				Msg:     ret.Result,
			}
			continue
		}
		if ret.Message != nil {
			var msg string
			errtext, err := json.Marshal(ret.Message)
			if err != nil {
				msg = err.Error()
			} else {
				msg = string(errtext)
			}
			waitResults[key] = WaitResult{
				Msg: msg,
			}
		}
	}
	return waitResults, nil
}

func (connect *Connect) PostTxResult(name string, form *url.Values) (id int64, msg string, err error) {
	var contract getContractResult
	if err = connect.SendGet("contract/"+name, nil, &contract); err != nil {
		return
	}
	params := make(map[string]interface{})
	for _, field := range contract.Fields {
		name := field.Name
		value := form.Get(name)

		if len(value) == 0 {
			continue
		}

		switch field.Type {
		case "bool":
			params[name], err = strconv.ParseBool(value)
		case "int", "address":
			params[name], err = strconv.ParseInt(value, 10, 64)
		case "float":
			params[name], err = strconv.ParseFloat(value, 64)
		case "array":
			var v interface{}
			err = json.Unmarshal([]byte(value), &v)
			params[name] = v
		case "map":
			var v map[string]interface{}
			err = json.Unmarshal([]byte(value), &v)
			params[name] = v
		case "string", "money":
			params[name] = value
		}

		if err != nil {
			err = fmt.Errorf("Parse param '%s': %s", name, err)
			return
		}
	}

	var publicKey []byte
	if publicKey, err = crypto.PrivateToPublic(connect.PrivateKey); err != nil {
		return
	}
	txTime := time.Now().Unix()

	if newTime := form.Get("txtime"); len(newTime) > 0 {
		txTime = converter.StrToInt64(newTime)
	}

	data, _, err := tx.NewTransaction(tx.SmartContract{
		Header: tx.Header{
			ID:          int(contract.ID),
			Time:        txTime,
			EcosystemID: 1,
			KeyID:       crypto.Address(publicKey),
			NetworkID:   conf.Config.NetworkID,
		},
		Params: params,
	}, connect.PrivateKey)
	if err != nil {
		return 0, "", err
	}

	ret := &sendTxResult{}
	err = connect.SendMultipart("sendTx", map[string][]byte{
		"data": data,
	}, &ret)
	if err != nil {
		return
	}
	if len(form.Get("nowait")) > 0 {
		msg = ret.Hashes["data"]
		return
	}
	id, err = connect.WaitTx(ret.Hashes["data"])
	if id != 0 && err != nil {
		msg = err.Error()
		err = nil
	}

	return
}

func (connect *Connect) Login() error {
	var (
		sign []byte
		ret  getUIDResult
		err  error
	)
	if err = connect.SendGet(`getuid`, nil, &ret); err != nil {
		return err
	}
	if len(ret.UID) == 0 {
		return nil
	}
	connect.Auth = ret.Token
	sign, err = crypto.SignString(hex.EncodeToString(connect.PrivateKey), `LOGIN`+ret.NetworkID+ret.UID)
	if err != nil {
		return err
	}
	form := url.Values{"pubkey": {connect.PublicKey}, "signature": {hex.EncodeToString(sign)},
		`ecosystem`: {`1`}, "role_id": {"0"}}
	var logret loginResult
	err = connect.SendPost(`login`, &form, &logret)
	connect.Auth = logret.Token
	return err
}
