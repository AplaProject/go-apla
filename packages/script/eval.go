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
	//	"fmt"
	"hash/crc64"

	"github.com/EGaaS/go-egaas-mvp/packages/lib"
)

type EvalCode struct {
	Source string
	Code   *Block
}

var (
	evals = make(map[uint64]*EvalCode)
)

func (vm *VM) CompileEval(input string, state uint32) error {
	source := `func eval bool { return ` + input + `}`
	block, err := vm.CompileBlock([]rune(source), state, false, 0)
	//	fmt.Println(`Compile Eval`, err, input)
	if err == nil {
		crc := crc64.Checksum([]byte(input), lib.Table64)
		evals[crc] = &EvalCode{Source: input, Code: block}
		return nil
	}
	return err

}

func (vm *VM) EvalIf(input string, state uint32, vars *map[string]interface{}) (bool, error) {
	if len(input) == 0 {
		return true, nil
	}
	crc := crc64.Checksum([]byte(input), lib.Table64)
	if eval, ok := evals[crc]; !ok || eval.Source != input {
		if err := vm.CompileEval(input, state); err != nil {
			return false, err
		}
	}
	rt := vm.RunInit(COST_DEFAULT)
	ret, err := rt.Run(evals[crc].Code.Children[0], nil, vars)
	if err == nil {
		return valueToBool(ret[0]), nil
	}
	return false, err
}
