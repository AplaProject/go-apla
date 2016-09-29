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

package parser

import (
	"fmt"
	"reflect"
	//	"strings"

	"github.com/DayLightProject/go-daylight/packages/consts"
	"github.com/DayLightProject/go-daylight/packages/script"
)

type Contract struct {
	Name   string
	Called uint32
	parser *Parser //interface{}
	block  *script.Block
}

const (
	CALL_INIT  = 0x01
	CALL_FRONT = 0x02
	CALL_MAIN  = 0x04
)

var (
	smartVM *script.VM
)

func init() {
	smartVM = script.VMInit(map[string]interface{}{
		"Println":  fmt.Println,
		"Sprintf":  fmt.Sprintf,
		"DBUpdate": DBUpdate,
	}, map[string]string{
		`*parser.Parser`: `parser`,
	})
}

// Compiles contract source code
func Compile(src string) error {
	return smartVM.Compile([]rune(src))
}

// Returns true if the contract exists
func GetContract(name string, p *Parser /*data interface{}*/) *Contract {
	obj, ok := smartVM.Objects[name]
	//	fmt.Println(`Get`, ok, obj, obj.Type, script.OBJ_CONTRACT)
	if ok && obj.Type == script.OBJ_CONTRACT {
		return &Contract{Name: name, parser: p, block: obj.Value.(*script.Block)}
	}
	return nil
}

func (contract *Contract) getFunc(name string) *script.Block {
	if block, ok := (*contract).block.Objects[name]; ok && block.Type == script.OBJ_FUNC {
		return block.Value.(*script.Block)
	}
	return nil
}

func (contract *Contract) getExtend() *map[string]interface{} {
	head := consts.HeaderNew(contract.parser.TxPtr)
	var citizenId, walletId int64
	if head.StateId > 0 {
		citizenId = head.UserId
	} else {
		walletId = head.UserId
	}
	extend := map[string]interface{}{`type`: head.Type, `time`: head.Type, `state`: head.StateId,
		`block`: contract.parser.BlockData.BlockId, `citizen`: citizenId, `wallet`: walletId,
		`parser`: contract.parser}
	v := reflect.ValueOf(contract.parser.TxPtr).Elem()
	t := v.Type()
	for i := 1; i < t.NumField(); i++ {
		extend[t.Field(i).Name] = v.Field(i).Interface()
	}
	//	fmt.Println(`Extend`, extend)
	return &extend
}

func (contract *Contract) Call(flags int) (err error) {
	methods := []string{`init`, `front`, `main`}
	extend := contract.getExtend()
	for i := uint32(0); i < 3; i++ {
		if (flags & (1 << i)) > 0 {
			cfunc := contract.getFunc(methods[i])
			if cfunc == nil {
				continue
			}
			rt := smartVM.RunInit()
			contract.Called = 1 << i
			_, err = rt.Run(cfunc, nil, extend)
			if err != nil {
				return
			}
		}
	}
	return
}

func DBUpdate(p *Parser, tblname string, params string, val ...interface{}) (err error) { // map[string]interface{}) {
	//	err := p.selectiveLoggingAndUpd(strings.Split(params, `,`), val, tblname, nil, nil, true)
	fmt.Println(`DBUpdate`, p, tblname, params, val)
	return
}
