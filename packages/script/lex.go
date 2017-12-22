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

	"github.com/AplaProject/go-apla/packages/consts"

	"github.com/shopspring/decimal"
	log "github.com/sirupsen/logrus"
)

// The lexical analysis of the incoming program is implemented in this file. It is the first phase of compilation
// where the incoming text is divided into a sequence of lexemes.

const (
	//	lexUnknown = iota
	// Here are all the created lexemes
	lexSys     = iota + 1 // a system lexeme is different bracket, =, comma and so on.
	lexOper               // Operator is +, -, *, /
	lexNumber             // Number
	lexIdent              // Identifier
	lexNewLine            // Line translation
	lexString             // String
	lexComment            // Comment
	lexKeyword            // Key word
	lexType               // Name of the type
	lexExtend             // Referring to an external variable or function - $myname

	lexError = 0xff
	// flags of lexical states
	lexfNext = 1
	lexfPush = 2
	lexfPop  = 4
	lexfSkip = 8

	// Constants for system lexemes
	isLPar   = 0x2801 // (
	isRPar   = 0x2901 // )
	isComma  = 0x2c01 // ,
	isDot    = 0x2e01 // .
	isEq     = 0x3d01 // =
	isLCurly = 0x7b01 // {
	isRCurly = 0x7d01 // }
	isLBrack = 0x5b01 // [
	isRBrack = 0x5d01 // ]

	// Constants for operations
	isNot      = 0x0021 // !
	isAsterisk = 0x002a // *
	isPlus     = 0x002b // +
	isMinus    = 0x002d // -
	isSign     = 0x012d // - unary
	isSolidus  = 0x002f // /
	isLess     = 0x003c // <
	isGreat    = 0x003e // >
	isNotEq    = 0x213d // !=
	isAnd      = 0x2626 // &&
	isLessEq   = 0x3c3d // <=
	isEqEq     = 0x3d3d // ==
	isGrEq     = 0x3e3d // >=
	isOr       = 0x7c7c // ||

)

const (
	// The list of keyword identifiers
	// Constants for keywords
	//	keyUnknown = iota
	keyContract = iota + 1
	keyFunc
	keyReturn
	keyIf
	keyElse
	keyWhile
	keyTrue
	keyFalse
	keyVar
	keyTX
	keySettings
	keyBreak
	keyContinue
	keyWarning
	keyInfo
	keyNil
	keyAction
	keyCond
	keyTail
	keyError
)

const (
	msgWarning = `warning`
	msgError   = `error`
	msgInfo    = `info`
)

var (
	// The list of key words
	keywords = map[string]uint32{`contract`: keyContract, `func`: keyFunc, `return`: keyReturn,
		`if`: keyIf, `else`: keyElse, msgError: keyError, msgWarning: keyWarning, msgInfo: keyInfo,
		`while`: keyWhile, `data`: keyTX, `settings`: keySettings, `nil`: keyNil, `action`: keyAction, `conditions`: keyCond,
		`true`: keyTrue, `false`: keyFalse, `break`: keyBreak, `continue`: keyContinue,
		`var`: keyVar, `...`: keyTail}
	// list of available types
	// The list of types which save the corresponding 'reflect' type
	types = map[string]reflect.Type{`bool`: reflect.TypeOf(true), `bytes`: reflect.TypeOf([]byte{}),
		`int`: reflect.TypeOf(int64(0)), `address`: reflect.TypeOf(uint64(0)),
		`array`: reflect.TypeOf([]interface{}{}),
		`map`:   reflect.TypeOf(map[string]interface{}{}), `money`: reflect.TypeOf(decimal.New(0, 0)),
		`float`: reflect.TypeOf(float64(0.0)), `string`: reflect.TypeOf(``)}
)

// Lexem contains information about language item
type Lexem struct {
	Type   uint32      // Type of the lexem
	Value  interface{} // Value of lexem
	Line   uint32      // Line of the lexem
	Column uint32      // Position inside the line
}

// GetLogger returns logger
func (l Lexem) GetLogger() *log.Entry {
	return log.WithFields(log.Fields{"lex_type": l.Type, "lex_line": l.Line, "lex_column": l.Column})
}

// Lexems is a slice of lexems
type Lexems []*Lexem

// The lexical analysis is based on the finite machine which is described in the file
// tools/lextable/lextable.go. lextable.go generates a representation of a finite machine as an array
// and records it in the file lex_table.go. In fact, the lexTable array is a set of states and
// depending on the next sign, the machine goes into a new state.
// lexParser parsers the input language source code
func lexParser(input []rune) (Lexems, error) {
	var (
		curState                                        uint8
		length, line, off, offline, flags, start, lexID uint32
	)

	lexems := make(Lexems, 0, len(input)/4)
	irune := len(alphabet) - 1

	// This function according to the next symbol looks with help of lexTable what new state we will have,
	// whether we got the lexeme and what flags are displayed
	todo := func(r rune) {
		var letter uint8
		if r > 127 {
			letter = alphabet[irune]
		} else {
			letter = alphabet[r]
		}
		val := lexTable[curState][letter]
		curState = uint8(val >> 16)
		lexID = (val >> 8) & 0xff
		flags = val & 0xff
	}
	length = uint32(len(input)) + 1
	line = 1
	skip := false
	for off < length {
		// Here we go through the symbols one by one
		if off == length-1 {
			todo(rune(' '))
		} else {
			todo(input[off])
		}
		if curState == lexError {
			return nil, fmt.Errorf(`unknown lexem %s [Ln:%d Col:%d]`,
				string(input[off:off+1]), line, off-offline+1)
		}
		if (flags & lexfSkip) != 0 {
			off++
			skip = true
			continue
		}
		// If machine determined the completed lexeme, we record it in the list of lexemes.
		if lexID > 0 {
			// We do not start a stack for symbols but memorize the displacement when the parse of lexeme began.
			// To get a string of a lexeme we take a substring from the initial displacement to the current one.
			// We immediately write a string as values, a number or a binary representation of operations.
			lexOff := off
			if (flags & lexfPop) != 0 {
				lexOff = start
			}
			right := off
			if (flags & lexfNext) != 0 {
				right++
			}
			var value interface{}
			switch lexID {
			case lexNewLine:
				if input[lexOff] == rune(0x0a) {
					line++
					offline = off
				}
			case lexSys:
				ch := uint32(input[lexOff])
				lexID |= ch << 8
				value = ch
			case lexString, lexComment:
				value = string(input[lexOff+1 : right-1])
				if lexID == lexString && skip {
					skip = false
					value = strings.Replace(value.(string), `\"`, `"`, -1)
					value = strings.Replace(strings.Replace(value.(string), `\r`, "\r", -1), `\n`, "\n", -1)
				}
				for i, ch := range value.(string) {
					if ch == 0xa {
						line++
						offline = off + uint32(i) + 1
					}
				}
			case lexOper:
				oper := []byte(string(input[lexOff:right]))
				value = binary.BigEndian.Uint32(append(make([]byte, 4-len(oper)), oper...))
			case lexNumber:
				name := string(input[lexOff:right])
				if strings.ContainsAny(name, `.`) {
					if val, err := strconv.ParseFloat(name, 64); err == nil {
						value = val
					} else {
						log.WithFields(log.Fields{"error": err, "value": name, "lex_line": line, "lex_col": off - offline + 1, "type": consts.ConversionError}).Error("converting lex number to float")
						return nil, fmt.Errorf(`%v %s [Ln:%d Col:%d]`, err, name, line, off-offline+1)
					}
				} else if val, err := strconv.ParseInt(name, 10, 64); err == nil {
					value = val
				} else {
					log.WithFields(log.Fields{"error": err, "value": name, "lex_line": line, "lex_col": off - offline + 1, "type": consts.ConversionError}).Error("converting lex number to int")
					return nil, fmt.Errorf(`%v %s [Ln:%d Col:%d]`, err, name, line, off-offline+1)
				}
			case lexIdent:
				name := string(input[lexOff:right])
				if name[0] == '$' {
					lexID = lexExtend
					value = name[1:]
				} else if keyID, ok := keywords[name]; ok {
					switch keyID {
					case keyAction, keyCond:
						if len(lexems) > 0 {
							lexf := *lexems[len(lexems)-1]
							if lexf.Type&0xff != lexKeyword || lexf.Value.(uint32) != keyFunc {
								lexems = append(lexems, &Lexem{lexKeyword | (keyFunc << 8),
									keyFunc, line, lexOff - offline + 1})
							}
						}
						value = name
					case keyTrue:
						lexID = lexNumber
						value = true
					case keyFalse:
						lexID = lexNumber
						value = false
					case keyNil:
						lexID = lexNumber
						value = nil
					default:
						lexID = lexKeyword | (keyID << 8)
						value = keyID
					}
				} else if typeID, ok := types[name]; ok {
					lexID = lexType
					value = typeID
				} else {
					value = name
				}
			}
			if lexID != lexComment {
				lexems = append(lexems, &Lexem{lexID, value, line, lexOff - offline + 1})
			}
		}
		if (flags & lexfPush) != 0 {
			start = off
		}
		if (flags & lexfNext) != 0 {
			off++
		}
	}
	return lexems, nil
}
