// MIT License
//
// Copyright (c) 2016-2018 GenesisKernel
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

package script

import (
	log "github.com/sirupsen/logrus"

	"github.com/GenesisKernel/go-genesis/packages/consts"
	"github.com/GenesisKernel/go-genesis/packages/crypto"
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
	source := `func eval bool { return ` + input + `}`
	block, err := vm.CompileBlock([]rune(source), &OwnerInfo{StateID: state})
	if err == nil {
		crc, err := crypto.CalcChecksum([]byte(input))
		if err != nil {
			log.WithFields(log.Fields{"type": consts.CryptoError, "error": err}).Fatal("calculating compile eval input checksum")
		}
		evals[crc] = &evalCode{Source: input, Code: block}
		return nil
	}
	return err

}

// EvalIf runs the conditional expression. It compiles the source code before that if that's necessary.
func (vm *VM) EvalIf(input string, state uint32, vars *map[string]interface{}) (bool, error) {
	if len(input) == 0 {
		return true, nil
	}
	crc, err := crypto.CalcChecksum([]byte(input))
	if err != nil {
		log.WithFields(log.Fields{"type": consts.CryptoError, "error": err}).Fatal("calculating compile eval checksum")
	}
	if eval, ok := evals[crc]; !ok || eval.Source != input {
		if err := vm.CompileEval(input, state); err != nil {
			log.WithFields(log.Fields{"type": consts.EvalError, "error": err}).Error("compiling eval")
			return false, err
		}
	}
	rt := vm.RunInit(CostDefault)
	ret, err := rt.Run(evals[crc].Code.Children[0], nil, vars)
	if err == nil {
		if len(ret) == 0 {
			return false, nil
		}
		return valueToBool(ret[0]), nil
	}
	return false, err
}
