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

type ValStack struct {
	Value interface{}
}

type Stack []*ValStack

type VM struct {
	stack Stack
	vars  *map[string]interface{}
}

func VMFunc(vm *VM, name string) error {
	var (
		ok bool
		f  interface{}
	)
	if f, ok = (*vm.vars)[name]; !ok || reflect.ValueOf(f).Kind().String() != `func` {
		return fmt.Errorf(`unknown function %s`, name)
	}
	size := len(vm.stack)
	foo := reflect.ValueOf(f)
	/*	if count != foo.Type().NumIn() {
		return fmt.Errorf(`The number of params %s is wrong`, name)
	}*/
	count := foo.Type().NumIn()
	pars := make([]reflect.Value, count)
	for i := count; i > 0; i-- {
		pars[count-i] = reflect.ValueOf(vm.stack[size-i].Value)
	}
	result := foo.Call(pars)
	if result[len(result)-1].Interface() != nil {
		return result[len(result)-1].Interface().(error)
	}
	vm.stack[size-count] = &ValStack{Value: result[0].Interface()}
	vm.stack = vm.stack[:size-count+1]
	return nil
}

func ValueToBool(v interface{}) bool {
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

func Eval(input string, vars *map[string]interface{}) interface{} {
	vm := VM{make(Stack, 0, 1024), vars}
	bytecode := Compile([]rune(input))
	last := bytecode[len(bytecode)-1]
	if last.Cmd == CMD_ERROR {
		return fmt.Errorf(`%v [%d:%d]`, last.Value, last.Lex.Line, last.Lex.Column)
	}
	top := make([]interface{}, 8)
	for _, cmd := range bytecode {
		var bin interface{}
		size := len(vm.stack)
		if size < int(cmd.Cmd>>8) {
			return fmt.Errorf(`stack is empty [%d:%d]`, last.Lex.Line, last.Lex.Column)
		}
		for i := 1; i <= int(cmd.Cmd>>8); i++ {
			top[i-1] = vm.stack[size-i].Value
		}
		switch cmd.Cmd {
		case CMD_PUSH:
			vm.stack = append(vm.stack, &ValStack{Value: cmd.Value})
		case CMD_PUSHSTR:
			vm.stack = append(vm.stack, &ValStack{Value: cmd.Value.(string)})
		case CMD_TABLE, CMD_CALL:
			funcname := `Table`
			if cmd.Cmd == CMD_CALL {
				funcname = cmd.Value.(string)
			}
			err := VMFunc(&vm, funcname)
			if err != nil {
				return fmt.Errorf(`%s [%d:%d]`, err.Error(), last.Lex.Line, last.Lex.Column)
			}
		case CMD_VAR:
			if val, ok := (*vm.vars)[cmd.Value.(string)]; ok {
				var number int64
				switch varVal := val.(type) {
				case int:
					number = int64(varVal)
				case int64:
					number = varVal
				}
				vm.stack = append(vm.stack, &ValStack{Value: number})
			} else {
				return fmt.Errorf(`unknown identifier %s [%d:%d]`, cmd.Value.(string), last.Lex.Line, last.Lex.Column)
			}
		case CMD_NOT:
			vm.stack[size-1] = &ValStack{Value: !ValueToBool(top[0])}

		case CMD_ADD:
			bin = top[1].(int64) + top[0].(int64)
		case CMD_SUB:
			bin = top[1].(int64) - top[0].(int64)
		case CMD_MUL:
			bin = top[1].(int64) * top[0].(int64)
		case CMD_DIV:
			if top[0].(int64) == 0 {
				return fmt.Errorf(`divided by zero [%d:%d]`, last.Lex.Line, last.Lex.Column)
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
			return fmt.Errorf(`Unknown command [%d:%d]`, last.Lex.Line, last.Lex.Column)
		}
		if (cmd.Cmd >> 8) == 2 {
			vm.stack[size-2] = &ValStack{Value: bin}
			vm.stack = vm.stack[:size-1]
		}
	}
	if len(vm.stack) == 0 {
		return fmt.Errorf(`Stack empty`)
	}
	return vm.stack[len(vm.stack)-1].Value
}

func EvalIf(input string, vars *map[string]interface{}) (bool, error) {
	ret := Eval(input, vars)
	if err, ok := ret.(error); ok {
		return false, err
	}
	return ValueToBool(ret), nil
}
