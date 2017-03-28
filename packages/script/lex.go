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
	"encoding/binary"
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"github.com/shopspring/decimal"
)

const (
	LEX_UNKNOWN = iota
	LEX_SYS
	LEX_OPER
	LEX_NUMBER
	LEX_IDENT
	LEX_NEWLINE
	LEX_STRING
	LEX_COMMENT
	LEX_KEYWORD
	LEX_TYPE
	LEX_EXTEND

	LEX_ERROR = 0xff
	LEXF_NEXT = 1
	LEXF_PUSH = 2
	LEXF_POP  = 4
	LEXF_SKIP = 8

	// System characters
	IS_LPAR   = 0x2801 // (
	IS_RPAR   = 0x2901 // )
	IS_COMMA  = 0x2c01 // ,
	IS_EQ     = 0x3d01 // =
	IS_LCURLY = 0x7b01 // {
	IS_RCURLY = 0x7d01 // }
	IS_LBRACK = 0x5b01 // [
	IS_RBRACK = 0x5d01 // ]

	// Operators
	IS_NOT      = 0x0021 // !
	IS_ASTERISK = 0x002a // *
	IS_PLUS     = 0x002b // +
	IS_MINUS    = 0x002d // -
	IS_SIGN     = 0x012d // - unary
	IS_SOLIDUS  = 0x002f // /
	IS_LESS     = 0x003c // <
	IS_GREAT    = 0x003e // >
	IS_NOTEQ    = 0x213d // !=
	IS_AND      = 0x2626 // &&
	IS_LESSEQ   = 0x3c3d // <=
	IS_EQEQ     = 0x3d3d // ==
	IS_GREQ     = 0x3e3d // >=
	IS_OR       = 0x7c7c // ||

)

const (
	KEY_UNKNOWN = iota
	KEY_CONTRACT
	KEY_FUNC
	KEY_RETURN
	KEY_IF
	KEY_ELSE
	KEY_WHILE
	KEY_TRUE
	KEY_FALSE
	KEY_VAR
	KEY_TX
	KEY_BREAK
	KEY_CONTINUE
	KEY_WARNING
	KEY_INFO
	KEY_NIL
	KEY_ACTION
	KEY_COND
	KEY_ERROR
)

var (
	KEYWORDS = map[string]uint32{`contract`: KEY_CONTRACT, `func`: KEY_FUNC, `return`: KEY_RETURN,
		`if`: KEY_IF, `else`: KEY_ELSE, `error`: KEY_ERROR, `warning`: KEY_WARNING, `info`: KEY_INFO,
		`while`: KEY_WHILE, `data`: KEY_TX, `nil`: KEY_NIL, `action`: KEY_ACTION, `conditions`: KEY_COND,
		`true`: KEY_TRUE, `false`: KEY_FALSE, `break`: KEY_BREAK, `continue`: KEY_CONTINUE, `var`: KEY_VAR}
	TYPES = map[string]reflect.Type{`bool`: reflect.TypeOf(true), `bytes`: reflect.TypeOf([]byte{}),
		`int`: reflect.TypeOf(int64(0)), `address`: reflect.TypeOf(uint64(0)),
		`array`: reflect.TypeOf([]interface{}{}),
		`map`:   reflect.TypeOf(map[string]interface{}{}), `money`: reflect.TypeOf(decimal.New(0, 0)),
		`float`: reflect.TypeOf(float64(0.0)), `string`: reflect.TypeOf(``)}
)

type Lexem struct {
	Type   uint32      // Type of the lexem
	Value  interface{} // Value of lexem
	Line   uint32      // Line of the lexem
	Column uint32      // Position inside the line
}

type Lexems []*Lexem

func LexParser(input []rune) (Lexems, error) {
	var (
		curState                                        uint8
		length, line, off, offline, flags, start, lexId uint32
	)

	lexems := make(Lexems, 0, len(input)/4)
	irune := len(ALPHABET) - 1

	todo := func(r rune) {
		var letter uint8
		if r > 127 {
			letter = ALPHABET[irune]
		} else {
			letter = ALPHABET[r]
		}
		val := LEXTABLE[curState][letter]
		curState = uint8(val >> 16)
		lexId = (val >> 8) & 0xff
		flags = val & 0xff
	}
	length = uint32(len(input)) + 1
	line = 1
	skip := false
	for off < length {
		if off == length-1 {
			todo(rune(' '))
		} else {
			todo(input[off])
		}
		if curState == LEX_ERROR {
			return nil, fmt.Errorf(`unknown lexem %s [Ln:%d Col:%d]`,
				string(input[off:off+1]), line, off-offline+1)
		}
		if (flags & LEXF_SKIP) != 0 {
			off++
			skip = true
			continue
		}

		if lexId > 0 {
			lexOff := off
			if (flags & LEXF_POP) != 0 {
				lexOff = start
			}
			right := off
			if (flags & LEXF_NEXT) != 0 {
				right++
			}
			var value interface{}
			switch lexId {
			case LEX_NEWLINE:
				if input[lexOff] == rune(0x0a) {
					line++
					offline = off
				}
			case LEX_SYS:
				ch := uint32(input[lexOff])
				lexId |= ch << 8
				value = ch
			case LEX_STRING, LEX_COMMENT:
				value = string(input[lexOff+1 : right-1])
				if lexId == LEX_STRING && skip {
					skip = false
					value = strings.Replace(value.(string), `\"`, `"`, -1)
				}
				for i, ch := range value.(string) {
					if ch == 0xa {
						line++
						offline = off + uint32(i) + 1
					}
				}
			case LEX_OPER:
				oper := []byte(string(input[lexOff:right]))
				value = binary.BigEndian.Uint32(append(make([]byte, 4-len(oper)), oper...))
			case LEX_NUMBER:
				name := string(input[lexOff:right])
				if strings.IndexAny(name, `.`) >= 0 {
					if val, err := strconv.ParseFloat(name, 64); err == nil {
						value = val
					} else {
						return nil, fmt.Errorf(`%v %s [Ln:%d Col:%d]`, err, name, line, off-offline+1)
					}
				} else if val, err := strconv.ParseInt(name, 10, 64); err == nil {
					value = val
				} else {
					return nil, fmt.Errorf(`%v %s [Ln:%d Col:%d]`, err, name, line, off-offline+1)
				}
			case LEX_IDENT:
				name := string(input[lexOff:right])
				if name[0] == '$' {
					lexId = LEX_EXTEND
					value = name[1:]
				} else if keyId, ok := KEYWORDS[name]; ok {
					switch keyId {
					case KEY_ACTION, KEY_COND:
						if len(lexems) > 0 {
							lexf := *lexems[len(lexems)-1]
							if lexf.Type&0xff != LEX_KEYWORD || lexf.Value.(uint32) != KEY_FUNC {
								lexems = append(lexems, &Lexem{LEX_KEYWORD | (KEY_FUNC << 8),
									KEY_FUNC, line, lexOff - offline + 1})
							}
						}
						value = name
					case KEY_TRUE:
						lexId = LEX_NUMBER
						value = true
					case KEY_FALSE:
						lexId = LEX_NUMBER
						value = false
					case KEY_NIL:
						lexId = LEX_NUMBER
						value = nil
					default:
						lexId = LEX_KEYWORD | (keyId << 8)
						value = keyId
					}
				} else if typeId, ok := TYPES[name]; ok {
					lexId = LEX_TYPE
					value = typeId
				} else {
					value = name
				}
			}
			lexems = append(lexems, &Lexem{lexId, value, line, lexOff - offline + 1})
		}
		if (flags & LEXF_PUSH) != 0 {
			start = off
		}
		if (flags & LEXF_NEXT) != 0 {
			off++
		}
	}
	return lexems, nil
}
