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

package exchangeapi

import (
	"encoding/hex"
	"net/http"

	"github.com/EGaaS/go-egaas-mvp/packages/converter"
	"github.com/EGaaS/go-egaas-mvp/packages/crypto"
)

// NewKey is an answer structure for newKey request
type NewKey struct {
	Error    string `json:"error"`
	Public   string `json:"public"`
	Address  string `json:"address"`
	WalletID int64  `json:"wallet_id"`
}

func newKey(r *http.Request) interface{} {
	var result NewKey

	pub, err := genNewKey()
	if err != nil {
		result.Error = err.Error()
		return result
	}
	result.WalletID = int64(crypto.Address(pub))
	result.Address = converter.AddressToString(result.WalletID)
	result.Public = hex.EncodeToString(pub)
	return result
}
