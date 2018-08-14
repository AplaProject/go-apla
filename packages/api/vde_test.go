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

package api

import (
	"fmt"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/GenesisKernel/go-genesis/packages/conf"
	"github.com/GenesisKernel/go-genesis/packages/consts"
	"github.com/GenesisKernel/go-genesis/packages/converter"
	"github.com/GenesisKernel/go-genesis/packages/crypto"
	taskContract "github.com/GenesisKernel/go-genesis/packages/scheduler/contract"
)

func TestVDETables(t *testing.T) {
	require.NoError(t, keyLogin(1))
	var res tableResult

	require.NoError(t, sendGet("/table/system_parameters", nil, &res))
	fmt.Println(res)
}

func TestVDECreate(t *testing.T) {
	require.NoError(t, keyLogin(1))

	form := url.Values{
		"VDEName":    {"myvde3"},
		"DBUser":     {"myvde3user"},
		"DBPassword": {"vdepassword"},
		"VDEAPIPort": {"8098"},
	}
	assert.NoError(t, postTx("NewVDE", &form))
}

func TestVDEList(t *testing.T) {
	require.NoError(t, keyLogin(1))

	fmt.Println(postTx("ListVDE", nil))
}

func TestVDECron(t *testing.T) {
	require.NoError(t, keyLogin(1))

	data := &url.Values{
		"Cron":       {"1 * * * *"},
		"Contract":   {"MainCondition"},
		"Conditions": {"true"},
	}

	require.NoError(t, postTx("NewCron", data))
}

func TestStopVDE(t *testing.T) {
	require.NoError(t, keyLogin(1))
	form := url.Values{
		"VDEName": {"myvde3"},
	}
	require.NoError(t, postTx("StopVDE", &form))
}

func TestRunVDE(t *testing.T) {
	require.NoError(t, keyLogin(1))
	form := url.Values{
		"VDEName": {"myvde3"},
	}
	require.NoError(t, postTx("RunVDE", &form))
}

func TestRemoveVDE(t *testing.T) {
	require.NoError(t, keyLogin(1))
	form := url.Values{
		"VDEName": {"myvde3"},
	}
	require.NoError(t, postTx("RemoveVDE", &form))
}

func TestCreateTable(t *testing.T) {
	require.NoError(t, keyLogin(1))

	sql1 := `new_column`

	form := url.Values{
		"Name":          {"my_test_table"},
		"Columns":       {"[{\"name\":\"" + sql1 + "\",\"type\":\"varchar\", \"index\": \"0\", \"conditions\":{\"update\":\"true\", \"read\":\"true\"}}]"},
		"ApplicationId": {"1"},
		"Permissions":   {"{\"insert\": \"true\", \"update\" : \"true\", \"new_column\": \"true\"}"},
	}

	require.NoError(t, postTx("NewTable", &form))
}

func TestVDEParams(t *testing.T) {
	assert.NoError(t, keyLogin(1))

	rnd := `rnd` + crypto.RandSeq(6)
	form := url.Values{`Name`: {rnd}, `Value`: {`Test value`}, `Conditions`: {`ContractConditions("MainCondition")`},
		`vde`: {`true`}}

	assert.NoError(t, postTx(`NewParameter`, &form))

	var ret ecosystemParamsResult
	assert.NoError(t, sendGet(`ecosystemparams?vde=true`, nil, &ret))
	if len(ret.List) < 5 {
		t.Errorf(`wrong count of parameters %d`, len(ret.List))
	}

	assert.NoError(t, sendGet(`ecosystemparams?vde=true&names=stylesheet,`+rnd, nil, &ret))
	assert.Len(t, ret.List, 2, fmt.Errorf(`wrong count of parameters %d`, len(ret.List)))

	var parValue paramValue
	assert.NoError(t, sendGet(`ecosystemparam/`+rnd+`?vde=true`, nil, &parValue))
	assert.Equal(t, rnd, parValue.Name)

	var tblResult tablesResult
	assert.NoError(t, sendGet(`tables?vde=true`, nil, &tblResult))
	if tblResult.Count < 5 {
		t.Error(fmt.Errorf(`wrong tables result`))
	}

	form = url.Values{"Name": {rnd}, `vde`: {`1`}, "Columns": {`[{"name":"MyName","type":"varchar", "index": "1",
		"conditions":"true"},
	  {"name":"Amount", "type":"number","index": "0", "conditions":"true"},
	  {"name":"Active", "type":"character","index": "0", "conditions":"true"}]`},
		"Permissions": {`{"insert": "true", "update" : "true", "new_column": "true"}`}}
	assert.NoError(t, postTx(`NewTable`, &form))

	var tResult tableResult
	assert.NoError(t, sendGet(`table/`+rnd+`?vde=true`, nil, &tResult))
	assert.Equal(t, rnd, tResult.Name)

	var retList listResult
	assert.NoError(t, sendGet(`list/contracts?vde=true`, nil, &retList))
	if converter.StrToInt64(retList.Count) < 7 {
		t.Errorf(`The number of records %s < 7`, retList.Count)
		return
	}

	var retRow rowResult
	assert.NoError(t, sendGet(`row/contracts/2?vde=true`, nil, &retRow))
	if !strings.Contains(retRow.Value[`value`], `VDEFunctions`) {
		t.Error(`wrong row result`)
		return
	}

	var retCont contractsResult
	assert.NoError(t, sendGet(`contracts?vde=true`, nil, &retCont))

	form = url.Values{`Value`: {`contract ` + rnd + ` {
		data {
			Par string
		}
		action { Test("active",  $Par)}}`}, `Conditions`: {`ContractConditions("MainCondition")`}, `vde`: {`true`}}

	assert.NoError(t, postTx(`NewContract`, &form))

	var cont getContractResult
	assert.NoError(t, sendGet(`contract/`+rnd+`?vde=true`, nil, &cont))
	if !strings.HasSuffix(cont.Name, rnd) {
		t.Error(`wrong contract result`)
		return
	}

	form = url.Values{"Name": {rnd}, "Value": {`Page`}, "Menu": {`government`},
		"Conditions": {`true`}, `vde`: {`1`}}
	assert.NoError(t, postTx(`NewPage`, &form))

	assert.NoError(t, sendPost(`content/page/`+rnd, &url.Values{`vde`: {`true`}}, &ret))

	form = url.Values{"Name": {rnd}, "Value": {`Menu`}, "Conditions": {`true`}, `vde`: {`1`}}
	assert.NoError(t, postTx(`NewMenu`, &form))

	assert.NoError(t, sendPost(`content/menu/`+rnd, &url.Values{`vde`: {`true`}}, &ret))

	name := randName(`lng`)
	value := `{"en": "My VDE test", "fr": "French VDE test"}`

	form = url.Values{"Name": {name}, "Trans": {value}, "vde": {`true`}}
	assert.NoError(t, postTx(`NewLang`, &form))

	input := fmt.Sprintf(`Span($%s$)+LangRes(%[1]s,fr)`, name)
	var retContent contentResult
	assert.NoError(t, sendPost(`content`, &url.Values{`template`: {input}, `vde`: {`true`}}, &retContent))
	assert.Equal(t, `[{"tag":"span","children":[{"tag":"text","text":"My VDE test"}]},{"tag":"text","text":"+French VDE test"}]`, RawToString(retContent.Tree))

	name = crypto.RandSeq(4)
	assert.NoError(t, postTx(`Import`, &url.Values{"vde": {`true`}, "Data": {fmt.Sprintf(vdeimp, name)}}))
}

var vdeimp = `{
    "pages": [
        {
            "Name": "imp_page2",
            "Conditions": "true",
            "Menu": "imp",
            "Value": "imp"
        }
    ],
    "blocks": [
        {
            "Name": "imp2",
            "Conditions": "true",
            "Value": "imp"
        }
    ],
    "menus": [
        {
            "Name": "imp2",
            "Conditions": "true",
            "Value": "imp"
        }
    ],
    "parameters": [
        {
            "Name": "founder_account2",
            "Value": "-6457397116804798941",
            "Conditions": "ContractConditions(\"MainCondition\")"
        },
        {
            "Name": "test_pa2",
            "Value": "1",
            "Conditions": "true"
        }
    ],
    "languages": [
        {
            "Name": "est2",
            "Trans": "{\"en\":\"yeye\",\"te\":\"knfek\"}"
        }
    ],
    "contracts": [
        {
            "Name": "testCont2",
            "Value": "contract testCont2 {\n    data {\n\n    }\n\n    conditions {\n\n    }\n\n    action {\n        $result=\"privet\"\n    }\n}",
            "Conditions": "true"
        }
    ],
    "tables": [
        {
            "Name": "tests2",
            "Columns": "[{\"name\":\"name\",\"type\":\"text\",\"conditions\":\"true\"}]",
            "Permissions": "{\"insert\":\"true\",\"update\":\"true\",\"new_column\":\"true\"}"
        }
    ],
    "data": [
        {
            "Table": "tests2",
            "Columns": [
                "name"
            ],
            "Data": []
        }
    ]
}`

func TestVDEImport(t *testing.T) {
	if err := keyLogin(1); err != nil {
		t.Error(err)
		return
	}
	err := postTx(`Import`, &url.Values{"vde": {`true`}, "Data": {vdeimp}})
	if err != nil {
		t.Error(err)
		return
	}
}

func TestHTTPRequest(t *testing.T) {
	if err := keyLogin(1); err != nil {
		t.Error(err)
		return
	}
	rnd := `rnd` + crypto.RandSeq(6)
	form := url.Values{`Value`: {`contract ` + rnd + ` {
		    data {
				Auth string
			}
			action {
				var ret string 
				var pars, heads, json map
				$ret_table = DBFind("tables").Columns("id").Where("name=$", "notable")
				ret = HTTPRequest("http://www.instagram.com/", "GET", heads, pars)
				if !Contains(ret, "react-root") {
					error "instagram error"
				}
				ret = HTTPRequest("http://www.google.com/search?q=exotic", "GET", heads, pars)
				if !Contains(ret, "exotic") {
					error "google error"
				}
				heads["Authorization"] = "Bearer " + $Auth
				pars["vde"] = "true"
				ret = HTTPRequest("http://localhost:7079` + consts.ApiPath + `content/page/` + rnd + `", "POST", heads, pars)
				json = JSONDecode(ret)
				if json["menu"] != "myvdemenu" {
					error "Wrong vde menu"
				}
				ret = HTTPRequest("http://localhost:7079` + consts.ApiPath + `contract/VDEFunctions?vde=true", "GET", heads, pars)
				json = JSONDecode(ret)
				if json["name"] != "@1VDEFunctions" {
					error "Wrong vde contract"
				}
			}}`}, `Conditions`: {`true`}, `vde`: {`true`}}

	if err := postTx(`NewContract`, &form); err != nil {
		t.Error(err)
		return
	}
	form = url.Values{"Name": {rnd}, "Value": {`Page`}, "Menu": {`myvdemenu`},
		"Conditions": {`true`}, `vde`: {`true`}}
	if err := postTx(`NewPage`, &form); err != nil {
		t.Error(err)
		return
	}
	if err := postTx(rnd, &url.Values{`vde`: {`true`}, `Auth`: {gAuth}}); err != nil {
		t.Error(err)
		return
	}
}

func TestNodeHTTPRequest(t *testing.T) {
	var err error
	assert.NoError(t, keyLogin(1))

	rnd := `rnd` + crypto.RandSeq(4)
	form := url.Values{`Value`: {`contract for` + rnd + ` {
		data {
			Par string
		}
		action { $result = "Test NodeContract " + $Par + " ` + rnd + `"}
    }`}, `Conditions`: {`ContractConditions("MainCondition")`}}
	assert.NoError(t, postTx(`NewContract`, &form))

	var ret getContractResult
	assert.NoError(t, sendGet(`contract/for`+rnd, nil, &ret))

	assert.NoError(t, postTx(`ActivateContract`, &url.Values{`Id`: {ret.TableID}}))

	form = url.Values{`Value`: {`contract ` + rnd + ` {
		    data {
				Par string
			}
			action {
				var ret string 
				var pars, heads, json map
				heads["Authorization"] = "Bearer " + $auth_token
				pars["vde"] = "false"
				pars["Par"] = $Par
				ret = HTTPRequest("http://localhost:7079` + consts.ApiPath + `node/for` + rnd + `", "POST", heads, pars)
				json = JSONDecode(ret)
				$result = json["hash"]
			}}`}, `Conditions`: {`true`}, `vde`: {`true`}}
	assert.NoError(t, postTx(`NewContract`, &form))

	var (
		msg string
		id  int64
	)
	_, msg, err = postTxResult(rnd, &url.Values{`vde`: {`true`}, `Par`: {`node`}})
	assert.NoError(t, err)

	id, err = waitTx(msg)
	if id != 0 && err != nil {
		msg = err.Error()
		err = nil
	}
	assert.Equal(t, `Test NodeContract node `+rnd, msg)

	form = url.Values{`Value`: {`contract node` + rnd + ` {
		data {
		}
		action { 
			var ret string 
			var pars, heads, json map
			heads["Authorization"] = "Bearer " + $auth_token
			pars["vde"] = "false"
			pars["Par"] = "NodeContract testing"
			ret = HTTPRequest("http://localhost:7079` + consts.ApiPath + `node/for` + rnd + `", "POST", heads, pars)
			json = JSONDecode(ret)
			$result = json["hash"]
		}
	}`}, `Conditions`: {`ContractConditions("MainCondition")`}, `vde`: {`true`}}
	assert.NoError(t, postTx(`NewContract`, &form))

	// You can specify the directory with NodePrivateKey & NodePublicKey files
	if len(conf.Config.KeysDir) > 0 {
		conf.Config.HTTP.Host = `localhost`
		conf.Config.HTTP.Port = 7079

		nodeResult, err := taskContract.NodeContract(`@1node` + rnd)
		assert.NoError(t, err)

		id, err = waitTx(nodeResult.Result)
		if id != 0 && err != nil {
			msg = err.Error()
			err = nil
		}
		assert.Equal(t, `Test NodeContract NodeContract testing `+rnd, msg)
	}
}

func TestCron(t *testing.T) {
	if err := keyLogin(1); err != nil {
		t.Error(err)
		return
	}

	err := postTx("NewCron", &url.Values{
		"Cron":       {"60 * * * *"},
		"Contract":   {"TestCron"},
		"Conditions": {`ContractConditions("MainCondition")`},
		"vde":        {"true"},
	})
	if err.Error() != `500 {"error": "E_SERVER", "msg": "{\"type\":\"panic\",\"error\":\"End of range (60) above maximum (59): 60\"}" }` {
		t.Error(err)
	}

	postTx("NewContract", &url.Values{
		"Value": {`
			contract TestCron {
				data {}
				action {
					return "Success"
				}
			}
		`},
		"Conditions": {`ContractConditions("MainCondition")`},
		"vde":        {"true"},
	})

	till := time.Now().Format(time.RFC3339)
	err = postTx("NewCron", &url.Values{
		"Cron":       {"* * * * *"},
		"Contract":   {"TestCron"},
		"Conditions": {`ContractConditions("MainCondition")`},
		"Till":       {till},
		"vde":        {"true"},
	})
	if err != nil {
		t.Error(err)
	}

	err = postTx("EditCron", &url.Values{
		"Id":         {"1"},
		"Cron":       {"*/3 * * * *"},
		"Contract":   {"TestCron"},
		"Conditions": {`ContractConditions("MainCondition")`},
		"Till":       {till},
		"vde":        {"true"},
	})
	if err != nil {
		t.Error(err)
	}
}
