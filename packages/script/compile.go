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
	STATE_FPARAMS
	STATE_FPARAM
	STATE_FRESULT
	STATE_IF
	STATE_VAR
	STATE_ASSIGNEVAL
	STATE_ASSIGN
	STATE_TX
	STATE_FIELDS

	STATE_EVAL

	STATE_PUSH    = 0x0100
	STATE_POP     = 0x0200
	STATE_STAY    = 0x0400
	STATE_TOBLOCK = 0x0800
	STATE_TOBODY  = 0x1000
	STATE_FORK    = 0x2000
	STATE_TOFORK  = 0x4000
)

const (
	ERR_NOERROR    = iota
	ERR_UNKNOWNCMD // unknown command
	ERR_MUSTNAME   // must be the name
	ERR_MUSTLCURLY // must be '{'
	ERR_MUSTRCURLY // must be '}'
	ERR_PARAMS     // wrong parameters
	ERR_VARS       // wrong variables
	ERR_ASSIGN     // must be '='
)

const (
	CF_NOTHING = iota
	CF_ERROR
	CF_NAMEBLOCK
	CF_FRESULT
	CF_RETURN
	CF_IF
	CF_ELSE
	CF_FPARAM
	CF_FTYPE
	CF_ASSIGNVAR
	CF_ASSIGN
	CF_TX
	CF_FIELD
	CF_FIELDTYPE
	CF_FIELDTAG
	CF_CMDERROR
	CF_EVAL
)

var (
	opers = map[uint32]Oper{
		IS_OR: {CMD_OR, 10}, IS_AND: {CMD_AND, 15}, IS_EQEQ: {CMD_EQUAL, 20}, IS_NOTEQ: {CMD_NOTEQ, 20},
		IS_LESS: {CMD_LESS, 22}, IS_GREQ: {CMD_NOTLESS, 22}, IS_GREAT: {CMD_GREAT, 22}, IS_LESSEQ: {CMD_NOTGREAT, 22},
		IS_PLUS: {CMD_ADD, 25}, IS_MINUS: {CMD_SUB, 25}, IS_ASTERISK: {CMD_MUL, 30},
		IS_SOLIDUS: {CMD_DIV, 30}, IS_SIGN: {CMD_SIGN, UNARY}, IS_NOT: {CMD_NOT, UNARY}, IS_LPAR: {CMD_SYS, 0xff}, IS_RPAR: {CMD_SYS, 0},
	}
	funcs = []FuncCompile{nil,
		fError,
		fNameBlock,
		fFuncResult,
		fReturn,
		fIf,
		fElse,
		fFparam,
		fFtype,
		fAssignVar,
		fAssign,
		fTx,
		fField,
		fFieldType,
		fFieldTag,
		fCmdError,
	}
	states = States{
		{ // STATE_ROOT
			LEX_NEWLINE:                       {STATE_ROOT, 0},
			LEX_KEYWORD | (KEY_CONTRACT << 8): {STATE_CONTRACT | STATE_PUSH, 0},
			LEX_KEYWORD | (KEY_FUNC << 8):     {STATE_FUNC | STATE_PUSH, 0},
			LEX_COMMENT:                       {STATE_ROOT, 0},
			0:                                 {ERR_UNKNOWNCMD, CF_ERROR},
		},
		{ // STATE_BODY
			LEX_NEWLINE:                     {STATE_BODY, 0},
			LEX_KEYWORD | (KEY_FUNC << 8):   {STATE_FUNC | STATE_PUSH, 0},
			LEX_KEYWORD | (KEY_RETURN << 8): {STATE_EVAL, CF_RETURN},
			LEX_KEYWORD | (KEY_IF << 8):     {STATE_EVAL | STATE_PUSH | STATE_TOBLOCK, CF_IF},
			LEX_KEYWORD | (KEY_ELSE << 8):   {STATE_BLOCK | STATE_PUSH, CF_ELSE},
			LEX_KEYWORD | (KEY_VAR << 8):    {STATE_VAR, 0},
			LEX_KEYWORD | (KEY_TX << 8):     {STATE_TX, CF_TX},
			LEX_KEYWORD | (KEY_ERROR << 8):  {STATE_EVAL, CF_CMDERROR},
			LEX_COMMENT:                     {STATE_BODY, 0},
			LEX_IDENT:                       {STATE_ASSIGNEVAL | STATE_FORK, 0},
			LEX_EXTEND:                      {STATE_ASSIGNEVAL | STATE_FORK, 0},
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
			LEX_IDENT:   {STATE_FPARAMS, CF_NAMEBLOCK},
			0:           {ERR_MUSTNAME, CF_ERROR},
		},
		{ // STATE_FPARAMS
			LEX_NEWLINE: {STATE_FPARAMS, 0},
			IS_LPAR:     {STATE_FPARAM, 0},
			0:           {STATE_FRESULT | STATE_STAY, 0},
		},
		{ // STATE_FPARAM
			LEX_NEWLINE: {STATE_FPARAM, 0},
			LEX_IDENT:   {STATE_FPARAM, CF_FPARAM},
			LEX_TYPE:    {STATE_FPARAM, CF_FTYPE},
			IS_COMMA:    {STATE_FPARAM, 0},
			IS_RPAR:     {STATE_FRESULT, 0},
			0:           {ERR_PARAMS, CF_ERROR},
		},
		{ // STATE_FRESULT
			LEX_NEWLINE: {STATE_FRESULT, 0},
			LEX_TYPE:    {STATE_FRESULT, CF_FRESULT},
			IS_COMMA:    {STATE_FRESULT, 0},
			0:           {STATE_BLOCK | STATE_STAY, 0},
		},
		{ // STATE_IF
			0: {STATE_EVAL | STATE_TOBLOCK | STATE_PUSH, CF_IF},
		},
		{ // STATE_VAR
			LEX_NEWLINE: {STATE_BODY, 0},
			LEX_IDENT:   {STATE_VAR, CF_FPARAM},
			LEX_TYPE:    {STATE_VAR, CF_FTYPE},
			IS_COMMA:    {STATE_VAR, 0},
			0:           {ERR_VARS, CF_ERROR},
		},
		{ // STATE_ASSIGNEVAL
			IS_LPAR: {STATE_EVAL | STATE_TOFORK | STATE_TOBODY, 0},
			0:       {STATE_ASSIGN | STATE_TOFORK | STATE_STAY, 0},
		},
		{ // STATE_ASSIGN
			IS_COMMA:   {STATE_ASSIGN, 0},
			LEX_IDENT:  {STATE_ASSIGN, CF_ASSIGNVAR},
			LEX_EXTEND: {STATE_ASSIGN, CF_ASSIGNVAR},
			IS_EQ:      {STATE_EVAL | STATE_TOBODY, CF_ASSIGN},
			0:          {ERR_ASSIGN, CF_ERROR},
		},
		{ // STATE_TX
			LEX_NEWLINE: {STATE_TX, 0},
			IS_LCURLY:   {STATE_FIELDS, 0},
			0:           {ERR_MUSTLCURLY, CF_ERROR},
		},
		{ // STATE_FIELDS
			LEX_NEWLINE: {STATE_FIELDS, 0},
			LEX_COMMENT: {STATE_FIELDS, 0},
			IS_COMMA:    {STATE_FIELDS, 0},
			LEX_IDENT:   {STATE_FIELDS, CF_FIELD},
			LEX_TYPE:    {STATE_FIELDS, CF_FIELDTYPE},
			LEX_STRING:  {STATE_FIELDS, CF_FIELDTAG},
			IS_RCURLY:   {STATE_TOBODY, 0},
			0:           {ERR_MUSTRCURLY, CF_ERROR},
		},
	}
)

func fError(buf *[]*Block, state int, lexem *Lexem) error {
	errors := []string{`no error`,
		`unknown command`,  // ERR_UNKNOWNCMD
		`must be the name`, // ERR_MUSTNAME
		`must be '{'`,      // ERR_MUSTLCURLY
		`must be '}'`,      // ERR_MUSTRCURLY
		`wrong parameters`, // ERR_PARAMS
		`wrong variables`,  // ERR_VARS
		`must be '='`,      // ERR_ASSIGN
	}
	fmt.Printf("%s %x %v [Ln:%d Col:%d]\r\n", errors[state], lexem.Type, lexem.Value, lexem.Line, lexem.Column)
	return fmt.Errorf(`%s %x %v [Ln:%d Col:%d]`, errors[state], lexem.Type, lexem.Value, lexem.Line, lexem.Column)
}

func fFuncResult(buf *[]*Block, state int, lexem *Lexem) error {
	fblock := (*buf)[len(*buf)-1].Info.(*FuncInfo)
	(*fblock).Results = append((*fblock).Results, lexem.Value.(reflect.Type))
	return nil
}

func fReturn(buf *[]*Block, state int, lexem *Lexem) error {
	//	fblock := (*buf)[len(*buf)-1].Info.(*FuncInfo)
	(*(*buf)[len(*buf)-1]).Code = append((*(*buf)[len(*buf)-1]).Code, &ByteCode{CMD_RETURN, 0}) //len(fblock.Results)})
	return nil
}

func fCmdError(buf *[]*Block, state int, lexem *Lexem) error {
	(*(*buf)[len(*buf)-1]).Code = append((*(*buf)[len(*buf)-1]).Code, &ByteCode{CMD_ERROR, 0})
	return nil
}

func fFparam(buf *[]*Block, state int, lexem *Lexem) error {
	block := (*buf)[len(*buf)-1]
	if block.Type == OBJ_FUNC && state == STATE_FPARAM {
		fblock := block.Info.(*FuncInfo)
		fblock.Params = append(fblock.Params, reflect.TypeOf(nil))
	}
	if block.Objects == nil {
		block.Objects = make(map[string]*ObjInfo)
	}
	block.Objects[lexem.Value.(string)] = &ObjInfo{Type: OBJ_VAR, Value: len(block.Vars)}
	block.Vars = append(block.Vars, reflect.TypeOf(nil))
	return nil
}

func fFtype(buf *[]*Block, state int, lexem *Lexem) error {
	block := (*buf)[len(*buf)-1]
	if block.Type == OBJ_FUNC && state == STATE_FPARAM {
		fblock := block.Info.(*FuncInfo)
		for pkey, param := range fblock.Params {
			if param == reflect.TypeOf(nil) {
				fblock.Params[pkey] = lexem.Value.(reflect.Type)
			}
		}
	}
	for vkey, ivar := range block.Vars {
		if ivar == reflect.TypeOf(nil) {
			block.Vars[vkey] = lexem.Value.(reflect.Type)
		}
	}
	//	fmt.Println(`VARS`, block.Vars)
	return nil
}

func fIf(buf *[]*Block, state int, lexem *Lexem) error {
	(*(*buf)[len(*buf)-2]).Code = append((*(*buf)[len(*buf)-2]).Code, &ByteCode{CMD_IF, (*buf)[len(*buf)-1]})
	return nil
}

func fAssignVar(buf *[]*Block, state int, lexem *Lexem) error {
	block := (*buf)[len(*buf)-1]
	var (
		prev []*VarInfo
		ivar VarInfo
	)
	if lexem.Type == LEX_EXTEND {
		ivar = VarInfo{&ObjInfo{OBJ_EXTEND, lexem.Value.(string)}, nil}
	} else {
		objInfo, tobj := findVar(lexem.Value.(string), buf)
		if objInfo == nil || objInfo.Type != OBJ_VAR {
			return fmt.Errorf(`unknown variable %s`, lexem.Value.(string))
		}
		ivar = VarInfo{objInfo, tobj}
	}
	if len(block.Code) > 0 {
		if block.Code[len(block.Code)-1].Cmd == CMD_ASSIGNVAR {
			prev = block.Code[len(block.Code)-1].Value.([]*VarInfo)
		}
	}
	prev = append(prev, &ivar)
	if len(prev) == 1 {
		(*(*buf)[len(*buf)-1]).Code = append((*block).Code, &ByteCode{CMD_ASSIGNVAR, prev})
	} else {
		(*(*buf)[len(*buf)-1]).Code[len(block.Code)-1] = &ByteCode{CMD_ASSIGNVAR, prev}
	}
	return nil
}

func fAssign(buf *[]*Block, state int, lexem *Lexem) error {
	(*(*buf)[len(*buf)-1]).Code = append((*(*buf)[len(*buf)-1]).Code, &ByteCode{CMD_ASSIGN, 0})
	return nil
}

func fTx(buf *[]*Block, state int, lexem *Lexem) error {
	contract := (*buf)[len(*buf)-1]
	if contract.Type != OBJ_CONTRACT {
		return fmt.Errorf(`tx can be only in contract`)
	}
	(*contract).Info.(*ContractInfo).Tx = new([]*FieldInfo)
	return nil
}

func fField(buf *[]*Block, state int, lexem *Lexem) error {
	tx := (*(*buf)[len(*buf)-1]).Info.(*ContractInfo).Tx
	*tx = append(*tx, &FieldInfo{Name: lexem.Value.(string), Type: reflect.TypeOf(nil)})
	return nil
}

func fFieldType(buf *[]*Block, state int, lexem *Lexem) error {
	tx := (*(*buf)[len(*buf)-1]).Info.(*ContractInfo).Tx
	for i, field := range *tx {
		if field.Type == reflect.TypeOf(nil) {
			(*tx)[i].Type = lexem.Value.(reflect.Type)
		}
	}
	return nil
}

func fFieldTag(buf *[]*Block, state int, lexem *Lexem) error {
	tx := (*(*buf)[len(*buf)-1]).Info.(*ContractInfo).Tx
	for i := len(*tx) - 1; i > 0; i-- {
		if len((*tx)[i].Tags) == 0 {
			(*tx)[i].Tags = lexem.Value.(string)
			break
		}
	}
	return nil
}

func fElse(buf *[]*Block, state int, lexem *Lexem) error {
	code := (*(*buf)[len(*buf)-2]).Code
	if code[len(code)-1].Cmd != CMD_IF {
		return fmt.Errorf(`there is not if before %v [Ln:%d Col:%d]`, lexem.Type, lexem.Line, lexem.Column)
	}
	(*(*buf)[len(*buf)-2]).Code = append(code, &ByteCode{CMD_ELSE, (*buf)[len(*buf)-1]})
	return nil
}

func fNameBlock(buf *[]*Block, state int, lexem *Lexem) error {
	var itype int

	prev := (*buf)[len(*buf)-2]
	fblock := (*buf)[len(*buf)-1]

	switch state {
	case STATE_BLOCK:
		itype = OBJ_CONTRACT
		fblock.Info = &ContractInfo{Id: uint32(len(prev.Children) - 1), Name: lexem.Value.(string)}
	default:
		itype = OBJ_FUNC
		fblock.Info = &FuncInfo{}
	}
	fblock.Type = itype
	prev.Objects[lexem.Value.(string)] = &ObjInfo{Type: itype, Value: fblock}
	return nil
}

func (vm *VM) CompileBlock(input []rune) (*Block, error) {
	root := &Block{}
	lexems, err := LexParser(input)
	if err != nil {
		return nil, err
	}
	if len(lexems) == 0 {
		return root, nil
	}
	curState := 0
	stack := make([]int, 0, 64)
	blockstack := make([]*Block, 1, 64)
	blockstack[0] = root
	fork := 0

	for i := 0; i < len(lexems); i++ {
		var (
			newState State
			ok       bool
		)
		lexem := lexems[i]
		if newState, ok = states[curState][int(lexem.Type)]; !ok {
			newState = states[curState][0]
		}
		nextState := newState.NewState & 0xff
		if (newState.NewState & STATE_FORK) > 0 {
			fork = i
			//			continue
		}
		if (newState.NewState & STATE_TOFORK) > 0 {
			i = fork
			fork = 0
			lexem = lexems[i]
			//			fmt.Printf("State %x %x %v %v\r\n", curState, newState.NewState, lexem, stack)
			//			continue
		}

		if (newState.NewState & STATE_STAY) > 0 {
			curState = nextState
			i--
			continue
		}
		if nextState == STATE_EVAL {
			if err := vm.compileEval(&lexems, &i, &blockstack); err != nil {
				return nil, err
			}
			nextState = curState
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
			//			fmt.Println(`PUSH`, curState)
		}
		if (newState.NewState & STATE_POP) > 0 {
			if len(stack) == 0 {
				return nil, fError(&blockstack, ERR_MUSTLCURLY, lexem)
			}
			nextState = stack[len(stack)-1]
			stack = stack[:len(stack)-1]
			blockstack = blockstack[:len(blockstack)-1]
			//	fmt.Println(`POP`, stack, newState.NewState)
			//			continue
		}
		if (newState.NewState & STATE_TOBLOCK) > 0 {
			nextState = STATE_BLOCK
		}
		if (newState.NewState & STATE_TOBODY) > 0 {
			nextState = STATE_BODY
		}
		//fmt.Println(`LEX`, curState, lexem, stack)
		if newState.Func > 0 {
			if err := funcs[newState.Func](&blockstack, nextState, lexem); err != nil {
				return nil, err
			}
			//		fmt.Println(`Block Func`, *blockstack[len(blockstack)-1], len(blockstack)-1)
		}
		curState = nextState
	}
	if len(stack) > 0 {
		return nil, fError(&blockstack, ERR_MUSTRCURLY, lexems[len(lexems)-1])
	}
	//	shift := len(vm.Children)
	//	fmt.Println(`Root`, blockstack[0])
	//	fmt.Println(`VM`, vm)
	return root, nil
}

func (vm *VM) FlushBlock(root *Block) {
	for key, item := range root.Objects {
		vm.Objects[key] = item
	}
	shift := len(vm.Children)
	for _, item := range root.Children {
		if item.Type == OBJ_CONTRACT {
			item.Info.(*ContractInfo).Id += uint32(shift)
		}
		vm.Children = append(vm.Children, item)
	}
}

func (vm *VM) Compile(input []rune) error {
	root, err := vm.CompileBlock(input)
	if err == nil {
		vm.FlushBlock(root)
	}
	return err
}

func findVar(name string, block *[]*Block) (ret *ObjInfo, owner *Block) {
	var ok bool
	i := len(*block) - 1
	for ; i >= 0; i-- {
		ret, ok = (*block)[i].Objects[name]
		if ok {
			return ret, (*block)[i]
		}
	}
	return nil, nil
}

func (vm *VM) findObj(name string, block *[]*Block) (ret *ObjInfo, owner *Block) {
	ret, owner = findVar(name, block)
	if ret != nil {
		return
	}
	return vm.getObjByName(name), nil
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
		var call bool
		lexem := (*lexems)[i]
		//		fmt.Println(i, parcount, lexem)
		switch lexem.Type {
		case IS_RCURLY, IS_LCURLY:
			i--
			break main
		case LEX_NEWLINE:
			if i > 0 && ((*lexems)[i-1].Type == IS_COMMA || (*lexems)[i-1].Type == LEX_OPER) {
				continue main
			}
			for k := len(buffer) - 1; k >= 0; k-- {
				if buffer[k].Cmd == CMD_SYS {
					continue main
				}
			}
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
				if oper.Cmd == CMD_SUB && (i == 0 || ((*lexems)[i-1].Type != LEX_NUMBER && (*lexems)[i-1].Type != LEX_IDENT &&
					(*lexems)[i-1].Type != LEX_STRING && (*lexems)[i-1].Type != IS_RCURLY)) {
					oper.Cmd = CMD_SIGN
					oper.Priority = UNARY
				}
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
		case LEX_EXTEND:
			if i < len(*lexems)-2 {
				if (*lexems)[i+1].Type == IS_LPAR {
					count := 0
					if (*lexems)[i+2].Type != IS_RPAR {
						count++
					}
					parcount = append(parcount, count)
					buffer = append(buffer, &ByteCode{CMD_CALLEXTEND, lexem.Value.(string)})
					call = true
				}
			}
			if !call {
				cmd = &ByteCode{CMD_EXTEND, lexem.Value.(string)}
			}
		case LEX_IDENT:
			objInfo, tobj := vm.findObj(lexem.Value.(string), block)
			if objInfo == nil {
				return fmt.Errorf(`unknown identifier %s`, lexem.Value.(string))
			}
			if i < len(*lexems)-2 {
				if (*lexems)[i+1].Type == IS_LPAR {
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
				cmd = &ByteCode{CMD_VAR, &VarInfo{objInfo, tobj}}
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
