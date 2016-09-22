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

const (
	LEXO_UNKNOWN = iota
	LEXO_SYS
	LEXO_OPER
	LEXO_NUMBER
	LEXO_IDENT

	LEXO_ERROR = 0xFF
	LEXFO_NEXT = 1
	LEXFO_PUSH = 2
	LEXFO_POP  = 4
)

type LexemOld struct {
	Type   uint8  // Type of the lexem
	Offset uint32 // Absolute offset
	Right  uint32 // Right Offset of the lexem
	Line   uint32 // Line of the lexem
	Column uint32 // Position inside the line
}

type LexemsOld []*LexemOld

func LexParserOld(input []rune) LexemsOld {
	var (
		curState, lexId                          uint8
		length, line, off, offline, flags, start uint32
	)

	lexems := make(LexemsOld, 0, len(input)/4)
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
		lexId = uint8((val >> 8) & 0xff)
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
		if curState == LEXO_ERROR {
			lexems = append(lexems, &LexemOld{LEXO_UNKNOWN, off, off + 1, line, off - offline + 1})
			break
		}
		if lexId > 0 {
			lexOff := off
			if (flags & LEXFO_POP) != 0 {
				lexOff = start
			}
			right := off
			if (flags & LEXFO_NEXT) != 0 {
				right++
			}
			lexems = append(lexems, &LexemOld{lexId, lexOff, right, line, lexOff - offline + 1})
			if lexId == LEXO_SYS && input[lexOff] == rune(0x0a) {
				line++
				offline = off
			}
		}
		if (flags & LEXFO_PUSH) != 0 {
			start = off
		}
		if (flags & LEXFO_NEXT) != 0 {
			off++
		}
	}
	return lexems
}
