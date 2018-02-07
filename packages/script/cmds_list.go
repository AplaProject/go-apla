// MIT License
//
// Copyright (c) 2016 GenesisKernel
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.
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
