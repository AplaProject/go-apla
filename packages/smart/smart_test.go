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
	"encoding/hex"
	"fmt"
	"testing"
	"time"

	"github.com/DayLightProject/go-daylight/packages/consts"
)

type TestSmart struct {
	Input  string
	Output string
}

func TestNewContract(t *testing.T) {
	var err error
	test := []TestSmart{
		{`contract NewCitizen {
			func front {
				Println("NewCitizen Front")
			}
			func main {
				Println("NewCitizen Main")
			}
}			
		`, ``},
	}
	for _, item := range test {
		if err := Compile(item.Input); err != nil {
			t.Error(err)
		}
	}
	sign, _ := hex.DecodeString(`3276233276237115`)
	public, _ := hex.DecodeString(`12456788999900087676`)
	data := &consts.TXNewCitizen{
		consts.TXHeader{4, uint32(time.Now().Unix()), 1, 1, sign}, public,
	}
	fmt.Println(`Data`, data)
	cnt := GetContract(`NewCitizen`, data)
	if cnt == nil {
		t.Error(`GetContract error`)
	}
	if err = cnt.Init(); err != nil {
		t.Error(err.Error())
	}
	if err = cnt.Front(); err != nil {
		t.Error(err.Error())
	}
	if err = cnt.Main(); err != nil {
		t.Error(err.Error())
	}

}
