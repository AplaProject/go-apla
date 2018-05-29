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

package main

import (
	"fmt"
	"net/url"
	"time"

	"github.com/GenesisKernel/go-genesis/packages/converter"
	"github.com/GenesisKernel/go-genesis/packages/crypto"
)

var (
	Nodes      = `/home/losaped/go/src/github.com/GenesisKernel/go-genesis/`
	PathNodes  = Nodes + `node%d-data/`
	apiAddress = "http://localhost:7079"
	port       = []int{7079, 7081, 7083}
)

func main() {
	var (
		err error
	)
	if err = KeyLogin(1, 1); err != nil {
		fmt.Println(`Login`, err)
		return
	}

	if err = updateFullNodes(); err != nil {
		fmt.Println(err)
		return
	}

	for i := 0; i < 10; i++ {
		fmt.Println(`Step`, i)
		node := 0 //Random(0, 2)
		rnd := crypto.RandSeq(5)
		apiAddress = fmt.Sprintf("http://localhost:%d", port[node])
		if node == 0 && i&1 == 0 {
			if err = CreateEcosystem(rnd); err != nil {
				fmt.Println(err)
				break
			}
		}
		for j := 0; j < 90; j++ {
			fmt.Print(j)
			postTx("NewLang", &url.Values{
				"Name":          {rnd + converter.IntToStr(j)},
				"Trans":         {`{"en": "My test", "ru": "Русский текст", "en-US": "US locale"}`},
				"ApplicationId": {"1"},
			})
			form := url.Values{"ApplicationId": {`1`}, "Name": {rnd + converter.IntToStr(j)}, "Value": {`P(paragraph)`},
				"Menu": {`default_menu`}, "Conditions": {"ContractConditions(`MainCondition`)"},
				`nowait`: {`true`}}
			if err = postTx(`@1NewPage`, &form); err != nil {
				fmt.Println(err)
				break
			}
			//time.Sleep(time.Duration(Random(1, 50)*10) * time.Millisecond)
			form = url.Values{"ApplicationId": {`1`}, "Id": {`1`}, "Value": {`Div(paragraph)`},
				"Menu": {`default_menu`}, "Conditions": {"ContractConditions(`MainCondition`)"},
				`nowait`: {`true`}}
			if err = postTx(`@1EditPage`, &form); err != nil {
				fmt.Println(err)
				break
			}
		}
		time.Sleep(time.Duration(Random(1, 50)*20) * time.Millisecond)
		fmt.Println(`upd`)
		for j := 1; j < 50; j++ {
			fmt.Print(j)
			postTx("EditLang", &url.Values{
				"Id":            {converter.IntToStr(j)},
				"Trans":         {`{"en": "My test", "ru": "Русский текст новый", "en-US": "US locale"}`},
				"ApplicationId": {"1"},
			})
		}
		time.Sleep(time.Duration(Random(1, 50)*50) * time.Millisecond)
		for i := 0; i < 3; i++ {
			apiAddress = fmt.Sprintf("http://localhost:%d", port[i])
			var ret checkResult
			if err = sendGet(`check`, nil, &ret); err != nil {
				fmt.Println(err)
				break
			}
			fmt.Println(`Counts`, ret)
		}
	}
	time.Sleep(time.Duration(10000 * time.Millisecond))
	fmt.Println(`=======`)
	for k := 0; k < 200; k++ {
		fmt.Println(`---------- {rollback_tx, ecosystems, blocks}`)
		for i := 0; i < 3; i++ {
			apiAddress = fmt.Sprintf("http://localhost:%d", port[i])
			var ret checkResult
			if err = sendGet(`check`, nil, &ret); err != nil {
				fmt.Println(err)
				break
			}
			fmt.Println(`node`, i, `counts`, ret)
		}
		time.Sleep(time.Duration(3000 * time.Millisecond))
	}
	fmt.Println(`OK`)
}

func CreateEcosystem(name string) error {
	var (
		result string
		err    error
	)
	form := url.Values{`Name`: {name}}
	if _, result, err = postTxResult(`NewEcosystem`, &form); err != nil {
		return err
	}
	fmt.Println(result)
	return nil
}

func updateFullNodes() error {
	byteNodes := `[{"tcp_address":"127.0.0.1:7078", "api_address":"https://127.0.0.1:7079", "key_id":"-4466900793776865315", "public_key":"ca901a97e84d76f8d46e2053028f709074b3e60d3e2e33495840586567a0c961820d789592666b67b05c6ae120d5bd83d4388b2f1218638d8226d40ced0bb208"},`
	byteNodes += `{"tcp_address":"127.0.0.1:7080", "api_address":"https://127.0.0.1:7081", "key_id":"542353610328569127", "public_key":"a8ada71764fd2f0c9fa1d2986455288f11f0f3931492d27dc62862fdff9c97c38923ef46679488ad1cd525342d4d974621db58f809be6f8d1c19fdab50abc06b"},`
	byteNodes += `{"tcp_address":"127.0.0.1:7082", "api_address":"https://127.0.0.1:7083", "key_id":"5972241339967729614", "public_key":"de1b74d36ae39422f2478cba591f4d14eb017306f6ffdc3b577cc52ee50edb8fe7c7b2eb191a24c8ddfc567cef32152bab17de698ed7b3f2ab75f3bcc8b9b372"}]`

	form := &url.Values{
		"Name":  {"full_nodes"},
		"Value": {string(byteNodes)},
	}

	return postTx(`UpdateSysParam`, form)
}

func addKeys() error {
	form := url.Values{
		`Value`: {`contract InsertNodeKey {
			data {
				KeyID string
				PubKey string
			}
			conditions {}
			action {
				DBInsert("keys", "id,pub,amount", $KeyID, $PubKey, "100000000000000000000")
			}
		}`},
		`ApplicationId`: {`1`},
		`Conditions`:    {`true`},
	}

	if err := postTx(`NewContract`, &form); err != nil {
		return err
	}

	nodes := []url.Values{
		url.Values{
			`KeyID`:  {"542353610328569127"},
			`PubKey`: {"a8ada71764fd2f0c9fa1d2986455288f11f0f3931492d27dc62862fdff9c97c38923ef46679488ad1cd525342d4d974621db58f809be6f8d1c19fdab50abc06b"},
		},
		url.Values{
			`KeyID`:  {"5972241339967729614"},
			`PubKey`: {"de1b74d36ae39422f2478cba591f4d14eb017306f6ffdc3b577cc52ee50edb8fe7c7b2eb191a24c8ddfc567cef32152bab17de698ed7b3f2ab75f3bcc8b9b372"},
		},
	}

	for _, form := range nodes {
		if err := postTx(`InsertNodeKey`, &form); err != nil {
			return err
		}
	}

	return nil
}
