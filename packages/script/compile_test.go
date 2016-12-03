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
	Func   string
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

/*			if (111> 10) { //01 Commment
				if 0==1 {
					Println("TRUE TRUE temp function")
				} else { // 02 Commment
				eeee

3232 Комментарий
				}
			} else {
				Println("FALSE temp function")
			}
			return "OK"*/

func TestVMCompile(t *testing.T) {
	test := []TestVM{
		{`func formap string {
			var my map
			
			return Sprintf("result=%v %s", my, my["test"] )
		}`, `formap`, ``},
		{`func runtime string {
			var i int
			i = 50
			return Sprintf("val=%d", i 0)
		}`, `runtime`, `runtime panic error`},
		{`func nop {
			return
		}
		
		func loop string {
			var i int
			while true {//i < 10 {
				i=i+1
				if i==5 {
					continue
				}
				if i == 121 {
					i = i+ 4
					break
				}
			}
			nop()
			return Sprintf("val=%d", i)
		}`, `loop`, `val=125`},
		{`contract my {
							tx {
								Par1 int
								Par2 string
							}
							func front {
								Println("Front", $Par1)
				//				my("Par1,Par2,ext", 123, "Parameter 2", "extended" )
							}
							func main {
								Println("Main", $Par2, $ext)
							}
						}
						contract empty {
							func main {
								Println("Empty")
							}
						}
						contract mytest {
							func init string {
								my("Par1,Par2,ext", 123, "Parameter 2", "extended" )
								my("Par1,Par2,ext", 33123, "Parameter 332", "33extended" )
								empty("test",10)
								Println( "mytest")
								return "OK"
							}
						}
						`, `mytest.init`, `OK`},
		{`func money_test string {
					var my2, m1 money
					my2 = 100
					m1 = 1.2
					return Sprintf( "Account %v %v", my2 - 5.6, m1*5 + my2)
				}`, `money_test`, `Account 94.4 106`},

		{`func line_test string {
						return "Start " +
						Sprintf( "My String %s %d %d",
						      "Param 1", 24,
							345 + 789)
					}`, `line_test`, `Start My String Param 1 24 1134`},

		{`func err_test string {
						if 1001.02 {
							error "Error message err_test"
						}
						return "OK"
					}`, `err_test`, `Error message err_test`},
		{`contract my {
							tx {
								PublicKey  bytes
								FirstName  string
								MiddleName string "optional"
								LastName   string
							}
							func init string {
								return "OK"
							}
						}`, `my.init`, `OK`},

		{`func temp3 string {
							var i1 i2 int, s1 string, s2 string
							i2, i1 = 348, 7
							if i1 > 5 {
								var i5 int, s3 string
								i5 = 26788
								s1 = "s1 string"
								i2 = (i1+2)*i5+i2
								s2 = Sprintf("temp 3 function %s %d", Sprintf("%s + %d", s1, i2), -1 )
							}
							return s2
						}`, `temp3`, `temp 3 function s1 string + 241440 -1`},
		{`func params2(myval int, mystr string ) string {
							if 101>myval {
								if myval == 90 {
								} else {
									return Sprintf("myval=%d + %s", myval, mystr )
								}
							}
							return "OOPs"
						}
						func temp2 string {
							if true {
								return params2(51, "Params 2 test")
							}
						}
						`, `temp2`, `myval=51 + Params 2 test`},

		{`func params(myval int, mystr string ) string {
							return Sprintf("Params function %d %s", 33 + myval + $test1, mystr + " end" )
						}
						func temp string {
							return "Prefix " + params(20, "Test string " + $test2) + $test3( 202 )
						}
						`, `temp`, `Prefix Params function 154 Test string test 2 endtest=202=test`},
		{`func my_test string {
										return Sprintf("Called my_test %s %d", "Ooops", 777)
									}

							contract my {
									func initf string {
										return Sprintf("%d %s %s %s", 65123 + (1001-500)*11, my_test(), "Тестовая строка", Sprintf("> %s %d <","OK", 999 ))
									}
							}`, `my.initf`, `70634 Called my_test Ooops 777 Тестовая строка > OK 999 <`},
	}
	vm := NewVM()
	vm.Extend(&ExtendData{map[string]interface{}{"Println": fmt.Println, "Sprintf": fmt.Sprintf}, nil})

	for _, item := range test {
		source := []rune(item.Input)
		if err := vm.Compile(source); err != nil {
			t.Error(err)
		} else {
			if out, err := vm.Call(item.Func, nil, &map[string]interface{}{
				`test1`: 101, `test2`: `test 2`, `test3`: func(param int64) string {
					return fmt.Sprintf("test=%d=test", param)
				},
			}); err == nil {
				if out[0].(string) != item.Output {
					fmt.Println(out[0].(string))
					t.Error(`error vm ` + item.Input)
				}
			} else if err.Error() != item.Output {
				t.Error(err)
			}

		}
	}
	//	vm.Call(`Println`, []interface{}{"Qwerty", 100, `OOOPS`}, nil)
	//ret, _ := vm.Call(`Sprintf`, []interface{}{"Value %d %s OK", 100, `String value`}, nil)
	//fmt.Println(ret[0].(string))
	//	fmt.Println(`Result`, err)
}
