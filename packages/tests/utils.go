package parser

import (
	"fmt"
	"github.com/DayLightProject/go-daylight/packages/utils"
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"fmt"
	"github.com/astaxie/beego/config"
)



func (db *DCDB) MakeFrontTest(transactionArray []string, time int64, dataForSign string, txType string, userId int64) error {
	var err error
	nodeArr := []string{"new_admin", "votes_node_new_miner"}
	MY_PREFIX := utils.Int64ToStr(userId)+"_"
	var binSign []byte
	if utils.InSliceString(txType, nodeArr) {
		k, err := db.GetNodePrivateKey()
		privateKey, err := MakePrivateKey(k)
		if err != nil {
			return ErrInfo(err)
		}
		binSign, err = rsa.SignPKCS1v15(rand.Reader, privateKey, crypto.SHA1, utils.HashSha1(dataForSign))
	} else {
		k, err := db.GetPrivateKey(MY_PREFIX)
		privateKey, err := MakePrivateKey(k)
		if err != nil {
			return ErrInfo(err)
		}
		binSign, err = rsa.SignPKCS1v15(rand.Reader, privateKey, crypto.SHA1, utils.HashSha1(dataForSign))
		binSign = EncodeLengthPlusData(binSign)
	}
	transactionArray = append(transactionArray, binSign)

	parser := new(Parser)
	parser.DCDB = db
	parser.GoroutineName = "test"
	parser.TxSlice = transactionArray
	parser.BlockData = &utils.BlockData{BlockId: 190006, Time: time, UserId: userId}
	parser.TxHash = "111111111111111"

	err0 := utils.CallMethod(parser, txType+"Init")
	if i, ok := err0.(error); ok {
	fmt.Println(err0.(error), i)
	return err0.(error)
	}
	err0 := utils.CallMethod(parser, txType+"Front")
	if i, ok := err0.(error); ok {
	fmt.Println(err0.(error), i)
	return err0.(error)
	}
	err0 := utils.CallMethod(parser, txType+"RollbackFront")
	if i, ok := err0.(error); ok {
	fmt.Println(err0.(error), i)
	return err0.(error)
	}
	return nil
}
