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

	"github.com/EGaaS/go-egaas-mvp/packages/lib"
	"github.com/EGaaS/go-egaas-mvp/packages/utils"
	"github.com/EGaaS/go-egaas-mvp/packages/utils/sql"
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

	sender := lib.StringToAddress(r.FormValue(`sender`))
	if sender == 0 {
		result.Error = `Sender is invalid`
		return result
	}
	recipient := lib.StringToAddress(r.FormValue(`recipient`))
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
		encpriv := tx.Bucket(bucket).Get([]byte(utils.Int64ToStr(sender)))
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

	fPrice, err := sql.DB.Single(`SELECT value->'dlt_transfer' FROM system_parameters WHERE name = ?`, "op_price").String()
	if err != nil {
		result.Error = err.Error()
		return result
	}
	fuelRate := sql.DB.GetFuel()
	if fuelRate.Cmp(decimal.New(0, 0)) <= 0 {
		result.Error = `fuel rate must be greater than 0`
		return result
	}
	fPriceDecemal, err := decimal.NewFromString(fPrice)
	if err != nil {
		result.Error = err.Error()
		return result
	}
	commission := fPriceDecemal.Mul(fuelRate)

	total, err := sql.DB.Single(`SELECT amount FROM dlt_wallets WHERE wallet_id = ?`, sender).String()
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
	wallet := lib.AddressToString(recipient)
	txType := utils.TypeInt(`DLTTransfer`)
	txTime := time.Now().Unix()
	forSign := fmt.Sprintf("%d,%d,%d,%s,%s,%s,%s", txType, txTime, sender,
		wallet, amount.String(), commission.String(), `api`)
	signature, err := lib.SignECDSA(hex.EncodeToString(priv), forSign)
	if err != nil {
		result.Error = err.Error()
		return result
	}
	sign := make([]byte, 0)
	sign = append(sign, utils.EncodeLengthPlusData(signature)...)
	binsign := utils.EncodeLengthPlusData(sign)

	data := make([]byte, 0)
	data = utils.DecToBin(txType, 1)
	data = append(data, utils.DecToBin(txTime, 4)...)
	data = append(data, utils.EncodeLengthPlusData(sender)...)
	data = append(data, utils.EncodeLengthPlusData(0)...)
	data = append(data, utils.EncodeLengthPlusData([]byte(wallet))...)
	data = append(data, utils.EncodeLengthPlusData([]byte(amount.String()))...)
	data = append(data, utils.EncodeLengthPlusData([]byte(commission.String()))...)
	data = append(data, utils.EncodeLengthPlusData([]byte(`api`))...)
	data = append(data, utils.EncodeLengthPlusData(lib.PrivateToPublic(priv))...)
	data = append(data, binsign...)
	err = sql.DB.SendTx(txType, sender, data)
	if err != nil {
		result.Error = err.Error()
		return result
	}
	return result
}
