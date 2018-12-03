// Apla Software includes an integrated development
// environment with a multi-level system for the management
// of access rights to data, interfaces, and Smart contracts. The
// technical characteristics of the Apla Software are indicated in
// Apla Technical Paper.
//
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
//
// THE APLA SOFTWARE IS PROVIDED “AS IS”, WITHOUT WARRANTY
// OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED
// TO THE WARRANTIES OF MERCHANTABILITY, FITNESS FOR A
// PARTICULAR PURPOSE, ERROR FREE AND NONINFRINGEMENT. IN
// NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE
// LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER
// IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING
// FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR
// THE USE OR OTHER DEALINGS IN THE APLA SOFTWARE.

package script

import "errors"

const (
	eContractLoop    = `there is loop in %s contract`
	eSysVar          = `system variable $%s cannot be changed`
	eTypeParam       = `parameter %d has wrong type`
	eUndefinedParam  = `%s is not defined`
	eUnknownContract = `unknown contract %s`
	eWrongParams     = `function %s must have %d parameters`
	eArrIndex        = `index of array cannot be type %s`
	eMapIndex        = `index of map cannot be type %s`
	eUnknownIdent    = `unknown identifier %s`
	eWrongVar        = `wrong var %v`
	eDataType        = `expecting type of the data field [Ln:%d Col:%d]`
	eDataName        = `expecting name of the data field [Ln:%d Col:%d]`
	eDataTag         = `unexpected tag [Ln:%d Col:%d]`
)

var (
	errContractPars    = errors.New(`wrong contract parameters`)
	errWrongCountPars  = errors.New(`wrong count of parameters`)
	errDivZero         = errors.New(`divided by zero`)
	errUnsupportedType = errors.New(`unsupported combination of types in the operator`)
	errMaxArrayIndex   = errors.New(`The index is out of range`)
	errMaxMapCount     = errors.New(`The maxumim length of map`)
	errRecursion       = errors.New(`The contract can't call itself recursively`)
	errUnclosedArray   = errors.New(`unclosed array initialization`)
	errUnclosedMap     = errors.New(`unclosed map initialization`)
	errUnexpKey        = errors.New(`unexpected lexem; expecting string key`)
	errUnexpColon      = errors.New(`unexpected lexem; expecting colon`)
	errUnexpComma      = errors.New(`unexpected lexem; expecting comma`)
	errUnexpValue      = errors.New(`unexpected lexem; expecting string, int value or variable`)
	errCondWrite       = errors.New(`'conditions' cannot call contracts or functions which can modify the blockchain database.`)
)
