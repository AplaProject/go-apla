// Apla Software includes an integrated development
// environment with a multi-level system for the management
// of access rights to data, interfaces, and Smart contracts. The
// technical characteristics of the Apla Software are indicated in
// Apla Technical Paper.

// Apla Users are granted a permission to deal in the Apla
// Software without restrictions, including without limitation the
// rights to use, copy, modify, merge, publish, distribute, sublicense,
// and/or sell copies of Apla Software, and to permit persons
// to whom Apla Software is furnished to do so, subject to the
// following conditions:
// * the copyright notice of GenesisKernel and EGAAS S.A.
// and this permission notice shall be included in all copies or
// substantial portions of the software;
// * a result of the dealing in Apla Software cannot be
// implemented outside of the Apla Platform environment.

// THE APLA SOFTWARE IS PROVIDED “AS IS”, WITHOUT WARRANTY
// OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED
// TO THE WARRANTIES OF MERCHANTABILITY, FITNESS FOR A
// PARTICULAR PURPOSE, ERROR FREE AND NONINFRINGEMENT. IN
// NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE
// LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER
// IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING
// FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR
// THE USE OR OTHER DEALINGS IN THE APLA SOFTWARE.

package script

import (
	"fmt"
	"strings"
	"testing"

	"github.com/AplaProject/go-apla/packages/types"

	"github.com/shopspring/decimal"
)

type TestVM struct {
	Input  string
	Func   string
	Output string
}

func (block *Block) String() (ret string) {
	if (*block).Objects != nil {
		ret = fmt.Sprintf("Objects: %v\n", (*block).Objects)
	}
	ret += fmt.Sprintf("Type: %v \n", (*block).Type)
	if (*block).Children != nil {
		ret += fmt.Sprintf("Blocks: [\n")
		for i, item := range (*block).Children {
			ret += fmt.Sprintf("{%d: %v}\n", i, item.String())
		}
		ret += fmt.Sprintf("]")
	}
	return
}

func getMap() *types.Map {
	myMap := types.NewMap()
	myMap.Set(`par0`, `Parameter 0`)
	myMap.Set(`par1`, `Parameter 1`)
	return myMap
}

func getArray() []interface{} {
	myMap := types.NewMap()
	myMap.Set(`par0`, `Parameter 0`)
	myMap.Set(`par1`, `Parameter 1`)
	return []interface{}{myMap,
		"The second string", int64(2000)}
}

// Str converts the value to a string
func str(v interface{}) (ret string) {
	return fmt.Sprint(v)
}

func lenArray(par []interface{}) int64 {
	return int64(len(par))
}

func Money(v interface{}) (ret decimal.Decimal) {
	ret, _ = ValueToDecimal(v)
	return ret
}

func outMap(v *types.Map) string {
	return fmt.Sprint(v)
}

func TestVMCompile(t *testing.T) {
	test := []TestVM{
		{`contract sets {
			settings {
				val = 1.56
				rate = 100000000000
				name="Name parameter"
			}
			action {
				$result = Settings("@22sets","name")
			}
		}
		func result() string {
			var par map
			return CallContract("@22sets", par) + "=" + sets()
		}
		`, `result`, `Name parameter=Name parameter`},
		{`func proc(par string) string {
					return par + "proc"
					}
				func forarray string {
					var my map
					var ret array
					var myret array
					
					ret = GetArray()
					myret[1] = "Another "
					my = ret[0]
					my["par3"] = 3456
					ret[2] = "Test"
					return Sprintf("result=%s+%s+%d+%s", ret[1], my["par0"], my["par3"], myret[1] + ret[2])
				}`, `forarray`, `result=The second string+Parameter 0+3456+Another Test`},
		{`func proc(par string) string {
				return par + "proc"
				}
			func formap string {
				var my map
				var ret map

				ret = GetMap()
		//			Println(ret)
				//Println("Ooops", ret["par0"], ret["par1"])
				my["par1"] = "my value" + proc(" space ")
				my["par2"] = 203 * (100-86)
				return Sprintf("result=%s+%d+%s+%s+%d", ret["par1"], my["par2"] + 32, my["par1"], proc($glob["test"] ), $glob["number"] )
			}`, `formap`, `result=Parameter 1+2874+my value space proc+String valueproc+1001`},
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
							data {
								Par1 int
								Par2 string
							}
							func conditions {
								var q int
								Println("Front", $Par1, $parent)
				//				my("Par1,Par2,ext", 123, "Parameter 2", "extended" )
							}
							func action {
								Println("Main", $Par2, $ext)
							}
						}
						contract mytest {
							func init string {
								empty()
								my("Par1,Par2,ext", 123, "Parameter 2", "extended" )
								//my("Par1,Par2,ext", 33123, "Parameter 332", "33extended" )
								//@26empty("test",10)
								empty("toempty", 10)
								Println( "mytest", $parent)
								return "OK INIT"
							}
						}
						contract empty {
							conditions {Println("EmptyCond")
								}
							action {
								Println("Empty", $parent)
								if 1 {
									my("Par1,Par2,ext", 123, "Parameter 2", "extended" )
								}
							}
						}
						`, `mytest.init`, `OK INIT`},
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
					}`, `err_test`, `{"type":"error","error":"Error message err_test"}`},
		{`contract my {
					data {
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
				return Sprintf("%d %s %s %s", 65123 + (1001-500)*11, my_test(), "Test message", Sprintf("> %s %d <","OK", 999 ))
			}
	}`, `my.initf`, `70634 Called my_test Ooops 777 Test message > OK 999 <`},
		{`contract vars {
		func cond() string {return "vars"}
		func actions() { var test int}
	}`, `vars.cond`, `vars`},
		{`func mytail(name string, tail ...) string {
		if lenArray(tail) == 0 {
			return name
		}
		if lenArray(tail) == 1 {
			return Sprintf("%s=%v ", name, tail[0])
		}
		return Sprintf("%s=%v+%v ", name, tail[1], tail[0])
	}
	func emptytail(tail ...) string {
		return Sprintf("%d ", lenArray(tail))
	}
	func sum(out string, values ...) string {
		var i, res int
		while i < lenArray(values) {
		   res = res + values[i]
		   i = i+1
		}
		return Sprintf(out, res)
	}
	func calltail() string {
		var out string
		out = emptytail() + emptytail(10) + emptytail("name1", "name2")
		out = out + mytail("OK") + mytail("1=", 11) + mytail("2=", "name", 11)
		return out + sum("Sum: %d", 10, 20, 30, 40)
	}
	`, `calltail`, `0 1 2 OK1==11 2==11+name Sum: 100`},
		{`func DBFind( table string).Columns(columns string) 
		. Where(format string, tail ...). Limit(limit int).
		Offset(offset int) string  {
		Println("DBFind", table, tail)
		return Sprintf("%s %s %s %d %d=", table, columns, format, limit, offset)
	}
	func names() string {
		var out, cols string
		cols = "name,value"
		out = DBFind( "mytable") + DBFind( "keys"
			).Columns(cols)+ DBFind( "keys"
				).Offset(199).Columns("qq"+"my")
		out = out + DBFind( "table").Columns("name").Where("id=?", 
			100).Limit(10) + DBFind( "table").Where("request")
		return out
	}`, `names`, `mytable   0 0=keys name,value  0 0=keys qqmy  0 199=table name id=? 10 0=table  request 0 0=`},
		{`contract seterr {
				func getset string {
					var i int
					i = MyFunc("qqq", 10)
					return "OK"
				}
			}`, `seterr.getset`, `unknown identifier MyFunc`},
		{`func one() int {
				return 9
			}
			func signfunc string {
				var myarr array
				myarr[0] = 0
				myarr[1] = 1
				var i, k, j int
				k = one()-2
				j = /*comment*/-3
				i = lenArray(myarr) - 1
				return Sprintf("%s %d %d %d %d %d", "ok", lenArray(myarr)-1, i, k, j, -4)
			}`, `signfunc`, `ok 1 1 7 -3 -4`},
		{`func exttest() string {
				return Replace("text", "t")
			}
			`, `exttest`, `function Replace must have 4 parameters`},
		{`func mytest(first string, second int) string {
				return Sprintf("%s %d", first, second)
		}
		func test() {
			return mytest("one", "two")
		}
		`, `test`, `parameter 2 has wrong type`},
		{`func mytest(first string, second int) string {
								return Sprintf("%s %d", first, second)
						}
						func test() string {
							return mytest("one")
						}
						`, `test`, `wrong count of parameters`},
		{
			`func ifMap string {
				var m map
				if m {
					return "empty"
				}
				
				m["test"]=1
				if m {
					return "not empty"
				}

				return error "error"
			}`, "ifMap", "not empty",
		},
		{`func One(list array, name string) string {
			if list {
				var row map 
				row = list[0]
				return row[name]
			}
			return nil
		}
		func Row(list array) map {
			var ret map
			if list {
				ret = list[0]
			}
			return ret
		}
		func GetData().WhereId(id int) array {
			var par array
			var item map
			item["id"] = str(id)
			item["name"] = "Test value " + str(id)
			par[0] = item
			return par
		}
		func GetEmpty().WhereId(id int) array {
			var par array
			return par
		}
		func result() string {
			var m map
			var s string
			m = GetData().WhereId(123).Row()
			s = GetEmpty().WhereId(1).One("name") 
			if s != nil {
				return "problem"
			}
			return m["id"] + "=" + GetData().WhereId(100).One("name")
		}`, `result`, `123=Test value 100`},
		{`func mapbug() string {
			$data[10] = "extend ok"
			return $data[10]
			}`, `mapbug`, `extend ok`},
		{`func result() string {
				var myarr array
				myarr[0] = "string"
				myarr[1] = 7
				myarr[2] = "9th item"
				return Sprintf("RESULT=%s %d %v", myarr...)
			}`, `result`, `RESULT=string 7 9th item`},
		{`func find().Where(pattern string, params ...) string {
				return Sprintf(pattern, params ...)
			}
			func row().Where(pattern string, params ...) string {
				return find().Where(pattern, params ...)
			}
			func result() string {
				return row().Where("%d %d", 10, 20)
			}
			`, `result`, `10 20`},
		{`func result string {
				var arr array
				var mymap map
				arr[100000] = 0
				var i int
				while i < 100 {
					mymap[str(i)] = 10
					i = i + 1
				}
				i = i + "2" 
				i = (i - "10")/"2"*"3"
				return Sprintf("%T %[1]v", .21 + i)
			  }`, `result`, `float64 138.21`},
		{`func money_test string {
				var my2, m1 money
				my2 = 100
				m1 = 1.2
				return Sprintf( "Account %v %v %v", my2/Money(3),  my2 - Money(5.6), m1*Money(5) + Money(my2))
			}`, `money_test`, `Account 33 95 105`},
		{`func long() int {
				return  99999999999999999999
				}
				func result() string {
					return Sprintf("ok=%d", long())
					}`, `result`, `strconv.ParseInt: parsing "99999999999999999999": value out of range 99999999999999999999 [Ln:2 Col:34]`},
		{`func result() string {
			var i, result int
			
			if true {
				if false {
					result = 99
				} else {
					result = 5
				}
			}
			if i == 1 {
				result = 20
			} elif i> 0 {
				result = 30
			} 
			elif i == 0 
			{
				result = result + 50
				if true {
					i=10
				}
			} elif i==10 {
				Println("3")
				result = 0
				i=33
			} elif false {
				Println("4")
				result = 1
			} 
			else 
			{
				Println("5")
				result = 2
			}
			if i == 4 {
				result = result
			} elif i == 20 {
				result = 22
			} else {
				result = result + 23
				i = 11
			}
			if i == 11 {
				result = result + 7
			} else {
				result = 0
			}
			if result == 85 {
				if false {
					result = 1
				} elif 0 {
					result = 5
				} elif 1 {
					result = result + 10
				}
			}
			if result == 10 {
				result = 11
			} elif result == 95 {
				result = result + 1
				if false {
					result = 0
				} elif true {
					result = result + 4
				}
			}
			return Sprintf("%d", result)
		}
		`, `result`, `100`},
		{`func initerr string {
			var my map
			return {qqq
		`, `initerr`, `unclosed map initialization`},
		{`func initmap string {
			var my, sub map
			var list array
			var i int
			i = 256
			var s string
			$ext = "Ooops"
			s = "Spain"
			my = {conditions: "$Conditions"}
			list = [0, i, {"item": i}, [$ext]]
			sub = {"name": "John", "lastname": "Smith", myarr: []}
			my = {qqq: 10, "22": "MY STRING", /* comment*/ "float": 1.2, "ext": $ext,
			"in": true, "var": i, sub: sub, "Company": {"Name": "Ltd", Country: s, 
				Arr: [s, 20, "finish"]}}
			return outMap(my) + Sprintf("%v", list)
		}`, `initmap`, `map[qqq:10 22:MY STRING float:1.2 ext:Ooops in:true var:256 sub:map[name:John lastname:Smith myarr:[]] Company:map[Name:Ltd Country:Spain Arr:[Spain 20 finish]]][0 256 map[item:256] [Ooops]]`},
		{`func test() string {
			var where map
			where["name"] = {"$in": "menus_names"}
			return Sprintf("%v", where)
		 }`, `test`, `map[name:map[$in:menus_names]]`},
		{`contract TestCyr {
			data {}
			conditions { }
			action {
			   //тест
			   var a map
			   a["тест"] = "тест"
			   $result = a["тест"]
			}
		}
		func result() string {
			var par map
			return CallContract("TestCyr", par) 
		}`, `result`, `тест`},
		{`contract MainCond {
			conditions {
				error $test
			}
			action {
				$result = "OK"
			}
		}
		func result() bool {
			return MainCond
		}
		`, `result`, `unknown variable MainCond`},
		{`func myFunc(my string) string {
			return Sprintf("writable: %s", my)
		}
		contract mySet {
			conditions {
				myFunc("test")	
			}
			action {
				myFunc("test")	
			}
		}	
		contract myExec {
			conditions {
				mySet()
			}
			action {
				mySet()
				$result = "OK"
			}
		}
		func result() string {
			myExec()
			return "COND"
		}`, `result`, `'conditions' cannot call contracts or functions which can modify the blockchain database.`},
		{`func test string {
			var s string
			var m map
			m = {f: 5, b: 2, a: 1, d: 3, c: 0, e: 4}
			var i int
			while i<3{
				s = s + Sprintf("%v", m)
				i = i + 1
			}
			return s
		}
		`, `test`, `map[f:5 b:2 a:1 d:3 c:0 e:4]map[f:5 b:2 a:1 d:3 c:0 e:4]map[f:5 b:2 a:1 d:3 c:0 e:4]`},
		{`contract qqq3 {
			data {
				Name string "aaq"
				Temp
			}
			action {
				$result = $Name
			}
		}
		`, `qqq3.action`, `expecting type of the data field [Ln:5 Col:1]`},
		{`contract qqq2 {
			data {
				Name string "aaq"
				"awede"
			}
			action {
				$result = $Name
			}
		}
		`, `qqq2.action`, `unexpected tag [Ln:4 Col:6]`},
		{`contract qqq1 {
			data {
				string Name qwerty
			}
			action {
				$result = $Name
			}
		}
		`, `qqq1.action`, `expecting name of the data field [Ln:3 Col:6]`},
		{`contract qqq {
			data {
				Name qwerty
			}
			action {
				$result = $Name
			}
		}
		`, `qqq.action`, `expecting type of the data field [Ln:3 Col:11]`},
		{`contract qq3 {
			data {
				Id uint
			}
			action {
				$result = "OK"
			}
		}
		`, `qq3.action`, `expecting type of the data field [Ln:3 Col:9]`},
		{`contract qq2 {
			data {
				Id, ID2 int
			}
			action {
				$result = str($Id) + str($ID2)
			}
		}
		func getqq() string {
			return qq2("Id,ID2", 10,20)
		}`, `getqq`, `1020`},
		{`func IND() string {
			var a,b,d array
			a[0] = 100
			a[1] = 555
			b[0] = 200
			d[0] = a
			d[1] = b
			d[0][0] =  777
	}`, `IND`, `multi-index is not supported`},
	}
	vm := NewVM()
	vm.Extern = true
	vm.Extend(&ExtendData{map[string]interface{}{"Println": fmt.Println, "Sprintf": fmt.Sprintf,
		"GetMap": getMap, "GetArray": getArray, "lenArray": lenArray, "outMap": outMap,
		"str": str, "Money": Money, "Replace": strings.Replace}, nil,
		map[string]struct{}{"Sprintf": {}}})

	for ikey, item := range test {
		if ikey > 100 {
			break
		}
		source := []rune(item.Input)
		if err := vm.Compile(source, &OwnerInfo{StateID: uint32(ikey) + 22, Active: true, TableID: 1}); err != nil {
			if err.Error() != item.Output {
				t.Error(err)
				break
			}
		} else {
			glob := types.NewMap()
			glob.Set(`test`, `String value`)
			glob.Set(`number`, 1001)
			if out, err := vm.Call(item.Func, nil, &map[string]interface{}{
				`rt_state`: uint32(ikey) + 22, `data`: make([]interface{}, 0),
				`test1`: 101, `test2`: `test 2`,
				"glob": glob,
				`test3`: func(param int64) string {
					return fmt.Sprintf("test=%d=test", param)
				},
			}); err == nil {
				if out[0].(string) != item.Output {
					t.Error(`error vm ` + out[0].(string) + `!=` + item.Output)
					break
				}
			} else if err.Error() != item.Output {
				t.Error(err)
				break
			}

		}
	}
	t.Error(`OK`)
}

func TestContractList(t *testing.T) {
	test := []TestLexem{{`contract NewContract {
		conditions {
			ValidateCondition($Conditions,$ecosystem_id)
			while i < Len(list) {
				if IsObject(list[i], $ecosystem_id) {
					warning Sprintf("Contract or function %s exists", list[i] )
				}
			}
		}
		action {
		}
		func price() int {
			return  SysParamInt("price_create_contract")
		}
	}func MyFunc {}`,
		`NewContract,MyFunc`},
		{`contract demo_сontract {
			data {
				contract_txt str
			}
			func test() {
			}
			conditions {
				if $contract_txt="" {
					warning "Sorry, you do not have contract access to this action."
				}
			}
		} contract another_contract {} func main { func subfunc(){}}`,
			`demo_сontract,another_contract,main`},
	}
	for _, item := range test {
		list, _ := ContractsList(item.Input)
		if strings.Join(list, `,`) != item.Output {
			t.Error(`wrong names`, strings.Join(list, `,`))
			break
		}
	}
}
