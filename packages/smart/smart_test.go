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
	//	"encoding/hex"
	"fmt"
	"testing"
	//	"time"
	//	"github.com/EGaaS/go-egaas-mvp/packages/consts"
	//	"github.com/EGaaS/go-egaas-mvp/packages/utils"
)

type TestSmart struct {
	Input  string
	Output string
}

func TestNewContract(t *testing.T) {
	//	var err error
	test := []TestSmart{
		{`contract NewCitizen {
			data {
				Public bytes
				MyVal  string
			}
			func conditions {
				Println( "Front")//TxJson())
				//$tmp = "Test string"
//				Println("NewCitizen Front", $tmp, $citizen, $state, $PublicKey )
			}
			func action {
//				Println("NewCitizen Main", $tmp, $type, $wallet )
//				DBInsert(Sprintf( "%d_citizens", $state), "public_key,block_id", $PublicKey, $block)
			}
}			
		`, ``},
	}
	for _, item := range test {
		if err := Compile(item.Input, `1`, true, 1); err != nil {
			t.Error(err)
		}
	}
	cnt := GetContract(`NewCitizen`, 1)
	cfunc := cnt.GetFunc(`conditions`)
	_, err := Run(cfunc, nil, &map[string]interface{}{})
	fmt.Println(`Err`, err)
	//
	/*	if err = cnt.Call(CALL_INIT | CALL_FRONT | CALL_MAIN); err != nil {
			t.Error(err.Error())
	}*/
	//	sign, _ := hex.DecodeString(`3276233276237115`)
	//	public, _ := hex.DecodeString(`12456788999900087676`)
	//	p := Parser{BlockData: &utils.BlockData{BlockId: 133}}
	/*	p.TxPtr = &consts.TXNewCitizen{
			consts.TXHeader{4, uint32(time.Now().Unix()), 1, 1, sign}, public,
		}
		//	fmt.Println(`Data`, data)
		cnt := GetContract(`NewCitizen`, &p)
		if cnt == nil {
			t.Error(`GetContract error`)
		}
		if err = cnt.Call(CALL_INIT | CALL_FRONT | CALL_MAIN); err != nil {
			t.Error(err.Error())
		}*/
}
