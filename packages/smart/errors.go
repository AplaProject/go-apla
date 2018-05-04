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
	eAccessContract     = `%s can be only called from %s`
	eColumnExist        = `Column %s exists`
	eColumnNotExist     = `Column %s doesn't exist`
	eColumnType         = `Type '%s' of columns is not supported`
	eContractCondition  = `There is not conditions in contract %s`
	eEmptyCond          = `%v condition is empty`
	eIncorrectSignature = `incorrect signature %s`
	eItemNotFound       = `Item %d has not been found`
	eManyColumns        = `Too many columns. Limit is %d`
	eNotCondition       = `There is not %s in parameters`
	eParamNotFound      = `Parameter %s has not been found`
	eRecordNotFound     = `Record %s has not been found`
	eTableExists        = `Table %s exists`
	eTableNotFound      = `Table %s has not been found`
	eUnknownContract    = `Unknown contract %s`
	eUnsupportedType    = "Unsupported type %T"
	eWrongRandom        = `Wrong random parameters %d %d`
)

var (
	errAccessDenied      = errors.New(`Access denied`)
	errConditionEmpty    = errors.New(`Condition is empty`)
	errContractNotFound  = errors.New(`Contract has not been found`)
	errCurrentBalance    = errors.New(`Current balance is not enough`)
	errDeletedKey        = errors.New(`The key is deleted`)
	errDiffKeys          = errors.New(`Contract and user public keys are different`)
	errEmpty             = errors.New(`empty value and condition`)
	errEmptyCond         = errors.New(`The condition is empty`)
	errEmptyContract     = errors.New(`Empty contract name in ContractConditions`)
	errEmptyPublicKey    = errors.New(`Empty public key`)
	errFounderAccount    = errors.New(`Unknown founder account`)
	errFuelRate          = errors.New(`Fuel rate must be greater than 0`)
	errIncorrectSign     = errors.New(`Incorrect sign`)
	errIncorrectType     = errors.New(`Incorrect type`)
	errInvalidValue      = errors.New(`Invalid value`)
	errNegPrice          = errors.New(`Price value is negative`)
	errOneContract       = errors.New(`Оnly one contract must be in the record`)
	errPermEmpty         = errors.New(`Permissions are empty`)
	errSameColumns       = errors.New(`There are the same columns`)
	errSetPubKey         = errors.New(`SetPubKey can be only called from NewUser contract`)
	errTableName         = errors.New(`The name of the table cannot begin with @`)
	errUndefBlock        = errors.New(`It is impossible to write to DB when Block is undefined`)
	errUndefColumns      = errors.New(`Columns are undefined`)
	errUnknownNodeID     = errors.New(`Unknown node id`)
	errUpdNotExistRecord = errors.New(`Update for not existing record`)
	errValues            = errors.New(`Values are undefined`)
	errWrongColumn       = errors.New(`Parameters of column are wrong`)
	errWrongPriceFunc    = errors.New(`Wrong type of price function`)
	errWrongSignature    = errors.New(`wrong signature`)

	errMaxPrice = fmt.Errorf(`Price value is more than %d`, MaxPrice)
)
