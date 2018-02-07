//MIT License
//
//Copyright (c) 2016 GenesisKernel
//
//Permission is hereby granted, free of charge, to any person obtaining a copy
//of this software and associated documentation files (the "Software"), to deal
//in the Software without restriction, including without limitation the rights
//to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
//copies of the Software, and to permit persons to whom the Software is
//furnished to do so, subject to the following conditions:
//
//The above copyright notice and this permission notice shall be included in all
//copies or substantial portions of the Software.
//
//THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
//IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
//FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
//AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
//LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
//OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
//SOFTWARE.
package api

import (
	"encoding/hex"
	"errors"
	"net/http"

	"github.com/GenesisCommunity/go-genesis/packages/consts"
	"github.com/GenesisCommunity/go-genesis/packages/crypto"
	"github.com/GenesisCommunity/go-genesis/packages/smart"
	"github.com/GenesisCommunity/go-genesis/packages/utils"

	log "github.com/sirupsen/logrus"
)

func nodeContract(w http.ResponseWriter, r *http.Request, data *apiData, logger *log.Entry) error {
	var err error

	NodePrivateKey, NodePublicKey, err := utils.GetNodeKeys()
	if err != nil {
		return err
	}
	if len(NodePrivateKey) == 0 {
		logger.WithFields(log.Fields{"type": consts.EmptyObject}).Error("node private key is empty")
		return errors.New(`empty node private key`)
	}
	pubkey, err := hex.DecodeString(NodePublicKey)
	if err != nil {
		logger.WithFields(log.Fields{"type": consts.ConversionError, "error": err}).Error("decoding private key from hex")
		return err
	}
	data.params[`signed_by`] = smart.PubToID(NodePublicKey)
	prepareData := *data
	if err = prepareContract(w, r, &prepareData, logger); err != nil {
		return err
	}
	signature, err := crypto.Sign(NodePrivateKey, prepareData.result.(prepareResult).ForSign)
	if err != nil {
		logger.WithFields(log.Fields{"type": consts.CryptoError, "error": err}).Error("signing by node private key")
		return err
	}
	data.params[`signature`] = signature
	data.params[`pubkey`] = pubkey
	data.params[`time`] = prepareData.result.(prepareResult).Time
	if err = contract(w, r, data, logger); err != nil {
		return err
	}
	return nil
}
