package main

import (
	"fmt"
//	"database/sql"
	"github.com/DayLightProject/go-daylight/packages/utils"
	_ "github.com/lib/pq"
	//"encoding/binary"
	//"bytes"
	//"encoding/hex"
	//"crypto/rand"
	//"crypto/rsa"
	//"crypto/sha1"
	//"github.com/DayLightProject/go-daylight/packages/daemons"
//	"strconv"
	//"errors"
//	"github.com/DayLightProject/go-daylight/packages/parser"
	"log"
	"os"
	//"github.com/alyu/configparser"
	"github.com/astaxie/beego/config"
	//"strings"
	//"regexp"
	//"reflect"
	"io"
	"tests_utils"
)


func main() {

	f, err := os.OpenFile("dclog.txt", os.O_WRONLY | os.O_APPEND | os.O_CREATE, 0777)
	defer f.Close()
	//log.SetOutput(f)
	log.SetOutput(io.MultiWriter(f, os.Stdout))
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	txType := "ChangeHost";
	txTime := "1427383713";
	userId := []byte("2")
	var blockId int64 = 1288

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
	txSlice = append(txSlice, []byte(`http://sffs.ru/`))
	// sign
	txSlice = append(txSlice, []byte("11111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111"))

	blockData := new(utils.BlockData)
	blockData.BlockId = blockId
	blockData.Time = utils.StrToInt64(txTime)
	blockData.UserId = utils.BytesToInt64(userId)

	dir, err := utils.GetCurrentDir()
	if err != nil {
		log.Error("%v", utils.ErrInfo(err))
	}
	configIni_, err := config.NewConfig("ini", dir+"/config.ini")
	if err != nil {
		fmt.Println(err)
	}
	configIni, err := configIni_.GetSection("default")
	db := utils.DbConnect(configIni)

	// делаем снимок БД в виде хэшей до начала тестов
	hashesStart, err := tests_utils.AllHashes(db)

	err = tests_utils.MakeTest(txSlice, blockData, txType, hashesStart, db, "work_and_rollback");
	if err != nil {
		fmt.Println(err)
	}

}
