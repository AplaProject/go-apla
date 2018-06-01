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

package smart

import (
	"errors"
	"fmt"
)

const (
	eTableNotFound = `Table %s has not been found`
)

var (
	errAccessDenied           = errors.New(`Access denied`)
	errAccessRollbackContract = errors.New(`RollbackContract can be only called from Import or NewContract`)
	errConditionEmpty         = errors.New(`Conditions is empty`)
	errContractNotFound       = errors.New(`Contract has not been found`)
	errCommission             = errors.New("There is not enough money to pay the commission fee")
	errCurrentBalance         = errors.New(`Current balance is not enough`)
	errDeletedKey             = errors.New(`The key is deleted`)
	errDiffKeys               = errors.New(`Contract and user public keys are different`)
	errEmptyPublicKey         = errors.New(`Empty public key`)
	errFounderAccount         = errors.New(`Unknown founder account`)
	errFuelRate               = errors.New(`Fuel rate must be greater than 0`)
	errIncorrectSign          = errors.New(`Incorrect sign`)
	errInvalidValue           = errors.New(`Invalid value`)
	errNegPrice               = errors.New(`Price value is negative`)
	errUnknownNodeID          = errors.New(`Unknown node id`)
	errWrongPriceFunc         = errors.New(`Wrong type of price function`)

	errMaxPrice = fmt.Errorf(`Price value is more than %d`, MaxPrice)
)
