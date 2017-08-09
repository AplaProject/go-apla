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

package exchangeapi

import (
	"encoding/hex"
	"fmt"
	"net/http"
	"regexp"
	"time"

	"github.com/EGaaS/go-egaas-mvp/packages/config/syspar"
	"github.com/EGaaS/go-egaas-mvp/packages/converter"
	"github.com/EGaaS/go-egaas-mvp/packages/crypto"
	"github.com/EGaaS/go-egaas-mvp/packages/model"
	"github.com/EGaaS/go-egaas-mvp/packages/utils"
	"github.com/boltdb/bolt"
	"github.com/shopspring/decimal"
)

// Send is a answer structure for send handle function
type Send struct {
	Error string `json:"error"`
}

func send(r *http.Request) interface{} {
	var (
		result Send
		priv   []byte
	)

	sender := converter.StringToAddress(r.FormValue(`sender`))
	if sender == 0 {
		result.Error = `Sender is invalid`
		return result
	}
	recipient := converter.StringToAddress(r.FormValue(`recipient`))
	if recipient == 0 {
		result.Error = `Recipient is invalid`
		return result
	}
	money := r.FormValue(`amount`)
	re := regexp.MustCompile(`^\d+$`)
	if !re.Match([]byte(money)) {
		result.Error = fmt.Sprintf(`The value of money %s is not valid`, money)
		return result
	}
	amount, err := decimal.NewFromString(money)
	if err != nil {
		result.Error = err.Error()
		return result
	}

	err = boltDB.View(func(tx *bolt.Tx) error {
		var err error
		encpriv := tx.Bucket(bucket).Get([]byte(converter.Int64ToStr(sender)))
		if len(encpriv) == 0 {
			return fmt.Errorf(`Sender has not been found`)
		}
		priv, err = decryptBytes(encpriv)
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		result.Error = err.Error()
		return result
	}

	fPrice := sql.SysCost(`dlt_transfer`)
	systemParam := &model.SystemParameter{}
	err = systemParam.Get("fuel_rate")
	if err != nil {
		log.Fatal(err)
	}
	fuelRate := decimal.NewFromString(systemParam.Value)
	if fuelRate.Cmp(decimal.New(0, 0)) <= 0 {
		result.Error = `fuel rate must be greater than 0`
		return result
	}
	fPriceDecimal := decimal.New(fPrice, 0)
	commission := fPriceDecimal.Mul(fuelRate)

	total, err := model.Single(`SELECT amount FROM dlt_wallets WHERE wallet_id = ?`, sender).String()
	if err != nil {
		result.Error = err.Error()
		return result
	}
	totalAmount, err := decimal.NewFromString(total)
	if err != nil {
		result.Error = err.Error()
		return result
	}
	if totalAmount.Cmp(amount.Add(commission)) < 0 {
		result.Error = fmt.Sprintf(`There is not enough money. %v is less than %v`, totalAmount, amount.Add(commission))
		return result
	}
	wallet := converter.AddressToString(recipient)
	txType := utils.TypeInt(`DLTTransfer`)
	txTime := time.Now().Unix()
	forSign := fmt.Sprintf("%d,%d,%d,%s,%s,%s,%s", txType, txTime, sender,
		wallet, amount.String(), commission.String(), `api`)
	signature, err := crypto.Sign(hex.EncodeToString(priv), forSign)
	if err != nil {
		result.Error = err.Error()
		return result
	}
	sign := make([]byte, 0)
	sign = append(sign, converter.EncodeLengthPlusData(signature)...)
	binsign := converter.EncodeLengthPlusData(sign)

	data := make([]byte, 0)
	data = converter.DecToBin(txType, 1)
	data = append(data, converter.DecToBin(txTime, 4)...)
	data = append(data, converter.EncodeLengthPlusData(sender)...)
	data = append(data, converter.EncodeLengthPlusData(0)...)
	data = append(data, converter.EncodeLengthPlusData([]byte(wallet))...)
	data = append(data, converter.EncodeLengthPlusData([]byte(amount.String()))...)
	data = append(data, converter.EncodeLengthPlusData([]byte(commission.String()))...)
	data = append(data, converter.EncodeLengthPlusData([]byte(`api`))...)
	pub, err := crypto.PrivateToPublic(priv)
	if err != nil {
		log.Fatal(err)
	}
	data = append(data, converter.EncodeLengthPlusData(pub)...)
	data = append(data, binsign...)
	_, err = model.SendTx(txType, sender, data)
	if err != nil {
		result.Error = err.Error()
		return result
	}
	return result
}
