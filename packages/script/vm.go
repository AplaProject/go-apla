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
	"encoding/json"
	"fmt"
	"reflect"
	"runtime/debug"
	"strconv"
	"strings"

	"github.com/AplaProject/go-apla/packages/consts"

	"github.com/shopspring/decimal"
	log "github.com/sirupsen/logrus"
)

const (
	statusNormal = iota
	statusReturn
	statusContinue
	statusBreak

	// Decimal is the constant string for decimal type
	Decimal = `decimal.Decimal`
	// Interface is the constant string for interface type
	Interface = `interface`

	brackets = `[]`
)

type VMError struct {
	Type  string `json:"type"`
	Error string `json:"error"`
}

type blockStack struct {
	Block  *Block
	Offset int
}

// RunTime is needed for the execution of the byte-code
type RunTime struct {
	stack  []interface{}
	blocks []*blockStack
	vars   []interface{}
	extend *map[string]interface{}
	vm     *VM
	cost   int64
	err    error
}

func (rt *RunTime) callFunc(cmd uint16, obj *ObjInfo) (err error) {
	var (
		count, in int
	)
	size := len(rt.stack)
	in = rt.vm.getInParams(obj)
	if cmd == cmdCallVari {
		count = rt.stack[size-1].(int)
		size--
	} else {
		count = in
	}
	if obj.Type == ObjFunc {
		var imap map[string][]interface{}
		if obj.Value.(*Block).Info.(*FuncInfo).Names != nil {
			if rt.stack[size-1] != nil {
				imap = rt.stack[size-1].(map[string][]interface{})
			}
			rt.stack = rt.stack[:size-1]
		}
		if cmd == cmdCallVari {
			parcount := count + 1 - in
			if parcount < 0 {
				log.WithFields(log.Fields{"type": consts.VMError}).Error("wrong count of parameters")
				return fmt.Errorf(`wrong count of parameters`)
			}
			pars := make([]interface{}, parcount)
			shift := size - parcount
			for i := parcount; i > 0; i-- {
				pars[i-1] = rt.stack[size+i-parcount-1]
			}
			rt.stack = rt.stack[:shift]
			rt.stack = append(rt.stack, pars)
		}
		if obj.Value.(*Block).Info.(*FuncInfo).Names != nil {
			rt.stack = append(rt.stack, imap)
		}
		_, err = rt.RunCode(obj.Value.(*Block))
	} else {
		finfo := obj.Value.(ExtFuncInfo)
		foo := reflect.ValueOf(finfo.Func)
		var result []reflect.Value
		pars := make([]reflect.Value, in)
		limit := 0
		(*rt.extend)[`rt`] = rt
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
			if !pars[count-i].IsValid() {
				pars[count-i] = reflect.Zero(reflect.TypeOf(map[string]interface{}{}))
			}
		}
		if i > 0 {
			pars[in-1] = reflect.ValueOf(rt.stack[size-i : size])
		}
		if finfo.Name == `ExecContract` && (pars[2].Type().String() != `string` || !pars[3].IsValid()) {
			return fmt.Errorf(`unknown function %v`, pars[1])
		}
		if finfo.Variadic {
			result = foo.CallSlice(pars)
		} else {
			result = foo.Call(pars)
		}
		rt.stack = rt.stack[:shift]

		for i, iret := range result {
			// first return value of every extend function that makes queries to DB is cost
			if i == 0 && rt.vm.FuncCallsDB != nil {
				if _, ok := rt.vm.FuncCallsDB[finfo.Name]; ok {
					cost := iret.Int()
					if cost > rt.cost {
						rt.cost = 0
						rt.vm.logger.WithFields(log.Fields{"type": consts.VMError}).Error("paid CPU resource is over")
						return fmt.Errorf("paid CPU resource is over")
					}

					rt.cost -= cost
					continue
				}
			}
			if finfo.Results[i].String() == `error` {
				if iret.Interface() != nil {
					return iret.Interface().(error)
				}
			} else {
				rt.stack = append(rt.stack, iret.Interface())
			}
		}
	}
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

	count := foo.Type().NumIn()
	pars := make([]reflect.Value, count)
	for i := count; i > 0; i-- {
		pars[count-i] = reflect.ValueOf(rt.stack[size-i])
	}
	result := foo.Call(pars)

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
	case string:
		return len(val) > 0
	case []uint8:
		return len(val) > 0
	case []interface{}:
		return val != nil && len(val) > 0
	default:
		dec, _ := decimal.NewFromString(fmt.Sprintf(`%v`, val))
		return dec.Cmp(decimal.New(0, 0)) != 0
	}
	return false
}

// ValueToInt converts interface (string or int64) to int64
func ValueToInt(v interface{}) (ret int64) {
	var err error
	switch val := v.(type) {
	case int64:
		ret = val
	case string:
		ret, err = strconv.ParseInt(val, 10, 64)
		if err != nil {
			log.WithFields(log.Fields{"type": consts.ConversionError, "error": err, "value": val}).Error("converting value from string to int")
		}
	}
	return
}

// ValueToFloat converts interface (string, float64 or int64) to float64
func ValueToFloat(v interface{}) (ret float64) {
	var err error
	switch val := v.(type) {
	case float64:
		ret = val
	case int64:
		ret = float64(val)
	case string:
		ret, err = strconv.ParseFloat(val, 64)
		if err != nil {
			log.WithFields(log.Fields{"type": consts.ConversionError, "error": err, "value": val}).Error("converting value from string to float")
		}
	}
	return
}

// ValueToDecimal converts interface (string, float64, Decimal or int64) to Decimal
func ValueToDecimal(v interface{}) (ret decimal.Decimal) {
	var err error
	switch val := v.(type) {
	case float64:
		ret = decimal.NewFromFloat(val)
	case string:
		ret, err = decimal.NewFromString(val)
		if err != nil {
			log.WithFields(log.Fields{"type": consts.ConversionError, "error": err, "value": val}).Error("converting value from string to decimal")
		}
	case int64:
		ret = decimal.New(val, 0)
	default:
		ret = val.(decimal.Decimal)
	}
	return
}

// SetCost sets the max cost of the execution.
func (rt *RunTime) SetCost(cost int64) {
	rt.cost = cost
}

// Cost return the remain cost of the execution.
func (rt *RunTime) Cost() int64 {
	return rt.cost
}

// RunInit creates a new RunTime for the virtual machine
func (vm *VM) RunInit(cost int64) *RunTime {
	rt := RunTime{stack: make([]interface{}, 0, 1024), vm: vm, cost: cost}
	return &rt
}

func SetVMError(eType string, eText interface{}) error {
	out, err := json.Marshal(&VMError{Type: eType, Error: fmt.Sprintf(`%v`, eText)})
	if err != nil {
		log.WithFields(log.Fields{"type": consts.JSONMarshallError, "error": err}).Error("marshalling VMError")
		out = []byte(`{"type": "panic", "error": "marshalling VMError"}`)
	}
	return fmt.Errorf(string(out))
}

// RunCode executes Block
func (rt *RunTime) RunCode(block *Block) (status int, err error) {
	top := make([]interface{}, 8)
	rt.blocks = append(rt.blocks, &blockStack{block, len(rt.vars)})
	var namemap map[string][]interface{}
	if block.Type == ObjFunc && block.Info.(*FuncInfo).Names != nil {
		if rt.stack[len(rt.stack)-1] != nil {
			namemap = rt.stack[len(rt.stack)-1].(map[string][]interface{})
		}
		rt.stack = rt.stack[:len(rt.stack)-1]
	}
	start := len(rt.stack)
	varoff := len(rt.vars)
	for vkey, vpar := range block.Vars {
		rt.cost--
		var value interface{}
		if block.Type == ObjFunc && vkey < len(block.Info.(*FuncInfo).Params) {
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
	if namemap != nil {
		for key, item := range namemap {
			params := (*block.Info.(*FuncInfo).Names)[key]
			if params.Variadic {

			}
			for i, value := range item {
				if params.Variadic && i >= len(params.Params)-1 {
					off := varoff + params.Offset[len(params.Params)-1]
					rt.vars[off] = append(rt.vars[off].([]interface{}), value)
				} else {
					rt.vars[varoff+params.Offset[i]] = value
				}
			}
		}
	}
	if block.Type == ObjFunc {
		start -= len(block.Info.(*FuncInfo).Params)
	}
	var assign []*VarInfo
	labels := make([]int, 0)
	for ci := 0; ci < len(block.Code); ci++ {
		rt.cost--
		if rt.cost <= 0 {
			rt.vm.logger.WithFields(log.Fields{"type": consts.VMError}).Warn("paid CPU resource is over")
			return 0, fmt.Errorf(`paid CPU resource is over`)
		}
		cmd := block.Code[ci]
		var bin interface{}
		size := len(rt.stack)
		if size < int(cmd.Cmd>>8) {
			rt.vm.logger.WithFields(log.Fields{"type": consts.VMError}).Error("stack is empty")
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
			val := rt.stack[len(rt.stack)-1]
			rt.stack = rt.stack[:len(rt.stack)-1]
			if valueToBool(val) {
				status, err = rt.RunCode(cmd.Value.(*Block))
				newci := labels[len(labels)-1]
				labels = labels[:len(labels)-1]
				if status == statusContinue {
					ci = newci - 1
					status = statusNormal
					continue
				}
				if status == statusBreak {
					status = statusNormal
					break
				}
			}
		case cmdLabel:
			labels = append(labels, ci)
		case cmdContinue:
			status = statusContinue
		case cmdBreak:
			status = statusBreak
		case cmdAssignVar:
			assign = cmd.Value.([]*VarInfo)
		case cmdAssign:
			count := len(assign)
			for ivar, item := range assign {
				if item.Owner == nil {
					if (*item).Obj.Type == ObjExtend {
						(*rt.extend)[(*item).Obj.Value.(string)] = rt.stack[len(rt.stack)-count+ivar]
					}
				} else {
					var i int
					for i = len(rt.blocks) - 1; i >= 0; i-- {
						if item.Owner == rt.blocks[i].Block {
							switch rt.blocks[i].Block.Vars[item.Obj.Value.(int)].String() {
							case Decimal:
								rt.vars[rt.blocks[i].Offset+item.Obj.Value.(int)] = ValueToDecimal(rt.stack[len(rt.stack)-count+ivar])
							default:
								rt.vars[rt.blocks[i].Offset+item.Obj.Value.(int)] = rt.stack[len(rt.stack)-count+ivar]
							}
							break
						}
					}
				}
			}
		case cmdReturn:
			status = statusReturn
		case cmdError:
			eType := msgError
			if cmd.Value.(uint32) == keyWarning {
				eType = msgWarning
			} else if cmd.Value.(uint32) == keyInfo {
				eType = msgInfo
			}
			err = SetVMError(eType, rt.stack[len(rt.stack)-1])
		case cmdFuncName:
			ifunc := cmd.Value.(FuncNameCmd)
			mapoff := len(rt.stack) - 1 - ifunc.Count
			if rt.stack[mapoff] == nil {
				rt.stack[mapoff] = make(map[string][]interface{})
			}
			params := make([]interface{}, ifunc.Count)
			for i := 0; i < ifunc.Count; i++ {
				params[i] = rt.stack[mapoff+1+i]
			}
			rt.stack[mapoff].(map[string][]interface{})[ifunc.Name] = params
			rt.stack = rt.stack[:mapoff+1]
			continue
		case cmdCallVari, cmdCall:
			if cmd.Value.(*ObjInfo).Type == ObjExtFunc {
				finfo := cmd.Value.(*ObjInfo).Value.(ExtFuncInfo)
				if rt.vm.ExtCost != nil {
					cost := rt.vm.ExtCost(finfo.Name)
					if cost > rt.cost {
						rt.cost = 0
						rt.vm.logger.WithFields(log.Fields{"type": consts.VMError}).Warning("paid CPU resource is over")
						return 0, fmt.Errorf(`paid CPU resource is over`)
					} else if cost == -1 {
						rt.cost -= CostCall
					} else {
						rt.cost -= cost
					}
				}
			} else {
				rt.cost -= CostCall
			}
			err = rt.callFunc(cmd.Cmd, cmd.Value.(*ObjInfo))

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
				rt.vm.logger.WithFields(log.Fields{"type": consts.VMError, "var": ivar.Obj.Value}).Error("wrong var")
				return 0, fmt.Errorf(`wrong var %v`, ivar.Obj.Value)
			}
		case cmdExtend, cmdCallExtend:
			if val, ok := (*rt.extend)[cmd.Value.(string)]; ok {
				rt.cost -= CostExtend
				if cmd.Cmd == cmdCallExtend {
					err = rt.extendFunc(cmd.Value.(string))
					if err != nil {
						rt.vm.logger.WithFields(log.Fields{"type": consts.VMError, "error": err, "cmd": cmd.Value.(string)}).Error("executing extended function")
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
				rt.vm.logger.WithFields(log.Fields{"type": consts.VMError, "cmd": cmd.Value.(string)}).Error("unknown extend identifier")
				err = fmt.Errorf(`unknown extend identifier %s`, cmd.Value.(string))
			}
		case cmdIndex:
			itype := reflect.TypeOf(rt.stack[size-2]).String()
			switch {
			case itype[:3] == `map`:
				if strings.Contains(itype, Interface) {
					rt.stack[size-2] = rt.stack[size-2].(map[string]interface{})[rt.stack[size-1].(string)]
				} else {
					rt.stack[size-2] = rt.stack[size-2].(map[string]string)[rt.stack[size-1].(string)]
				}
				rt.stack = rt.stack[:size-1]
			case itype[:2] == brackets:
				if strings.Contains(itype, Interface) {
					rt.stack[size-2] = rt.stack[size-2].([]interface{})[rt.stack[size-1].(int64)]
				} else {
					rt.stack[size-2] = rt.stack[size-2].([]map[string]string)[rt.stack[size-1].(int64)]
				}
				rt.stack = rt.stack[:size-1]
			default:
				rt.vm.logger.WithFields(log.Fields{"type": consts.VMError, "vm_type": itype}).Error("type does not support indexing")
				err = fmt.Errorf(`Type %s doesn't support indexing`, itype)
			}
		case cmdSetIndex:
			itype := reflect.TypeOf(rt.stack[size-3]).String()
			switch {
			case itype[:3] == `map`:
				if strings.Contains(itype, Interface) {
					rt.stack[size-3].(map[string]interface{})[rt.stack[size-2].(string)] = rt.stack[size-1]
				} else {
					rt.stack[size-3].(map[string]string)[rt.stack[size-2].(string)] = rt.stack[size-1].(string)
				}
				rt.stack = rt.stack[:size-2]
			case itype[:2] == brackets:
				ind := rt.stack[size-2].(int64)
				if strings.Contains(itype, Interface) {
					slice := rt.stack[size-3].([]interface{})
					if int(ind) >= len(slice) {
						slice = append(slice, make([]interface{}, int(ind)-len(slice)+1)...)
						for i := 0; i < len(rt.vars); i++ {
							if reflect.TypeOf(rt.vars[i]).String()[:2] == brackets {
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
				rt.vm.logger.WithFields(log.Fields{"type": consts.VMError, "vm_type": itype}).Error("type does not support indexing")
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
			switch top[1].(type) {
			case string:
				switch top[0].(type) {
				case int64:
					bin = ValueToInt(top[1]) + top[0].(int64)
				case float64:
					bin = ValueToFloat(top[1]) + top[0].(float64)
				default:
					if reflect.TypeOf(top[0]).String() == Decimal {
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
				case Decimal:
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
					if reflect.TypeOf(top[0]).String() == Decimal {
						bin = ValueToDecimal(top[1]).Sub(top[0].(decimal.Decimal))
					}
				}
			case float64:
				bin = top[1].(float64) - ValueToFloat(top[0])
			case int64:
				bin = top[1].(int64) - top[0].(int64)
			default:
				switch reflect.TypeOf(top[1]).String() {
				case Decimal:
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
					if reflect.TypeOf(top[0]).String() == Decimal {
						bin = ValueToDecimal(top[1]).Mul(top[0].(decimal.Decimal))
					}
				}
			case float64:
				if reflect.TypeOf(top[0]).String() == Decimal {
					bin = ValueToDecimal(top[1]).Mul(top[0].(decimal.Decimal))
				} else {
					bin = top[1].(float64) * ValueToFloat(top[0])
				}
			case int64:
				if reflect.TypeOf(top[0]).String() == Decimal {
					bin = ValueToDecimal(top[1]).Mul(top[0].(decimal.Decimal))
				} else {
					bin = top[1].(int64) * top[0].(int64)
				}
			default:
				switch reflect.TypeOf(top[1]).String() {
				case Decimal:
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
					if reflect.TypeOf(top[0]).String() == Decimal {
						bin = ValueToDecimal(top[1]).Div(top[0].(decimal.Decimal))
					}
				}
			case float64:
				bin = top[1].(float64) / ValueToFloat(top[0])
			case int64:
				if top[0].(int64) == 0 {
					log.WithFields(log.Fields{"type": consts.DivisionByZero}).Error("divided by zero")
					return 0, fmt.Errorf(`divided by zero`)
				}
				bin = top[1].(int64) / top[0].(int64)
			default:
				switch reflect.TypeOf(top[1]).String() {
				case Decimal:
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
					if reflect.TypeOf(top[0]).String() == Decimal {
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
					if reflect.TypeOf(top[0]).String() == Decimal {
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
					if reflect.TypeOf(top[0]).String() == Decimal {
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
			rt.vm.logger.WithFields(log.Fields{"type": consts.VMError, "vm_cmd": cmd.Cmd}).Error("Unknown command")
			err = fmt.Errorf(`Unknown command %d`, cmd.Cmd)
		}
		if err != nil {
			rt.err = err
			break
		}
		if status == statusReturn || status == statusContinue || status == statusBreak {
			break
		}
		if (cmd.Cmd >> 8) == 2 {
			rt.stack[size-2] = bin
			rt.stack = rt.stack[:size-1]
		}
	}
	last := rt.blocks[len(rt.blocks)-1]
	rt.blocks = rt.blocks[:len(rt.blocks)-1]
	if status == statusReturn {
		if last.Block.Type == ObjFunc {
			for count := len(last.Block.Info.(*FuncInfo).Results); count > 0; count-- {
				rt.stack[start] = rt.stack[len(rt.stack)-count]
				start++
			}
			status = statusNormal
		} else {
			return
		}
	}
	rt.stack = rt.stack[:start]
	if err != nil {
		rt.vm.logger.WithFields(log.Fields{"type": consts.VMError, "error": err}).Error("error in vm")
	}
	return
}

// Run executes Block with the specified parameters and extended variables and functions
func (rt *RunTime) Run(block *Block, params []interface{}, extend *map[string]interface{}) (ret []interface{}, err error) {
	defer func() {
		if r := recover(); r != nil {
			rt.vm.logger.WithFields(log.Fields{"type": consts.PanicRecoveredError, "stack": string(debug.Stack())}).Error("runtime panic error")
			err = fmt.Errorf(`runtime panic error`)
		}
	}()
	info := block.Info.(*FuncInfo)
	rt.extend = extend
	if _, err = rt.RunCode(block); err == nil {
		off := len(rt.stack) - len(info.Results)
		for i := 0; i < len(info.Results); i++ {
			ret = append(ret, rt.stack[off+i])
		}
	}
	return
}
