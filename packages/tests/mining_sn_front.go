package main

import (
	"fmt"
	"github.com/DayLightProject/go-daylight/packages/utils"
	"github.com/DayLightProject/go-daylight/packages/tests_utils"
)

func main() {

	f:=tests_utils.InitLog()
	defer f.Close()

	txType := "MiningSn";
	txTime := "1462215741";
	userId := []byte("2")
	var blockId int64 = 1415

	var txSlice [][]byte
	// hash
	txSlice = append(txSlice, []byte("22cb812e53e22ee539af4a1d39b4596d"))
	// type
	txSlice = append(txSlice,  utils.Int64ToByte(utils.TypeInt(txType)))
	// time
	txSlice = append(txSlice, []byte(txTime))
	// user_id
	txSlice = append(txSlice, userId)

	dataForSign := fmt.Sprintf("%s,%s,%s", utils.TypeInt(txType), txTime, userId)

	err := tests_utils.MakeFrontTest(txSlice, utils.StrToInt64(txTime), dataForSign, txType, utils.BytesToInt64(userId), "", blockId)
	if err != nil {
		fmt.Println(err)
	}
}
