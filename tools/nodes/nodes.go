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
	Nodes      = `/home/ak/two/data/`
	PathNodes  = Nodes + `node%d/`
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
	for i := 0; i < 1; i++ {
		fmt.Println(`Step`, i)
		node := 0 //Random(0, 2)
		rnd := crypto.RandSeq(5)
		apiAddress = fmt.Sprintf("http://localhost:%d", port[node])
		/*		if node == 0 && i&1 == 0 {
				if err = CreateEcosystem(rnd); err != nil {
					fmt.Println(err)
					break
				}
			}*/
		for j := 0; j < 90; j++ {
			fmt.Print(j)
			postTx("NewLang", &url.Values{
				"Name":          {rnd + converter.IntToStr(j)},
				"Trans":         {`{"en": "My test", "ru": "Русский текст", "en-US": "US locale"}`},
				"ApplicationId": {"1"},
			})
			/*			form := url.Values{"ApplicationId": {`1`}, "Name": {rnd + converter.IntToStr(j)}, "Value": {`P(paragraph)`},
							"Menu": {`default_menu`}, "Conditions": {"ContractConditions(`MainCondition`)"},
							`nowait`: {`true`}}
						if err = postTx(`@1NewPage`, &form); err != nil {
							fmt.Println(err)
							break
						}*/
			//			time.Sleep(time.Duration(Random(1, 50)*10) * time.Millisecond)
			/*			form = url.Values{"ApplicationId": {`1`}, "Id": {converter.IntToStr(j + 1)}, "Value": {`Div(paragraph)`},
							"Menu": {`default_menu`}, "Conditions": {"ContractConditions(`MainCondition`)"},
							`nowait`: {`true`}}
						if err = postTx(`@1EditPage`, &form); err != nil {
							fmt.Println(err)
							break
						}*/
		}
		/*		rets := make([]checkResult, 5)
				time.Sleep(time.Duration(10000 * time.Millisecond))
				for i := 0; i < 3; i++ {
					apiAddress = fmt.Sprintf("http://localhost:%d", port[i])
					if err = sendGet(`check`, nil, &rets[i]); err != nil {
						fmt.Println(err)
						break
					}
					fmt.Println(`Counts`, rets[i])
					if i > 0 && rets[i] != rets[i-1] {
						fmt.Println(`Problem!`)
						return
					}
				}*/
		/*		time.Sleep(time.Duration(Random(1, 50)*20) * time.Millisecond)
				fmt.Println(`upd`)
				for j := 1; j < 50; j++ {
					fmt.Print(j)
					postTx("EditLang", &url.Values{
						"Id":            {converter.IntToStr(j)},
						"Trans":         {`{"en": "My test", "ru": "Русский текст новый", "en-US": "US locale"}`},
						"ApplicationId": {"1"},
					})
				}*/
		//		time.Sleep(time.Duration(Random(1, 50)*50) * time.Millisecond)
	}
	time.Sleep(time.Duration(8000 * time.Millisecond))
	fmt.Println(`=======`)
	for i := 0; i < 3; i++ {
		apiAddress = fmt.Sprintf("http://localhost:%d", port[i])
		var ret checkResult
		if err = sendGet(`check`, nil, &ret); err != nil {
			fmt.Println(err)
			break
		}
		fmt.Println(`Counts`, ret)
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
