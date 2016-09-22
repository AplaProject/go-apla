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
)

type ExtFuncInfo struct {
	Params   []reflect.Kind
	Results  []reflect.Kind
	Variadic bool
	Func     interface{}
}

type ObjInfo struct {
	Type  int
	Value interface{}
}

type Block struct {
	Objects  map[string]*ObjInfo
	Code     ByteCodes
	Children Blocks
}

type Blocks []*Block

type VM struct {
	Block
}

func VMInit(obj map[string]interface{}) *VM {
	vm := VM{}
	vm.Objects = make(map[string]*ObjInfo)

	for key, item := range obj {
		fobj := reflect.ValueOf(item).Type()
		switch fobj.Kind() {
		case reflect.Func:
			data := ExtFuncInfo{make([]reflect.Kind, fobj.NumIn()),
				make([]reflect.Kind, fobj.NumOut()),
				fobj.IsVariadic(), item}
			for i := 0; i < fobj.NumIn(); i++ {
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

func (vm *VM) getObjByName(name string) *ObjInfo {
	ret, ok := vm.Objects[name]
	if !ok {
		return nil
	}
	return ret
}

func (vm *VM) Call(name string, params []interface{}, extend map[string]interface{}) ([]interface{}, error) {
	obj := vm.getObjByName(name)
	if obj == nil {
		return nil, fmt.Errorf(`unknown function`, name)
	}
	switch obj.Type {
	case OBJ_EXTFUNC:
		finfo := obj.Value.(ExtFuncInfo)
		foo := reflect.ValueOf(finfo.Func)
		pars := make([]reflect.Value, len(finfo.Params))
		for i := 0; i < len(pars); i++ {
			pars[i] = reflect.ValueOf(params[i])
		}
		if finfo.Variadic {

			for i := len(pars); i < len(params); i++ {
				pars = append(pars, reflect.ValueOf(params[i]))
			}
			fmt.Println(`Pars`, pars)
			result := foo.CallSlice(pars)
			fmt.Println(`Result`, result)
		} else {
			result := foo.Call(pars)
			fmt.Println(`Result`, result)
		}
	default:
		return nil, fmt.Errorf(`unknown function`, name)
	}
	return nil, nil
}
