package main

import (
	"fmt"
	"github.com/DayLightProject/go-daylight/packages/utils"
	"tests_utils"
	"encoding/json"
)

type vComplex struct {
	Currency map[string][]float64 `json:"currency"`
	Referral map[string]int64 `json:"referral"`
	Admin int64 `json:"admin"`
}

func main() {
/*	Currency map[string][]float64 `json:"currency"`
	Referral map[string]int64 `json:"referral"`
	Admin int64 `json:"admin"`
*/
	f:=tests_utils.InitLog()
	defer f.Close()

	txType := "VotesComplex";
	txTime := "1427383713";
	userId := []byte("2")
	var blockId int64 = 128008
	newPct:=new(vComplex)
	newPct.Currency = make(map[string][]float64)
	newPct.Referral = make(map[string]int64)
	newPct.Currency["1"] = []float64{0.0000000760368,0.0000000497405,1000,55,10}
	newPct.Currency["33"] = []float64{0.0000000760368,0.0000000497405,1000,55,10}
	newPct.Currency["2"] = []float64{0.0000000760368,0.0000000497405,1000,55,10}
	newPct.Referral["first"] = 30;
	newPct.Referral["second"] = 0;
	newPct.Referral["third"] = 30;
	newPct.Admin = 100;
	newPctJson, _ := json.Marshal(newPct)

	//newPct1:=new(vComplex)
	//err := json.Unmarshal([]byte(`{"currency":{"1":[7.60368e-08,4.97405e-08,1000,55,10],"2":[7.60368e-08,4.97405e-08,1000,55,10],"33":[7.60368e-08,4.97405e-08,1000,55,10]},"referral":{"first":30,"second":0,"third":30},"admin":100}`), &newPct1)
	//fmt.Println(newPct1)
	//fmt.Println(err)


	var txSlice [][]byte
	// hash
	txSlice = append(txSlice, []byte("22cb812e53e22ee539af4a1d39b4596d"))
	// type
	txSlice = append(txSlice,  utils.Int64ToByte(utils.TypeInt(txType)))
	// time
	txSlice = append(txSlice, []byte(txTime))
	// user_id
	txSlice = append(txSlice, userId)
	// newPctJson
	txSlice = append(txSlice, []byte(newPctJson))
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
