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
	"strconv"
)

const (
	LEX_UNKNOWN = iota
	LEX_SYS
	LEX_OPER
	LEX_NUMBER
	LEX_IDENT
	LEX_NEWLINE
	LEX_STRING
	LEX_KEYWORD

	LEX_ERROR = 0xff
	LEXF_NEXT = 1
	LEXF_PUSH = 2
	LEXF_POP  = 4

	// System characters
	IS_LPAR   = 0x2801 // (
	IS_RPAR   = 0x2901 // )
	IS_COMMA  = 0x2c01 // ,
	IS_LCURLY = 0x7b01 // {
	IS_RCURLY = 0x7d01 // }

	// Operators
	IS_NOT    = 0x0021 // !
	IS_PLUS   = 0x002b // +
	IS_MINUS  = 0x002d // -
	IS_NOTEQ  = 0x213d // !=
	IS_AND    = 0x2626 // &&
	IS_LESSEQ = 0x3c3d // <=
	IS_EQEQ   = 0x3d3d // ==
	IS_OR     = 0x7c7c // ||

)

const (
	KEY_UNKNOWN = iota
	KEY_CONTRACT
	KEY_FUNC
	KEY_RETURN
	KEY_IF
	KEY_WHILE
)

var (
	KEYWORDS = map[string]uint32{`contract`: KEY_CONTRACT, `func`: KEY_FUNC, `return`: KEY_RETURN,
		`if`: KEY_IF, `while`: KEY_WHILE}
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
			case LEX_STRING:
				value = string(input[lexOff+1 : right-1])
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
				if val, err := strconv.ParseInt(name, 10, 64); err == nil {
					value = val
				} else {
					return nil, fmt.Errorf(`%v %s [Ln:%d Col:%d]`, err, name, line, off-offline+1)
				}
			case LEX_IDENT:
				name := string(input[lexOff:right])
				if keyId, ok := KEYWORDS[name]; ok {
					lexId = LEX_KEYWORD | (keyId << 8)
					value = keyId
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
