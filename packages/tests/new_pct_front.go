package main

import (
	"fmt"
	"github.com/DayLightProject/go-daylight/packages/utils"
	"tests_utils"
)


func main() {

	f:=tests_utils.InitLog()
	defer f.Close()

	txType := "NewPct";
	txTime := "1599278817";
	userId := []byte("2")
	var blockId int64 = 140015
	//data:=`{"currency":{"1":{"miner_pct":"0.0000000617044","user_pct":"0.0000000439591"},"72":{"miner_pct":"0.0000000617044","user_pct":"0.0000000439591"}},"referral":{"first":10,"second":0,"third":0}}`
	data := `{"currency":{"1":{"miner_pct":"0.0000000617044","user_pct":"0.0000000435602"},"72":{"miner_pct":"0.0000000760368","user_pct":"0.0000000562834"}},"referral":{"first":30,"second":20,"third":5}}`

	var txSlice [][]byte
	// hash
	txSlice = append(txSlice, []byte("22cb812e53e22ee539af4a1d39b4596d"))
	// type
	txSlice = append(txSlice,  utils.Int64ToByte(utils.TypeInt(txType)))
	// time
	txSlice = append(txSlice, []byte(txTime))
	// user_id
	txSlice = append(txSlice, userId)
	// promised_amount_id
	txSlice = append(txSlice, []byte(data))

	dataForSign := fmt.Sprintf("%v,%v,%s,%s", utils.TypeInt(txType), txTime, userId, data)

	err := tests_utils.MakeFrontTest(txSlice, utils.StrToInt64(txTime), dataForSign, txType, utils.BytesToInt64(userId), "", blockId)
	if err != nil {
		fmt.Println(err)
	}

}
