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
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/AplaProject/go-apla/packages/conf"
	"github.com/AplaProject/go-apla/packages/consts"
	"github.com/AplaProject/go-apla/packages/converter"
	"github.com/AplaProject/go-apla/packages/crypto"
	"github.com/AplaProject/go-apla/packages/utils/tx"
)

var apiAddress = "http://localhost:7079"

var (
	gAuth             string
	gAddress          string
	gPrivate, gPublic string
	gMobile           bool
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
	return crypto.PubToHex(pubKey), nil
}

func sendRawRequest(rtype, url string, form *url.Values) ([]byte, error) {
	client := &http.Client{}
	var ioform io.Reader
	if form != nil {
		ioform = strings.NewReader(form.Encode())
	}
	req, err := http.NewRequest(rtype, apiAddress+consts.ApiPath+url, ioform)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	if len(gAuth) > 0 {
		req.Header.Set("Authorization", jwtPrefix+gAuth)
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

func sendRequest(rtype, url string, form *url.Values, v interface{}) error {
	data, err := sendRawRequest(rtype, url, form)
	if err != nil {
		return err
	}

	return json.Unmarshal(data, v)
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

	sign, err = crypto.SignString(string(key), `LOGIN`+ret.NetworkID+ret.UID)
	if err != nil {
		return
	}
	pub, err = PrivateToPublicHex(string(key))
	if err != nil {
		return
	}
	form := url.Values{"pubkey": {pub}, "signature": {hex.EncodeToString(sign)},
		`ecosystem`: {converter.Int64ToStr(state)}, "role_id": {"0"}}
	if gMobile {
		form[`mobile`] = []string{`true`}
	}
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
	sign, err := crypto.SignString(gPrivate, forSign)
	if err != nil {
		return ``, err
	}
	return hex.EncodeToString(sign), nil
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
	data, err := json.Marshal(&txstatusRequest{
		Hashes: []string{hash},
	})
	if err != nil {
		return 0, err
	}

	for i := 0; i < 15; i++ {
		var multiRet multiTxStatusResult
		err := sendPost(`txstatus`, &url.Values{
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

func randName(prefix string) string {
	return fmt.Sprintf(`%s%d`, prefix, time.Now().Unix())
}

type getter interface {
	Get(string) string
}

type contractParams map[string]interface{}

func (cp *contractParams) Get(key string) string {
	if _, ok := (*cp)[key]; !ok {
		return ""
	}
	return fmt.Sprintf("%v", (*cp)[key])
}

func (cp *contractParams) GetRaw(key string) interface{} {
	return (*cp)[key]
}

func postTxResult(name string, form getter) (id int64, msg string, err error) {
	var contract getContractResult
	if err = sendGet("contract/"+name, nil, &contract); err != nil {
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
		case "file", "bytes":
			if cp, ok := form.(*contractParams); !ok {
				err = fmt.Errorf("Form is not *contractParams type")
			} else {
				params[name] = cp.GetRaw(name)
			}
		}

		if err != nil {
			err = fmt.Errorf("Parse param '%s': %s", name, err)
			return
		}
	}

	var privateKey, publicKey []byte
	if privateKey, err = hex.DecodeString(gPrivate); err != nil {
		return
	}
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
	err = sendMultipart("sendTx", map[string][]byte{
		"data": data,
	}, &ret)
	if err != nil {
		return
	}

	if len(form.Get("nowait")) > 0 {
		return
	}
	id, err = waitTx(ret.Hashes["data"])
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

func TestGetAvatar(t *testing.T) {

	err := keyLogin(1)
	assert.NoError(t, err)

	url := `http://localhost:7079` + consts.ApiPath + "avatar/-1744264011260937456"
	req, err := http.NewRequest(http.MethodGet, url, nil)
	assert.NoError(t, err)

	if len(gAuth) > 0 {
		req.Header.Set("Authorization", jwtPrefix+gAuth)
	}

	cli := http.DefaultClient
	resp, err := cli.Do(req)
	assert.NoError(t, err)

	defer resp.Body.Close()
	mime := resp.Header.Get("Content-Type")
	expectedMime := "image/png"
	assert.Equal(t, expectedMime, mime, "content type must be a '%s' but returns '%s'", expectedMime, mime)
}

func sendMultipart(url string, files map[string][]byte, v interface{}) error {
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

	req, err := http.NewRequest("POST", apiAddress+consts.ApiPath+url, body)
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", writer.FormDataContentType())

	if len(gAuth) > 0 {
		req.Header.Set("Authorization", jwtPrefix+gAuth)
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
