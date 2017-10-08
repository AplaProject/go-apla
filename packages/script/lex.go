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

// В данном файле реализован лексический анализ входящей программы. Это первый этап компиляции,
// при котором входящий текст разбивается на последовательность лексем.
// The lexical analysis of the incoming program is implemented in this file. It is the first phase of compilation
// where the incoming text is divided into a sequence of lexemes.

const (
	//	lexUnknown = iota
	// Здесь перечислены все создаваемые лексемы
	// Here are all the created lexemes
	lexSys     = iota + 1 // системная лексема - это разные скобки, =, запятая и т.п. // a system lexeme is different bracket, =, comma and so on.
	lexOper               // Оператор - это всякие +, -, *, / // Operator is +, -, *, /
	lexNumber             // Число // Number
	lexIdent              // Идентификатор // Identifier
	lexNewLine            // Перевод строки // Line translation
	lexString             // Строка // String
	lexComment            // Комментарий // Comment
	lexKeyword            // Ключевое слово // Key word
	lexType               // Имя типа // Name of the type
	lexExtend             // Обращение к внешней переменной или функции - $myname // Referring to an external variable or function - $myname

	lexError = 0xff
	// flags of lexical states
	lexfNext = 1
	lexfPush = 2
	lexfPop  = 4
	lexfSkip = 8

	// System characters    константы для системных лексем
	// Constants for system lexemes
	isLPar   = 0x2801 // (
	isRPar   = 0x2901 // )
	isComma  = 0x2c01 // ,
	isEq     = 0x3d01 // =
	isLCurly = 0x7b01 // {
	isRCurly = 0x7d01 // }
	isLBrack = 0x5b01 // [
	isRBrack = 0x5d01 // ]

	// Operators  константы для операций
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
	// Константы для ключевых слов
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
	keyError
)

var (
	// Список ключевых слов
	// The list of key words
	keywords = map[string]uint32{`contract`: keyContract, `func`: keyFunc, `return`: keyReturn,
		`if`: keyIf, `else`: keyElse, `error`: keyError, `warning`: keyWarning, `info`: keyInfo,
		`while`: keyWhile, `data`: keyTX, `settings`: keySettings, `nil`: keyNil, `action`: keyAction, `conditions`: keyCond,
		`true`: keyTrue, `false`: keyFalse, `break`: keyBreak, `continue`: keyContinue, `var`: keyVar}
	// list of available types
	// Список типов которые хранят соответствующие reflect типы
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

// Lexems is a slice of lexems
type Lexems []*Lexem

// Лексический разбор происходит на основе конечного автомата, который описан в файле
// tools/lextable/lextable.go. lextable.go генерирует представление конечного автомата в виде массива
// и записывает его в файл lex_table.go. По сути, массив lexTable - это набор состояний и
// в зависимости от очередного символа автомат переходит в новое состояние.
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

	// Эта функция по очередному символу смотрит с помощью lexTable какое у нас будет новое состояние,
	// получили ли лексему и какие флаги выставлены
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
		// Здесь мы перебираем символы один за другим
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
		// Если у нас автомат определил законченную лексему, то мы записываем ее в список лексем.
		// If machine determined the completed lexeme, we record it in the list of lexemes.
		if lexID > 0 {
			// Мы не заводим стэк для символов, а запоминаем смещение, когда начался разбор лексемы.
			// Для получения строки лексемы мы берем подстроку от начального смещения до текущего.
			// В качестве значений мы сразу пишем строку, число или двоичное представление операций.
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
						return nil, fmt.Errorf(`%v %s [Ln:%d Col:%d]`, err, name, line, off-offline+1)
					}
				} else if val, err := strconv.ParseInt(name, 10, 64); err == nil {
					value = val
				} else {
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
			lexems = append(lexems, &Lexem{lexID, value, line, lexOff - offline + 1})
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
