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
	"github.com/DayLightProject/go-daylight/packages/dcparser"
	"log"
	"os"
	//"github.com/alyu/configparser"
	"github.com/astaxie/beego/config"
	//"strings"
	//"regexp"
	//"reflect"
	"github.com/DayLightProject/go-daylight/packages/consts"
)
type Config struct {
	Section struct {
		Name string
		Flag bool
	}
}
type Data struct {
	id int32
	name [16]byte
}



func main() {

	f, err := os.OpenFile("dclog.txt", os.O_WRONLY | os.O_APPEND | os.O_CREATE, 0777)
	defer f.Close()
	log.SetOutput(f)
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	txType := "NewPct";
	txTime := "1426283713";
	userId := []byte("1")

	var txSlice [][]byte
	// hash
	txSlice = append(txSlice, []byte("22cb812e53e22ee539af4a1d39b4596d"))
	// type
	txSlice = append(txSlice,  utils.Int64ToByte(TypeInt(txType)))
	// time
	txSlice = append(txSlice, []byte(txTime))
	// user_id
	txSlice = append(txSlice, userId)
	//new_pct
	txSlice = append(txSlice, []byte(`{"1":{"miner_pct":"0.0000000044318","user_pct":"0.0000000027036"},"72":{"miner_pct":"0.0000000047610","user_pct":"0.0000000029646"}}`))
	// sign
	txSlice = append(txSlice, []byte("11111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111"))

	blockData := new(utils.BlockData)
	blockData.BlockId = 1200
	blockData.Time = utils.StrToInt64(txTime)
	blockData.UserId = 1

	dir, err := utils.GetCurrentDir()
	if err != nil {
		fmt.Println(err)
	}
	configIni_, err := config.NewConfig("ini", dir+"/config.ini")
	if err != nil {
		fmt.Println(err)
	}
	configIni, err := configIni_.GetSection("default")
	db := utils.DbConnect(configIni)
	parser := new(dcparser.Parser)
	parser.DCDB = db
	parser.TxSlice = txSlice;
	parser.BlockData = blockData;

	// делаем снимок БД в виде хэшей до начала тестов
	hashesStart, err := parser.AllHashes()

	err = dcparser.MakeTest(parser, txType, hashesStart);
	if err != nil {
		fmt.Println(err)
	}
	//go daemons.candidateBlock_is_ready()

	//parser.Db.HashTableData("holidays", "", "")
	//HashTableData(parser.Db.DB,"holidays", "", "")
	//hashes, err := parser.Db.AllHashes()
	utils.CheckErr(err);
	//fmt.Println(hashes)
	fmt.Println()


}
