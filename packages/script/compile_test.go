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
	"fmt"
	"testing"
)

type TestVM struct {
	Input  string
	Output string
}

func (block *Block) String() (ret string) {
	/*	for _, item := range lexems {
		slex := string(source[item.Offset:item.Right])
		if item.Type == 0 {
			slex = `error`
		}
		ret += fmt.Sprintf("[%d %s]", item.Type, slex)
	}*/
	if (*block).Objects != nil {
		ret = fmt.Sprintf("Objects: %v", (*block).Objects)
	}
	if (*block).Children != nil {
		ret += fmt.Sprintf("Blocks: [\n")
		for i, item := range (*block).Children {
			ret += fmt.Sprintf("{%d: %v}\n", i, item.String())
		}
		ret += fmt.Sprintf("]")
	}
	return
}

func TestVMCompile(t *testing.T) {
	test := []TestLexem{
		{`func my_test string {
						return Sprintf("Called my_test %s %d", "Ooops", 777)
					}

			contract my {
					func temp {
						if (111> 10) { //01 Commment
if 0==1 {
								Println("TRUE TRUE temp function")
							} else { // 02 Commment
								Println("TRUE FALSE temp function" /*comment*/ )
								Println("Post") /* 
3232 Комментарий	*/
							}
						} else {
							Println("FALSE temp function")
						}
					}

					func init {
						Println(65123 + (1001-500)*11, my_test(), "Тестовая строка", Sprintf("> %s %d <","OK", 999 ))
						temp()
					}
			}`,
			``},
	}
	vm := VMInit(map[string]interface{}{"Println": fmt.Println, "Sprintf": fmt.Sprintf})

	for _, item := range test {
		source := []rune(item.Input)
		var out string
		if err := vm.Compile(source); err != nil {
			t.Error(err)
		} else {
			out = vm.String()
			if out != item.Output {
				//			t.Error(`error of vm compile ` + item.Input)
			}
		}
		//		fmt.Println(`%s`, out)
		//fmt.Printf("%s", item.Output)
	}
	vm.Call(`Println`, []interface{}{"Qwerty", 100, `OOOPS`}, nil)
	ret, _ := vm.Call(`Sprintf`, []interface{}{"Value %d %s OK", 100, `String value`}, nil)
	fmt.Println(ret[0].(string))
	_, err := vm.Call(`my.init`, nil, nil)
	fmt.Println(`Result`, err)
}
