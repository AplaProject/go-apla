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
	"regexp"
	"strconv"
	"strings"

	"github.com/AplaProject/go-apla/packages/consts"

	log "github.com/sirupsen/logrus"
)

// ByteCode stores a command and an additional parameter.
type ByteCode struct {
	Cmd   uint16
	Value interface{}
}

// ByteCodes is the slice of ByteCode items
type ByteCodes []*ByteCode

const (
	// Types of the compiled objects

	// ObjUnknown is an unknown object. It means something wrong.
	ObjUnknown = iota
	// ObjContract is a contract object.
	ObjContract
	// ObjFunc is a function object. myfunc()
	ObjFunc
	// ObjExtFunc is an extended function object. $myfunc()
	ObjExtFunc
	// ObjVar is a variable. myvar
	ObjVar
	// ObjExtend is an extended variable. $myvar
	ObjExtend

	// CostCall is the cost of the function calling
	CostCall = 50
	// CostContract is the cost of the contract calling
	CostContract = 100
	// CostExtend is the cost of the extend function calling
	CostExtend = 10
	// CostDefault is the default maximum cost of F
	CostDefault = int64(10000000)
)

// ExtFuncInfo is the structure for the extrended function
type ExtFuncInfo struct {
	Name     string
	Params   []reflect.Type
	Results  []reflect.Type
	Auto     []string
	Variadic bool
	Func     interface{}
}

// FieldInfo describes the field of the data structure
type FieldInfo struct {
	Name string
	Type reflect.Type
	Tags string
}

// ContractInfo contains the contract information
type ContractInfo struct {
	ID       uint32
	Name     string
	Owner    *OwnerInfo
	Used     map[string]bool // Called contracts
	Tx       *[]*FieldInfo
	Settings map[string]interface{}
}

// FuncNameCmd for cmdFuncName
type FuncNameCmd struct {
	Name  string
	Count int
}

// FuncName is storing param of FuncName
type FuncName struct {
	Params   []reflect.Type
	Offset   []int
	Variadic bool
}

// FuncInfo contains the function information
type FuncInfo struct {
	Params   []reflect.Type
	Results  []reflect.Type
	Names    *map[string]FuncName
	Variadic bool
}

// VarInfo contains the variable information
type VarInfo struct {
	Obj   *ObjInfo
	Owner *Block
}

// ObjInfo is the common object type
type ObjInfo struct {
	Type  int
	Value interface{}
}

// OwnerInfo storing info about owner
type OwnerInfo struct {
	StateID  uint32 `json:"state"`
	Active   bool   `json:"active"`
	TableID  int64  `json:"tableid"`
	WalletID int64  `json:"walletid"`
	TokenID  int64  `json:"tokenid"`
}

// Block contains all information about compiled block {...} and its children
type Block struct {
	Objects  map[string]*ObjInfo
	Type     int
	Owner    *OwnerInfo
	Info     interface{}
	Parent   *Block
	Vars     []reflect.Type
	Code     ByteCodes
	Children Blocks
}

// Blocks is a slice of blocks
type Blocks []*Block

// VM is the main type of the virtual machine
type VM struct {
	Block
	ExtCost     func(string) int64
	FuncCallsDB map[string]struct{}
	Extern      bool // extern mode of compilation
	logger      *log.Entry
}

// ExtendData is used for the definition of the extended functions and variables
type ExtendData struct {
	Objects  map[string]interface{}
	AutoPars map[string]string
}

// ParseContract gets a state identifier and the name of the contract from the full name like @[id]name
func ParseContract(in string) (id uint64, name string) {
	var err error
	re := regexp.MustCompile(`(?is)^@(\d+)(\w[_\w\d]*)$`)
	ret := re.FindStringSubmatch(in)
	if len(ret) == 3 {
		id, err = strconv.ParseUint(ret[1], 10, 32)
		if err != nil {
			log.WithFields(log.Fields{"type": consts.ConversionError, "error": err, "value": ret[1]}).Error("converting state identifier from string to int while parsing contract")
		}
		name = ret[2]
	}
	return
}

// ExecContract runs the name contract where txs contains the list of parameters and
// params are the values of parameters
func ExecContract(rt *RunTime, name, txs string, params ...interface{}) (string, error) {
	var result string

	contract, ok := rt.vm.Objects[name]
	if !ok {
		log.WithFields(log.Fields{"contract_name": name, "type": consts.ContractError}).Error("unknown contract")
		return ``, fmt.Errorf(eUnknownContract, name)
	}
	logger := log.WithFields(log.Fields{"contract_name": name, "type": consts.ContractError})
	cblock := contract.Value.(*Block)
	parnames := make(map[string]bool)
	pars := strings.Split(txs, `,`)
	if len(pars) != len(params) {
		logger.WithFields(log.Fields{"contract_params_len": len(pars), "contract_params_len_needed": len(params), "type": consts.ContractError}).Error("wrong contract parameters pars")
		return ``, errContractPars
	}
	for _, ipar := range pars {
		parnames[ipar] = true
	}
	var isSignature bool
	if cblock.Info.(*ContractInfo).Tx != nil {
		for _, tx := range *cblock.Info.(*ContractInfo).Tx {
			if !parnames[tx.Name] && !strings.Contains(tx.Tags, `optional`) {
				logger.WithFields(log.Fields{"transaction_name": tx.Name, "type": consts.ContractError}).Error("transaction not defined")
				return ``, fmt.Errorf(eUndefinedParam, tx.Name)
			}
			if tx.Name == `Signature` {
				isSignature = true
			}
		}
	}
	if _, ok := (*rt.extend)[`loop_`+name]; ok {
		logger.WithFields(log.Fields{"type": consts.ContractError, "contract_name": name}).Error("there is loop in contract")
		return ``, fmt.Errorf(eContractLoop, name)
	}
	(*rt.extend)[`loop_`+name] = true
	defer delete(*rt.extend, `loop_`+name)
	for i, ipar := range pars {
		(*rt.extend)[ipar] = params[i]
	}
	prevparent := (*rt.extend)[`parent`]
	parent := ``
	for i := len(rt.blocks) - 1; i >= 0; i-- {
		if rt.blocks[i].Block.Type == ObjFunc && rt.blocks[i].Block.Parent != nil &&
			rt.blocks[i].Block.Parent.Type == ObjContract {
			parent = rt.blocks[i].Block.Parent.Info.(*ContractInfo).Name
			fid, fname := ParseContract(parent)
			cid, _ := ParseContract(name)
			if len(fname) > 0 {
				if fid == 0 {
					parent = `@` + fname
				} else if fid == cid {
					parent = fname
				}
			}
			break
		}
	}
	rt.cost -= CostContract
	var stackCont func(interface{}, string)
	if stack, ok := (*rt.extend)[`stack_cont`]; ok && (*rt.extend)[`sc`] != nil {
		stackCont = stack.(func(interface{}, string))
		stackCont((*rt.extend)[`sc`], name)
	}
	if (*rt.extend)[`sc`] != nil && isSignature {
		obj := rt.vm.Objects[`check_signature`]
		finfo := obj.Value.(ExtFuncInfo)
		if err := finfo.Func.(func(*map[string]interface{}, string) error)(rt.extend, name); err != nil {
			logger.WithFields(log.Fields{"error": err, "func_name": finfo.Name, "type": consts.ContractError}).Error("executing exended function")
			return ``, err
		}
	}
	for _, method := range []string{`init`, `conditions`, `action`} {
		if block, ok := (*cblock).Objects[method]; ok && block.Type == ObjFunc {
			rtemp := rt.vm.RunInit(rt.cost)
			(*rt.extend)[`parent`] = parent
			_, err := rtemp.Run(block.Value.(*Block), nil, rt.extend)
			rt.cost = rtemp.cost
			if err != nil {
				logger.WithFields(log.Fields{"error": err, "method_name": method, "type": consts.ContractError}).Error("executing contract method")
				return ``, err
			}
		}
	}
	if stackCont != nil {
		stackCont((*rt.extend)[`sc`], ``)
	}
	(*rt.extend)[`parent`] = prevparent
	if (*rt.extend)[`result`] != nil {
		result = fmt.Sprint((*rt.extend)[`result`])
	}
	return result, nil
}

// NewVM creates a new virtual machine
func NewVM() *VM {
	vm := VM{}
	vm.Objects = make(map[string]*ObjInfo)
	// Reserved 256 indexes for system purposes
	vm.Children = make(Blocks, 256, 1024)
	vm.Extend(&ExtendData{map[string]interface{}{"ExecContract": ExecContract, "CallContract": ExContract,
		"Settings": GetSettings},
		map[string]string{
			`*script.RunTime`: `rt`,
		}})
	vm.logger = log.WithFields(log.Fields{"extern": vm.Extern, "vm_block_type": vm.Block.Type})
	return &vm
}

// Extend sets the extended variables and functions
func (vm *VM) Extend(ext *ExtendData) {
	for key, item := range ext.Objects {
		fobj := reflect.ValueOf(item).Type()
		switch fobj.Kind() {
		case reflect.Func:
			data := ExtFuncInfo{key, make([]reflect.Type, fobj.NumIn()),
				make([]reflect.Type, fobj.NumOut()), make([]string, fobj.NumIn()),
				fobj.IsVariadic(), item}
			for i := 0; i < fobj.NumIn(); i++ {
				if isauto, ok := ext.AutoPars[fobj.In(i).String()]; ok {
					data.Auto[i] = isauto
				}
				data.Params[i] = fobj.In(i)
			}
			for i := 0; i < fobj.NumOut(); i++ {
				data.Results[i] = fobj.Out(i)
			}
			vm.Objects[key] = &ObjInfo{ObjExtFunc, data}
		}
	}
}

func (vm *VM) getObjByName(name string) (ret *ObjInfo) {
	var ok bool
	names := strings.Split(name, `.`)
	block := &vm.Block
	for i, name := range names {
		ret, ok = block.Objects[name]
		if !ok {
			return nil
		}
		if i == len(names)-1 {
			return
		}
		if ret.Type != ObjContract && ret.Type != ObjFunc {
			return nil
		}
		block = ret.Value.(*Block)
	}
	return
}

func (vm *VM) getObjByNameExt(name string, state uint32) (ret *ObjInfo) {
	sname := StateName(state, name)
	if ret = vm.getObjByName(name); ret == nil && len(sname) > 0 {
		ret = vm.getObjByName(sname)
	}
	return
}

func (vm *VM) getInParams(ret *ObjInfo) int {
	if ret.Type == ObjExtFunc {
		return len(ret.Value.(ExtFuncInfo).Params)
	}
	return len(ret.Value.(*Block).Info.(*FuncInfo).Params)
}

// Call executes the name object with the specified params and extended variables and functions
func (vm *VM) Call(name string, params []interface{}, extend *map[string]interface{}) (ret []interface{}, err error) {
	var obj *ObjInfo
	if state, ok := (*extend)[`rt_state`]; ok {
		obj = vm.getObjByNameExt(name, state.(uint32))
	} else {
		obj = vm.getObjByName(name)
	}
	if obj == nil {
		vm.logger.WithFields(log.Fields{"type": consts.VMError, "vm_func_name": name}).Error("unknown function")
		return nil, fmt.Errorf(`unknown function %s`, name)
	}
	switch obj.Type {
	case ObjFunc:
		rt := vm.RunInit(CostDefault)
		ret, err = rt.Run(obj.Value.(*Block), params, extend)
	case ObjExtFunc:
		finfo := obj.Value.(ExtFuncInfo)
		foo := reflect.ValueOf(finfo.Func)
		var result []reflect.Value
		pars := make([]reflect.Value, len(finfo.Params))
		if finfo.Variadic {
			for i := 0; i < len(pars)-1; i++ {
				pars[i] = reflect.ValueOf(params[i])
			}
			pars[len(pars)-1] = reflect.ValueOf(params[len(pars)-1:])
			result = foo.CallSlice(pars)
		} else {
			for i := 0; i < len(pars); i++ {
				pars[i] = reflect.ValueOf(params[i])
			}
			result = foo.Call(pars)
		}
		for _, iret := range result {
			ret = append(ret, iret.Interface())
		}
	default:
		vm.logger.WithFields(log.Fields{"type": consts.VMError, "vm_func_name": name}).Error("unknown function")
		return nil, fmt.Errorf(`unknown function %s`, name)
	}
	return ret, err
}

// ExContract executes the name contract in the state with spoecified parameters
func ExContract(rt *RunTime, state uint32, name string, params map[string]interface{}) (string, error) {

	name = StateName(state, name)
	contract, ok := rt.vm.Objects[name]
	if !ok {
		log.WithFields(log.Fields{"contract_name": name, "type": consts.ContractError}).Error("unknown contract")
		return ``, fmt.Errorf(eUnknownContract, name)
	}
	if params == nil {
		params = make(map[string]interface{})
	}
	logger := log.WithFields(log.Fields{"contract_name": name, "type": consts.ContractError})
	names := make([]string, 0)
	vals := make([]interface{}, 0)
	cblock := contract.Value.(*Block)
	if cblock.Info.(*ContractInfo).Tx != nil {
		for _, tx := range *cblock.Info.(*ContractInfo).Tx {
			val, ok := params[tx.Name]
			if !ok && !strings.Contains(tx.Tags, `optional`) {
				logger.WithFields(log.Fields{"transaction_name": tx.Name, "type": consts.ContractError}).Error("transaction not defined")
				return ``, fmt.Errorf(eUndefinedParam, tx.Name)
			}
			names = append(names, tx.Name)
			vals = append(vals, val)
		}
	}
	if len(vals) == 0 {
		vals = append(vals, ``)
	}
	return ExecContract(rt, name, strings.Join(names, `,`), vals...)
}

// GetSettings returns the value of the parameter
func GetSettings(rt *RunTime, cntname, name string) (interface{}, error) {
	contract, ok := rt.vm.Objects[cntname]
	if !ok {
		log.WithFields(log.Fields{"contract_name": name, "type": consts.ContractError}).Error("unknown contract")
		return nil, fmt.Errorf(`unknown contract %s`, cntname)
	}
	cblock := contract.Value.(*Block)
	if cblock.Info.(*ContractInfo).Settings != nil {
		if val, ok := cblock.Info.(*ContractInfo).Settings[name]; ok {
			return val, nil
		}
	}
	return ``, nil
}
