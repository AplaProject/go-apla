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
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/EGaaS/go-egaas-mvp/packages/converter"
	"github.com/EGaaS/go-egaas-mvp/packages/crypto"
	"github.com/EGaaS/go-egaas-mvp/packages/script"
	"github.com/EGaaS/go-egaas-mvp/packages/smart"
	"github.com/EGaaS/go-egaas-mvp/packages/utils/sql"
	"github.com/EGaaS/go-egaas-mvp/packages/utils/tx"
	"gopkg.in/vmihailenco/msgpack.v2"
)

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
}

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

// PrepareTxJSON is a structure for the answer of ajax_prepare_tx ajax request
type PrepareTxJSON struct {
	ForSign string            `json:"forsign"`
	Signs   []TxSignJSON      `json:"signs"`
	Values  map[string]string `json:"values"`
	Time    string            `json:"time"`
	//	Error   string            `json:"error"`
}

// EncryptKey is a structure for the answer of ajax_encrypt_key ajax request
type EncryptKey struct {
	Encrypted string `json:"encrypted"` //hex
	Public    string `json:"public"`    //hex
	WalletID  int64  `json:"wallet_id"`
	Address   string `json:"address"`
	Error     string `json:"error"`
}

func getSmartContract(w http.ResponseWriter, r *http.Request, data *apiData) error {

	cntname := data.params[`name`].(string)
	contract := smart.GetContract(cntname, uint32(data.sess.Get(`state`).(int64)))
	if contract == nil {
		return errorAPI(w, fmt.Sprintf(`there is not %s contract`, cntname), http.StatusBadRequest)
	}
	info := (*contract).Block.Info.(*script.ContractInfo)
	fields := make([]smartField, 0)
	result := smartFieldsResult{Name: info.Name, Active: info.Active}

	if info.Tx != nil {
		for _, fitem := range *info.Tx {
			field := smartField{Name: fitem.Name, Type: fitem.Type.String(), Tags: fitem.Tags}

			if strings.Contains(fitem.Tags, `hidden`) || strings.Contains(fitem.Tags, `signature`) {
				field.HTML = `hidden`
			} else {
				for _, tag := range []string{`date`, `polymap`, `map`, `image`, `text`, `address`} {
					if strings.Contains(fitem.Tags, tag) {
						field.HTML = tag
						break
					}
				}
				if len(field.HTML) == 0 {
					if fitem.Type.String() == script.Decimal {
						field.HTML = `money`
					} else if fitem.Type.String() == `string` || fitem.Type.String() == `int64` || fitem.Type.String() == `float64` {
						field.HTML = `textinput`
					}
				}
			}
			fields = append(fields, field)
		}
	}
	result.Fields = fields
	data.result = result
	return nil
}

func validateSmartContract(r *http.Request, data *apiData, result *PrepareTxJSON) (contract *smart.Contract, err error) {
	cntname := data.params[`name`].(string)
	state := data.sess.Get(`state`).(int64)
	contract = smart.GetContract(cntname, uint32(state))
	if contract == nil {
		return nil, fmt.Errorf(`there is not %s contract`, cntname)
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
					value, err = sql.DB.Single(fmt.Sprintf(`select value from "%s_signatures" where name=?`, pref), ret[1]).String()
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
					sign.ForSign = fmt.Sprintf(`%d,%d`, (*result).Time, uint64(data.sess.Get(`wallet`).(int64)))
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
	pubKey, err := sql.DB.Single(`select public_key_0 from dlt_wallets where wallet_id=?`, id).String()
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

		exist, err := sql.DB.Single(`select wallet_id from dlt_wallets where wallet_id=?`, idnew).Int64()
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

func txPreSmartContract(w http.ResponseWriter, r *http.Request, data *apiData) error {
	var (
		result                   PrepareTxJSON
		stateID, userID, timeNow int64
		smartTx                  tx.SmartContract
	)

	timeNow = time.Now().Unix()
	result.Time = converter.Int64ToStr(timeNow)
	result.Values = make(map[string]string)
	contract, err := validateSmartContract(r, data, &result)
	if err != nil {
		return errorAPI(w, err.Error(), http.StatusBadRequest)
	}
	info := (*contract).Block.Info.(*script.ContractInfo)
	userID = data.sess.Get(`wallet`).(int64)
	if data.sess.Get(`state`) != nil {
		stateID = data.sess.Get(`state`).(int64)
	}
	smartTx.Header = tx.Header{Type: int(info.ID), Time: timeNow, UserID: userID, StateID: stateID}
	forsign := smartTx.ForSign()
	if info.Tx != nil {
		for _, fitem := range *info.Tx {
			if strings.Contains(fitem.Tags, `image`) || strings.Contains(fitem.Tags, `signature`) {
				continue
			}
			var val string
			if strings.Contains(fitem.Tags, `crypt`) {
				var wallet string
				if ret := regexp.MustCompile(`(?is)crypt:([\w_\d]+)`).FindStringSubmatch(fitem.Tags); len(ret) == 2 {
					wallet = r.FormValue(ret[1])
				} else {
					wallet = converter.Int64ToStr(userID)
				}
				key := EncryptNewKey(wallet)
				if len(key.Error) != 0 {
					return errorAPI(w, key.Error, http.StatusConflict)
				}
				result.Values[fitem.Name] = key.Encrypted
				val = key.Encrypted
			} else if fitem.Type.String() == `[]interface {}` {
				for key, values := range r.Form {
					if key == fitem.Name+`[]` {
						var list []string
						for _, value := range values {
							list = append(list, value)
						}
						val = strings.Join(list, `,`)
					}
				}
			} else {
				val = strings.TrimSpace(r.FormValue(fitem.Name))
				if strings.Contains(fitem.Tags, `address`) {
					val = converter.Int64ToStr(converter.StringToAddress(val))
				} else if fitem.Type.String() == script.Decimal {
					val = strings.TrimLeft(val, `0`)
				}
			}
			forsign += fmt.Sprintf(",%v", val)
		}
	}
	result.ForSign = forsign
	data.result = result
	return nil
}

func txSmartContract(w http.ResponseWriter, r *http.Request, data *apiData) error {
	var (
		stateID, userID           int64
		isPublic, hash, publicKey []byte
		toSerialize               interface{}
	)
	contract, err := validateSmartContract(r, data, nil)
	if err != nil {
		return errorAPI(w, err.Error(), http.StatusBadRequest)
	}
	info := (*contract).Block.Info.(*script.ContractInfo)
	if data.sess.Get(`state`) != nil {
		stateID = data.sess.Get(`state`).(int64)
	}
	userID = data.sess.Get(`wallet`).(int64)

	isPublic, err = sql.DB.Single(`select public_key_0 from dlt_wallets where wallet_id=?`, userID).Bytes()
	if err != nil {
		return errorAPI(w, err.Error(), http.StatusConflict)
	}
	if len(isPublic) == 0 {
		if _, ok := data.params[`pubkey`]; ok && len(data.params[`pubkey`].([]byte)) > 0 {
			publicKey = data.params[`pubkey`].([]byte)
			lenpub := len(publicKey)
			if lenpub > 64 {
				publicKey = publicKey[lenpub-64:]
			}
		}
		if len(publicKey) == 0 {
			return errorAPI(w, `empty public key`, http.StatusConflict)
		}
	} else {
		publicKey = []byte("null")
	}
	signature := data.params[`signature`].([]byte)
	if len(signature) == 0 {
		return errorAPI(w, `signature is empty`, http.StatusBadRequest)
	}
	idata := make([]byte, 0)
	if info.Tx != nil {
	fields:
		for _, fitem := range *info.Tx {
			val := strings.TrimSpace(r.FormValue(fitem.Name))
			if strings.Contains(fitem.Tags, `address`) {
				val = converter.Int64ToStr(converter.StringToAddress(val))
			}
			switch fitem.Type.String() {
			case `[]interface {}`:
				var list []string
				for key, values := range r.Form {
					if key == fitem.Name+`[]` {
						for _, value := range values {
							list = append(list, value)
						}
					}
				}
				idata = append(idata, converter.EncodeLength(int64(len(list)))...)
				for _, ilist := range list {
					blist := []byte(ilist)
					idata = append(append(idata, converter.EncodeLength(int64(len(blist)))...), blist...)
				}
			case `uint64`:
				converter.BinMarshal(&idata, converter.StrToUint64(val))
			case `int64`:
				converter.EncodeLenInt64(&idata, converter.StrToInt64(val))
			case `float64`:
				converter.BinMarshal(&idata, converter.StrToFloat64(val))
			case `string`, script.Decimal:
				idata = append(append(idata, converter.EncodeLength(int64(len(val)))...), []byte(val)...)
			case `[]uint8`:
				var bytes []byte
				bytes, err = hex.DecodeString(val)
				if err != nil {
					break fields
				}
				idata = append(append(idata, converter.EncodeLength(int64(len(bytes)))...), bytes...)
			}
		}
	}
	toSerialize = tx.SmartContract{
		Header: tx.Header{Type: int(info.ID), Time: converter.StrToInt64(data.params[`time`].(string)),
			UserID: userID, StateID: stateID, PublicKey: publicKey,
			BinSignatures: converter.EncodeLengthPlusData(signature)},
		Data: idata,
	}
	serializedData, err := msgpack.Marshal(toSerialize)
	if err != nil {
		return errorAPI(w, err.Error(), http.StatusConflict)
	}
	if hash, err = sql.DB.SendTx(int64(info.ID), userID,
		append([]byte{128}, serializedData...)); err != nil {
		return errorAPI(w, err.Error(), http.StatusConflict)
	}
	data.result = &hashTx{Hash: string(hash)}
	return nil
}
