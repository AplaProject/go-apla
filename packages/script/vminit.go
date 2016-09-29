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
	Params   []reflect.Kind
	Results  []reflect.Kind
	Auto     []string
	Variadic bool
	Func     interface{}
}

type FuncInfo struct {
	Params   []reflect.Kind
	Results  []reflect.Kind
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
	Vars     []reflect.Kind
	Code     ByteCodes
	Children Blocks
}

type Blocks []*Block

type VM struct {
	Block
}

func VMInit(obj map[string]interface{}, autopar map[string]string) *VM {
	vm := VM{}
	vm.Objects = make(map[string]*ObjInfo)

	for key, item := range obj {
		fobj := reflect.ValueOf(item).Type()
		switch fobj.Kind() {
		case reflect.Func:
			data := ExtFuncInfo{make([]reflect.Kind, fobj.NumIn()),
				make([]reflect.Kind, fobj.NumOut()), make([]string, fobj.NumIn()),
				fobj.IsVariadic(), item}
			for i := 0; i < fobj.NumIn(); i++ {
				if isauto, ok := autopar[fobj.In(i).String()]; ok {
					data.Auto[i] = isauto
				}
				data.Params[i] = fobj.In(i).Kind()
			}
			for i := 0; i < fobj.NumOut(); i++ {
				data.Results[i] = fobj.Out(i).Kind()
			}
			vm.Objects[key] = &ObjInfo{OBJ_EXTFUNC, data}
		}
	}
	return &vm
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
