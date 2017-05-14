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

package parser

import (
	"github.com/EGaaS/go-egaas-mvp/packages/utils"
)

func (p *Parser) ClearIncompatibleTx(binaryTx []byte, myTx bool) (string, string, int64, int64, int64, int64, int64) {

	var fatalError, waitError string
	var thirdVar int64

	// получим тип тр-ии и юзера
	txType, walletID, citizenID := utils.GetTxTypeAndUserId(binaryTx)
	if walletID == 0 && citizenID == 0 {
		fatalError = "undefined walletId and citizenId"
	}
	var forSelfUse int64

	return fatalError, waitError, forSelfUse, txType, walletID, citizenID, thirdVar
}
