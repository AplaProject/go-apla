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

const (
	//	cmdUnknown = iota // error
	// here are described the commands of bytecode
	cmdPush       = iota + 1 // Push value to stack
	cmdVar                   // Push variable to stack
	cmdExtend                // Push extend variable to stack
	cmdCallExtend            // Call extend function
	cmdPushStr               // Push ident as string
	cmdCall                  // call a function
	cmdCallVari              // call a variadic function
	cmdReturn                // return from function
	cmdIf                    // run block if Value is true
	cmdElse                  // run block if Value is false
	cmdAssignVar             // list of assigned var
	cmdAssign                // assign
	cmdLabel                 // label for continue
	cmdContinue              // continue from label
	cmdWhile                 // while
	cmdBreak                 // break
	cmdIndex                 // get index []
	cmdSetIndex              // set index []
	cmdFuncName              // set func name Func(...).Name(...)
	cmdUnwrapArr             // unwrap array to stack
	cmdMapInit               // map initialization
	cmdArrayInit             // array initialization
	cmdError                 // error command
)

// the commands for operations in expressions are listed below
const (
	cmdNot = iota | 0x0100
	cmdSign
)

const (
	cmdAdd = iota | 0x0200
	cmdSub
	cmdMul
	cmdDiv
	cmdAnd
	cmdOr
	cmdEqual
	cmdNotEq
	cmdLess
	cmdNotLess
	cmdGreat
	cmdNotGreat

	cmdSys          = 0xff
	cmdUnary uint16 = 50
)
