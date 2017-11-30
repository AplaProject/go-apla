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
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"strings"

	"github.com/AplaProject/go-apla/packages/consts"
	"github.com/AplaProject/go-apla/packages/converter"
	"github.com/AplaProject/go-apla/packages/model"
	"github.com/AplaProject/go-apla/packages/script"
	"github.com/AplaProject/go-apla/packages/smart"

	log "github.com/sirupsen/logrus"
)

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
	contract = smart.VMGetContract(data.vm, cntname, uint32(data.ecosystemId))
	if contract == nil {
		return nil, cntname, fmt.Errorf(`E_CONTRACT`)
	}

	if contract.Block.Info.(*script.ContractInfo).Tx != nil {
		for _, fitem := range *(*contract).Block.Info.(*script.ContractInfo).Tx {
			if strings.Contains(fitem.Tags, `image`) || strings.Contains(fitem.Tags, `crypt`) {
				continue
			}
			if strings.Contains(fitem.Tags, `signature`) && result != nil {
				if ret := regexp.MustCompile(`(?is)signature:([\w_\d]+)`).FindStringSubmatch(fitem.Tags); len(ret) == 2 {
					pref := getPrefix(data)
					signature := &model.Signature{}
					signature.SetTablePrefix(pref)
					found, err := signature.Get(ret[1])
					if err != nil {
						log.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("selecting signature by name")
						break
					}
					if !found {
						log.WithFields(log.Fields{"type": consts.NotFound, "signature": ret[1]}).Error("unknown signature")
						return contract, ret[1], fmt.Errorf(errors["E_UNKNOWNSIGN"])
					}
					var sign TxSignJSON
					err = json.Unmarshal([]byte(signature.Value), &sign)
					if err != nil {
						log.WithFields(log.Fields{"type": consts.JSONUnmarshallError, "error": err}).Error("unmarshalling sign from json")
						break
					}
					sign.ForSign = fmt.Sprintf(`%s,%d`, (*result).Time, uint64(data.keyId))
					for _, isign := range sign.Params {
						sign.ForSign += fmt.Sprintf(`,%v`, strings.TrimSpace(r.FormValue(isign.Param)))
					}
					sign.Field = fitem.Name
					(*result).Signs = append((*result).Signs, sign)
				}
			} else {
				var val string

				val = strings.TrimSpace(r.FormValue(fitem.Name))
				if len(val) == 0 && !strings.Contains(fitem.Tags, `optional`) &&
					!strings.Contains(fitem.Tags, `signature`) {
					log.WithFields(log.Fields{"type": consts.EmptyObject, "item_name": fitem.Name}).Error("route item is empty")
					err = fmt.Errorf(`%s is empty`, fitem.Name)
					break
				}
				if strings.Contains(fitem.Tags, `address`) {
					addr := converter.StringToAddress(val)
					if addr == 0 {
						log.WithFields(log.Fields{"type": consts.ConversionError, "value": val}).Error("converting string to address")
						err = fmt.Errorf(`Address %s is not valid`, val)
						break
					}
				}
				if fitem.Type.String() == script.Decimal {
					re := regexp.MustCompile(`^\d+$`)
					if !re.Match([]byte(val)) {
						log.WithFields(log.Fields{"type": consts.InvalidObject, "value": val}).Error("The value of money is not valid")
						err = fmt.Errorf(`The value of money %s is not valid`, val)
						break
					}
				}
			}
		}
	}
	return
}
