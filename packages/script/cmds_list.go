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
