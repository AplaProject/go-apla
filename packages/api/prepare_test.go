package api

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

type txForm struct {
	files     map[string][]byte
	contracts []prepareRequestItem
	noWait    bool
}

func (txf *txForm) body() (io.Reader, string, error) {
	data, err := json.Marshal(&prepareRequest{Contracts: txf.contracts})
	if err != nil {
		return nil, "", err
	}

	if len(txf.files) == 0 {
		body := strings.NewReader("data=" + url.QueryEscape(string(data)))
		return body, "application/x-www-form-urlencoded", nil
	}

	body := new(bytes.Buffer)
	writer := multipart.NewWriter(body)

	for key, data := range txf.files {
		part, err := writer.CreateFormFile(key, key)
		if err != nil {
			return nil, "", err
		}
		if _, err := part.Write(data); err != nil {
			return nil, "", err
		}
	}
	writer.WriteField("data", string(data))
	if err := writer.Close(); err != nil {
		return nil, "", err
	}

	return body, writer.FormDataContentType(), nil
}

func (txf *txForm) prepareRequest() (*prepareResult, error) {
	body, contentType, err := txf.body()
	if err != nil {
		return nil, err
	}

	data, err := sendRawRequest("POST", "prepareMultiple", contentType, body)
	if err != nil {
		return nil, err
	}

	prepareRes := &prepareResult{}
	if err = json.Unmarshal(data, prepareRes); err != nil {
		return nil, err
	}

	return prepareRes, nil
}

func (txf *txForm) Send() ([]txResult, error) {
	res, err := txf.prepareRequest()
	if err != nil {
		return nil, err
	}

	data, err := getSignParams(res)
	if err != nil {
		return nil, err
	}

	hashes, err := sendRawForm("POST", "contractMultiple/"+res.ID, &url.Values{"data": {data}})
	if err != nil {
		return nil, err
	}

	if txf.noWait {
		return nil, nil
	}

	return multiWaitTxStatus(string(hashes))
}

func (txf *txForm) Add(contract string, params map[string]string, files map[string][]byte) {
	if params == nil {
		params = make(map[string]string)
	}

	for key, data := range files {
		fileKey := fmt.Sprintf("file_%d_%s", len(txf.contracts), key)
		params[key] = fileKey
		txf.files[fileKey] = data
	}

	txf.contracts = append(txf.contracts, prepareRequestItem{
		Contract: contract,
		Params:   params,
	})
}

func (txf *txForm) NoWait() {
	txf.noWait = true
}

func newTxForm() *txForm {
	return &txForm{
		files:     make(map[string][]byte),
		contracts: make([]prepareRequestItem, 0),
	}
}

func getSignParams(p *prepareResult) (string, error) {
	req := contractRequest{
		Time:       p.Time,
		Signatures: make([]string, 0, len(p.ForSigns)),
	}
	for _, v := range p.ForSigns {
		sign, err := getSign(v)
		if err != nil {
			return "", err
		}
		req.Signatures = append(req.Signatures, sign)
	}
	b, err := json.Marshal(req)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

type txResult struct {
	BlockID    string          `json:"blockid"`
	ErrMessage json.RawMessage `json:"errmsg,omitempty"`
	Result     string          `json:"result"`
}

func multiWaitTxStatus(hashes string) ([]txResult, error) {
	req := struct {
		Hashes []string `json:"hashes"`
	}{}
	if err := json.Unmarshal([]byte(hashes), &req); err != nil {
		return nil, err
	}

	res := struct {
		Results map[string]txResult `json:"results"`
	}{}

	for i := 0; i < 15; i++ {
		err := sendPost("txstatusMultiple", &url.Values{"data": {hashes}}, &res)
		if err != nil {
			return nil, err
		}

		count := len(res.Results)
		for _, v := range res.Results {
			if len(v.BlockID) > 0 || len(v.ErrMessage) > 0 {
				count--
			}
		}

		if count == 0 {
			var lastErr error
			txs := make([]txResult, 0, len(req.Hashes))
			for _, hash := range req.Hashes {
				if v, ok := res.Results[hash]; ok {
					if len(v.ErrMessage) > 0 {
						lastErr = errors.New(string(v.ErrMessage))
					}
					txs = append(txs, v)
				}
			}
			return txs, lastErr
		}

		time.Sleep(time.Second)
	}

	return nil, fmt.Errorf("TxStatus timeout")
}

func TestMultiRequest(t *testing.T) {
	assert.NoError(t, keyLogin(1))

	tx := newTxForm()
	tx.Add("NewLang", map[string]string{
		"Name":          randName("lang1"),
		"Trans":         `{"en":"test"}`,
		"ApplicationId": "1",
	}, nil)
	tx.Add("NewLang", map[string]string{
		"Name":          randName("lang2"),
		"Trans":         `{"en":"test"}`,
		"ApplicationId": "1",
	}, nil)

	_, err := tx.Send()
	assert.NoError(t, err)
}
