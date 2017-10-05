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

package apiv2

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"strings"

	"github.com/EGaaS/go-egaas-mvp/packages/converter"
	"github.com/EGaaS/go-egaas-mvp/packages/crypto"
	"github.com/EGaaS/go-egaas-mvp/packages/model"
	"github.com/EGaaS/go-egaas-mvp/packages/script"
	"github.com/EGaaS/go-egaas-mvp/packages/smart"
)

/*
type smartField struct {
	Name string `json:"name"`
	HTML string `json:"htmltype"`
	Type string `json:"txtype"`
	Tags string `json:"tags"`
}

type smartFieldsResult struct {
	Fields []smartField `json:"fields"`
	Name   string       `json:"name"`
	Active bool         `json:"active"`
}*/

//SignRes contains the data of the signature
type SignRes struct {
	Param string `json:"name"`
	Text  string `json:"text"`
}

// TxSignJSON is a structure for additional signs of transaction
type TxSignJSON struct {
	ForSign string    `json:"forsign"`
	Field   string    `json:"field"`
	Title   string    `json:"title"`
	Params  []SignRes `json:"params"`
}

// EncryptKey is a structure for the answer of ajax_encrypt_key ajax request
type EncryptKey struct {
	Encrypted string `json:"encrypted"` //hex
	Public    string `json:"public"`    //hex
	WalletID  int64  `json:"wallet_id"`
	Address   string `json:"address"`
	Error     string `json:"error"`
}

func validateSmartContract(r *http.Request, data *apiData, result *prepareResult) (contract *smart.Contract, parerr interface{}, err error) {
	cntname := data.params[`name`].(string)
	contract = smart.GetContract(cntname, uint32(data.state))
	if contract == nil {
		return nil, cntname, fmt.Errorf(`E_CONTRACT`) //fmt.Errorf(`there is not %s contract`, cntname)
	}

	if contract.Block.Info.(*script.ContractInfo).Tx != nil {
		for _, fitem := range *(*contract).Block.Info.(*script.ContractInfo).Tx {
			if strings.Contains(fitem.Tags, `image`) || strings.Contains(fitem.Tags, `crypt`) {
				continue
			}
			if strings.Contains(fitem.Tags, `signature`) && result != nil {
				if ret := regexp.MustCompile(`(?is)signature:([\w_\d]+)`).FindStringSubmatch(fitem.Tags); len(ret) == 2 {
					pref := getPrefix(data)
					var value string
					value, err = model.Single(fmt.Sprintf(`select value from "%s_signatures" where name=?`, pref), ret[1]).String()
					if err != nil {
						break
					}
					if len(value) == 0 {
						err = fmt.Errorf(`%s is unknown signature`, ret[1])
						break
					}
					var sign TxSignJSON
					err = json.Unmarshal([]byte(value), &sign)
					if err != nil {
						break
					}
					sign.ForSign = fmt.Sprintf(`%s,%d`, (*result).Time, uint64(data.wallet))
					for _, isign := range sign.Params {
						sign.ForSign += fmt.Sprintf(`,%v`, strings.TrimSpace(r.FormValue(isign.Param)))
					}
					sign.Field = fitem.Name
					(*result).Signs = append((*result).Signs, sign)
				}
			} else {
				var val string

				val = strings.TrimSpace(r.FormValue(fitem.Name))
				if len(val) == 0 && !strings.Contains(fitem.Tags, `optional`) {
					err = fmt.Errorf(`%s is empty`, fitem.Name)
					break
				}
				if strings.Contains(fitem.Tags, `address`) {
					addr := converter.StringToAddress(val)
					if addr == 0 {
						err = fmt.Errorf(`Address %s is not valid`, val)
						break
					}
				}
				if fitem.Type.String() == script.Decimal {
					re := regexp.MustCompile(`^\d+$`) //`^\d+\.?\d+?$`
					if !re.Match([]byte(val)) {
						err = fmt.Errorf(`The value of money %s is not valid`, val)
						break
					}
				}
			}
		}
	}
	return
}

// EncryptNewKey creates a shared key, generates and crypts a new private key
func EncryptNewKey(walletID string) (result EncryptKey) {
	var (
		err error
		id  int64
	)

	if len(walletID) == 0 {
		result.Error = `unknown wallet id`
		return result
	}
	id = converter.StringToAddress(walletID)
	pubKey, err := model.Single(`select public_key_0 from dlt_wallets where wallet_id=?`, id).String()
	if err != nil {
		result.Error = err.Error()
		return result
	}
	if len(pubKey) == 0 {
		result.Error = `unknown wallet id`
		return result
	}
	var private string

	for result.WalletID == 0 {
		private, result.Public, _ = crypto.GenHexKeys()

		pub, _ := hex.DecodeString(result.Public)
		idnew := crypto.Address(pub)

		exist, err := model.Single(`select wallet_id from dlt_wallets where wallet_id=?`, idnew).Int64()
		if err != nil {
			result.Error = err.Error()
			return result
		}
		if exist == 0 {
			result.WalletID = idnew
		}
	}
	priv, _ := hex.DecodeString(private)
	encrypted, err := crypto.SharedEncrypt([]byte(pubKey), priv)
	if err != nil {
		result.Error = err.Error()
		return result
	}
	result.Encrypted = hex.EncodeToString(encrypted)
	result.Address = converter.AddressToString(result.WalletID)

	return
}
