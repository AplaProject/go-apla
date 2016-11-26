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

//	"fmt"
//	"strconv"

const (
	CMD_UNKNOWN    = iota // error
	CMD_PUSH              // Push value to stack
	CMD_VAR               // Push variable to stack
	CMD_EXTEND            // Push extend variable to stack
	CMD_CALLEXTEND        // Call extend function
	CMD_PUSHSTR           // Push ident as string
	CMD_TABLE             // #table_name[id_column_name = value].column_name
	CMD_CALL              // call a function
	CMD_CALLVARI          // call a variadic function
	CMD_RETURN            // return from function
	CMD_IF                // run block if Value is true
	CMD_ELSE              // run block if Value is false
	CMD_ASSIGNVAR         // list of assigned var
	CMD_ASSIGN            // assign
	CMD_LABEL             // label for continue
	CMD_CONTINUE          // continue from label
	CMD_WHILE             // while
	CMD_ERROR             // error command
)

const (
	CMD_NOT = iota | 0x0100
	CMD_SIGN
)

const (
	CMD_ADD = iota | 0x0200
	CMD_SUB
	CMD_MUL
	CMD_DIV
	CMD_AND
	CMD_OR
	CMD_EQUAL
	CMD_NOTEQ
	CMD_LESS
	CMD_NOTLESS
	CMD_GREAT
	CMD_NOTGREAT

	CMD_SYS           = 0xff
	UNARY      uint16 = 50
	MODE_TABLE        = 1
)

var (
	OPERS = map[string]Oper{
		`||`: {CMD_OR, 10}, `&&`: {CMD_AND, 15}, `==`: {CMD_EQUAL, 20}, `!=`: {CMD_NOTEQ, 20},
		`<`: {CMD_LESS, 22}, `>=`: {CMD_NOTLESS, 22}, `>`: {CMD_GREAT, 22}, `<=`: {CMD_NOTGREAT, 22},
		`+`: {CMD_ADD, 25}, `-`: {CMD_SUB, 25}, `*`: {CMD_MUL, 30},
		`/`: {CMD_DIV, 30}, `!`: {CMD_NOT, UNARY}, `(`: {CMD_SYS, 0xff}, `)`: {CMD_SYS, 0},
	}
)
