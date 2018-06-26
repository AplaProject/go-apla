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
	"net/http"

	"github.com/GenesisKernel/go-genesis/packages/crypto"
	"github.com/GenesisKernel/go-genesis/packages/utils"

	"github.com/gorilla/mux"
)

func (h *contractHandlers) ContractNodeHandler(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)

	if err := checkNodeAccess(r); err != nil {
		errorResponse(w, err)
		return
	}

	contract := getContract(r, params[keyName])
	if contract == nil {
		errorResponse(w, errContract.Errorf(params[keyName]))
		return
	}

	if err := contract.CreateTx(); err != nil {
		errorResponse(w, err)
		return
	}
}

func checkNodeAccess(r *http.Request) error {
	client := getClient(r)

	if !client.IsVDE {
		return errNotFound
	}

	_, publicKey, err := utils.GetNodeKeys()
	if err != nil {
		return err
	}

	publicKeyBytes, err := hex.DecodeString(publicKey)
	if err != nil {
		return err
	}

	keyID := crypto.Address(publicKeyBytes)

	if keyID != client.KeyID {
		return errPermission
	}

	return nil
}
