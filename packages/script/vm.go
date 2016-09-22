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

/*type ValStack struct {
	Value interface{}
}*/

type RunTime struct {
	stack []interface{}
	vm    *VM
	//	vars  *map[string]interface{}
}

func (rt *RunTime) CallFunc(cmd uint16, obj *ObjInfo) error {
	var (
		count, in int
		//		f         interface{}
	)
	size := len(rt.stack)
	in = rt.vm.getInParams(obj)
	if cmd == CMD_CALLVARI {
		count = rt.stack[size-1].(int)
		size--
	} else {
		count = in
	}
	//	fmt.Println(`Count`, count, `In`, in)

	finfo := obj.Value.(ExtFuncInfo)
	foo := reflect.ValueOf(finfo.Func)
	var result []reflect.Value
	pars := make([]reflect.Value, in)
	i := count
	limit := 0
	if finfo.Variadic {
		limit = count - in + 1
	}
	for ; i > limit; i-- {
		pars[count-i] = reflect.ValueOf(rt.stack[size-i])
	}
	if i > 0 {
		pars[in-1] = reflect.ValueOf(rt.stack[in-i : size])
	}
	//	fmt.Println(`Pars`, len(pars), pars, pars[0].Interface())
	if finfo.Variadic {
		result = foo.CallSlice(pars)
	} else {
		result = foo.Call(pars)
	}
	fmt.Println(`Result`, result)
	/*
		if finfo.Variadic {
			for i := 0; i < len(pars)-1; i++ {
				pars[i] = reflect.ValueOf(rt.stack[size - count+i-1 ])
			}
			pars[len(pars)-1] = reflect.ValueOf(params[len(pars)-1:])
			result = foo.CallSlice(pars)
		} else {
			for i := 0; i < len(pars); i++ {
				pars[i] = reflect.ValueOf(params[i])
			}
			result = foo.Call(pars)
		}*/
	/*		for _, iret := range result {
			ret = append(ret, iret.Interface())
		}*/

	/*	if

		if f, ok = (*rt.vars)[name]; !ok || reflect.ValueOf(f).Kind().String() != `func` {
			return fmt.Errorf(`unknown function %s`, name)
		}
		foo := reflect.ValueOf(f)
		//	if count != foo.Type().NumIn() {
		//	return fmt.Errorf(`The number of params %s is wrong`, name)
		//
		count := foo.Type().NumIn()
		pars := make([]reflect.Value, count)
		for i := count; i > 0; i-- {
			pars[count-i] = reflect.ValueOf(rt.stack[size-i].Value)
		}
		result := foo.Call(pars)
		if result[len(result)-1].Interface() != nil {
			return result[len(result)-1].Interface().(error)
		}
		rt.stack[size-count] = &ValStack{Value: result[0].Interface()}
		rt.stack = rt.stack[:size-count+1]*/
	return nil
}

func valueToBool(v interface{}) bool {
	switch val := v.(type) {
	case int:
		if val != 0 {
			return true
		}
	case int64:
		if val != 0 {
			return true
		}
	case bool:
		return val
	}
	return false
}

func (vm *VM) RunInit() *RunTime {
	rt := RunTime{stack: make([]interface{}, 0, 1024), vm: vm}
	return &rt
}

func (rt *RunTime) RunCode(block *Block) error {
	top := make([]interface{}, 8)
	for _, cmd := range block.Code {
		var bin interface{}
		size := len(rt.stack)
		if size < int(cmd.Cmd>>8) {
			return fmt.Errorf(`stack is empty`)
		}
		for i := 1; i <= int(cmd.Cmd>>8); i++ {
			top[i-1] = rt.stack[size-i]
		}
		switch cmd.Cmd {
		case CMD_PUSH:
			rt.stack = append(rt.stack, cmd.Value)
		case CMD_PUSHSTR:
			rt.stack = append(rt.stack, cmd.Value.(string))
		case CMD_CALLVARI, CMD_CALL:
			rt.CallFunc(cmd.Cmd, cmd.Value.(*ObjInfo))
			/*			if err != nil {
						return fmt.Errorf(`%s [%d:%d]`, err.Error(), last.Lex.Line, last.Lex.Column)
					}*/

			//		case CMD_CALL:
			/*			if cmd.Cmd == CMD_CALL {
							funcname = cmd.Value.(string)
						}
						VMFunc(&vm, funcname)*/
			/*			if err != nil {
						return fmt.Errorf(`%s [%d:%d]`, err.Error(), last.Lex.Line, last.Lex.Column)
					}*/
		case CMD_VAR:
			/*			if val, ok := (*rt.vars)[cmd.Value.(string)]; ok {
							var number int64
							switch varVal := val.(type) {
							case int:
								number = int64(varVal)
							case int64:
								number = varVal
							}
							rt.stack = append(rt.stack, number)
						} else {
							return nil, fmt.Errorf(`unknown identifier %s`, cmd.Value.(string), last.Lex.Line, last.Lex.Column)
						}*/
		case CMD_NOT:
			rt.stack[size-1] = !ValueToBool(top[0])

		case CMD_ADD:
			bin = top[1].(int64) + top[0].(int64)
		case CMD_SUB:
			bin = top[1].(int64) - top[0].(int64)
		case CMD_MUL:
			bin = top[1].(int64) * top[0].(int64)
		case CMD_DIV:
			if top[0].(int64) == 0 {
				return fmt.Errorf(`divided by zero`)
			}
			bin = top[1].(int64) / top[0].(int64)
		case CMD_AND:
			bin = ValueToBool(top[1]) && ValueToBool(top[0])
		case CMD_OR:
			bin = ValueToBool(top[1]) || ValueToBool(top[0])
		case CMD_EQUAL, CMD_NOTEQ:
			bin = top[1].(int64) == top[0].(int64)
			if cmd.Cmd == CMD_NOTEQ {
				bin = !bin.(bool)
			}
		case CMD_LESS, CMD_NOTLESS:
			bin = top[1].(int64) < top[0].(int64)
			if cmd.Cmd == CMD_NOTLESS {
				bin = !bin.(bool)
			}
		case CMD_GREAT, CMD_NOTGREAT:
			bin = top[1].(int64) > top[0].(int64)
			if cmd.Cmd == CMD_NOTGREAT {
				bin = !bin.(bool)
			}
		default:
			return fmt.Errorf(`Unknown command %d`, cmd.Cmd)
		}
		if (cmd.Cmd >> 8) == 2 {
			rt.stack[size-2] = bin
			rt.stack = rt.stack[:size-1]
		}
	}
	return nil
}

func (rt *RunTime) Run(block *Block, params []interface{}, extend map[string]interface{}) ([]interface{}, error) {
	rt.RunCode(block)
	/*	if len(rt.stack) == 0 {
		return fmt.Errorf(`Stack empty`)
	}*/
	return nil, nil //rt.stack[len(rt.stack)-1].Value
}

/*
func EvalIf(input string, vars *map[string]interface{}) (bool, error) {
	ret := Eval(input, vars)
	if err, ok := ret.(error); ok {
		return false, err
	}
	return ValueToBool(ret), nil
}*/
