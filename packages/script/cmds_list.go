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

package script

const (
	//	cmdUnknown = iota // error
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
	cmdError                 // error command
)

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
