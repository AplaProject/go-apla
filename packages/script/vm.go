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

const (
	STATUS_NORMAL = iota
	STATUS_RETURN
)

type BlockStack struct {
	Block  *Block
	Offset int
}

type RunTime struct {
	stack  []interface{}
	blocks []*BlockStack
	vars   []interface{}
	extend *map[string]interface{}
	vm     *VM
	//	vars  *map[string]interface{}
}

func (rt *RunTime) CallFunc(cmd uint16, obj *ObjInfo) (err error) {
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
	if obj.Type == OBJ_FUNC {
		_, err = rt.RunCode(obj.Value.(*Block))
	} else {
		finfo := obj.Value.(ExtFuncInfo)
		foo := reflect.ValueOf(finfo.Func)
		var result []reflect.Value
		pars := make([]reflect.Value, in)
		limit := 0
		//	fmt.Println(`CALL`, count, i, in, limit)
		auto := 0
		for k := 0; k < in; k++ {
			if len(finfo.Auto[k]) > 0 {
				auto++
			}
		}
		shift := size - count + auto
		if finfo.Variadic {
			shift = size - count
			count += auto
			limit = count - in + 1
		}
		i := count
		for ; i > limit; i-- {
			if len(finfo.Auto[count-i]) > 0 {
				pars[count-i] = reflect.ValueOf((*rt.extend)[finfo.Auto[count-i]])
				auto--
			} else {
				pars[count-i] = reflect.ValueOf(rt.stack[size-i+auto])
			}
		}
		if i > 0 {
			pars[in-1] = reflect.ValueOf(rt.stack[size-i : size])
		}
		//		fmt.Println(`Pars`, shift, count, i, size, pars)
		if finfo.Variadic {
			result = foo.CallSlice(pars)
		} else {
			result = foo.Call(pars)
		}
		rt.stack = rt.stack[:shift]
		//	fmt.Println(`Result`, result)
		for _, iret := range result {
			rt.stack = append(rt.stack, iret.Interface())
		}
	}
	/*	if
		if result[len(result)-1].Interface() != nil {
			return result[len(result)-1].Interface().(error)
		}
	*/
	return
}

func (rt *RunTime) extendFunc(name string) error {
	var (
		ok bool
		f  interface{}
	)
	if f, ok = (*rt.extend)[name]; !ok || reflect.ValueOf(f).Kind().String() != `func` {
		return fmt.Errorf(`unknown function %s`, name)
	}
	size := len(rt.stack)
	foo := reflect.ValueOf(f)
	//	if count != foo.Type().NumIn() {
	//	return fmt.Errorf(`The number of params %s is wrong`, name)
	//
	count := foo.Type().NumIn()
	pars := make([]reflect.Value, count)
	for i := count; i > 0; i-- {
		pars[count-i] = reflect.ValueOf(rt.stack[size-i])
	}
	result := foo.Call(pars)
	/*	if result[len(result)-1].Interface() != nil {
		return result[len(result)-1].Interface().(error)
	}*/
	rt.stack = rt.stack[:size-count]
	for _, iret := range result {
		rt.stack = append(rt.stack, iret.Interface())
	}
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

func (rt *RunTime) RunCode(block *Block) (status int, err error) {

	top := make([]interface{}, 8)
	start := len(rt.stack)
	rt.blocks = append(rt.blocks, &BlockStack{block, len(rt.vars)})
	for vkey, vpar := range block.Vars {
		var value interface{}
		if block.Type == OBJ_FUNC && vkey < len(block.Info.(*FuncInfo).Params) {
			value = rt.stack[start-len(block.Info.(*FuncInfo).Params)+vkey]
		} else {
			var vtype reflect.Type
			switch vpar {
			case reflect.Int64:
				vtype = reflect.TypeOf(int64(0))
			case reflect.String:
				vtype = reflect.TypeOf(``)
			case reflect.Bool:
				vtype = reflect.TypeOf(true)
			}
			value = reflect.New(vtype).Elem().Interface()
		}
		rt.vars = append(rt.vars, value)
	}
	if block.Type == OBJ_FUNC {
		start -= len(block.Info.(*FuncInfo).Params)
	}
	var assign []*VarInfo
	//main:
	for _, cmd := range block.Code {
		var bin interface{}
		size := len(rt.stack)
		if size < int(cmd.Cmd>>8) {
			return 0, fmt.Errorf(`stack is empty`)
		}
		for i := 1; i <= int(cmd.Cmd>>8); i++ {
			top[i-1] = rt.stack[size-i]
		}
		switch cmd.Cmd {
		case CMD_PUSH:
			rt.stack = append(rt.stack, cmd.Value)
		case CMD_PUSHSTR:
			rt.stack = append(rt.stack, cmd.Value.(string))
		case CMD_IF:
			if ValueToBool(rt.stack[len(rt.stack)-1]) {
				status, err = rt.RunCode(cmd.Value.(*Block))
			}
		case CMD_ELSE:
			if !ValueToBool(rt.stack[len(rt.stack)-1]) {
				status, err = rt.RunCode(cmd.Value.(*Block))
			}
		case CMD_ASSIGNVAR:
			assign = cmd.Value.([]*VarInfo)
		case CMD_ASSIGN:
			count := len(assign)
			for ivar, item := range assign {
				if item.Owner == nil {
					if (*item).Obj.Type == OBJ_EXTEND {
						(*rt.extend)[(*item).Obj.Value.(string)] = rt.stack[len(rt.stack)-count+ivar]
					}
				} else {
					var i int
					//				fmt.Println(`Var`, ivar, item.Obj.Value.(int))
					for i = len(rt.blocks) - 1; i >= 0; i-- {
						if item.Owner == rt.blocks[i].Block {
							rt.vars[rt.blocks[i].Offset+item.Obj.Value.(int)] = rt.stack[len(rt.stack)-count+ivar]
							break
						}
					}
				}
			}
			//			fmt.Println(`CMD ASSIGN`, count, rt.stack, rt.vars)
		case CMD_RETURN:
			status = STATUS_RETURN
			/*			for count := cmd.Value.(int); count > 0; count-- {
						rt.stack[start] = rt.stack[len(rt.stack)-count]
						start++
					}*/
		case CMD_CALLVARI, CMD_CALL:
			err = rt.CallFunc(cmd.Cmd, cmd.Value.(*ObjInfo))

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
			ivar := cmd.Value.(*VarInfo)
			var i int
			for i = len(rt.blocks) - 1; i >= 0; i-- {
				if ivar.Owner == rt.blocks[i].Block {
					rt.stack = append(rt.stack, rt.vars[rt.blocks[i].Offset+ivar.Obj.Value.(int)])
					break
				}
			}
			if i < 0 {
				return 0, fmt.Errorf(`wrong var`)
			}
			//			fmt.Println(`VAR`, voff, *ivar.Obj, ivar.Owner.Vars, rt.vars)
			//rt.stack = append(rt.stack, rt.vars[voff+ivar.Obj.Value.(int)])
		case CMD_EXTEND, CMD_CALLEXTEND:
			if val, ok := (*rt.extend)[cmd.Value.(string)]; ok {
				if cmd.Cmd == CMD_CALLEXTEND {
					err := rt.extendFunc(cmd.Value.(string))
					if err != nil {
						return 0, fmt.Errorf(`extend function %s %s`, cmd.Value.(string), err.Error())
					}
				} else {
					switch varVal := val.(type) {
					case int:
						val = int64(varVal)
					}
					rt.stack = append(rt.stack, val)
				}
			} else {
				return 0, fmt.Errorf(`unknown extend identifier %s`, cmd.Value.(string))
			}
		case CMD_NOT:
			rt.stack[size-1] = !ValueToBool(top[0])

		case CMD_ADD:
			/*			fmt.Println(`Stack`)
						for _, item := range rt.stack {
							fmt.Printf("|%v|", item)
						}
						fmt.Println(`Stack`, rt.stack)*/
			switch top[1].(type) {
			case string:
				bin = top[1].(string) + top[0].(string)
			default:
				bin = top[1].(int64) + top[0].(int64)
			}
		case CMD_SUB:
			bin = top[1].(int64) - top[0].(int64)
		case CMD_MUL:
			bin = top[1].(int64) * top[0].(int64)
		case CMD_DIV:
			if top[0].(int64) == 0 {
				return 0, fmt.Errorf(`divided by zero`)
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
			return 0, fmt.Errorf(`Unknown command %d`, cmd.Cmd)
		}
		if err != nil {
			return 0, err
		}
		if status == STATUS_RETURN {
			break
		}
		if (cmd.Cmd >> 8) == 2 {
			rt.stack[size-2] = bin
			rt.stack = rt.stack[:size-1]
		}
	}
	if status == STATUS_RETURN {
		//		fmt.Println(`Status`, rt.stack)
		if rt.blocks[len(rt.blocks)-1].Block.Type == OBJ_FUNC {
			for count := len(rt.blocks[len(rt.blocks)-1].Block.Info.(*FuncInfo).Results); count > 0; count-- {
				rt.stack[start] = rt.stack[len(rt.stack)-count]
				start++
			}
			status = STATUS_NORMAL
			rt.blocks = rt.blocks[:len(rt.blocks)-1]

			//fmt.Println(`Ret function`, rt.stack)
		} else {
			rt.blocks = rt.blocks[:len(rt.blocks)-1]
			return
		}
	}
	rt.stack = rt.stack[:start]
	return
}

func (rt *RunTime) Run(block *Block, params []interface{}, extend *map[string]interface{}) (ret []interface{}, err error) {
	info := block.Info.(*FuncInfo)
	rt.extend = extend
	if _, err = rt.RunCode(block); err == nil {
		off := len(rt.stack) - len(info.Results)
		//		fmt.Println(`RUN`, len(rt.stack), len(info.Results))
		for i := 0; i < len(info.Results); i++ {
			ret = append(ret, rt.stack[off+i])
		}
	}
	return
}

/*
func EvalIf(input string, vars *map[string]interface{}) (bool, error) {
	ret := Eval(input, vars)
	if err, ok := ret.(error); ok {
		return false, err
	}
	return ValueToBool(ret), nil
}*/
