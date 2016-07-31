package main

import (
	"fmt"
	"database/sql"
	"github.com/DayLightProject/go-daylight/packages/utils"
	_ "github.com/lib/pq"
	//"encoding/binary"
	//"bytes"
	//"encoding/hex"
	//"crypto/rand"
	//"crypto/rsa"
	//"crypto/sha1"
	//"github.com/DayLightProject/go-daylight/packages/daemons"
	"strconv"
	//"errors"
	"github.com/DayLightProject/go-daylight/packages/dcparser"
	"log"
	"os"
	"github.com/alyu/configparser"
	//"strings"
	//"regexp"
	//"reflect"
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

func TypeInt (txType string) int32 {
	x := make([]string, 67)
	// новый юзер
	x[1] = "new_user"
	x[48] = "cf_send_dc"
	for k, v := range x {
		if v == txType {
			return int32(k)
		}
	}
	return 0
}

type Parser struct {
	db *sql.DB
	txSlice []string
	txMap map[string]string
	blockData map[string]string
}

var configIni *configparser.Section

func main() {

	f, err := os.OpenFile("dclog.txt", os.O_WRONLY | os.O_APPEND | os.O_CREATE, 0777)
	defer f.Close()
	log.SetOutput(f)
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	dir, err := utils.GetCurrentDir()
	if err != nil {
		log.Fatal(err)
	}
	config, err := configparser.Read(dir+"/config.ini")
	if err != nil {
		log.Fatal(err)
	}
	configIni, err := config.Section("main")



	txType := "NewHolidays";
	txTime := "1426283713";
	blockData := make(map[string]string)

	var txSlice []string
	// hash
	txSlice = append(txSlice, "22cb812e53e22ee539af4a1d39b4596d")
	// type
	txSlice = append(txSlice,  strconv.Itoa(int(TypeInt(txType))))
	// time
	txSlice = append(txSlice, txTime)
	// user_id
	txSlice = append(txSlice, strconv.FormatInt(1, 10));
	//start
	txSlice = append(txSlice, strconv.FormatInt(100000, 10));
	//end
	txSlice = append(txSlice, strconv.FormatInt(4545, 10));
	// sign
	txSlice = append(txSlice, "11111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111")

	blockData["block_id"] = strconv.FormatInt(185510, 10);
	blockData["time"] = txTime;
	blockData["user_id"] = strconv.FormatInt(1, 10);

	//fmt.Println(txSlice)

	parser := new(dcparser.Parser)
	parser.DCDB = utils.NewDbConnect(configIni)
	parser.TxSlice = txSlice;
	parser.BlockData = blockData;

	/*for i:=0; i<10000; i++ {

		x := func() {
			stmt, err := parser.DCDB.Prepare(`INSERT INTO main_lock(lock_time,script_name) VALUES($1,$2)`)
			defer stmt.Close()
			if err!=nil {
				fmt.Println(err)
			}
			_, err = stmt.Exec(11111, "testblock_generator")
			if err!=nil {
				fmt.Println(err)
			}
		}
		x()
		//stmt, _ := parser.DCDB.Prepare(`INSERT INTO main_lock(lock_time,script_name) VALUES($1,$2)`)
		//fmt.Println(err)
		//defer stmt.Close()
		//_, _ = stmt.Exec(11111, "testblock_generator")
		//fmt.Println(err)
		//_, _ = parser.DCDB.Query("INSERT INTO main_lock(lock_time,script_name) VALUES($1,$2)", 11111, "testblock_generator")
		x2 := func() {
			row, err := parser.DCDB.Query("DELETE FROM main_lock WHERE script_name='testblock_generator'")
			defer row.Close()
			if err!=nil {
				fmt.Println(err)
			}
		}
		x2()
		//parser.DCDB.DbLock()
		//parser.DCDB.DbUnlock()
		//fmt.Println(i)
	}*/
	fmt.Println()


	err = dcparser.MakeTest(parser, txType, hashesStart);
	if err != nil {
		fmt.Println("err", err)
	}
	//go daemons.Testblock_is_ready()

	//parser.Db.HashTableData("holidays", "", "")
	//HashTableData(parser.Db.DB,"holidays", "", "")
	//hashes, err := parser.Db.AllHashes()
	utils.CheckErr(err);
	//fmt.Println(hashes)
	fmt.Println()
/*
	var ptr reflect.Value
	var value reflect.Value
	//var finalMethod reflect.Value

	i := Test{Start: "start"}

	value = reflect.ValueOf(i)

	// if we start with a pointer, we need to get value pointed to
	// if we start with a value, we need to get a pointer to that value
	if value.Type().Kind() == reflect.Ptr {
		ptr = value
		value = ptr.Elem()
	} else {
		ptr = reflect.New(reflect.TypeOf(i))
		temp := ptr.Elem()
		temp.Set(value)
	}
	fmt.Println(value)
/*
	// check for method on value
	method := value.MethodByName("Finish")
	fmt.Println(method)
	// check for method on pointer
	method = ptr.MethodByName("Finish")
	fmt.Println(method)*/

}
