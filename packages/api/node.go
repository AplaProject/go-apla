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

	"github.com/AplaProject/go-apla/packages/conf"
	"github.com/AplaProject/go-apla/packages/consts"
	"github.com/AplaProject/go-apla/packages/converter"
	"github.com/AplaProject/go-apla/packages/crypto"
	"github.com/AplaProject/go-apla/packages/smart"
	"github.com/AplaProject/go-apla/packages/utils"

	log "github.com/sirupsen/logrus"
)

func nodeContract(w http.ResponseWriter, r *http.Request, data *apiData, logger *log.Entry) error {
	var err error

	NodePrivateKey, NodePublicKey, err := utils.GetNodeKeys()
	if err != nil || len(NodePrivateKey) == 0 {
		if err == nil {
			logger.WithFields(log.Fields{"type": consts.EmptyObject}).Error("node private key is empty")
			err = errors.New(`empty node private key`)
		}
		return err
	}
	pubkey, err := hex.DecodeString(NodePublicKey)
	if err != nil {
		logger.WithFields(log.Fields{"type": consts.ConversionError, "error": err}).Error("decoding private key from hex")
		return err
	}
	data.params[`signed_by`] = smart.PubToID(NodePublicKey)
	prepareData := *data
	if err = prepareContract(w, r, &prepareData, logger); err != nil {
		logger.WithFields(log.Fields{"type": consts.APIError}).Error("can't prepare contract")
		return err
	}
	signed, err := crypto.Sign(NodePrivateKey, prepareData.result.(prepareResult).ForSign)
	data.params[`signature`] = signed
	data.params[`pubkey`] = pubkey
	data.params[`time`] = prepareData.result.(prepareResult).Time
	if err = contract(w, r, data, logger); err != nil {
		logger.WithFields(log.Fields{"type": consts.APIError}).Error("can't call contract")
		return err
	}
	return nil
}

func NodeContract(Name string) (result contractResult, err error) {
	var (
		sign                          []byte
		ret                           getUIDResult
		NodePrivateKey, NodePublicKey string
	)
	err = sendAPIRequest(`GET`, `getuid`, nil, &ret, ``)
	if err != nil {
		log.WithFields(log.Fields{"type": consts.APIError, "error": err}).Error("calling getuid")
		return
	}
	auth := ret.Token
	if len(ret.UID) == 0 {
		err = fmt.Errorf(`getuid has returned empty uid`)
		log.WithFields(log.Fields{"type": consts.APIError, "error": err}).Error("empty uid")
		return
	}
	NodePrivateKey, NodePublicKey, err = utils.GetNodeKeys()
	if err != nil || len(NodePrivateKey) == 0 {
		if err == nil {
			log.WithFields(log.Fields{"type": consts.EmptyObject}).Error("node private key is empty")
			err = errors.New(`empty node private key`)
		}
		return
	}
	sign, err = crypto.Sign(NodePrivateKey, ret.UID)
	if err != nil {
		log.WithFields(log.Fields{"type": consts.CryptoError, "error": err}).Error("signing node uid")
		return
	}
	form := url.Values{"pubkey": {NodePublicKey}, "signature": {hex.EncodeToString(sign)},
		`ecosystem`: {converter.Int64ToStr(1)}}
	var logret loginResult
	err = sendAPIRequest(`POST`, `login`, &form, &logret, auth)
	if err != nil {
		log.WithFields(log.Fields{"type": consts.APIError, "error": err}).Error("login node")
		return
	}
	auth = logret.Token
	form = url.Values{`vde`: {`true`}}
	err = sendAPIRequest(`POST`, `node/`+Name, &form, &result, auth)
	if err != nil {
		log.WithFields(log.Fields{"type": consts.APIError, "error": err}).Error("node request")
		return
	}
	return
}

func sendAPIRequest(rtype, url string, form *url.Values, v interface{}, auth string) error {
	client := &http.Client{}
	var ioform io.Reader
	if form != nil {
		ioform = strings.NewReader(form.Encode())
	}
	req, err := http.NewRequest(rtype, fmt.Sprintf(`http://%s:%d%s%s`, conf.Config.HTTP.Host,
		conf.Config.HTTP.Port, consts.ApiPath, url), ioform)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	if len(auth) > 0 {
		req.Header.Set("Authorization", jwtPrefix+auth)
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
