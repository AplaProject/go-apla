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
	"strings"

	"github.com/shopspring/decimal"
)

/*type ValStack struct {
	Value interface{}
}*/

const (
	STATUS_NORMAL = iota
	STATUS_RETURN
	STATUS_CONTINUE
	STATUS_BREAK

//	STATUS_ERROR
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
	cost   int64
	err    error
	//	vars  *map[string]interface{}
}

func (rt *RunTime) CallFunc(cmd uint16, obj *ObjInfo) (err error) {
	var (
		count, in int
		//		f         interface{}
	)
	size := len(rt.stack)
	in = rt.vm.getInParams(obj)
	if cmd == cmdCallVari {
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
		(*rt.extend)[`rt`] = rt
		//		fmt.Println(`CALL`, finfo, count, in)
		auto := 0
		for k := 0; k < in; k++ {
			if len(finfo.Auto[k]) > 0 {
				auto++
			}
		}
		//		fmt.Println(`Extend`, auto, *rt.extend, finfo.Auto)
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
			if !pars[count-i].IsValid() {
				pars[count-i] = reflect.Zero(reflect.TypeOf(map[string]interface{}{}))
				//reflect.ValueOf((*interface{})(nil)) //reflect.Zero(reflect.TypeOf(&MyType{}))
			}
		}
		if i > 0 {
			pars[in-1] = reflect.ValueOf(rt.stack[size-i : size])
		}
		//fmt.Println(`Pars`, shift, count, limit, i, size, pars)
		if finfo.Variadic {
			result = foo.CallSlice(pars)
		} else {
			result = foo.Call(pars)
		}
		rt.stack = rt.stack[:shift]

		//	fmt.Println(`Result`, result)
		for i, iret := range result {
			if finfo.Results[i].String() == `error` {
				if iret.Interface() != nil {
					return iret.Interface().(error)
				}
			} else {
				rt.stack = append(rt.stack, iret.Interface())
			}
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
	for i, iret := range result {
		if foo.Type().Out(i).String() == `error` {
			if iret.Interface() != nil {
				return iret.Interface().(error)
			}
		} else {
			rt.stack = append(rt.stack, iret.Interface())
		}
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
	case float64:
		if val != 0.0 {
			return true
		}
	case bool:
		return val
	default:
		return val.(decimal.Decimal).Cmp(decimal.New(0, 0)) != 0
	}
	return false
}

func ValueToInt(v interface{}) (ret int64) {
	switch val := v.(type) {
	case int64:
		ret = val
	case string:
		ret, _ = strconv.ParseInt(val, 10, 64)
		/*	default:
			ret = val.(decimal.Decimal)*/
	}
	return
}

func ValueToFloat(v interface{}) (ret float64) {
	switch val := v.(type) {
	case float64:
		ret = val
	case int64:
		ret = float64(val)
	case string:
		ret, _ = strconv.ParseFloat(val, 64)
		/*	default:
			ret = val.(decimal.Decimal)*/
	}
	return
}

func ValueToDecimal(v interface{}) (ret decimal.Decimal) {
	switch val := v.(type) {
	case float64:
		ret = decimal.NewFromFloat(val)
	case string:
		ret, _ = decimal.NewFromString(val)
	case int64:
		ret = decimal.New(val, 0)
	default:
		ret = val.(decimal.Decimal)
	}
	return
}

func (rt *RunTime) SetCost(cost int64) {
	rt.cost = cost
}

func (rt *RunTime) Cost() int64 {
	return rt.cost
}

func (vm *VM) RunInit(cost int64) *RunTime {
	rt := RunTime{stack: make([]interface{}, 0, 1024), vm: vm, cost: cost}
	return &rt
}

func (rt *RunTime) RunCode(block *Block) (status int, err error) {
	top := make([]interface{}, 8)
	start := len(rt.stack)
	rt.blocks = append(rt.blocks, &BlockStack{block, len(rt.vars)})
	for vkey, vpar := range block.Vars {
		rt.cost--
		var value interface{}
		if block.Type == OBJ_FUNC && vkey < len(block.Info.(*FuncInfo).Params) {
			value = rt.stack[start-len(block.Info.(*FuncInfo).Params)+vkey]
		} else {
			value = reflect.New(vpar).Elem().Interface()
			if vpar == reflect.TypeOf(map[string]interface{}{}) {
				value = make(map[string]interface{})
			} else if vpar == reflect.TypeOf([]interface{}{}) {
				value = make([]interface{}, 0, len(rt.vars)+1)
			}
		}
		rt.vars = append(rt.vars, value)
	}
	if block.Type == OBJ_FUNC {
		start -= len(block.Info.(*FuncInfo).Params)
	}
	var assign []*VarInfo
	labels := make([]int, 0)
	//main:
	for ci := 0; ci < len(block.Code); ci++ { //_, cmd := range block.Code {
		rt.cost--
		if rt.cost <= 0 {
			return 0, fmt.Errorf(`paid CPU resource is over`)
		}
		cmd := block.Code[ci]
		var bin interface{}
		size := len(rt.stack)
		if size < int(cmd.Cmd>>8) {
			return 0, fmt.Errorf(`stack is empty`)
		}
		for i := 1; i <= int(cmd.Cmd>>8); i++ {
			top[i-1] = rt.stack[size-i]
		}
		switch cmd.Cmd {
		case cmdPush:
			rt.stack = append(rt.stack, cmd.Value)
		case cmdPushStr:
			rt.stack = append(rt.stack, cmd.Value.(string))
		case cmdIf:
			if valueToBool(rt.stack[len(rt.stack)-1]) {
				status, err = rt.RunCode(cmd.Value.(*Block))
			}
		case cmdElse:
			if !valueToBool(rt.stack[len(rt.stack)-1]) {
				status, err = rt.RunCode(cmd.Value.(*Block))
			}
		case cmdWhile:
			if valueToBool(rt.stack[len(rt.stack)-1]) {
				status, err = rt.RunCode(cmd.Value.(*Block))
				newci := labels[len(labels)-1]
				labels = labels[:len(labels)-1]
				if status == STATUS_CONTINUE {
					ci = newci - 1
					status = STATUS_NORMAL
					continue
				}
				if status == STATUS_BREAK {
					status = STATUS_NORMAL
					break
				}
			}
		case cmdLabel:
			labels = append(labels, ci)
		case cmdContinue:
			status = STATUS_CONTINUE
		case cmdBreak:
			status = STATUS_BREAK
		case cmdAssignVar:
			assign = cmd.Value.([]*VarInfo)
		case cmdAssign:
			count := len(assign)
			for ivar, item := range assign {
				if item.Owner == nil {
					if (*item).Obj.Type == OBJ_EXTEND {
						(*rt.extend)[(*item).Obj.Value.(string)] = rt.stack[len(rt.stack)-count+ivar]
					}
				} else {
					var i int
					for i = len(rt.blocks) - 1; i >= 0; i-- {
						if item.Owner == rt.blocks[i].Block {
							//							fmt.Println(`Var`, item.Obj.Type, item.Obj.Value, rt.blocks[i].Block.Vars[item.Obj.Value.(int)])

							switch rt.blocks[i].Block.Vars[item.Obj.Value.(int)].String() {
							case `decimal.Decimal`:
								rt.vars[rt.blocks[i].Offset+item.Obj.Value.(int)] = ValueToDecimal(rt.stack[len(rt.stack)-count+ivar])
							default:
								rt.vars[rt.blocks[i].Offset+item.Obj.Value.(int)] = rt.stack[len(rt.stack)-count+ivar]
							}
							break
						}
					}
				}
			}
			//			fmt.Println(`CMD ASSIGN`, count, rt.stack, rt.vars)
		case cmdReturn:
			status = STATUS_RETURN
		case cmdError:
			pattern := `%v`
			if cmd.Value.(uint32) == KEY_WARNING {
				pattern = `!%v`
			} else if cmd.Value.(uint32) == KEY_INFO {
				pattern = `*%v`
			}
			err = fmt.Errorf(pattern, rt.stack[len(rt.stack)-1])
		case cmdCallVari, cmdCall:
			if cmd.Value.(*ObjInfo).Type == OBJ_EXTFUNC {
				finfo := cmd.Value.(*ObjInfo).Value.(ExtFuncInfo)
				if rt.vm.ExtCost != nil {
					cost := rt.vm.ExtCost(finfo.Name)
					if cost > rt.cost {
						rt.cost = 0
						return 0, fmt.Errorf(`paid CPU resource is over`)
					} else if cost == -1 {
						rt.cost -= COST_CALL
					} else {
						rt.cost -= cost
					}
				}
			} else {
				rt.cost -= COST_CALL
			}
			err = rt.CallFunc(cmd.Cmd, cmd.Value.(*ObjInfo))

		case cmdVar:
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
		case cmdExtend, cmdCallExtend:
			if val, ok := (*rt.extend)[cmd.Value.(string)]; ok {
				rt.cost -= COST_EXTEND
				if cmd.Cmd == cmdCallExtend {
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
				err = fmt.Errorf(`unknown extend identifier %s`, cmd.Value.(string))
			}
		case cmdIndex:
			itype := reflect.TypeOf(rt.stack[size-2]).String()
			switch {
			case itype[:3] == `map`:
				if strings.Index(itype, `interface`) >= 0 {
					rt.stack[size-2] = rt.stack[size-2].(map[string]interface{})[rt.stack[size-1].(string)]
				} else {
					rt.stack[size-2] = rt.stack[size-2].(map[string]string)[rt.stack[size-1].(string)]
				}
				rt.stack = rt.stack[:size-1]
			case itype[:2] == `[]`:
				if strings.Index(itype, `interface`) >= 0 {
					rt.stack[size-2] = rt.stack[size-2].([]interface{})[rt.stack[size-1].(int64)]
				} else {
					rt.stack[size-2] = rt.stack[size-2].([]map[string]string)[rt.stack[size-1].(int64)]
				}
				rt.stack = rt.stack[:size-1]
			default:
				err = fmt.Errorf(`Type %s doesn't support indexing`, itype)
			}
		case cmdSetIndex:
			itype := reflect.TypeOf(rt.stack[size-3]).String()
			switch {
			case itype[:3] == `map`:
				if strings.Index(itype, `interface`) >= 0 {
					rt.stack[size-3].(map[string]interface{})[rt.stack[size-2].(string)] = rt.stack[size-1]
				} else {
					rt.stack[size-3].(map[string]string)[rt.stack[size-2].(string)] = rt.stack[size-1].(string)
				}
				rt.stack = rt.stack[:size-2]
			case itype[:2] == `[]`:
				ind := rt.stack[size-2].(int64)
				if strings.Index(itype, `interface`) >= 0 {
					slice := rt.stack[size-3].([]interface{})
					if int(ind) >= len(slice) {
						slice = append(slice, make([]interface{}, int(ind)-len(slice)+1)...)
						for i := 0; i < len(rt.vars); i++ {
							if reflect.TypeOf(rt.vars[i]).String()[:2] == `[]` {
								if len(rt.stack[size-3].([]interface{})) == len(rt.vars[i].([]interface{})) &&
									((len(rt.vars[i].([]interface{})) > 0 &&
										&rt.stack[size-3].([]interface{})[0] == &rt.vars[i].([]interface{})[0]) ||
										len(rt.vars[i].([]interface{})) == 0 && cap(rt.vars[i].([]interface{})) == i+1) {
									rt.vars[i] = slice
									break
								}
							}
						}
						rt.stack[size-3] = slice
					}
					slice[ind] = rt.stack[size-1]
				} else {
					slice := rt.stack[size-3].([]map[string]string)
					slice[ind] = rt.stack[size-1].(map[string]string)
				}
				rt.stack = rt.stack[:size-2]
			default:
				err = fmt.Errorf(`Type %s doesn't support indexing`, itype)
			}
		case cmdSign:
			switch top[0].(type) {
			case float64:
				rt.stack[size-1] = -top[0].(float64)
			default:
				rt.stack[size-1] = -top[0].(int64)
			}
		case cmdNot:
			rt.stack[size-1] = !valueToBool(top[0])

		case cmdAdd:
			/*			fmt.Println(`Stack`)
						for _, item := range rt.stack {
							fmt.Printf("|%v|", item)
						}
						fmt.Println(`Stack`, rt.stack)*/
			switch top[1].(type) {
			case string:
				switch top[0].(type) {
				case int64:
					bin = ValueToInt(top[1]) + top[0].(int64)
				case float64:
					bin = ValueToFloat(top[1]) + top[0].(float64)
				default:
					if reflect.TypeOf(top[0]).String() == `decimal.Decimal` {
						bin = ValueToDecimal(top[1]).Add(top[0].(decimal.Decimal))
					} else {
						bin = top[1].(string) + top[0].(string)
					}
				}
			case float64:
				bin = top[1].(float64) + ValueToFloat(top[0])
			case int64:
				bin = top[1].(int64) + top[0].(int64)
			default:
				switch reflect.TypeOf(top[1]).String() {
				case `decimal.Decimal`:
					bin = top[1].(decimal.Decimal).Add(ValueToDecimal(top[0]))
				}
			}
		case cmdSub:
			switch top[1].(type) {
			case string:
				switch top[0].(type) {
				case int64:
					bin = ValueToInt(top[1]) - top[0].(int64)
				case float64:
					bin = ValueToFloat(top[1]) - top[0].(float64)
				default:
					if reflect.TypeOf(top[0]).String() == `decimal.Decimal` {
						bin = ValueToDecimal(top[1]).Sub(top[0].(decimal.Decimal))
					}
				}
			case float64:
				bin = top[1].(float64) - ValueToFloat(top[0])
			case int64:
				bin = top[1].(int64) - top[0].(int64)
			default:
				switch reflect.TypeOf(top[1]).String() {
				case `decimal.Decimal`:
					bin = top[1].(decimal.Decimal).Sub(ValueToDecimal(top[0]))
				}
			}
		case cmdMul:
			switch top[1].(type) {
			case string:
				switch top[0].(type) {
				case int64:
					bin = ValueToInt(top[1]) * top[0].(int64)
				case float64:
					bin = ValueToFloat(top[1]) * top[0].(float64)
				default:
					if reflect.TypeOf(top[0]).String() == `decimal.Decimal` {
						bin = ValueToDecimal(top[1]).Mul(top[0].(decimal.Decimal))
					}
				}
			case float64:
				if reflect.TypeOf(top[0]).String() == `decimal.Decimal` {
					bin = ValueToDecimal(top[1]).Mul(top[0].(decimal.Decimal))
				} else {
					bin = top[1].(float64) * ValueToFloat(top[0])
				}
			case int64:
				if reflect.TypeOf(top[0]).String() == `decimal.Decimal` {
					bin = ValueToDecimal(top[1]).Mul(top[0].(decimal.Decimal))
				} else {
					bin = top[1].(int64) * top[0].(int64)
				}
			default:
				switch reflect.TypeOf(top[1]).String() {
				case `decimal.Decimal`:
					bin = top[1].(decimal.Decimal).Mul(ValueToDecimal(top[0]))
				}
			}
		case cmdDiv:
			switch top[1].(type) {
			case string:
				switch top[0].(type) {
				case int64:
					bin = ValueToInt(top[1]) / top[0].(int64)
				case float64:
					bin = ValueToFloat(top[1]) / top[0].(float64)
				default:
					if reflect.TypeOf(top[0]).String() == `decimal.Decimal` {
						bin = ValueToDecimal(top[1]).Div(top[0].(decimal.Decimal))
					}
				}
			case float64:
				bin = top[1].(float64) / ValueToFloat(top[0])
			case int64:
				if top[0].(int64) == 0 {
					return 0, fmt.Errorf(`divided by zero`)
				}
				bin = top[1].(int64) / top[0].(int64)
			default:
				switch reflect.TypeOf(top[1]).String() {
				case `decimal.Decimal`:
					bin = top[1].(decimal.Decimal).Div(ValueToDecimal(top[0]))
				}
			}
		case cmdAnd:
			bin = valueToBool(top[1]) && valueToBool(top[0])
		case cmdOr:
			bin = valueToBool(top[1]) || valueToBool(top[0])
		case cmdEqual, cmdNotEq:
			switch top[1].(type) {
			case string:
				switch top[0].(type) {
				case int64:
					bin = ValueToInt(top[1]) == top[0].(int64)
				case float64:
					bin = ValueToFloat(top[1]) == top[0].(float64)
				default:
					if reflect.TypeOf(top[0]).String() == `decimal.Decimal` {
						bin = ValueToDecimal(top[1]).Cmp(top[0].(decimal.Decimal)) == 0
					} else {
						bin = top[1].(string) == top[0].(string)
					}
				}
			case float64:
				bin = top[1].(float64) == ValueToFloat(top[0])
			case int64:
				bin = top[1].(int64) == top[0].(int64)
			default:
				bin = top[1].(decimal.Decimal).Cmp(ValueToDecimal(top[0])) == 0
			}
			if cmd.Cmd == cmdNotEq {
				bin = !bin.(bool)
			}
		case cmdLess, cmdNotLess:
			switch top[1].(type) {
			case string:
				switch top[0].(type) {
				case int64:
					bin = ValueToInt(top[1]) < top[0].(int64)
				case float64:
					bin = ValueToFloat(top[1]) < top[0].(float64)
				default:
					if reflect.TypeOf(top[0]).String() == `decimal.Decimal` {
						bin = ValueToDecimal(top[1]).Cmp(top[0].(decimal.Decimal)) < 0
					} else {
						bin = top[1].(string) < top[0].(string)
					}
				}
			case float64:
				bin = top[1].(float64) < ValueToFloat(top[0])
			case int64:
				bin = top[1].(int64) < top[0].(int64)
			default:
				bin = top[1].(decimal.Decimal).Cmp(ValueToDecimal(top[0])) < 0
			}
			if cmd.Cmd == cmdNotLess {
				bin = !bin.(bool)
			}
		case cmdGreat, cmdNotGreat:
			switch top[1].(type) {
			case string:
				switch top[0].(type) {
				case int64:
					bin = ValueToInt(top[1]) > top[0].(int64)
				case float64:
					bin = ValueToFloat(top[1]) > top[0].(float64)
				default:
					if reflect.TypeOf(top[0]).String() == `decimal.Decimal` {
						bin = ValueToDecimal(top[1]).Cmp(top[0].(decimal.Decimal)) > 0
					} else {
						bin = top[1].(string) > top[0].(string)
					}
				}
			case float64:
				bin = top[1].(float64) > ValueToFloat(top[0])
			case int64:
				bin = top[1].(int64) > top[0].(int64)
			default:
				bin = top[1].(decimal.Decimal).Cmp(ValueToDecimal(top[0])) > 0
			}
			if cmd.Cmd == cmdNotGreat {
				bin = !bin.(bool)
			}
		default:
			err = fmt.Errorf(`Unknown command %d`, cmd.Cmd)
		}
		if err != nil {
			rt.err = err
			//			status = STATUS_ERROR
			break
		}
		if status == STATUS_RETURN || status == STATUS_CONTINUE || status == STATUS_BREAK {
			break
		}
		if (cmd.Cmd >> 8) == 2 {
			rt.stack[size-2] = bin
			rt.stack = rt.stack[:size-1]
		}
	}
	/*	if status == STATUS_BREAK {
		status = STATUS_NORMAL
	}*/
	if status == STATUS_RETURN {
		//		fmt.Println(`Status`, start, rt.stack)
		if rt.blocks[len(rt.blocks)-1].Block.Type == OBJ_FUNC {
			for count := len(rt.blocks[len(rt.blocks)-1].Block.Info.(*FuncInfo).Results); count > 0; count-- {
				rt.stack[start] = rt.stack[len(rt.stack)-count]
				start++
			}
			status = STATUS_NORMAL
			rt.blocks = rt.blocks[:len(rt.blocks)-1]

			//			fmt.Println(`Ret function`, start, rt.stack)
		} else {
			rt.blocks = rt.blocks[:len(rt.blocks)-1]
			return
		}
	}
	rt.stack = rt.stack[:start]
	return
}

func (rt *RunTime) Run(block *Block, params []interface{}, extend *map[string]interface{}) (ret []interface{}, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf(`runtime panic error`)
		}
	}()
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
