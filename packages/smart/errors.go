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
	eContractLoop        = `There is loop in %s contract`
	eContractExist       = `Contract %s already exists`
	eContractNotFound    = `Contract %s has not been found`
	eLatin               = `Name %s must only contain latin, digit and '_', '-' characters`
	eAccessContract      = `%s can be only called from %s`
	eColumnExist         = `column %s exists`
	eColumnNotExist      = `column %s doesn't exist`
	eColumnType          = `Type '%s' of columns is not supported`
	eContractCondition   = `There is not conditions in contract %s`
	eNotCustomTable      = `%s is not a custom table`
	eEmptyCond           = `%v condition is empty`
	eIncorrectEcosys     = `Incorrect ecosystem id %s != %d`
	eIncorrectSignature  = `incorrect signature %s`
	eItemNotFound        = `Item %d has not been found`
	eManyColumns         = `Too many columns. Limit is %d`
	eNotCondition        = `There is not %s in parameters`
	eParamNotFound       = `Parameter %s has not been found`
	eRecordNotFound      = `Record %s has not been found`
	eTableExists         = `table %s exists`
	eTableNotFound       = `Table %s has not been found`
	eTypeJSON            = `Type %T doesn't support json marshalling`
	eUnknownContract     = `Unknown contract %s`
	eUnsupportedType     = "Unsupported type %T"
	eWrongRandom         = `Wrong random parameters %d %d`
	eConditionNotAllowed = `Condition %s is not allowed`
	eTableNotEmpty       = `Table %s is not empty`
	eColumnNotDeleted    = `Column %s cannot be deleted`
)

var (
	errDelayedContract    = errors.New(`Incorrect delayed contract`)
	errAccessDenied       = errors.New(`Access denied`)
	errConditionEmpty     = errors.New(`Conditions is empty`)
	errContractNotFound   = errors.New(`Contract has not been found`)
	errCommission         = errors.New("There is not enough money to pay the commission fee")
	errEmptyColumn        = errors.New(`Column name is empty`)
	errWrongColumn        = errors.New(`Column name cannot begin with digit`)
	errNotFound           = errors.New(`Record has not been found`)
	errNow                = errors.New(`It is prohibited to use NOW() or current time functions`)
	errContractChange     = errors.New(`Contract cannot be removed or inserted`)
	errCurrentBalance     = errors.New(`Current balance is not enough`)
	errDeletedKey         = errors.New(`The key is deleted`)
	errDiffKeys           = errors.New(`Contract and user public keys are different`)
	errEmpty              = errors.New(`empty value and condition`)
	errEmptyCond          = errors.New(`The condition is empty`)
	errEmptyContract      = errors.New(`empty contract name in ContractConditions`)
	errEmptyPublicKey     = errors.New(`Empty public key`)
	errFounderAccount     = errors.New(`Unknown founder account`)
	errFuelRate           = errors.New(`Fuel rate must be greater than 0`)
	errIncorrectSign      = errors.New(`Incorrect sign`)
	errIncorrectType      = errors.New(`incorrect type`)
	errInvalidValue       = errors.New(`Invalid value`)
	errNameChange         = errors.New(`Contracts or functions names cannot be changed`)
	errNegPrice           = errors.New(`Price value is negative`)
	errOneContract        = errors.New(`Оnly one contract must be in the record`)
	errPermEmpty          = errors.New(`Permissions are empty`)
	errRecursion          = errors.New("Recursion detected")
	errSameColumns        = errors.New(`There are the same columns`)
	errTableName          = errors.New(`The name of the table cannot begin with @`)
	errTableEmptyName     = errors.New(`The table name cannot be empty`)
	errUndefBlock         = errors.New(`It is impossible to write to DB when Block is undefined`)
	errUndefColumns       = errors.New(`Columns are undefined`)
	errUnknownNodeID      = errors.New(`Unknown node id`)
	errUpdNotExistRecord  = errors.New(`Update for not existing record`)
	errValues             = errors.New(`Values are undefined`)
	errWrongPriceFunc     = errors.New(`Wrong type of price function`)
	errWrongSignature     = errors.New(`wrong signature`)
	errIncorrectParameter = errors.New(`Incorrect parameter of the condition function`)
	errParseTransaction   = errors.New(`parse transaction`)
	errInputSlice         = errors.New(`input slice is short`)

	errMaxPrice = fmt.Errorf(`Price value is more than %d`, MaxPrice)
)
