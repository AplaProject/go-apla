package api

import (
	"encoding/json"
	"fmt"
	"net/url"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestMultiRequest(t *testing.T) {
	assert.NoError(t, keyLogin(1))

	res, err := multiPrepare(&multiPrepareRequest{
		Contracts: []multiPrepareRequestItem{
			{"NewLang", map[string]string{"Name": randName("lang1"), "Trans": `{"en":"test"}`, "ApplicationId": "1"}},
			{"NewLang", map[string]string{"Name": randName("lang2"), "Trans": `{"en":"test"}`, "ApplicationId": "1"}},
		},
	})
	assert.NoError(t, err)

	hashes, err := multiRequest(res)
	assert.NoError(t, err)

	assert.NoError(t, multiWaitTxStatus(hashes))
}

func multiPrepare(req *multiPrepareRequest) (*multiPrepareResult, error) {
	b, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	prepareRes := &multiPrepareResult{}
	err = sendPost("/prepareMultiple", &url.Values{"data": {string(b)}}, prepareRes)
	if err != nil {
		return nil, err
	}

	return prepareRes, nil
}

func multiRequest(res *multiPrepareResult) (string, error) {
	req := contractMultiRequest{
		Time:       res.Time,
		Signatures: make([]string, 0, len(res.ForSigns)),
	}
	for _, v := range res.ForSigns {
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

	hashes, err := sendRawRequest("POST", "/contractMultiple/"+res.ID, &url.Values{"data": {string(b)}})
	if err != nil {
		return "", err
	}

	return string(hashes), err
}

func multiWaitTxStatus(hashes string) error {
	for i := 0; i < 15; i++ {
		txRes := multiTxStatusResult{}
		err := sendPost("/txstatusMultiple", &url.Values{"data": {string(hashes)}}, &txRes)
		if err != nil {
			return err
		}

		count := len(txRes.Results)
		for _, v := range txRes.Results {
			if len(v.BlockID) > 0 {
				count--
			}
		}

		if count == 0 {
			return nil
		}

		time.Sleep(time.Second)
	}

	return fmt.Errorf("TxStatus timeout")
}
