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

import "errors"

const (
	eTableNotFound = `Table %s has not been found`
)

var (
	errAccessDenied           = errors.New(`Access denied`)
	errConditionEmpty         = errors.New(`Conditions is empty`)
	errContractNotFound       = errors.New(`Contract has not been found`)
	errAccessRollbackContract = errors.New(`RollbackContract can be only called from Import or NewContract`)
	errCommission             = errors.New("There is not enough money to pay the commission fee")
)
