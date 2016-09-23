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
	"reflect"
)

type Oper struct {
	Cmd      uint16
	Priority uint16
}

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
	STATE_FRESULT

	STATE_EVAL

	STATE_PUSH = 0x0100
	STATE_POP  = 0x0200
	STATE_STAY = 0x0400
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
	CF_FRESULT
	CF_RETURN
	CF_EVAL
)

var (
	opers = map[uint32]Oper{
		IS_OR: {CMD_OR, 10}, IS_AND: {CMD_AND, 15}, IS_EQEQ: {CMD_EQUAL, 20}, IS_NOTEQ: {CMD_NOTEQ, 20},
		IS_LESS: {CMD_LESS, 22}, IS_GREQ: {CMD_NOTLESS, 22}, IS_GREAT: {CMD_GREAT, 22}, IS_LESSEQ: {CMD_NOTGREAT, 22},
		IS_PLUS: {CMD_ADD, 25}, IS_MINUS: {CMD_SUB, 25}, IS_ASTERISK: {CMD_MUL, 30},
		IS_SOLIDUS: {CMD_DIV, 30}, IS_NOT: {CMD_NOT, UNARY}, IS_LPAR: {CMD_SYS, 0xff}, IS_RPAR: {CMD_SYS, 0},
	}
	funcs = []FuncCompile{nil,
		fError,
		fNameBlock,
		fFuncResult,
		fReturn,
	}
	states = States{
		{ // STATE_ROOT
			LEX_NEWLINE:                       {STATE_ROOT, 0},
			LEX_KEYWORD | (KEY_CONTRACT << 8): {STATE_CONTRACT | STATE_PUSH, 0},
			LEX_KEYWORD | (KEY_FUNC << 8):     {STATE_FUNC | STATE_PUSH, 0},
			0: {ERR_UNKNOWNCMD, CF_ERROR},
		},
		{ // STATE_BODY
			LEX_NEWLINE:                     {STATE_BODY, 0},
			LEX_KEYWORD | (KEY_FUNC << 8):   {STATE_FUNC | STATE_PUSH, 0},
			LEX_KEYWORD | (KEY_RETURN << 8): {STATE_EVAL, CF_RETURN},
			LEX_IDENT:                       {STATE_EVAL, 0},
			IS_RCURLY:                       {STATE_POP, 0},
			0:                               {ERR_MUSTRCURLY, CF_ERROR},
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
			LEX_IDENT:   {STATE_FRESULT, CF_NAMEBLOCK},
			0:           {ERR_MUSTNAME, CF_ERROR},
		},
		{ // STATE_FRESULT
			LEX_NEWLINE: {STATE_FRESULT, 0},
			LEX_TYPE:    {STATE_FRESULT, CF_FRESULT},
			IS_COMMA:    {STATE_FRESULT, 0},
			0:           {STATE_BLOCK | STATE_STAY, 0},
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

func fFuncResult(buf *[]*Block, state int, lexem *Lexem) error {
	fblock := (*buf)[len(*buf)-1].Info.(*FuncInfo)
	(*fblock).Results = append((*fblock).Results, lexem.Value.(reflect.Kind))
	return nil
}

func fReturn(buf *[]*Block, state int, lexem *Lexem) error {
	fblock := (*buf)[len(*buf)-1].Info.(*FuncInfo)
	(*(*buf)[len(*buf)-1]).Code = append((*(*buf)[len(*buf)-1]).Code, &ByteCode{CMD_RETURN, len(fblock.Results)})
	return nil
}

func fNameBlock(buf *[]*Block, state int, lexem *Lexem) error {
	itype := OBJ_FUNC
	switch state {
	case STATE_CONTRACT:
		itype = OBJ_CONTRACT
	}
	prev := (*buf)[len(*buf)-2]
	fblock := (*buf)[len(*buf)-1]
	if itype == OBJ_FUNC {
		fblock.Info = &FuncInfo{}
	}
	prev.Objects[lexem.Value.(string)] = &ObjInfo{Type: itype, Value: fblock}
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
		if (newState.NewState & STATE_STAY) > 0 {
			curState = newState.NewState & 0xff
			i--
			continue
		}
		if newState.NewState == STATE_EVAL {
			if err := vm.compileEval(&lexems, &i, &blockstack); err != nil {
				return err
			}
			newState.NewState = curState
			//			fmt.Println(`Block`, *blockstack[len(blockstack)-1], len(blockstack)-1)
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
			//fmt.Println(`PUSH`, curState)
		}
		if (newState.NewState & STATE_POP) > 0 {
			if len(stack) == 0 {
				return fError(&blockstack, ERR_MUSTLCURLY, lexem)
			}
			newState.NewState = stack[len(stack)-1]
			stack = stack[:len(stack)-1]
			blockstack = blockstack[:len(blockstack)-1]
			//	fmt.Println(`POP`, stack, newState.NewState)
			//			continue
		}
		//		fmt.Println(`LEX`, curState, lexem, stack)
		if newState.Func > 0 {
			if err := funcs[newState.Func](&blockstack, curState, lexem); err != nil {
				return err
			}
		}
		curState = newState.NewState & 0xff
	}
	if len(stack) > 0 {
		return fError(&blockstack, ERR_MUSTRCURLY, lexems[len(lexems)-1])
	}
	//	shift := len(vm.Children)
	for key, item := range root.Objects {
		/*		if item.Type == OBJ_CONTRACT || item.Type == OBJ_FUNC {
				item.Value = item.Value.(int) + shift
			}*/
		vm.Objects[key] = item
	}
	for _, item := range root.Children {
		vm.Children = append(vm.Children, item)
	}

	//	fmt.Println(`Root`, blockstack[0])
	//	fmt.Println(`VM`, vm)
	return nil
}

func (vm *VM) findObj(name string, block *[]*Block) (ret *ObjInfo) {
	var ok bool
	i := len(*block) - 1
	for ; i >= 0; i-- {
		ret, ok = (*block)[i].Objects[name]
		if ok {
			return ret
		}
	}
	return vm.getObjByName(name)
}

func (vm *VM) compileEval(lexems *Lexems, ind *int, block *[]*Block) error {
	i := *ind
	curBlock := (*block)[len(*block)-1]

	buffer := make(ByteCodes, 0, 20)
	bytecode := make(ByteCodes, 0, 100)
	parcount := make([]int, 0, 20)
	//	mode := 0
main:
	for ; i < len(*lexems); i++ {
		var cmd *ByteCode
		lexem := (*lexems)[i]
		//		fmt.Println(i, parcount, lexem)
		switch lexem.Type {
		case IS_RCURLY, LEX_NEWLINE:
			break main
		case IS_LPAR:
			buffer = append(buffer, &ByteCode{CMD_SYS, uint16(0xff)})
		case IS_COMMA:
			if len(parcount) > 0 {
				parcount[len(parcount)-1]++
			}
			for len(buffer) > 0 {
				prev := buffer[len(buffer)-1]
				if prev.Cmd == CMD_SYS && prev.Value.(uint16) == 0xff {
					break
				} else {
					bytecode = append(bytecode, prev)
					buffer = buffer[:len(buffer)-1]
				}
			}
		case IS_RPAR:
			for {
				if len(buffer) == 0 {
					return fmt.Errorf(`there is not pair`)
				} else {
					prev := buffer[len(buffer)-1]
					buffer = buffer[:len(buffer)-1]
					if prev.Value.(uint16) == 0xff {
						break
					} else {
						bytecode = append(bytecode, prev)
					}
				}
			}
			if len(buffer) > 0 {
				if prev := buffer[len(buffer)-1]; prev.Cmd == CMD_CALL || prev.Cmd == CMD_CALLVARI {
					count := parcount[len(parcount)-1]
					parcount = parcount[:len(parcount)-1]
					if prev.Cmd == CMD_CALLVARI {
						bytecode = append(bytecode, &ByteCode{CMD_PUSH, count})
					}
					buffer = buffer[:len(buffer)-1]
					bytecode = append(bytecode, prev)
				}
			}
		case LEX_OPER:
			if oper, ok := opers[lexem.Value.(uint32)]; ok {
				byteOper := &ByteCode{oper.Cmd, oper.Priority}
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
				return fmt.Errorf(`unknown operator %s`, lexem.Value.(uint32))
			}
		case LEX_NUMBER, LEX_STRING:
			cmd = &ByteCode{CMD_PUSH, lexem.Value}
		case LEX_IDENT:
			var call bool
			if i < len(*lexems)-2 {
				if (*lexems)[i+1].Type == IS_LPAR {
					objInfo := vm.findObj(lexem.Value.(string), block)
					if objInfo == nil || (objInfo.Type != OBJ_EXTFUNC && objInfo.Type != OBJ_FUNC) {
						return fmt.Errorf(`unknown function %s`, lexem.Value.(string))
					}
					cmdCall := uint16(CMD_CALL)
					if objInfo.Type == OBJ_EXTFUNC && objInfo.Value.(ExtFuncInfo).Variadic { /*||
						(objInfo.Type == OBJ_FUNC && objInfo.Value.(*Block).Info.(FuncInfo).Variadic )*/
						cmdCall = CMD_CALLVARI
					}
					count := 0
					if (*lexems)[i+2].Type != IS_RPAR {
						count++
					}
					parcount = append(parcount, count)
					buffer = append(buffer, &ByteCode{cmdCall, objInfo})
					call = true
				}
			}
			if !call {
				cmd = &ByteCode{CMD_VAR, lexem.Value}
			}
		}
		if cmd != nil {
			bytecode = append(bytecode, cmd)
		}
	}
	*ind = i
	for i := len(buffer) - 1; i >= 0; i-- {
		if buffer[i].Cmd == CMD_SYS {
			return fmt.Errorf(`there is not pair`)
		} else {
			bytecode = append(bytecode, buffer[i])
		}
	}
	curBlock.Code = append(curBlock.Code, bytecode...)
	return nil
}
