package main

import (
	"fmt"
	"github.com/DayLightProject/go-daylight/packages/utils"
	"tests_utils"
)

func main() {

	f:=tests_utils.InitLog()
	defer f.Close()

	txType := "CashRequestOut";
	txTime := "1427383713";
	userId := []byte("2")
	var blockId int64 = 128008

	var txSlice [][]byte
	// hash
	txSlice = append(txSlice, []byte("1111111111"))
	// type
	txSlice = append(txSlice,  utils.Int64ToByte(utils.TypeInt(txType)))
	// time
	txSlice = append(txSlice, []byte(txTime))
	// user_id
	txSlice = append(txSlice, []byte("2"))
	//to_user_id
	txSlice = append(txSlice, []byte("4"))
	//amount
	txSlice = append(txSlice, []byte("600"))
	//comment
	txSlice = append(txSlice, []byte("111111111"))
	//currency_id
	txSlice = append(txSlice, []byte("21"))
	//hash_code
	txSlice = append(txSlice, []byte("11111111111"))
	// sign
	txSlice = append(txSlice, []byte("11111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111"))

	blockData := new(utils.BlockData)
	blockData.BlockId = blockId
	blockData.Time = utils.StrToInt64(txTime)
	blockData.UserId = utils.BytesToInt64(userId)

	err := tests_utils.MakeTest(txSlice, blockData, txType, "work_and_rollback");
	if err != nil {
		fmt.Println(err)
	}

}
