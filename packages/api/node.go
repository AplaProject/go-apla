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
	"errors"
	"net/http"

	"github.com/GenesisKernel/go-genesis/packages/consts"
	"github.com/GenesisKernel/go-genesis/packages/crypto"
	"github.com/GenesisKernel/go-genesis/packages/smart"
	"github.com/GenesisKernel/go-genesis/packages/utils"

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
