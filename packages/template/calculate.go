//MIT License
//
//Copyright (c) 2016-2018 GenesisKernel
//
//Permission is hereby granted, free of charge, to any person obtaining a copy
//of this software and associated documentation files (the "Software"), to deal
//in the Software without restriction, including without limitation the rights
//to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
//copies of the Software, and to permit persons to whom the Software is
//furnished to do so, subject to the following conditions:
//
//The above copyright notice and this permission notice shall be included in all
//copies or substantial portions of the Software.
//
//THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
//IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
//FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
//AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
//LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
//OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
//SOFTWARE.

package template

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"unicode"

	"github.com/shopspring/decimal"
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

type opFunc func()

var (
	errExp = errors.New(`wrong expression`)
	errDiv = errors.New(`dividing by zero`)
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
		if unicode.IsDigit(ch) || ch == '.' {
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
			case expMoney:
				val, err = decimal.NewFromString(input[off-numlen : off])
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

func calcExp(tokens []token, resType, prec int) string {
	var top int

	stack := make([]interface{}, 0, 16)

	addInt := func() {
		stack[top-1] = stack[top-1].(int64) + stack[top].(int64)
	}
	addFloat := func() {
		stack[top-1] = stack[top-1].(float64) + stack[top].(float64)
	}
	addMoney := func() {
		stack[top-1] = stack[top-1].(decimal.Decimal).Add(stack[top].(decimal.Decimal))
	}
	subInt := func() {
		stack[top-1] = stack[top-1].(int64) - stack[top].(int64)
	}
	subFloat := func() {
		stack[top-1] = stack[top-1].(float64) - stack[top].(float64)
	}
	subMoney := func() {
		stack[top-1] = stack[top-1].(decimal.Decimal).Sub(stack[top].(decimal.Decimal))
	}
	mulInt := func() {
		stack[top-1] = stack[top-1].(int64) * stack[top].(int64)
	}
	mulFloat := func() {
		stack[top-1] = stack[top-1].(float64) * stack[top].(float64)
	}
	mulMoney := func() {
		stack[top-1] = stack[top-1].(decimal.Decimal).Mul(stack[top].(decimal.Decimal))
	}
	divInt := func() {
		stack[top-1] = stack[top-1].(int64) / stack[top].(int64)
	}
	divFloat := func() {
		stack[top-1] = stack[top-1].(float64) / stack[top].(float64)
	}
	divMoney := func() {
		stack[top-1] = stack[top-1].(decimal.Decimal).Div(stack[top].(decimal.Decimal))
	}

	funcs := map[int][]opFunc{
		tkAdd: {addInt, addFloat, addMoney},
		tkSub: {subInt, subFloat, subMoney},
		tkMul: {mulInt, mulFloat, mulMoney},
		tkDiv: {divInt, divFloat, divMoney},
	}
	for _, item := range tokens {
		if item.Type == tkNumber {
			stack = append(stack, item.Value)
		} else {
			if len(stack) < 2 {
				return errExp.Error()
			}
			top = len(stack) - 1
			if item.Type == tkDiv {
				switch resType {
				case expInt:
					if stack[top].(int64) == 0 {
						return errDiv.Error()
					}
				case expFloat:
					if stack[top].(float64) == 0 {
						return errDiv.Error()
					}
				case expMoney:
					if stack[top].(decimal.Decimal).Cmp(decimal.New(0, 0)) == 0 {
						return errDiv.Error()
					}
				}
			}
			funcs[item.Type][resType]()
			stack = stack[:top]
		}
	}
	if len(stack) != 1 {
		return errExp.Error()
	}
	if prec > 0 {
		if resType == expFloat {
			return strconv.FormatFloat(stack[0].(float64), 'f', prec, 64)
		}
		if resType == expMoney {
			money := fmt.Sprint(stack[0])
			if len(money) < prec+1 {
				money = strings.Repeat(`0`, prec+1-len(money)) + money
			}
			money = money[:len(money)-prec] + `.` + money[len(money)-prec:]
			return strings.TrimRight(strings.TrimRight(money, `0`), `.`)
		}
	}
	return fmt.Sprint(stack[0])
}

func calculate(exp, etype string, prec int) string {
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
			if len(buf) > 0 {
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
	return calcExp(stack, resType, prec)
}
