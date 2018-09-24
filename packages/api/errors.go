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

var (
	apiErrors = map[string]string{
		`E_CONTRACT`:        `There is not %s contract`,
		`E_DBNIL`:           `DB is nil`,
		`E_DELETEDKEY`:      `The key is deleted`,
		`E_ECOSYSTEM`:       `Ecosystem %d doesn't exist`,
		`E_EMPTYPUBLIC`:     `Public key is undefined`,
		`E_EMPTYSIGN`:       `Signature is undefined`,
		`E_HASHWRONG`:       `Hash is incorrect`,
		`E_HASHNOTFOUND`:    `Hash has not been found`,
		`E_HEAVYPAGE`:       `This page is heavy`,
		`E_INSTALLED`:       `Apla is already installed`,
		`E_INVALIDWALLET`:   `Wallet %s is not valid`,
		`E_LIMITFORSIGN`:    `Length of forsign is too big (%d)`,
		`E_LIMITTXSIZE`:     `The size of tx is too big (%d)`,
		`E_NOTFOUND`:        `Page not found`,
		`E_NOTINSTALLED`:    `Apla is not installed`,
		`E_PARAMNOTFOUND`:   `Parameter %s has not been found`,
		`E_PERMISSION`:      `Permission denied`,
		`E_QUERY`:           `DB query is wrong`,
		`E_RECOVERED`:       `API recovered`,
		`E_REFRESHTOKEN`:    `Refresh token is not valid`,
		`E_SERVER`:          `Server error`,
		`E_SIGNATURE`:       `Signature is incorrect`,
		`E_UNKNOWNSIGN`:     `Unknown signature`,
		`E_STATELOGIN`:      `%s is not a membership of ecosystem %s`,
		`E_TABLENOTFOUND`:   `Table %s has not been found`,
		`E_TOKEN`:           `Token is not valid`,
		`E_TOKENEXPIRED`:    `Token is expired by %s`,
		`E_UNAUTHORIZED`:    `Unauthorized`,
		`E_UNDEFINEVAL`:     `Value %s is undefined`,
		`E_UNKNOWNUID`:      `Unknown uid`,
		`E_VDE`:             `Virtual Dedicated Ecosystem %d doesn't exist`,
		`E_VDECREATED`:      `Virtual Dedicated Ecosystem is already created`,
		`E_REQUESTNOTFOUND`: `Request %s doesn't exist`,
		`E_UPDATING`:        `Node is updating blockchain`,
		`E_STOPPING`:        `Network is stopping`,
		`E_NOTIMPLEMENTED`:  `Not implemented`,
	}
)
