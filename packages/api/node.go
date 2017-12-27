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

	"github.com/AplaProject/go-apla/packages/consts"
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
		log.WithFields(log.Fields{"type": consts.ConversionError, "error": err}).Error("decoding private key from hex")
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

func NodeContract(Name string) error {

	return nil
}
