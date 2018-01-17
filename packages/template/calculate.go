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

package template

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

const (
	tkNumber = iota
	tkAdd
	tkSub
	tkMul
	tkDiv
	tkLPar
	tkRPar

	expInt   = 0
	expFloat = 1
	expMoney = 2
)

type token struct {
	Type  int
	Value interface{}
}

var (
	errExp = errors.New(`wrong expression`)
)

func parsing(input string, itype int) (*[]token, error) {
	var err error

	tokens := make([]token, 0)
	newToken := func(itype int, value interface{}) {
		tokens = append(tokens, token{itype, value})
	}
	prevNumber := func() bool {
		return len(tokens) > 0 && tokens[len(tokens)-1].Type == tkNumber
	}
	prevOper := func() bool {
		return len(tokens) > 0 && (tokens[len(tokens)-1].Type >= tkAdd &&
			tokens[len(tokens)-1].Type <= tkDiv)
	}
	var (
		numlen int
	)
	ops := map[rune]struct {
		id int
		pr int
	}{
		'+': {tkAdd, 1},
		'-': {tkSub, 1},
		'*': {tkMul, 2},
		'/': {tkDiv, 2},
	}
	for off, ch := range input {
		if (ch >= '0' && ch <= '9') || ch == '.' {
			numlen++
			continue
		}
		if numlen > 0 {
			var val interface{}

			switch itype {
			case expInt:
				val, err = strconv.ParseInt(input[off-numlen:off], 10, 64)
			case expFloat:
				val, err = strconv.ParseFloat(input[off-numlen:off], 64)
			}
			if err != nil {
				return nil, err
			}
			if prevNumber() {
				return nil, errExp
			}
			newToken(tkNumber, val)
			numlen = 0
		}
		if item, ok := ops[ch]; ok {
			if prevOper() {
				return nil, errExp
			}
			newToken(item.id, item.pr)
			continue
		}
		switch ch {
		case '(':
			if prevNumber() {
				return nil, errExp
			}
			newToken(tkLPar, 3)
		case ')':
			if prevOper() {
				return nil, errExp
			}
			newToken(tkRPar, 3)
		case ' ', '\t', '\n', '\r':
		default:
			return nil, errExp
		}
	}
	return &tokens, nil
}

func calcExp(tokens []token, eType int) string {
	stack := make([]interface{}, 0, 16)
	for _, item := range tokens {
		switch item.Type {
		case tkNumber:
			stack = append(stack, item.Value)
		case tkAdd:

		}
	}
	if len(stack) != 1 {
		return errExp.Error()
	}
	return fmt.Sprint(stack[0])
}

func calculate(exp, etype string) string {
	var resType int
	if len(etype) == 0 && strings.Contains(exp, `.`) {
		etype = `float`
	}
	switch etype {
	case `float`:
		resType = expFloat
	case `money`:
		resType = expMoney
	}
	tk, err := parsing(exp+` `, resType)
	fmt.Println(`Parse`, err, tk)
	if err != nil {
		return err.Error()
	}
	stack := make([]token, 0, len(*tk))
	buf := make([]token, 0, 10)
	for _, item := range *tk {
		switch item.Type {
		case tkNumber:
			stack = append(stack, item)
		case tkLPar:
			buf = append(buf, item)
		case tkRPar:
			i := len(buf) - 1
			for i >= 0 && buf[i].Type != tkLPar {
				stack = append(stack, buf[i])
				i--
			}
			if i < 0 {
				return errExp.Error()
			}
			buf = buf[:i]
		default:
			if len(buf) > 1 {
				last := buf[len(buf)-1]
				if last.Type != tkLPar && last.Value.(int) >= item.Value.(int) {
					stack = append(stack, last)
					buf[len(buf)-1] = item
					continue
				}
			}
			buf = append(buf, item)
		}
	}
	for i := len(buf) - 1; i >= 0; i-- {
		last := buf[i]
		if last.Type >= tkAdd && last.Type <= tkDiv {
			stack = append(stack, last)
		} else {
			return errExp.Error()
		}
	}
	fmt.Println(resType, `Exp`, stack, exp)
	return calcExp(stack, resType)
}
