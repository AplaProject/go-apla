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

// operPrior contains command and its priority
type operPrior struct {
	Cmd      uint16 // identifier of the command
	Priority uint16 // priority of the command
}

// State contains a new state and a handle function
type compileState struct {
	NewState int // a new state
	Func     int // a handle function
}

type stateLine map[int]compileState

// The list of compile states
type compileStates []stateLine

type compileFunc func(*[]*Block, int, *Lexem) error

// Компилятор преобразует последовательность лексем в байт-код с помощью конечного автомата, подобно тому как
// это было реализовано при лексическом анализе. Отличие заключается в том, что мы не конвертируем список
// состояний и переходов в промежуточный массив.
// The compiler converts the sequence of lexemes into the bytecodes using a finite state machine the same as
// it was implemented in lexical analysis. The difference lays in that we do not convert the list of
// states and transitions to the intermediate array.

/* Байт-код из себя представляет дерево - на самом верхнем уровне функции контракты, и далее идет вложенность
 в соответствии с вложенностью фигурных скобок. Узлами дерева являются структуры типа Block.
 Например,
// Byte code could be described as a tree where functions and contracts are on the top level and nesting goes further according to nesting of bracketed brackets. Tree nodes are structures of 'Block' type. For instance,
 func a {
	 if b {
		 while d {

		 }
	 }
	 if c {
	 }
 }
будет скомпилировано в Block(a) у которого будут два дочерних блока Block(b) и Block(c), которые
      отвечают за выполнение байт-кода внутри if, а Block(b) в свою очередь будет иметь дочерний
	  блок Block(d) с циклом.
// will be compiled into Block(a) which will have two child blocks Block (b) and Block (c) that are responsible for executing bytecode inside if. Block (b) will have a child Block (d) with a cycle.
*/

const (
	// The list of state types Список состояний
	stateRoot = iota
	stateBody
	stateBlock
	stateContract
	stateFunc
	stateFParams
	stateFParam
	stateFParamTYPE
	stateFResult
	stateVar
	stateVarType
	stateAssignEval
	stateAssign
	stateTX
	stateFields
	stateEval

	// The list of state flags Список флагов
	statePush     = 0x0100
	statePop      = 0x0200
	stateStay     = 0x0400
	stateToBlock  = 0x0800
	stateToBody   = 0x1000
	stateFork     = 0x2000
	stateToFork   = 0x4000
	stateLabel    = 0x8000
	stateMustEval = 0x010000
)

const (
	// Ошибки компиляции
	// Errors of compilation
	//	errNoError    = iota
	errUnknownCmd = iota + 1 // unknown command
	errMustName              // must be the name
	errMustLCurly            // must be '{'
	errMustRCurly            // must be '}'
	errParams                // wrong parameters
	errVars                  // wrong variables
	errVarType               // must be type
	errAssign                // must be '='
)

const (
	// Это список идентификаторов для функций, которые будут генерировать байт-код для соответствующих случаев
	// This is a list of identifiers for functions that will generate a bytecode for the corresponding cases
	// Indexes of handle functions funcs = CompileFunc[]
	//	cfNothing = iota
	cfError = iota + 1
	cfNameBlock
	cfFResult
	cfReturn
	cfIf
	cfElse
	cfFParam
	cfFType
	cfAssignVar
	cfAssign
	cfTX
	cfField
	cfFieldType
	cfFieldTag
	cfWhile
	cfContinue
	cfBreak
	cfCmdError

//	cfEval
)

var (
	// Массив операций и их приоритет
	// Array of operations and their priority
	opers = map[uint32]operPrior{
		isOr: {cmdOr, 10}, isAnd: {cmdAnd, 15}, isEqEq: {cmdEqual, 20}, isNotEq: {cmdNotEq, 20},
		isLess: {cmdLess, 22}, isGrEq: {cmdNotLess, 22}, isGreat: {cmdGreat, 22}, isLessEq: {cmdNotGreat, 22},
		isPlus: {cmdAdd, 25}, isMinus: {cmdSub, 25}, isAsterisk: {cmdMul, 30},
		isSolidus: {cmdDiv, 30}, isSign: {cmdSign, cmdUnary}, isNot: {cmdNot, cmdUnary}, isLPar: {cmdSys, 0xff}, isRPar: {cmdSys, 0},
	}
	// Массив функций, соответствующий константам cf...
	// The array of functions corresponding to the constants cf...
	funcs = []compileFunc{nil,
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
		fWhile,
		fContinue,
		fBreak,
		fCmdError,
	}
	// states описывает конечный автомат с состояниями, на основе которого будет генерироваться байт-код
	// 'states' describes a finite machine with states on the base of which a bytecode will be generated
	states = compileStates{
		{ // stateRoot
			lexNewLine:                      {stateRoot, 0},
			lexKeyword | (keyContract << 8): {stateContract | statePush, 0},
			lexKeyword | (keyFunc << 8):     {stateFunc | statePush, 0},
			lexComment:                      {stateRoot, 0},
			0:                               {errUnknownCmd, cfError},
		},
		{ // stateBody
			lexNewLine:                      {stateBody, 0},
			lexKeyword | (keyFunc << 8):     {stateFunc | statePush, 0},
			lexKeyword | (keyReturn << 8):   {stateEval, cfReturn},
			lexKeyword | (keyContinue << 8): {stateBody, cfContinue},
			lexKeyword | (keyBreak << 8):    {stateBody, cfBreak},
			lexKeyword | (keyIf << 8):       {stateEval | statePush | stateToBlock | stateMustEval, cfIf},
			lexKeyword | (keyWhile << 8):    {stateEval | statePush | stateToBlock | stateLabel | stateMustEval, cfWhile},
			lexKeyword | (keyElse << 8):     {stateBlock | statePush, cfElse},
			lexKeyword | (keyVar << 8):      {stateVar, 0},
			lexKeyword | (keyTX << 8):       {stateTX, cfTX},
			lexKeyword | (keyError << 8):    {stateEval, cfCmdError},
			lexKeyword | (keyWarning << 8):  {stateEval, cfCmdError},
			lexKeyword | (keyInfo << 8):     {stateEval, cfCmdError},
			lexComment:                      {stateBody, 0},
			lexIdent:                        {stateAssignEval | stateFork, 0},
			lexExtend:                       {stateAssignEval | stateFork, 0},
			isRCurly:                        {statePop, 0},
			0:                               {errMustRCurly, cfError},
		},
		{ // stateBlock
			lexNewLine: {stateBlock, 0},
			isLCurly:   {stateBody, 0},
			0:          {errMustLCurly, cfError},
		},
		{ // stateContract
			lexNewLine: {stateContract, 0},
			lexIdent:   {stateBlock, cfNameBlock},
			0:          {errMustName, cfError},
		},
		{ // stateFunc
			lexNewLine: {stateFunc, 0},
			lexIdent:   {stateFParams, cfNameBlock},
			0:          {errMustName, cfError},
		},
		{ // stateFParams
			lexNewLine: {stateFParams, 0},
			isLPar:     {stateFParam, 0},
			0:          {stateFResult | stateStay, 0},
		},
		{ // stateFParam
			lexNewLine: {stateFParam, 0},
			lexIdent:   {stateFParamTYPE, cfFParam},
			// lexType:    {stateFParam, cfFType},
			isComma: {stateFParam, 0},
			isRPar:  {stateFResult, 0},
			0:       {errParams, cfError},
		},
		{ // stateFParamTYPE
			lexIdent: {stateFParamTYPE, cfFParam},
			lexType:  {stateFParam, cfFType},
			isComma:  {stateFParamTYPE, 0},
			//			isRPar:   {stateFResult, 0},
			0: {errVarType, cfError},
		},
		{ // stateFResult
			lexNewLine: {stateFResult, 0},
			lexType:    {stateFResult, cfFResult},
			isComma:    {stateFResult, 0},
			0:          {stateBlock | stateStay, 0},
		},
		{ // stateVar
			lexNewLine: {stateBody, 0},
			lexIdent:   {stateVarType, cfFParam},
			//			lexIdent:   {stateVar, cfFParam},
			//			lexType:    {stateVar, cfFType},
			isComma: {stateVar, 0},
			0:       {errVars, cfError},
		},
		{ // stateVarType
			lexIdent: {stateVarType, cfFParam},
			lexType:  {stateVar, cfFType},
			isComma:  {stateVarType, 0},
			0:        {errVarType, cfError},
		},
		{ // stateAssignEval
			isLPar:   {stateEval | stateToFork | stateToBody, 0},
			isLBrack: {stateEval | stateToFork | stateToBody, 0},
			0:        {stateAssign | stateToFork | stateStay, 0},
		},
		{ // stateAssign
			isComma:   {stateAssign, 0},
			lexIdent:  {stateAssign, cfAssignVar},
			lexExtend: {stateAssign, cfAssignVar},
			isEq:      {stateEval | stateToBody, cfAssign},
			0:         {errAssign, cfError},
		},
		{ // stateTX
			lexNewLine: {stateTX, 0},
			isLCurly:   {stateFields, 0},
			0:          {errMustLCurly, cfError},
		},
		{ // stateFields
			lexNewLine: {stateFields, 0},
			lexComment: {stateFields, 0},
			isComma:    {stateFields, 0},
			lexIdent:   {stateFields, cfField},
			lexType:    {stateFields, cfFieldType},
			lexString:  {stateFields, cfFieldTag},
			isRCurly:   {stateToBody, 0},
			0:          {errMustRCurly, cfError},
		},
	}
)

func fError(buf *[]*Block, state int, lexem *Lexem) error {
	errors := []string{`no error`,
		`unknown command`,  // errUnknownCmd
		`must be the name`, // errMustName
		`must be '{'`,      // errMustLCurly
		`must be '}'`,      // errMustRCurly
		`wrong parameters`, // errParams
		`wrong variables`,  // errVars
		`must be type`,     // errVarType
		`must be '='`,      // errAssign
	}
	fmt.Printf("%s %x %v [Ln:%d Col:%d]\r\n", errors[state], lexem.Type, lexem.Value, lexem.Line, lexem.Column)
	if lexem.Type == lexNewLine {
		return fmt.Errorf(`%s (unexpected new line) [Ln:%d]`, errors[state], lexem.Line-1)
	}
	return fmt.Errorf(`%s %x %v [Ln:%d Col:%d]`, errors[state], lexem.Type, lexem.Value, lexem.Line, lexem.Column)
}

func fFuncResult(buf *[]*Block, state int, lexem *Lexem) error {
	fblock := (*buf)[len(*buf)-1].Info.(*FuncInfo)
	(*fblock).Results = append((*fblock).Results, lexem.Value.(reflect.Type))
	return nil
}

func fReturn(buf *[]*Block, state int, lexem *Lexem) error {
	//	fblock := (*buf)[len(*buf)-1].Info.(*FuncInfo)
	(*(*buf)[len(*buf)-1]).Code = append((*(*buf)[len(*buf)-1]).Code, &ByteCode{cmdReturn, 0}) //len(fblock.Results)})
	return nil
}

func fCmdError(buf *[]*Block, state int, lexem *Lexem) error {
	(*(*buf)[len(*buf)-1]).Code = append((*(*buf)[len(*buf)-1]).Code, &ByteCode{cmdError, lexem.Value})
	return nil
}

func fFparam(buf *[]*Block, state int, lexem *Lexem) error {
	block := (*buf)[len(*buf)-1]
	if block.Type == ObjFunc && (state == stateFParam || state == stateFParamTYPE) {
		fblock := block.Info.(*FuncInfo)
		fblock.Params = append(fblock.Params, reflect.TypeOf(nil))
	}
	if block.Objects == nil {
		block.Objects = make(map[string]*ObjInfo)
	}
	block.Objects[lexem.Value.(string)] = &ObjInfo{Type: ObjVar, Value: len(block.Vars)}
	block.Vars = append(block.Vars, reflect.TypeOf(nil))
	return nil
}

func fFtype(buf *[]*Block, state int, lexem *Lexem) error {
	block := (*buf)[len(*buf)-1]
	if block.Type == ObjFunc && state == stateFParam {
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
	return nil
}

func fIf(buf *[]*Block, state int, lexem *Lexem) error {
	(*(*buf)[len(*buf)-2]).Code = append((*(*buf)[len(*buf)-2]).Code, &ByteCode{cmdIf, (*buf)[len(*buf)-1]})
	return nil
}

func fWhile(buf *[]*Block, state int, lexem *Lexem) error {
	(*(*buf)[len(*buf)-2]).Code = append((*(*buf)[len(*buf)-2]).Code, &ByteCode{cmdWhile, (*buf)[len(*buf)-1]})
	(*(*buf)[len(*buf)-2]).Code = append((*(*buf)[len(*buf)-2]).Code, &ByteCode{cmdContinue, 0})
	return nil
}

func fContinue(buf *[]*Block, state int, lexem *Lexem) error {
	(*(*buf)[len(*buf)-1]).Code = append((*(*buf)[len(*buf)-1]).Code, &ByteCode{cmdContinue, 0})
	return nil
}

func fBreak(buf *[]*Block, state int, lexem *Lexem) error {
	(*(*buf)[len(*buf)-1]).Code = append((*(*buf)[len(*buf)-1]).Code, &ByteCode{cmdBreak, 0})
	return nil
}

func fAssignVar(buf *[]*Block, state int, lexem *Lexem) error {
	block := (*buf)[len(*buf)-1]
	var (
		prev []*VarInfo
		ivar VarInfo
	)
	if lexem.Type == lexExtend {
		ivar = VarInfo{&ObjInfo{ObjExtend, lexem.Value.(string)}, nil}
	} else {
		objInfo, tobj := findVar(lexem.Value.(string), buf)
		if objInfo == nil || objInfo.Type != ObjVar {
			return fmt.Errorf(`unknown variable %s`, lexem.Value.(string))
		}
		//		fmt.Println(`Assign Var`, lexem.Value.(string), objInfo, objInfo.Type, reflect.TypeOf(objInfo.Value), tobj)
		ivar = VarInfo{objInfo, tobj}
	}
	if len(block.Code) > 0 {
		if block.Code[len(block.Code)-1].Cmd == cmdAssignVar {
			prev = block.Code[len(block.Code)-1].Value.([]*VarInfo)
		}
	}
	prev = append(prev, &ivar)
	if len(prev) == 1 {
		(*(*buf)[len(*buf)-1]).Code = append((*block).Code, &ByteCode{cmdAssignVar, prev})
	} else {
		(*(*buf)[len(*buf)-1]).Code[len(block.Code)-1] = &ByteCode{cmdAssignVar, prev}
	}
	return nil
}

func fAssign(buf *[]*Block, state int, lexem *Lexem) error {
	(*(*buf)[len(*buf)-1]).Code = append((*(*buf)[len(*buf)-1]).Code, &ByteCode{cmdAssign, 0})
	return nil
}

func fTx(buf *[]*Block, state int, lexem *Lexem) error {
	contract := (*buf)[len(*buf)-1]
	if contract.Type != ObjContract {
		return fmt.Errorf(`data can only be in contract`)
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
	for i := len(*tx) - 1; i >= 0; i-- {
		if len((*tx)[i].Tags) == 0 {
			(*tx)[i].Tags = lexem.Value.(string)
			break
		}
	}
	return nil
}

func fElse(buf *[]*Block, state int, lexem *Lexem) error {
	code := (*(*buf)[len(*buf)-2]).Code
	if code[len(code)-1].Cmd != cmdIf {
		return fmt.Errorf(`there is not if before %v [Ln:%d Col:%d]`, lexem.Type, lexem.Line, lexem.Column)
	}
	(*(*buf)[len(*buf)-2]).Code = append(code, &ByteCode{cmdElse, (*buf)[len(*buf)-1]})
	return nil
}

// StateName checks the name of the contract and modifies it to @[state]name if it is necessary.
func StateName(state uint32, name string) string {
	if name[0] != '@' {
		return fmt.Sprintf(`@%d%s`, state, name)
	} else if name[1] < '0' || name[1] > '9' {
		name = `@0` + name[1:]
	}
	return name
}

func fNameBlock(buf *[]*Block, state int, lexem *Lexem) error {
	var itype int

	prev := (*buf)[len(*buf)-2]
	fblock := (*buf)[len(*buf)-1]
	name := lexem.Value.(string)
	switch state {
	case stateBlock:
		itype = ObjContract
		name = StateName((*buf)[0].Info.(uint32), name)
		fblock.Info = &ContractInfo{ID: uint32(len(prev.Children) - 1), Name: name, Active: (*buf)[0].Active, TableID: (*buf)[0].TableID} //lexem.Value.(string)}
	default:
		itype = ObjFunc
		fblock.Info = &FuncInfo{}
	}
	fblock.Type = itype
	prev.Objects[name] = &ObjInfo{Type: itype, Value: fblock}
	return nil
}

// CompileBlock compile the source code into the Block structure with a byte-code
func (vm *VM) CompileBlock(input []rune, idstate uint32, active bool, tblid int64) (*Block, error) {
	root := &Block{Info: idstate, Active: active, TableID: tblid}
	lexems, err := lexParser(input)
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
			newState compileState
			ok       bool
		)
		lexem := lexems[i]
		if newState, ok = states[curState][int(lexem.Type)]; !ok {
			newState = states[curState][0]
		}
		nextState := newState.NewState & 0xff
		if (newState.NewState & stateFork) > 0 {
			fork = i
			//			continue
		}
		if (newState.NewState & stateToFork) > 0 {
			i = fork
			fork = 0
			lexem = lexems[i]
			//			fmt.Printf("State %x %x %v %v\r\n", curState, newState.NewState, lexem, stack)
			//			continue
		}

		if (newState.NewState & stateStay) > 0 {
			curState = nextState
			i--
			continue
		}
		if nextState == stateEval {
			if newState.NewState&stateLabel > 0 {
				(*blockstack[len(blockstack)-1]).Code = append((*blockstack[len(blockstack)-1]).Code, &ByteCode{cmdLabel, 0})
			}
			curlen := len((*blockstack[len(blockstack)-1]).Code)
			if err := vm.compileEval(&lexems, &i, &blockstack); err != nil {
				return nil, err
			}
			if (newState.NewState&stateMustEval) > 0 && curlen == len((*blockstack[len(blockstack)-1]).Code) {
				return nil, fmt.Errorf("there is not eval expression")
			}
			nextState = curState
			//			fmt.Println(`Block`, *blockstack[len(blockstack)-1], len(blockstack)-1)
		}
		if (newState.NewState & statePush) > 0 {
			stack = append(stack, curState)
			top := blockstack[len(blockstack)-1]
			if top.Objects == nil {
				top.Objects = make(map[string]*ObjInfo)
			}
			block := &Block{Parent: top}
			top.Children = append(top.Children, block)
			blockstack = append(blockstack, block)
		}
		if (newState.NewState & statePop) > 0 {
			if len(stack) == 0 {
				return nil, fError(&blockstack, errMustLCurly, lexem)
			}
			nextState = stack[len(stack)-1]
			stack = stack[:len(stack)-1]

			if len(blockstack) >= 2 {
				prev := blockstack[len(blockstack)-2]
				if len(prev.Code) > 0 && (*prev).Code[len((*prev).Code)-1].Cmd == cmdContinue {
					(*prev).Code = (*prev).Code[:len((*prev).Code)-1]
					prev = blockstack[len(blockstack)-1]
					(*prev).Code = append((*prev).Code, &ByteCode{cmdContinue, 0})
				}
			}
			blockstack = blockstack[:len(blockstack)-1]
		}
		if (newState.NewState & stateToBlock) > 0 {
			nextState = stateBlock
		}
		if (newState.NewState & stateToBody) > 0 {
			nextState = stateBody
		}
		//fmt.Println(`LEX`, curState, lexem, stack)
		if newState.Func > 0 {
			if err := funcs[newState.Func](&blockstack, nextState, lexem); err != nil {
				return nil, err
			}
		}
		curState = nextState
	}
	if len(stack) > 0 {
		return nil, fError(&blockstack, errMustRCurly, lexems[len(lexems)-1])
	}
	//	shift := len(vm.Children)
	//	fmt.Println(`Root`, blockstack[0])
	//	fmt.Println(`VM`, vm)
	return root, nil
}

// FlushBlock loads the compiled Block into the virtual machine
func (vm *VM) FlushBlock(root *Block) {
	shift := len(vm.Children)
	for key, item := range root.Objects {
		if cur, ok := vm.Objects[key]; ok && item.Type == ObjContract {
			root.Objects[key].Value.(*Block).Info.(*ContractInfo).ID = cur.Value.(*Block).Info.(*ContractInfo).ID + 0xFFFF
		}
		vm.Objects[key] = item
	}
	for _, item := range root.Children {
		if item.Type == ObjContract {
			if item.Info.(*ContractInfo).ID > 0xFFFF {
				item.Info.(*ContractInfo).ID -= 0xFFFF
				vm.Children[item.Info.(*ContractInfo).ID] = item
				shift--
				continue
			}
			item.Parent = &vm.Block
			item.Info.(*ContractInfo).ID += uint32(shift)
		}
		vm.Children = append(vm.Children, item)
	}
}

// FlushExtern switches off the extern mode of the compilation
func (vm *VM) FlushExtern() {
	/*	if !vm.Extern {
		return
	}*/
	vm.Extern = false
	return
}

// Compile compiles a source code and loads the byte-code into the virtual machine
func (vm *VM) Compile(input []rune, state uint32, active bool, tblid int64) error {
	root, err := vm.CompileBlock(input, state, active, tblid)
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
	sname := StateName((*block)[0].Info.(uint32), name)
	ret, owner = findVar(name, block)
	if ret != nil {
		return
	} else if len(sname) > 0 {
		if ret, owner = findVar(sname, block); ret != nil {
			return
		}
	}
	if ret = vm.getObjByName(name); ret == nil && len(sname) > 0 {
		ret = vm.getObjByName(sname)
	}
	return
}

// Данная функиця отвечает за компиляцию выражений
// This function is responsible for the compilation of expressions
func (vm *VM) compileEval(lexems *Lexems, ind *int, block *[]*Block) error {
	i := *ind
	curBlock := (*block)[len(*block)-1]

	buffer := make(ByteCodes, 0, 20)
	bytecode := make(ByteCodes, 0, 100)
	parcount := make([]int, 0, 20)
	setIndex := false
	//	mode := 0
main:
	for ; i < len(*lexems); i++ {
		var cmd *ByteCode
		var call bool
		lexem := (*lexems)[i]
		//		fmt.Println(i, parcount, lexem)
		switch lexem.Type {
		case isRCurly, isLCurly:
			i--
			break main
		case lexNewLine:
			if i > 0 && ((*lexems)[i-1].Type == isComma || (*lexems)[i-1].Type == lexOper) {
				continue main
			}
			for k := len(buffer) - 1; k >= 0; k-- {
				if buffer[k].Cmd == cmdSys {
					continue main
				}
			}
			break main
		case isLPar:
			buffer = append(buffer, &ByteCode{cmdSys, uint16(0xff)})
		case isLBrack:
			buffer = append(buffer, &ByteCode{cmdSys, uint16(0xff)})
		case isComma:
			if len(parcount) > 0 {
				parcount[len(parcount)-1]++
			}
			for len(buffer) > 0 {
				prev := buffer[len(buffer)-1]
				if prev.Cmd == cmdSys && prev.Value.(uint16) == 0xff {
					break
				} else {
					bytecode = append(bytecode, prev)
					buffer = buffer[:len(buffer)-1]
				}
			}
		case isRPar:
			for {
				if len(buffer) == 0 {
					return fmt.Errorf(`there is not pair`)
				}
				prev := buffer[len(buffer)-1]
				buffer = buffer[:len(buffer)-1]
				if prev.Value.(uint16) == 0xff {
					break
				} else {
					bytecode = append(bytecode, prev)
				}
			}
			if len(buffer) > 0 {
				if prev := buffer[len(buffer)-1]; prev.Cmd == cmdCall || prev.Cmd == cmdCallVari {
					count := parcount[len(parcount)-1]
					parcount = parcount[:len(parcount)-1]
					if prev.Cmd == cmdCallVari {
						bytecode = append(bytecode, &ByteCode{cmdPush, count})
					}
					buffer = buffer[:len(buffer)-1]
					bytecode = append(bytecode, prev)
				}
			}
		case isRBrack:
			for {
				if len(buffer) == 0 {
					return fmt.Errorf(`there is not pair`)
				}
				prev := buffer[len(buffer)-1]
				buffer = buffer[:len(buffer)-1]
				if prev.Value.(uint16) == 0xff {
					break
				} else {
					bytecode = append(bytecode, prev)
				}
			}
			if len(buffer) > 0 {
				if prev := buffer[len(buffer)-1]; prev.Cmd == cmdIndex {
					buffer = buffer[:len(buffer)-1]
					if i < len(*lexems)-1 && (*lexems)[i+1].Type == isEq {
						i++
						setIndex = true
						continue
					}
					bytecode = append(bytecode, prev)
				}
			}
		case lexOper:
			if oper, ok := opers[lexem.Value.(uint32)]; ok {
				if oper.Cmd == cmdSub && (i == 0 || ((*lexems)[i-1].Type != lexNumber && (*lexems)[i-1].Type != lexIdent &&
					(*lexems)[i-1].Type != lexExtend &&
					(*lexems)[i-1].Type != lexString && (*lexems)[i-1].Type != isRCurly && (*lexems)[i-1].Type != isRBrack)) {
					oper.Cmd = cmdSign
					oper.Priority = cmdUnary
				}
				byteOper := &ByteCode{oper.Cmd, oper.Priority}
				for {
					if len(buffer) == 0 {
						buffer = append(buffer, byteOper)
						break
					} else {
						prev := buffer[len(buffer)-1]
						if prev.Value.(uint16) >= oper.Priority && oper.Priority != cmdUnary && prev.Cmd != cmdSys {
							if prev.Value.(uint16) == cmdUnary { // Right to left
								unar := len(buffer) - 1
								for ; unar > 0 && buffer[unar-1].Value.(uint16) == cmdUnary; unar-- {
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
				return fmt.Errorf(`unknown operator %d`, lexem.Value.(uint32))
			}
		case lexNumber, lexString:
			cmd = &ByteCode{cmdPush, lexem.Value}
		case lexExtend:
			if i < len(*lexems)-2 {
				if (*lexems)[i+1].Type == isLPar {
					count := 0
					if (*lexems)[i+2].Type != isRPar {
						count++
					}
					parcount = append(parcount, count)
					buffer = append(buffer, &ByteCode{cmdCallExtend, lexem.Value.(string)})
					call = true
				}
			}
			if !call {
				cmd = &ByteCode{cmdExtend, lexem.Value.(string)}
				if i < len(*lexems)-1 && (*lexems)[i+1].Type == isLBrack {
					buffer = append(buffer, &ByteCode{cmdIndex, 0})
				}
			}
		case lexIdent:
			objInfo, tobj := vm.findObj(lexem.Value.(string), block)
			if objInfo == nil && (!vm.Extern || i >= len(*lexems)-2 || (*lexems)[i+1].Type != isLPar) {
				return fmt.Errorf(`unknown identifier %s`, lexem.Value.(string))
			}
			if i < len(*lexems)-2 {
				if (*lexems)[i+1].Type == isLPar {
					var isContract bool
					if vm.Extern && objInfo == nil {
						objInfo = &ObjInfo{Type: ObjContract}
					}
					if objInfo == nil || (objInfo.Type != ObjExtFunc && objInfo.Type != ObjFunc &&
						objInfo.Type != ObjContract) {
						return fmt.Errorf(`unknown function %s`, lexem.Value.(string))
					}
					if objInfo.Type == ObjContract {
						objInfo, tobj = vm.findObj(`ExecContract`, block)
						isContract = true
					}
					cmdCall := uint16(cmdCall)
					if objInfo.Type == ObjExtFunc && objInfo.Value.(ExtFuncInfo).Variadic { /*||
						(objInfo.Type == ObjFunc && objInfo.Value.(*Block).Info.(FuncInfo).Variadic )*/
						cmdCall = cmdCallVari
					}
					count := 0
					if (*lexems)[i+2].Type != isRPar {
						count++
					}
					buffer = append(buffer, &ByteCode{cmdCall, objInfo})
					if isContract {
						name := StateName((*block)[0].Info.(uint32), lexem.Value.(string))
						for j := len(*block) - 1; j >= 0; j-- {
							topblock := (*block)[j]
							if topblock.Type == ObjContract {
								if topblock.Info.(*ContractInfo).Used == nil {
									topblock.Info.(*ContractInfo).Used = make(map[string]bool)
								}
								topblock.Info.(*ContractInfo).Used[name] = true
							}
						}
						bytecode = append(bytecode, &ByteCode{cmdPush, name})
						if count == 0 {
							count = 2
							bytecode = append(bytecode, &ByteCode{cmdPush, ""})
							bytecode = append(bytecode, &ByteCode{cmdPush, ""})
						}
						count++
					}
					if lexem.Value.(string) == `CallContract` {
						bytecode = append(bytecode, &ByteCode{cmdPush, (*block)[0].Info.(uint32)})
					}
					parcount = append(parcount, count)
					call = true
				}
				if (*lexems)[i+1].Type == isLBrack {
					if objInfo == nil || objInfo.Type != ObjVar {
						return fmt.Errorf(`unknown variable %s`, lexem.Value.(string))
					}
					buffer = append(buffer, &ByteCode{cmdIndex, 0})
				}
			}
			if !call {
				cmd = &ByteCode{cmdVar, &VarInfo{objInfo, tobj}}
			}
		}
		if cmd != nil {
			bytecode = append(bytecode, cmd)
		}
	}
	*ind = i
	for i := len(buffer) - 1; i >= 0; i-- {
		if buffer[i].Cmd == cmdSys {
			return fmt.Errorf(`there is not pair`)
		}
		bytecode = append(bytecode, buffer[i])
	}
	if setIndex {
		bytecode = append(bytecode, &ByteCode{cmdSetIndex, 0})
	}
	curBlock.Code = append(curBlock.Code, bytecode...)
	return nil
}
