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
)

type State struct {
	NewState int
	Func     int
}

type StateLine map[int]State

type States []StateLine

type FuncCompile func(*[]*Block, int, *Lexem) error

const (
	STATE_ROOT = iota
	STATE_BODY
	STATE_BLOCK
	STATE_CONTRACT
	STATE_FUNC
	STATE_PUSH = 0x0100
	STATE_POP  = 0x0200
)

const (
	ERR_NOERROR    = iota
	ERR_UNKNOWNCMD // unknown command
	ERR_MUSTNAME   // must be the name
	ERR_MUSTLCURLY // must be '{'
	ERR_MUSTRCURLY // must be '}'
)

const (
	CF_NOTHING = iota
	CF_ERROR
	CF_NAMEBLOCK
)

var (
	funcs = []FuncCompile{nil,
		fError,
		fNameBlock,
	}
	states = States{
		{ // STATE_ROOT
			LEX_NEWLINE:                       {STATE_ROOT, 0},
			LEX_KEYWORD | (KEY_CONTRACT << 8): {STATE_CONTRACT | STATE_PUSH, 0},
			0: {ERR_UNKNOWNCMD, CF_ERROR},
		},
		{ // STATE_BODY
			LEX_NEWLINE:                   {STATE_BODY, 0},
			LEX_KEYWORD | (KEY_FUNC << 8): {STATE_FUNC | STATE_PUSH, 0},
			IS_RCURLY:                     {STATE_POP, 0},
			0:                             {ERR_MUSTRCURLY, CF_ERROR},
		},
		{ // STATE_BLOCK
			LEX_NEWLINE: {STATE_BLOCK, 0},
			IS_LCURLY:   {STATE_BODY, 0},
			0:           {ERR_MUSTLCURLY, CF_ERROR},
		},
		{ // STATE_CONTRACT
			LEX_NEWLINE: {STATE_CONTRACT, 0},
			LEX_IDENT:   {STATE_BLOCK, CF_NAMEBLOCK},
			0:           {ERR_MUSTNAME, CF_ERROR},
		},
		{ // STATE_FUNC
			LEX_NEWLINE: {STATE_FUNC, 0},
			LEX_IDENT:   {STATE_BLOCK, CF_NAMEBLOCK},
			0:           {ERR_MUSTNAME, CF_ERROR},
		},
	}
)

func fError(buf *[]*Block, state int, lexem *Lexem) error {
	errors := []string{`no error`,
		`unknown command`,  // ERR_UNKNOWNCMD
		`must be the name`, // ERR_MUSTNAME
		`must be '{'`,      // ERR_MUSTLCURLY
		`must be '}'`,      // ERR_MUSTRCURLY
	}
	return fmt.Errorf(`%s %v [Ln:%d Col:%d]`, errors[state], lexem.Value, lexem.Line, lexem.Column)
}

func fNameBlock(buf *[]*Block, state int, lexem *Lexem) error {
	itype := OBJ_FUNC
	switch state {
	case STATE_CONTRACT:
		itype = OBJ_CONTRACT
	}
	prev := (*buf)[len(*buf)-2]
	prev.Objects[lexem.Value.(string)] = &ObjInfo{Type: itype, Value: len(prev.Children) - 1}
	return nil
}

func (vm *VM) Compile(input []rune) error {

	lexems, err := LexParser(input)
	if err != nil {
		return err
	}
	if len(lexems) == 0 {
		return nil
	}
	curState := 0
	root := &Block{}
	stack := make([]int, 0, 64)
	blockstack := make([]*Block, 1, 64)
	blockstack[0] = root

	for i := 0; i < len(lexems); i++ {
		var (
			newState State
			ok       bool
		)
		lexem := lexems[i]
		if newState, ok = states[curState][int(lexem.Type)]; !ok {
			newState = states[curState][0]
		}
		if (newState.NewState & STATE_PUSH) > 0 {
			stack = append(stack, curState)
			top := blockstack[len(blockstack)-1]
			if top.Objects == nil {
				top.Objects = make(map[string]*ObjInfo)
			}
			block := &Block{}
			top.Children = append(top.Children, block)
			blockstack = append(blockstack, block)
		}
		if (newState.NewState & STATE_POP) > 0 {
			if len(stack) == 0 {
				return fError(&blockstack, ERR_MUSTLCURLY, lexem)
			}
			newState.NewState = stack[len(stack)-1]
			stack = stack[:len(stack)-1]
			blockstack = blockstack[:len(blockstack)-1]
			//			fmt.Println(`POP`, stack, newState.NewState)
			continue
		}
		if newState.Func > 0 {
			if err := funcs[newState.Func](&blockstack, curState, lexem); err != nil {
				return err
			}
		}
		//		fmt.Println(`LEX`, curState, lexem, buf, stack)
		curState = newState.NewState & 0xff
	}
	if len(stack) > 0 {
		return fError(&blockstack, ERR_MUSTRCURLY, lexems[len(lexems)-1])
	}
	shift := len(vm.Children)
	for key, item := range root.Objects {
		if item.Type == OBJ_CONTRACT || item.Type == OBJ_FUNC {
			item.Value = item.Value.(int) + shift
		}
		vm.Objects[key] = item
	}
	for _, item := range root.Children {
		vm.Children = append(vm.Children, item)
	}

	fmt.Println(`Root`, blockstack[0])
	fmt.Println(`VM`, vm)
	/*	getName := func(i int) string {
				return `name` //string(input[lexems[i].Offset:lexems[i].Right])
			}
			getNameErr := func(msg string, i int) error {
				return fmt.Errorf(`%s %s [Ln:%d Col:%d]`, msg, getName(i), lexems[i].Line, lexems[i].Column)
			}
		getNext := func(soft bool) (string, *Lexem, error) {
			i++
			if soft {
				for i < len(lexems) && lexems[i].Type == LEX_NEWLINE {
					i++
				}
			}
			if i >= len(lexems) {
				return ``, nil, fmt.Errorf(`end of source code`)
			}
			return getName(i), lexems[i], nil
		}
		for i = 0; i < len(lexems); i++ {
			lexem := lexems[i]
			if lexem.Type == LEX_NEWLINE {
				continue
			}
			if lexem.Type != LEX_KEYWORD && lexem.Value != KEY_CONTRACT {
				return getNameErr(`unknown lexem`, i)
			}
			name, next, err := getNext(true)
			if err != nil {
				return err
			}
			if next.Type != LEX_IDENT {
				return getNameErr(`must be identifier here`, i)
			}
			if _, next, err = getNext(true); err != nil {
				return err
			}
			if next.Type != LEX_SYS || next.Value != '{' {
				return getNameErr(`must be '{' here`, i)
			}
			fmt.Println(`ops`)
			i++
			block, err := vm.compileContract(&input, &lexems, &i)
			if err != nil {
				return err
			}
			vm.Children = append(vm.Children, block)
			vm.Objects[name] = len(vm.Children) - 1
		}*/
	return nil
}
