// Copyright (C) 2017, 2018, 2019 EGAAS S.A.
//
// This program is free software; you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation; either version 2 of the License, or (at
// your option) any later version.
//
// This program is distributed in the hope that it will be useful, but
// WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the GNU
// General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program; if not, write to the Free Software
// Foundation, Inc., 51 Franklin Street, Fifth Floor, Boston, MA 02110-1301, USA.

package smart

import (
	"errors"
)

const (
	eContractLoop        = `There is loop in %s contract`
	eContractExist       = `Contract %s already exists`
	eLatin               = `Name %s must only contain latin, digit and '_', '-' characters`
	eAccessContract      = `%s can only be called with condition: %s`
	eColumnExist         = `column %s exists`
	eColumnNotExist      = `column %s doesn't exist`
	eColumnType          = `Type '%s' of columns is not supported`
	eNotCustomTable      = `%s is not a custom table`
	eEmptyCond           = `%v condition is empty`
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
	eRollbackContract    = `Wrong rollback of the latest contract %d != %d`
	eExternalNet         = `External network %s is not defined`
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
	errOneContract        = errors.New(`Оnly one contract must be in the record`)
	errPermEmpty          = errors.New(`Permissions are empty`)
	errRecursion          = errors.New("Recursion detected")
	errSameColumns        = errors.New(`There are the same columns`)
	errTableName          = errors.New(`The name of the table cannot begin with @`)
	errTableEmptyName     = errors.New(`The table name cannot be empty`)
	errUndefColumns       = errors.New(`Columns are undefined`)
	errUpdNotExistRecord  = errors.New(`Update for not existing record`)
	errWrongSignature     = errors.New(`wrong signature`)
	errIncorrectParameter = errors.New(`Incorrect parameter of the condition function`)
	errParseTransaction   = errors.New(`parse transaction`)
	errWhereUpdate        = errors.New(`There is not Where in Update request`)
	errNotValidUTF        = errors.New(`Result is not valid utf-8 string`)
	errFloat              = errors.New(`incorrect float value`)
	errFloatResult        = errors.New(`incorrect float result`)
)
