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
	"github.com/EGaaS/go-egaas-mvp/packages/consts"
	logger "github.com/EGaaS/go-egaas-mvp/packages/log"

	"github.com/EGaaS/go-egaas-mvp/packages/crypto"
)

type evalCode struct {
	Source string
	Code   *Block
}

var (
	evals = make(map[uint64]*evalCode)
)

// CompileEval compiles conditional exppression
func (vm *VM) CompileEval(input string, state uint32) error {
	logger.LogDebug(consts.FuncStarted, "")
	source := `func eval bool { return ` + input + `}`
	block, err := vm.CompileBlock([]rune(source), &OwnerInfo{StateID: state})
	if err == nil {
		crc, err := crypto.CalcChecksum([]byte(input))
		if err != nil {
			logger.LogFatal(consts.CryptoError, err)
		}
		evals[crc] = &evalCode{Source: input, Code: block}
		return nil
	}
	if err != nil {
		logger.LogError(consts.VMError, err)
	}
	return err

}

// EvalIf runs the conditional expression. It compiles the source code before that if that's necessary.
func (vm *VM) EvalIf(input string, state uint32, vars *map[string]interface{}) (bool, error) {
	logger.LogDebug(consts.FuncStarted, "")
	if len(input) == 0 {
		return true, nil
	}
	crc, err := crypto.CalcChecksum([]byte(input))
	if err != nil {
		logger.LogFatal(consts.CryptoError, err)
	}
	if eval, ok := evals[crc]; !ok || eval.Source != input {
		if err := vm.CompileEval(input, state); err != nil {
			logger.LogError(consts.VMError, err)
			return false, err
		}
	}
	rt := vm.RunInit(CostDefault)
	ret, err := rt.Run(evals[crc].Code.Children[0], nil, vars)
	if err == nil {
		return valueToBool(ret[0]), nil
	}
	logger.LogError(consts.VMError, err)
	return false, err
}
