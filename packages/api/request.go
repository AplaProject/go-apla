// Apla Software includes an integrated development
// environment with a multi-level system for the management
// of access rights to data, interfaces, and Smart contracts. The
// technical characteristics of the Apla Software are indicated in
// Apla Technical Paper.

// Apla Users are granted a permission to deal in the Apla
// Software without restrictions, including without limitation the
// rights to use, copy, modify, merge, publish, distribute, sublicense,
// and/or sell copies of Apla Software, and to permit persons
// to whom Apla Software is furnished to do so, subject to the
// following conditions:
// * the copyright notice of GenesisKernel and EGAAS S.A.
// and this permission notice shall be included in all copies or
// substantial portions of the software;
// * a result of the dealing in Apla Software cannot be
// implemented outside of the Apla Platform environment.

// THE APLA SOFTWARE IS PROVIDED “AS IS”, WITHOUT WARRANTY
// OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED
// TO THE WARRANTIES OF MERCHANTABILITY, FITNESS FOR A
// PARTICULAR PURPOSE, ERROR FREE AND NONINFRINGEMENT. IN
// NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE
// LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER
// IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING
// FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR
// THE USE OR OTHER DEALINGS IN THE APLA SOFTWARE.

package api

import (
	"bytes"
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

type GetUIDResult struct {
	UID         string `json:"uid,omitempty"`
	Token       string `json:"token,omitempty"`
	Expire      string `json:"expire,omitempty"`
	EcosystemID string `json:"ecosystem_id,omitempty"`
	KeyID       string `json:"key_id,omitempty"`
	Address     string `json:"address,omitempty"`
	NetworkID   string `json:"network_id,omitempty"`
}

type LoginResult struct {
	Token       string        `json:"token,omitempty"`
	EcosystemID string        `json:"ecosystem_id,omitempty"`
	KeyID       string        `json:"key_id,omitempty"`
	Address     string        `json:"address,omitempty"`
	NotifyKey   string        `json:"notify_key,omitempty"`
	IsNode      bool          `json:"isnode,omitempty"`
	IsOwner     bool          `json:"isowner,omitempty"`
	IsOBS       bool          `json:"obs,omitempty"`
	Timestamp   string        `json:"timestamp,omitempty"`
	Roles       []rolesResult `json:"roles,omitempty"`
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

func SendGet(url, auth string, form *url.Values, v interface{}) error {
	return SendRequest("GET", url, auth, form, v)
}

func SendPost(url, auth string, form *url.Values, v interface{}) error {
	return SendRequest("POST", url, auth, form, v)
}

func SendMultipart(url, auth string, files map[string][]byte, v interface{}) error {
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

	req, err := http.NewRequest("POST", url, body)
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", writer.FormDataContentType())

	if len(auth) > 0 {
		req.Header.Set("Authorization", jwtPrefix+auth)
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

func WaitTx(urlPath, auth, hash string) (int64, error) {
	data, err := json.Marshal(&txstatusRequest{
		Hashes: []string{hash},
	})
	if err != nil {
		return 0, err
	}

	for i := 0; i < 15; i++ {
		var multiRet multiTxStatusResult
		err := SendPost(urlPath+`txstatus`, auth, &url.Values{
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

func PostTxResult(url, auth, name string, privateKey []byte,
	form *url.Values) (id int64, msg string, err error) {
	var contract getContractResult
	if err = SendGet(url+"contract/"+name, auth, nil, &contract); err != nil {
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
	if publicKey, err = crypto.PrivateToPublic(privateKey); err != nil {
		return
	}

	data, _, err := tx.NewTransaction(tx.SmartContract{
		Header: tx.Header{
			ID:          int(contract.ID),
			Time:        time.Now().Unix(),
			EcosystemID: 1,
			KeyID:       crypto.Address(publicKey),
			NetworkID:   conf.Config.NetworkID,
		},
		Params: params,
	}, privateKey)
	if err != nil {
		return 0, "", err
	}

	ret := &sendTxResult{}
	err = SendMultipart(url+"sendTx", auth, map[string][]byte{
		"data": data,
	}, &ret)
	if err != nil {
		return
	}
	if len(form.Get("nowait")) > 0 {
		return
	}
	id, err = WaitTx(url, auth, ret.Hashes["data"])
	if id != 0 && err != nil {
		msg = err.Error()
		err = nil
	}

	return
}
