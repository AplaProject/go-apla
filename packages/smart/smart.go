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

	"github.com/DayLightProject/go-daylight/packages/script"
)

type Contract struct {
	Name  string
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
func GetContract(name string) *Contract {
	obj, ok := smartVM.Objects[name]
	if ok && obj.Type == script.OBJ_CONTRACT {
		return &Contract{Name: name, block: obj.Value.(*script.Block)}
	}
	return nil
}

func (contract *Contract) Init() error {
	return nil
}

func (contract *Contract) Front() error {
	return nil
}

func (contract *Contract) Main() error {
	return nil
}
