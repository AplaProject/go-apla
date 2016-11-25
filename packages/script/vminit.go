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
	"strings"
)

type ByteCode struct {
	Cmd   uint16
	Value interface{}
}

type ByteCodes []*ByteCode

const (
	OBJ_UNKNOWN = iota
	OBJ_CONTRACT
	OBJ_FUNC
	OBJ_EXTFUNC
	OBJ_VAR
	OBJ_EXTEND
)

type ExtFuncInfo struct {
	Params   []reflect.Type
	Results  []reflect.Type
	Auto     []string
	Variadic bool
	Func     interface{}
}

type FieldInfo struct {
	Name string
	Type reflect.Type
	Tags string
}

type ContractInfo struct {
	Id   uint32
	Name string
	Tx   *[]*FieldInfo
}

type FuncInfo struct {
	Params   []reflect.Type
	Results  []reflect.Type
	Variadic bool
}

type VarInfo struct {
	Obj   *ObjInfo
	Owner *Block
}

type ObjInfo struct {
	Type  int
	Value interface{}
}

type Block struct {
	Objects  map[string]*ObjInfo
	Type     int
	Info     interface{}
	Vars     []reflect.Type
	Code     ByteCodes
	Children Blocks
}

type Blocks []*Block

type VM struct {
	Block
}

type ExtendData struct {
	Objects  map[string]interface{}
	AutoPars map[string]string
}

func ExecContract(rt *RunTime, name, txs string, params ...interface{}) error {
	//	fmt.Println(`ExecContract`, rt, name, txs, params)

	contract, ok := rt.vm.Objects[name]
	if !ok {
		return fmt.Errorf(`unknown contract %s`, name)
	}
	cblock := contract.Value.(*Block)
	parnames := make(map[string]bool)
	pars := strings.Split(txs, `,`)
	if len(pars) != len(params) {
		return fmt.Errorf(`wrong contract parameters`)
	}
	for _, ipar := range pars {
		parnames[ipar] = true
	}
	for _, tx := range *cblock.Info.(*ContractInfo).Tx {
		if !parnames[tx.Name] {
			return fmt.Errorf(`%s is not defined`, tx.Name)
		}
	}
	for i, ipar := range pars {
		(*rt.extend)[ipar] = params[i]
	}
	for _, method := range []string{`init`, `front`, `main`} {
		if block, ok := (*cblock).Objects[method]; ok && block.Type == OBJ_FUNC {
			rtemp := rt.vm.RunInit()
			_, err := rtemp.Run(block.Value.(*Block), nil, rt.extend)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func NewVM() *VM {
	vm := VM{}
	vm.Objects = make(map[string]*ObjInfo)
	// Reserved 256 indexes for system purposes
	vm.Children = make(Blocks, 256, 1024)
	vm.Extend(&ExtendData{map[string]interface{}{"ExecContract": ExecContract},
		map[string]string{
			`*script.RunTime`: `rt`,
		}})
	//	vm.Extend(&ExtendData{map[string]interface{}{"Bool": valueToBool}, nil})
	return &vm
}

func (vm *VM) Extend(ext *ExtendData) {
	for key, item := range ext.Objects {
		fobj := reflect.ValueOf(item).Type()
		switch fobj.Kind() {
		case reflect.Func:
			data := ExtFuncInfo{make([]reflect.Type, fobj.NumIn()),
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
			//			fmt.Println(`Extend`, data)
			vm.Objects[key] = &ObjInfo{OBJ_EXTFUNC, data}
		}
	}
}

func (vm *VM) getObjByName(name string) (ret *ObjInfo) {
	var ok bool
	names := strings.Split(name, `.`)
	block := &vm.Block
	//	fmt.Println(block.Objects)
	for i, name := range names {
		ret, ok = block.Objects[name]
		if !ok {
			return nil
		}
		if i == len(names)-1 {
			return
		}
		if ret.Type != OBJ_CONTRACT && ret.Type != OBJ_FUNC {
			return nil
		}
		block = ret.Value.(*Block)
	}
	return
}

func (vm *VM) getInParams(ret *ObjInfo) int {
	if ret.Type == OBJ_EXTFUNC {
		return len(ret.Value.(ExtFuncInfo).Params)
	}
	return len(ret.Value.(*Block).Info.(*FuncInfo).Params)
}

func (vm *VM) Call(name string, params []interface{}, extend *map[string]interface{}) (ret []interface{}, err error) {
	obj := vm.getObjByName(name)
	if obj == nil {
		return nil, fmt.Errorf(`unknown function`, name)
	}
	switch obj.Type {
	case OBJ_FUNC:
		rt := vm.RunInit()
		ret, err = rt.Run(obj.Value.(*Block), params, extend)
	case OBJ_EXTFUNC:
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
		return nil, fmt.Errorf(`unknown function`, name)
	}
	return ret, err
}
