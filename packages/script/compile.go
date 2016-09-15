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

import (
	"fmt"
	"strconv"
)

const (
	CMD_ERROR   = iota // error
	CMD_PUSH           // Push value to stack
	CMD_VAR            // Push variable to stack
	CMD_PUSHSTR        // Push ident as string
	CMD_TABLE          // #table_name[id_column_name = value].column_name
)

const (
	CMD_NOT = iota | 0x0100
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

type Oper struct {
	Cmd      uint16
	Priority uint16
}

var (
	OPERS = map[string]Oper{
		`||`: {CMD_OR, 10}, `&&`: {CMD_AND, 15}, `==`: {CMD_EQUAL, 20}, `!=`: {CMD_NOTEQ, 20},
		`<`: {CMD_LESS, 22}, `>=`: {CMD_NOTLESS, 22}, `>`: {CMD_GREAT, 22}, `<=`: {CMD_NOTGREAT, 22},
		`+`: {CMD_ADD, 25}, `-`: {CMD_SUB, 25}, `*`: {CMD_MUL, 30},
		`/`: {CMD_DIV, 30}, `!`: {CMD_NOT, UNARY}, `(`: {CMD_SYS, 0xff}, `)`: {CMD_SYS, 0},
	}
)

type Bytecode struct {
	Cmd   uint16
	Value interface{}
	Lex   *Lexem
}

type Bytecodes []*Bytecode

func Compile(input []rune) Bytecodes {
	var i int
	bytecode := make(Bytecodes, 0, 100)

	lexems := LexParser(input)
	if len(lexems) == 0 {
		return append(bytecode, &Bytecode{CMD_ERROR, `empty program`, nil})
	}
	last := lexems[len(lexems)-1]
	if last.Type == LEX_UNKNOWN {
		return append(bytecode, &Bytecode{CMD_ERROR, fmt.Sprintf(`unknown lexem %s`,
			string(input[last.Offset:last.Right])), last})
	}
	getNext := func() (string, *Lexem) {
		i++
		return string(input[lexems[i].Offset:lexems[i].Right]), lexems[i]
	}
	buffer := make(Bytecodes, 0, 20)
	mode := 0
	for i = 0; i < len(lexems); i++ {
		var cmd *Bytecode
		lexem := lexems[i]
		//		fmt.Println(i, lexem, buffer, bytecode)
		strlex := string(input[lexem.Offset:lexem.Right])
		switch lexem.Type {
		case LEX_SYS:
			switch strlex {
			case `#`:
				mode = MODE_TABLE
				buffer = append(buffer, &Bytecode{CMD_TABLE, UNARY, lexem})

				strnext, next := getNext()
				bytecode = append(bytecode, &Bytecode{CMD_PUSHSTR, strnext, next})
				strnext, next = getNext()
				if strnext != `[` {
					cmd = &Bytecode{CMD_ERROR, `must be [`, next}
				} else {
					strnext, next = getNext()
					bytecode = append(bytecode, &Bytecode{CMD_PUSHSTR, strnext, next})
					strnext, next = getNext()
					if strnext != `=` {
						cmd = &Bytecode{CMD_ERROR, `must be =`, next}
					}
				}
			case `(`:
				buffer = append(buffer, &Bytecode{CMD_SYS, uint16(0xff), lexem})
			case `)`, `]`:
				for {
					if len(buffer) == 0 {
						cmd = &Bytecode{CMD_ERROR, `there is not pair`, lexem}
						break
					} else {
						prev := buffer[len(buffer)-1]
						buffer = buffer[:len(buffer)-1]
						if (strlex == `)` && prev.Value.(uint16) == 0xff) ||
							(strlex == `]` && prev.Cmd == CMD_TABLE) {
							break
						} else {
							bytecode = append(bytecode, prev)
						}
					}

				}
				if mode == MODE_TABLE && strlex == `]` {
					strnext, next := getNext()
					if strnext != `.` {
						cmd = &Bytecode{CMD_ERROR, `must be .`, next}
					} else {
						strnext, next = getNext()
						bytecode = append(bytecode, &Bytecode{CMD_PUSHSTR, strnext, next})
						mode = 0
						bytecode = append(bytecode, &Bytecode{CMD_TABLE, UNARY, next})
					}
				}
			}
		case LEX_OPER:
			if oper, ok := OPERS[strlex]; ok {
				byteOper := &Bytecode{oper.Cmd, oper.Priority, lexem}
				for {
					if len(buffer) == 0 {
						buffer = append(buffer, byteOper)
						break
					} else {
						prev := buffer[len(buffer)-1]
						if prev.Value.(uint16) >= oper.Priority && oper.Priority != UNARY && prev.Cmd != CMD_SYS {
							if prev.Value.(uint16) == UNARY { // Right to left
								unar := len(buffer) - 1
								for ; unar > 0 && buffer[unar-1].Value.(uint16) == UNARY; unar-- {
								}
								bytecode = append(bytecode, buffer[unar:]...)
								buffer = buffer[:unar]
							} else {
								bytecode = append(bytecode, prev)
								buffer = buffer[:len(buffer)-1]
							}
						} else {
							buffer = append(buffer, byteOper)
							break
						}
					}
				}
			} else {
				cmd = &Bytecode{CMD_ERROR, `unknown operator`, lexem}
			}
		case LEX_NUMBER:
			if val, err := strconv.ParseInt(strlex, 10, 64); err == nil {
				cmd = &Bytecode{CMD_PUSH, val, lexem}
			} else {
				cmd = &Bytecode{CMD_ERROR, err.Error(), lexem}
			}
		case LEX_IDENT:
			cmd = &Bytecode{CMD_VAR, strlex, lexem}
		}
		if cmd != nil {
			bytecode = append(bytecode, cmd)
			if cmd.Cmd == CMD_ERROR {
				cmd.Value = fmt.Sprintf(`%s %s`, cmd.Value.(string), strlex)
				cmd.Lex = lexem
				break
			}
		}
	}
	for i := len(buffer) - 1; i >= 0; i-- {
		if buffer[i].Cmd == CMD_SYS {
			bytecode = append(bytecode, &Bytecode{CMD_ERROR, fmt.Sprintf(`there is not pair`), buffer[i].Lex})
			break
		} else {
			bytecode = append(bytecode, buffer[i])
		}
	}
	return bytecode
}
