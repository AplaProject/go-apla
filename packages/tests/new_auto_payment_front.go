package main

import (
	"fmt"
	"github.com/DayLightProject/go-daylight/packages/utils"
	"github.com/DayLightProject/go-daylight/packages/tests_utils"
)

type vComplex struct {
	Currency map[string][]float64 `json:"currency"`
	Referral map[string]string `json:"referral"`
	Admin int64 `json:"admin"`
}
func main() {

	f:=tests_utils.InitLog()
	defer f.Close()

	txType := "NewAutoPayment";
	txTime := "1499278817";
	userId := []byte("2")
	recipient := []byte("2")
	amount := []byte("2")
	commission := []byte("2")
	currency_id := []byte("2")
	period := []byte("2")
	comment := []byte("2")
	var blockId int64 = 128008

	var txSlice [][]byte
	// hash
	txSlice = append(txSlice, []byte("22cb812e53e22ee539af4a1d39b4596d"))
	// type
	txSlice = append(txSlice,  utils.Int64ToByte(utils.TypeInt(txType)))
	// time
	txSlice = append(txSlice, []byte(txTime))
	// user_id
	txSlice = append(txSlice, userId)
	// recipient
	txSlice = append(txSlice, recipient)
	// amount
	txSlice = append(txSlice, amount)
	// commission
	txSlice = append(txSlice, commission)
	// currency_id
	txSlice = append(txSlice, currency_id)
	// period
	txSlice = append(txSlice, period)
	// comment
	txSlice = append(txSlice, comment)

	dataForSign := fmt.Sprintf("%s,%s,%s,%s,%s,%s,%s,%s,%s", utils.TypeInt(txType), txTime, userId, recipient, amount, commission, currency_id, period, comment)

	err := tests_utils.MakeFrontTest(txSlice, utils.StrToInt64(txTime), dataForSign, txType, utils.BytesToInt64(userId), "", blockId)
	if err != nil {
		fmt.Println(err)
	}
}
