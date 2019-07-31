// Copyright (C) 2017, 2018, 2019 EGAAS S.A.
//
// This program is free software; you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation; either version 2 of the License, or (at
// your option) any later version.
//
// This program is distributed in the hope that it will be useful, but
// WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the GNU
// General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program; if not, write to the Free Software
// Foundation, Inc., 51 Franklin Street, Fifth Floor, Boston, MA 02110-1301, USA.

package script

import (
	log "github.com/sirupsen/logrus"

	"github.com/AplaProject/go-apla/packages/conf/syspar"
	"github.com/AplaProject/go-apla/packages/consts"
	"github.com/AplaProject/go-apla/packages/crypto"
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
	rt := vm.RunInit(syspar.GetMaxCost())
	ret, err := rt.Run(evals[crc].Code.Children[0], nil, vars)
	if err == nil {
		if len(ret) == 0 {
			return false, nil
		}
		return valueToBool(ret[0]), nil
	}
	return false, err
}
