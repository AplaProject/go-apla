package main

import (
	"fmt"
	"github.com/DayLightProject/go-daylight/packages/utils"
	"tests_utils"
	"encoding/json"
)

type vComplex struct {
	Currency map[string][]float64 `json:"currency"`
	Referral map[string]string `json:"referral"`
	Admin int64 `json:"admin"`
}
func main() {

	f:=tests_utils.InitLog()
	defer f.Close()

	txType := "VotesComplex";
	txTime := "1499278817";
	userId := []byte("2")
	var blockId int64 = 128008
	newPct:=new(vComplex)
	newPct.Currency = make(map[string][]float64)
	newPct.Referral = make(map[string]string)
	newPct.Currency["1"] = []float64{0.0000000760368,0.0000000497405,1000,55,10}
	newPct.Currency["72"] = []float64{0.0000000760368,0.0000000497405,1000,55,10}
	newPct.Referral["first"] = "30";
	newPct.Referral["second"] = "0";
	newPct.Referral["third"] = "30";
	newPct.Admin = 100;
	newPctJson, _ := json.Marshal(newPct)

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

	dataForSign := fmt.Sprintf("%v,%v,%s,%s", utils.TypeInt(txType), txTime, userId, newPctJson)

	err := tests_utils.MakeFrontTest(txSlice, utils.StrToInt64(txTime), dataForSign, txType, utils.BytesToInt64(userId), "", blockId)
	if err != nil {
		fmt.Println(err)
	}
}
