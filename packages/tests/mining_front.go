package main

import (
	"fmt"
//	"database/sql"
	//"github.com/DayLightProject/go-daylight/packages/parser"
	"github.com/DayLightProject/go-daylight/packages/utils"
	"tests_utils"
	//_ "github.com/lib/pq"
	//"encoding/binary"
	//"bytes"
	//"encoding/hex"
	//"crypto/rand"
	//"crypto/rsa"
	//"crypto/sha1"
	//"github.com/DayLightProject/go-daylight/packages/daemons"
//	"strconv"
	//"errors"
	"log"
	"os"
	//"github.com/alyu/configparser"
//	"github.com/astaxie/beego/config"
	//"strings"
	//"regexp"
	//"reflect"
//	"github.com/DayLightProject/go-daylight/packages/consts"
	"io"
)


func main() {

	f, err := os.OpenFile("dclog.txt", os.O_WRONLY | os.O_APPEND | os.O_CREATE, 0777)
	defer f.Close()
	//log.SetOutput(f)
	log.SetOutput(io.MultiWriter(f, os.Stdout))
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	txType := "Mining";
	txTime := "1406545938";
	userId := []byte("105")
	var blockId int64 = 123925
	promised_amount_id:="24"
	amount:="5.69"

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
	txSlice = append(txSlice, []byte(promised_amount_id))
	// amount
	txSlice = append(txSlice, []byte(amount))

	dataForSign := fmt.Sprintf("%v,%v,%s,%s,%s", utils.TypeInt(txType), txTime, userId, promised_amount_id, amount)

	blockData := new(utils.BlockData)
	blockData.BlockId = blockId
	blockData.Time = utils.StrToInt64(txTime)
	blockData.UserId = utils.BytesToInt64(userId)

	err = tests_utils.MakeFrontTest(txSlice, utils.StrToInt64(txTime), dataForSign, txType, utils.BytesToInt64(userId), "", blockId)
	if err != nil {
		fmt.Println(err)
	}

}
