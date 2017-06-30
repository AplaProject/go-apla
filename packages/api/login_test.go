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
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"testing"

	"github.com/EGaaS/go-egaas-mvp/packages/crypto"
)

var (
	gCookie string
)

func sendRequest(rtype, url string, form *url.Values) (map[string]interface{}, error) {
	client := &http.Client{}
	var ioform io.Reader
	if form != nil {
		ioform = strings.NewReader(form.Encode())
	}
	req, err := http.NewRequest(rtype, `http://localhost:7079/api/v1/`+url, ioform)
	if err != nil {
		return nil, err
	}
	if len(gCookie) > 0 {
		req.Header.Set("Cookie", gCookie)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	//	req.Header.Set("Connection", `keep-alive`)
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	var cookie []string
	for _, val := range resp.Cookies() {
		cookie = append(cookie, fmt.Sprintf(`%s=%s`, val.Name, val.Value))
	}
	if len(cookie) > 0 {
		gCookie = strings.Join(cookie, `;`)
	}
	defer resp.Body.Close()
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	fmt.Println(`ANSWER`, resp.StatusCode, strings.TrimSpace(string(data)))
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf(`%d %s`, resp.StatusCode, strings.TrimSpace(string(data)))
	}

	var v map[string]interface{}
	if err = json.Unmarshal(data, &v); err != nil {
		return nil, err
	}
	/*	if len(v[`error`].(string)) != 0 {
		return nil, fmt.Errorf(v[`error`].(string))
	}*/
	return v, nil
}

func sendGet(url string, form *url.Values) (map[string]interface{}, error) {
	return sendRequest("GET", url, form)
}

func sendPost(url string, form *url.Values) (map[string]interface{}, error) {
	return sendRequest("POST", url, form)
}

func TestAuth(t *testing.T) {
	ret, err := sendGet(`getuid`, nil)
	if err != nil {
		t.Error(err)
		return
	}
	if uid, ok := ret[`uid`]; !ok {
		t.Errorf(`getuid has returned empty uid`)
	} else {
		priv, pub, err := crypto.GenHexKeys()
		if err != nil {
			t.Error(err)
			return
		}
		sign, err := crypto.Sign(priv, uid.(string))
		if err != nil {
			t.Error(err)
			return
		}
		form := url.Values{"pubkey": {pub}, "signature": {hex.EncodeToString(sign)}}
		ret, err = sendPost(`login`, &form)
		if err != nil {
			t.Error(err)
			return
		}
	}
}

func TestBadRequest(t *testing.T) {
	form := url.Values{"pubkey": {`001122`}}
	ret, err := sendPost(`login`, &form)
	if ret != nil {
		t.Error(fmt.Errorf(`must be 400 error`))
		return
	}
	if err.Error() != `400 Value signature is undefined` {
		t.Error(fmt.Errorf(`wrong error: "%s"`, err.Error()))
		return
	}
}

func TestPageNotFound(t *testing.T) {
	ret, err := sendGet(`test`, nil)
	if ret != nil {
		t.Error(fmt.Errorf(`must be 404 error`))
		return
	}
	if err.Error() != `404 404 page not found` {
		t.Error(fmt.Errorf(`wrong error: "%s"`, err.Error()))
		return
	}
}

func TestBalance(t *testing.T) {
	ret, err := sendGet(`balance/qwert`, nil)
	if err != nil {
		t.Error(err)
		return
	}
	fmt.Println(`RET`, ret)
}
