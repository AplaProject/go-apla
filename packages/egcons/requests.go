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

package main

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
)

func sendRequest(rtype, url string, form *url.Values) (map[string]interface{}, error) {
	client := &http.Client{}
	var ioform io.Reader
	if form != nil {
		ioform = strings.NewReader(form.Encode())
	}
	req, err := http.NewRequest(rtype, gSettings.NodeURL+`/ajax?json=`+url, ioform)
	if err != nil {
		return nil, err
	}
	if len(gSettings.Cookie) > 0 {
		req.Header.Set("Cookie", gSettings.Cookie)
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
		gSettings.Cookie = strings.Join(cookie, `;`)
	}
	defer resp.Body.Close()
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	fmt.Println(`ANSWER`, string(data))
	var v map[string]interface{}
	if err = json.Unmarshal(data, &v); err != nil {
		return nil, err
	}
	if len(v[`error`].(string)) != 0 {
		return nil, fmt.Errorf(v[`error`].(string))
	}
	return v, nil
}

func sendGet(url string, form *url.Values) (map[string]interface{}, error) {
	return sendRequest("GET", url, form)
}

func sendPost(url string, form *url.Values) (map[string]interface{}, error) {
	return sendRequest("POST", url, form)
}
