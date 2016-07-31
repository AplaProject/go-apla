package main

import (
	"fmt"
	"github.com/DayLightProject/go-daylight/packages/utils"
	"tests_utils"
)

func main() {

	f:=tests_utils.InitLog()
	defer f.Close()

	txType := "NewCfProject";
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
	txSlice = append(txSlice, []byte("4"))
	//currency_id
	txSlice = append(txSlice, []byte("72"))
	//amount
	txSlice = append(txSlice, []byte("5000"))
	//end_time
	txSlice = append(txSlice, []byte("1427383713"))
	//latitude
	txSlice = append(txSlice, []byte("39.94801"))
	//langitude
	txSlice = append(txSlice, []byte("39.94801"))
	//category
	txSlice = append(txSlice, []byte("1"))
	//project_currency_name
	txSlice = append(txSlice, []byte("0VVDDDF"))
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
