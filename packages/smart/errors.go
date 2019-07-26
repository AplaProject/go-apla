// Apla Software includes an integrated development
// environment with a multi-level system for the management
// of access rights to data, interfaces, and Smart contracts. The
// technical characteristics of the Apla Software are indicated in
// Apla Technical Paper.

// Apla Users are granted a permission to deal in the Apla
// Software without restrictions, including without limitation the
// rights to use, copy, modify, merge, publish, distribute, sublicense,
// and/or sell copies of Apla Software, and to permit persons
// to whom Apla Software is furnished to do so, subject to the
// following conditions:
// * the copyright notice of GenesisKernel and EGAAS S.A.
// and this permission notice shall be included in all copies or
// substantial portions of the software;
// * a result of the dealing in Apla Software cannot be
// implemented outside of the Apla Platform environment.

// THE APLA SOFTWARE IS PROVIDED “AS IS”, WITHOUT WARRANTY
// OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED
// TO THE WARRANTIES OF MERCHANTABILITY, FITNESS FOR A
// PARTICULAR PURPOSE, ERROR FREE AND NONINFRINGEMENT. IN
// NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE
// LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER
// IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING
// FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR
// THE USE OR OTHER DEALINGS IN THE APLA SOFTWARE.

package smart

import (
	"errors"
)

const (
	eContractLoop        = `There is loop in %s contract`
	eContractExist       = `Contract %s already exists`
	eLatin               = `Name %s must only contain latin, digit and '_', '-' characters`
	eAccessContract      = `%s can be only called from %s`
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
