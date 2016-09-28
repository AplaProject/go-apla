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

package smart

import (
	"fmt"
	"reflect"

	"github.com/DayLightProject/go-daylight/packages/consts"
	"github.com/DayLightProject/go-daylight/packages/script"
)

type Contract struct {
	Name  string
	data  interface{}
	block *script.Block
}

var (
	smartVM *script.VM
)

func init() {
	smartVM = script.VMInit(map[string]interface{}{
		"Println": fmt.Println,
		"Sprintf": fmt.Sprintf,
	})
}

// Compiles contract source code
func Compile(src string) error {
	return smartVM.Compile([]rune(src))
}

// Returns true if the contract exists
func GetContract(name string, data interface{}) *Contract {
	obj, ok := smartVM.Objects[name]
	//	fmt.Println(`Get`, ok, obj, obj.Type, script.OBJ_CONTRACT)
	if ok && obj.Type == script.OBJ_CONTRACT {
		return &Contract{Name: name, data: data, block: obj.Value.(*script.Block)}
	}
	return nil
}

func (contract *Contract) getFunc(name string) *script.Block {
	if block, ok := (*contract).block.Objects[name]; ok && block.Type == script.OBJ_FUNC {
		return block.Value.(*script.Block)
	}
	return nil
}

func (contract *Contract) getExtend() map[string]interface{} {
	head := consts.HeaderNew(contract.data)
	var citizenId, walletId int64
	if head.StateId > 0 {
		citizenId = head.UserId
	} else {
		walletId = head.UserId
	}
	extend := map[string]interface{}{`type`: head.Type, `time`: head.Type, `stateId`: head.StateId,
		`citizenId`: citizenId, `walletId`: walletId}
	v := reflect.ValueOf(contract.data).Elem()
	t := v.Type()
	for i := 1; i < t.NumField(); i++ {
		extend[t.Field(i).Name] = v.Field(i).Interface()
	}
	//	fmt.Println(`Extend`, extend)
	return extend
}

func (contract *Contract) Init() error {
	init := contract.getFunc(`init`)
	if init == nil {
		return nil
	}
	rt := smartVM.RunInit()
	_, err := rt.Run(init, nil, contract.getExtend())
	return err
}

func (contract *Contract) Front() error {
	front := contract.getFunc(`front`)
	if front == nil {
		return nil
	}
	rt := smartVM.RunInit()
	_, err := rt.Run(front, nil, contract.getExtend())
	return err
}

func (contract *Contract) Main() error {
	main := contract.getFunc(`main`)
	if main == nil {
		return nil
	}
	rt := smartVM.RunInit()
	_, err := rt.Run(main, nil, contract.getExtend())
	return err
}
