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
	"strconv"

	"github.com/AplaProject/go-apla/packages/consts"

	log "github.com/sirupsen/logrus"
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

// The compiler converts the sequence of lexemes into the bytecodes using a finite state machine the same as
// it was implemented in lexical analysis. The difference lays in that we do not convert the list of
// states and transitions to the intermediate array.

/* Byte code could be described as a tree where functions and contracts are on the top level and
nesting goes further according to nesting of bracketed brackets. Tree nodes are structures of
'Block' type. For instance,
 func a {
	 if b {
		 while d {

		 }
	 }
	 if c {
	 }
 }
 will be compiled into Block(a) which will have two child blocks Block (b) and Block (c) that
 are responsible for executing bytecode inside if. Block (b) will have a child Block (d) with
 a cycle.
*/

const (
	// The list of state types
	stateRoot = iota
	stateBody
	stateBlock
	stateContract
	stateFunc
	stateFParams
	stateFParam
	stateFParamTYPE
	stateFTail
	stateFResult
	stateFDot
	stateVar
	stateVarType
	stateAssignEval
	stateAssign
	stateTX
	stateSettings
	stateConsts
	stateConstsAssign
	stateConstsValue
	stateFields
	stateEval

	// The list of state flags
	statePush     = 0x0100
	statePop      = 0x0200
	stateStay     = 0x0400
	stateToBlock  = 0x0800
	stateToBody   = 0x1000
	stateFork     = 0x2000
	stateToFork   = 0x4000
	stateLabel    = 0x8000
	stateMustEval = 0x010000

	flushMark = 0x100000
)

const (
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
	errStrNum                // must be number or string
)

const (
	// This is a list of identifiers for functions that will generate a bytecode for
	// the corresponding cases
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
	cfFTail
	cfFNameParam
	cfAssignVar
	cfAssign
	cfTX
	cfSettings
	cfConstName
	cfConstValue
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
	// Array of operations and their priority
	opers = map[uint32]operPrior{
		isOr: {cmdOr, 10}, isAnd: {cmdAnd, 15}, isEqEq: {cmdEqual, 20}, isNotEq: {cmdNotEq, 20},
		isLess: {cmdLess, 22}, isGrEq: {cmdNotLess, 22}, isGreat: {cmdGreat, 22}, isLessEq: {cmdNotGreat, 22},
		isPlus: {cmdAdd, 25}, isMinus: {cmdSub, 25}, isAsterisk: {cmdMul, 30},
		isSolidus: {cmdDiv, 30}, isSign: {cmdSign, cmdUnary}, isNot: {cmdNot, cmdUnary}, isLPar: {cmdSys, 0xff}, isRPar: {cmdSys, 0},
	}
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
		fFtail,
		fFNameParam,
		fAssignVar,
		fAssign,
		fTx,
		fSettings,
		fConstName,
		fConstValue,
		fField,
		fFieldType,
		fFieldTag,
		fWhile,
		fContinue,
		fBreak,
		fCmdError,
	}

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
			lexKeyword | (keySettings << 8): {stateSettings, cfSettings},
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
			lexComment: {stateFParams, 0},
			isLPar:     {stateFParam, 0},
			0:          {stateFResult | stateStay, 0},
		},
		{ // stateFParam
			lexNewLine: {stateFParam, 0},
			lexComment: {stateFParam, 0},
			lexIdent:   {stateFParamTYPE, cfFParam},
			// lexType:    {stateFParam, cfFType},
			isComma: {stateFParam, 0},
			isRPar:  {stateFResult, 0},
			0:       {errParams, cfError},
		},
		{ // stateFParamTYPE
			lexComment:                  {stateFParamTYPE, 0},
			lexIdent:                    {stateFParamTYPE, cfFParam},
			lexType:                     {stateFParam, cfFType},
			lexKeyword | (keyTail << 8): {stateFTail, cfFTail},
			isComma:                     {stateFParamTYPE, 0},
			//			isRPar:   {stateFResult, 0},
			0: {errVarType, cfError},
		},
		{ // stateFTail
			lexNewLine: {stateFTail, 0},
			isRPar:     {stateFResult, 0},
			0:          {errParams, cfError},
		},
		{ // stateFResult
			lexNewLine: {stateFResult, 0},
			isDot:      {stateFDot, 0},
			lexType:    {stateFResult, cfFResult},
			isComma:    {stateFResult, 0},
			0:          {stateBlock | stateStay, 0},
		},
		{ // stateFDot
			lexNewLine: {stateFDot, 0},
			lexIdent:   {stateFParams, cfFNameParam},
			0:          {errMustName, cfError},
		},
		{ // stateVar
			lexNewLine: {stateBody, 0},
			lexIdent:   {stateVarType, cfFParam},
			isRCurly:   {stateBody | stateStay, 0},
			isComma:    {stateVar, 0},
			0:          {errVars, cfError},
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
		{ // stateSettings
			lexNewLine: {stateSettings, 0},
			isLCurly:   {stateConsts, 0},
			0:          {errMustLCurly, cfError},
		},
		{ // stateConsts
			lexNewLine: {stateConsts, 0},
			lexComment: {stateConsts, 0},
			isComma:    {stateConsts, 0},
			lexIdent:   {stateConstsAssign, cfConstName},
			isRCurly:   {stateToBody, 0},
			0:          {errMustRCurly, cfError},
		},
		{ // stateConstsAssign
			isEq: {stateConstsValue, 0},
			0:    {errAssign, cfError},
		},
		{ // stateConstsValue
			lexString: {stateConsts, cfConstValue},
			lexNumber: {stateConsts, cfConstValue},
			0:         {errStrNum, cfError},
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
		`unknown command`,          // errUnknownCmd
		`must be the name`,         // errMustName
		`must be '{'`,              // errMustLCurly
		`must be '}'`,              // errMustRCurly
		`wrong parameters`,         // errParams
		`wrong variables`,          // errVars
		`must be type`,             // errVarType
		`must be '='`,              // errAssign
		`must be number or string`, // errStrNum
	}
	fmt.Printf("%s %x %v [Ln:%d Col:%d]\r\n", errors[state], lexem.Type, lexem.Value, lexem.Line, lexem.Column)
	logger := lexem.GetLogger()
	if lexem.Type == lexNewLine {
		logger.WithFields(log.Fields{"error": errors[state], "lex_value": lexem.Value, "type": consts.ParseError}).Error("unexpected new line")
		return fmt.Errorf(`%s (unexpected new line) [Ln:%d]`, errors[state], lexem.Line-1)
	}
	logger.WithFields(log.Fields{"error": errors[state], "lex_value": lexem.Value, "type": consts.ParseError}).Error("parsing error")
	return fmt.Errorf(`%s %x %v [Ln:%d Col:%d]`, errors[state], lexem.Type, lexem.Value, lexem.Line, lexem.Column)
}

func fFuncResult(buf *[]*Block, state int, lexem *Lexem) error {
	fblock := (*buf)[len(*buf)-1].Info.(*FuncInfo)
	(*fblock).Results = append((*fblock).Results, lexem.Value.(reflect.Type))
	return nil
}

func fReturn(buf *[]*Block, state int, lexem *Lexem) error {
	(*(*buf)[len(*buf)-1]).Code = append((*(*buf)[len(*buf)-1]).Code, &ByteCode{cmdReturn, 0})
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
		if fblock.Names == nil {
			fblock.Params = append(fblock.Params, reflect.TypeOf(nil))
		} else {
			for key := range *fblock.Names {
				if key[0] == '_' {
					name := key[1:]
					params := append((*fblock.Names)[name].Params, reflect.TypeOf(nil))
					offset := append((*fblock.Names)[name].Offset, len(block.Vars))
					(*fblock.Names)[name] = FuncName{Params: params, Offset: offset}
					break
				}
			}
		}
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
		if fblock.Names == nil {
			for pkey, param := range fblock.Params {
				if param == reflect.TypeOf(nil) {
					fblock.Params[pkey] = lexem.Value.(reflect.Type)
				}
			}
		} else {
			for key := range *fblock.Names {
				if key[0] == '_' {
					for pkey, param := range (*fblock.Names)[key[1:]].Params {
						if param == reflect.TypeOf(nil) {
							(*fblock.Names)[key[1:]].Params[pkey] = lexem.Value.(reflect.Type)
						}
					}
					break
				}
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

func fFtail(buf *[]*Block, state int, lexem *Lexem) error {
	var used bool
	block := (*buf)[len(*buf)-1]

	fblock := block.Info.(*FuncInfo)
	if fblock.Names == nil {
		for pkey, param := range fblock.Params {
			if param == reflect.TypeOf(nil) {
				if used {
					return fmt.Errorf(`... parameter must be one`)
				}
				fblock.Params[pkey] = reflect.TypeOf([]interface{}{})
				used = true
			}
		}
		block.Info.(*FuncInfo).Variadic = true
	} else {
		for key := range *fblock.Names {
			if key[0] == '_' {
				name := key[1:]
				for pkey, param := range (*fblock.Names)[name].Params {
					if param == reflect.TypeOf(nil) {
						if used {
							return fmt.Errorf(`... parameter must be one`)
						}
						(*fblock.Names)[name].Params[pkey] = reflect.TypeOf([]interface{}{})
						used = true
					}
				}
				offset := append((*fblock.Names)[name].Offset, len(block.Vars))
				(*fblock.Names)[name] = FuncName{Params: (*fblock.Names)[name].Params,
					Offset: offset, Variadic: true}
				break
			}
		}
	}
	for vkey, ivar := range block.Vars {
		if ivar == reflect.TypeOf(nil) {
			block.Vars[vkey] = reflect.TypeOf([]interface{}{})
		}
	}
	return nil
}

func fFNameParam(buf *[]*Block, state int, lexem *Lexem) error {
	block := (*buf)[len(*buf)-1]

	fblock := block.Info.(*FuncInfo)
	if fblock.Names == nil {
		names := make(map[string]FuncName)
		fblock.Names = &names
	}
	for key := range *fblock.Names {
		if key[0] == '_' {
			delete(*fblock.Names, key)
		}
	}
	(*fblock.Names)[`_`+lexem.Value.(string)] = FuncName{}

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
			logger := lexem.GetLogger()
			logger.WithFields(log.Fields{"type": consts.ParseError, "lex_value": lexem.Value.(string)}).Error("unknown variable")
			return fmt.Errorf(`unknown variable %s`, lexem.Value.(string))
		}
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
	logger := lexem.GetLogger()
	if contract.Type != ObjContract {
		logger.WithFields(log.Fields{"type": consts.ParseError, "contract_type": contract.Type, "lex_value": lexem.Value}).Error("data can only be in contract")
		return fmt.Errorf(`data can only be in contract`)
	}
	(*contract).Info.(*ContractInfo).Tx = new([]*FieldInfo)
	return nil
}

func fSettings(buf *[]*Block, state int, lexem *Lexem) error {
	contract := (*buf)[len(*buf)-1]
	if contract.Type != ObjContract {
		logger := lexem.GetLogger()
		logger.WithFields(log.Fields{"type": consts.ParseError, "contract_type": contract.Type, "lex_value": lexem.Value}).Error("data can only be in contract")
		return fmt.Errorf(`data can only be in contract`)
	}
	(*contract).Info.(*ContractInfo).Settings = make(map[string]interface{})
	return nil
}

func fConstName(buf *[]*Block, state int, lexem *Lexem) error {
	sets := (*(*buf)[len(*buf)-1]).Info.(*ContractInfo).Settings
	sets[lexem.Value.(string)] = nil
	return nil
}

func fConstValue(buf *[]*Block, state int, lexem *Lexem) error {
	sets := (*(*buf)[len(*buf)-1]).Info.(*ContractInfo).Settings
	for key, val := range sets {
		if val == nil {
			sets[key] = lexem.Value
			break
		}
	}
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
		logger := lexem.GetLogger()
		logger.WithFields(log.Fields{"type": consts.ParseError}).Error("there is not if before")
		return fmt.Errorf(`there is not if before %v [Ln:%d Col:%d]`, lexem.Type, lexem.Line, lexem.Column)
	}
	(*(*buf)[len(*buf)-2]).Code = append(code, &ByteCode{cmdElse, (*buf)[len(*buf)-1]})
	return nil
}

// StateName checks the name of the contract and modifies it to @[state]name if it is necessary.
func StateName(state uint32, name string) string {
	if len(name) < 3 {
		return name
	}
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
		fblock.Info = &ContractInfo{ID: uint32(len(prev.Children) - 1), Name: name,
			Owner: (*buf)[0].Owner}
	default:
		itype = ObjFunc
		fblock.Info = &FuncInfo{}
	}
	fblock.Type = itype
	prev.Objects[name] = &ObjInfo{Type: itype, Value: fblock}
	return nil
}

// CompileBlock compile the source code into the Block structure with a byte-code
func (vm *VM) CompileBlock(input []rune, owner *OwnerInfo) (*Block, error) {
	root := &Block{Info: owner.StateID, Owner: owner}
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
		}
		if (newState.NewState & stateToFork) > 0 {
			i = fork
			fork = 0
			lexem = lexems[i]
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
				log.WithFields(log.Fields{"type": consts.ParseError}).Error("there is not eval expression")
				return nil, fmt.Errorf("there is not eval expression")
			}
			nextState = curState
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
	return root, nil
}

// FlushBlock loads the compiled Block into the virtual machine
func (vm *VM) FlushBlock(root *Block) {
	shift := len(vm.Children)
	for key, item := range root.Objects {
		if cur, ok := vm.Objects[key]; ok {
			switch item.Type {
			case ObjContract:
				root.Objects[key].Value.(*Block).Info.(*ContractInfo).ID = cur.Value.(*Block).Info.(*ContractInfo).ID + flushMark
			case ObjFunc:
				root.Objects[key].Value.(*Block).Info.(*FuncInfo).ID = cur.Value.(*Block).Info.(*FuncInfo).ID + flushMark
				vm.Objects[key].Value = root.Objects[key].Value
			}
		}
		vm.Objects[key] = item
	}
	for _, item := range root.Children {
		switch item.Type {
		case ObjContract:
			if item.Info.(*ContractInfo).ID > flushMark {
				item.Info.(*ContractInfo).ID -= flushMark
				vm.Children[item.Info.(*ContractInfo).ID] = item
				shift--
				continue
			}
			item.Parent = &vm.Block
			item.Info.(*ContractInfo).ID += uint32(shift)
		case ObjFunc:
			if item.Info.(*FuncInfo).ID > flushMark {
				item.Info.(*FuncInfo).ID -= flushMark
				vm.Children[item.Info.(*FuncInfo).ID] = item
				shift--
				continue
			}
			item.Parent = &vm.Block
			item.Info.(*FuncInfo).ID += uint32(shift)
		}
		vm.Children = append(vm.Children, item)
	}
}

// FlushExtern switches off the extern mode of the compilation
func (vm *VM) FlushExtern() {
	vm.Extern = false
	return
}

// Compile compiles a source code and loads the byte-code into the virtual machine
func (vm *VM) Compile(input []rune, owner *OwnerInfo) error {
	root, err := vm.CompileBlock(input, owner)
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

// This function is responsible for the compilation of expressions
func (vm *VM) compileEval(lexems *Lexems, ind *int, block *[]*Block) error {
	i := *ind
	curBlock := (*block)[len(*block)-1]

	buffer := make(ByteCodes, 0, 20)
	bytecode := make(ByteCodes, 0, 100)
	parcount := make([]int, 0, 20)
	setIndex := false
main:
	for ; i < len(*lexems); i++ {
		var cmd *ByteCode
		var call bool
		lexem := (*lexems)[i]
		logger := lexem.GetLogger()
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
					logger.WithFields(log.Fields{"lex_value": lexem.Value.(string), "type": consts.ParseError}).Error("there is not pair")
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
				if prev := buffer[len(buffer)-1]; prev.Cmd == cmdFuncName {
					buffer = buffer[:len(buffer)-1]
					(*prev).Value = FuncNameCmd{Name: prev.Value.(FuncNameCmd).Name,
						Count: parcount[len(parcount)-1]}
					parcount = parcount[:len(parcount)-1]
					bytecode = append(bytecode, prev)
				}
				if prev := buffer[len(buffer)-1]; prev.Cmd == cmdCall || prev.Cmd == cmdCallVari {
					if prev.Value.(*ObjInfo).Type == ObjFunc && prev.Value.(*ObjInfo).Value.(*Block).Info.(*FuncInfo).Names != nil {
						if bytecode[len(bytecode)-1].Cmd != cmdFuncName {
							bytecode = append(bytecode, &ByteCode{cmdPush, nil})
						}
						if i < len(*lexems)-4 && (*lexems)[i+1].Type == isDot {
							if (*lexems)[i+2].Type != lexIdent {
								log.WithFields(log.Fields{"type": consts.ParseError}).Error("must be the name of the tail")
								return fmt.Errorf(`must be the name of the tail`)
							}
							names := prev.Value.(*ObjInfo).Value.(*Block).Info.(*FuncInfo).Names
							if _, ok := (*names)[(*lexems)[i+2].Value.(string)]; !ok {
								log.WithFields(log.Fields{"type": consts.ParseError, "tail": (*lexems)[i+2].Value.(string)}).Error("unknown function tail")
								return fmt.Errorf(`unknown function tail %s`, (*lexems)[i+2].Value.(string))
							}
							buffer = append(buffer, &ByteCode{cmdFuncName, FuncNameCmd{Name: (*lexems)[i+2].Value.(string)}})
							count := 0
							if (*lexems)[i+3].Type != isRPar {
								count++
							}
							parcount = append(parcount, count)
							i += 2
							break
						}
					}
					count := parcount[len(parcount)-1]
					parcount = parcount[:len(parcount)-1]
					var errtext string
					switch prev.Value.(*ObjInfo).Type {
					case ObjFunc:
						finfo := prev.Value.(*ObjInfo).Value.(*Block).Info.(*FuncInfo)
						if count != len(finfo.Params) && (!finfo.Variadic ||
							count < len(finfo.Params)-1) {
							errtext = fmt.Sprintf(eWrongParams, getNameByObj(prev.Value.(*ObjInfo)),
								len(finfo.Params))
						}
					case ObjExtFunc:
						extinfo := prev.Value.(*ObjInfo).Value.(ExtFuncInfo)
						wantlen := len(extinfo.Params)
						for _, v := range extinfo.Auto {
							if len(v) > 0 {
								wantlen--
							}
						}
						if count != wantlen && (!extinfo.Variadic || count < wantlen) {
							errtext = fmt.Sprintf(eWrongParams, extinfo.Name, wantlen)
						}
					}
					if len(errtext) > 0 {
						logger.WithFields(log.Fields{"error": errtext, "type": consts.ParseError}).Error(errtext)
						return fmt.Errorf(errtext)
					}
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
					logger.WithFields(log.Fields{"lex_value": lexem.Value.(string), "type": consts.ParseError}).Error("there is not pair")
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
				logger.WithFields(log.Fields{"lex_value": strconv.FormatUint(uint64(lexem.Value.(uint32)), 10), "type": consts.ParseError}).Error("unknown operator")
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
			if objInfo == nil && (!vm.Extern || i > *ind || i >= len(*lexems)-2 || (*lexems)[i+1].Type != isLPar) {
				logger.WithFields(log.Fields{"lex_value": lexem.Value.(string), "type": consts.ParseError}).Error("unknown identifier")
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
						logger.WithFields(log.Fields{"lex_value": lexem.Value.(string), "type": consts.ParseError}).Error("unknown function")
						return fmt.Errorf(`unknown function %s`, lexem.Value.(string))
					}
					if objInfo.Type == ObjContract {
						objInfo, tobj = vm.findObj(`ExecContract`, block)
						isContract = true
					}
					cmdCall := uint16(cmdCall)
					if (objInfo.Type == ObjExtFunc && objInfo.Value.(ExtFuncInfo).Variadic) ||
						(objInfo.Type == ObjFunc && objInfo.Value.(*Block).Info.(*FuncInfo).Variadic) {
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
						count++
						bytecode = append(bytecode, &ByteCode{cmdPush, (*block)[0].Info.(uint32)})
					}
					parcount = append(parcount, count)
					call = true
				}
				if (*lexems)[i+1].Type == isLBrack {
					if objInfo == nil || objInfo.Type != ObjVar {
						logger.WithFields(log.Fields{"lex_value": lexem.Value.(string), "type": consts.ParseError}).Error("unknown variable")
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
			log.WithFields(log.Fields{"type": consts.ParseError}).Error("there is not pair")
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
