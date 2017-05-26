// Copyright 2016 The go-daylight Authors
// This file is part of the go-daylight library.
//
// The go-daylight library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-daylight library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-daylight library. If not, see <http://www.gnu.org/licenses/>.

package tests_utils

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"fmt"
	"github.com/astaxie/beego/config"
	"github.com/EGaaS/go-egaas-mvp/packages/parser"
	"github.com/EGaaS/go-egaas-mvp/packages/utils"
	"io"
	"log"
	"os"
	"regexp"
	"strings"
	//	"crypto/rand"
	//	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
)

func genKeys() (string, string) {
	privatekey, _ := rsa.GenerateKey(rand.Reader, 1024)
	var pemkey = &pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(privatekey)}
	PrivBytes0 := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: pemkey.Bytes})

	PubASN1, _ := x509.MarshalPKIXPublicKey(&privatekey.PublicKey)
	pubBytes := pem.EncodeToMemory(&pem.Block{Type: "RSA PUBLIC KEY", Bytes: PubASN1})
	s := strings.Replace(string(pubBytes), "-----BEGIN RSA PUBLIC KEY-----", "", -1)
	s = strings.Replace(s, "-----END RSA PUBLIC KEY-----", "", -1)
	sDec, _ := base64.StdEncoding.DecodeString(s)

	return string(PrivBytes0), fmt.Sprintf("%x", sDec)
}

// для юнит-тестов. снимок всех данных в БД
// for unit tests. Snapshot of all data in the database
func AllHashes(db *utils.DCDB) (map[string]string, error) {
	//var orderBy string
	result := make(map[string]string)
	//var columns string;
	tables, err := db.GetAllTables()
	if err != nil {
		return result, err
	}
	/*rows, err := db.Query(`
		SELECT table_name
		FROM
		information_schema.tables
		WHERE
		table_type = 'BASE TABLE'
		AND
		table_schema NOT IN ('pg_catalog', 'information_schema');`)
	if err != nil {
		//fmt.Println(err)
		return result, err
	}
	for rows.Next() {
		var table string
		err = rows.Scan(&table)
		if err != nil {
			return result, err
		}
		//fmt.Println(table)
	*/
	for _, table := range tables {
		orderByFns := func(table string) string {
			// ошибки не проверяются т.к. некритичны
			// errors are not checking, because they are not important
			match, _ := regexp.MatchString("^(rb_forex_orders|rb_forex_orders_main|cf_comments|cf_currency|cf_funding|cf_lang|cf_projects|cf_projects_data)$", table)
			if match {
				return "id"
			}
			match, _ = regexp.MatchString("^rb_time_(.*)$", table)
			if match && table != "rb_time_money_orders" {
				return "user_id, time"
			}
			match, _ = regexp.MatchString("^log_transactions$", table)
			if match {
				return "time"
			}
			match, _ = regexp.MatchString("^rb_votes$", table)
			if match {
				return "user_id, voting_id"
			}
			match, _ = regexp.MatchString("^rb_(.*)$", table)
			if match && table != "rb_time_money_orders" && table != "rb_minute" {
				return "rb_id"
			}
			match, _ = regexp.MatchString("^wallets$", table)
			if match {
				return "last_update"
			}
			return ""
		}
		orderBy := orderByFns(table)
		hash, err := db.HashTableData(table, "", orderBy)
		if err != nil {
			return result, utils.ErrInfo(err)
		}
		result[table] = hash
	}
	return result, nil
}

func DbConn() *utils.DCDB {
	configIni_, err := config.NewConfig("ini", *utils.Dir+"/config.ini")
	if err != nil {
		fmt.Println(err)
	}
	configIni, err := configIni_.GetSection("default")
	db, err := utils.NewDbConnect(configIni)
	if err != nil {
		fmt.Println(err)
	}
	return db
}

func InitLog() *os.File {
	f, err := os.OpenFile("dclog.txt", os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0777)
	if err != nil {
		fmt.Println(err)
	}
	//log.SetOutput(f)
	log.SetOutput(io.MultiWriter(f, os.Stdout))
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	return f
}

func MakeFrontTest(transactionArray [][]byte, time int64, dataForSign string, txType string, userId int64, MY_PREFIX string, blockId int64) error {

	db := DbConn()

	priv, pub := genKeys()

	nodeArr := []string{"new_admin", "votes_node_new_miner", "NewPct"}
	var binSign []byte
	if utils.InSliceString(txType, nodeArr) {

		err := db.ExecSQL("UPDATE my_node_keys SET private_key = ?", priv)
		if err != nil {
			return utils.ErrInfo(err)
		}
		err = db.ExecSQL("UPDATE miners_data SET node_public_key = [hex] WHERE user_id = ?", pub, userId)
		if err != nil {
			return utils.ErrInfo(err)
		}

		k, err := db.GetNodePrivateKey()
		if err != nil {
			return utils.ErrInfo(err)
		}
		fmt.Println("k", k)
		privateKey, err := utils.MakePrivateKey(k)
		if err != nil {
			return utils.ErrInfo(err)
		}
		//fmt.Println("privateKey.PublicKey", privateKey.PublicKey)
		//fmt.Println("privateKey.D", privateKey.D)
		//fmt.Printf("privateKey.N %x\n", privateKey.N)
		//fmt.Println("privateKey.Public", privateKey.Public())
		binSign, err = rsa.SignPKCS1v15(rand.Reader, privateKey, crypto.SHA1, utils.HashSha1(dataForSign))
		//nodePublicKey, err := db.GetNodePublicKey(userId)
		//fmt.Println("nodePublicKey", nodePublicKey)
		//if err != nil {
		//	return utils.ErrInfo(err)
		//}
		//CheckSignResult, err := utils.CheckSign([][]byte{nodePublicKey}, dataForSign, binSign, true);
		//fmt.Printf("binSign: %x\n", binSign)
		//fmt.Println("err", err)
		//fmt.Println("CheckSignResult", CheckSignResult)

	} else {

		err := db.ExecSQL("UPDATE my_keys SET private_key = ?", priv)
		if err != nil {
			return utils.ErrInfo(err)
		}
		err = db.ExecSQL("UPDATE users SET public_key_0 = [hex]", pub)
		if err != nil {
			return utils.ErrInfo(err)
		}

		k, err := db.GetPrivateKey(MY_PREFIX)
		privateKey, err := utils.MakePrivateKey(k)
		if err != nil {
			return utils.ErrInfo(err)
		}
		binSign, err = rsa.SignPKCS1v15(rand.Reader, privateKey, crypto.SHA1, utils.HashSha1(dataForSign))
		binSign = utils.EncodeLengthPlusData(binSign)
	}

	//fmt.Println("HashSha1", utils.HashSha1(dataForSign))
	//fmt.Printf("binSign %x\n", binSign)
	//fmt.Println("dataForSign", dataForSign)
	transactionArray = append(transactionArray, binSign)

	parser := new(parser.Parser)
	parser.DCDB = db
	parser.GoroutineName = "test"
	parser.TxSlice = transactionArray
	parser.BlockData = &utils.BlockData{BlockId: blockId, Time: time, UserId: userId}
	parser.TxHash = "111111111111111"
	parser.Variables, _ = parser.DCDB.GetAllVariables()

	err0 := utils.CallMethod(parser, txType+"Init")
	if i, ok := err0.(error); ok {
		fmt.Println(err0.(error), i)
		return err0.(error)
	}
	err0 = utils.CallMethod(parser, txType+"Front")
	if i, ok := err0.(error); ok {
		fmt.Println(err0.(error), i)
		return err0.(error)
	}
	/*err0 = utils.CallMethod(parser, txType+"RollbackFront")
	if i, ok := err0.(error); ok {
		fmt.Println(err0.(error), i)
		return err0.(error)
	}*/
	return nil
}

func MakeTest(txSlice [][]byte, blockData *utils.BlockData, txType string, testType string) error {

	db := DbConn()

	parser := new(parser.Parser)
	parser.DCDB = db
	parser.TxSlice = txSlice
	parser.BlockData = blockData
	parser.TxHash = "111111111111111"
	parser.Variables, _ = db.GetAllVariables()

	// делаем снимок БД в виде хэшей до начала тестов
	// make a snapshot of database in a shape of hashes befor tests begin
	hashesStart, err := AllHashes(db)
	if err != nil {
		return err
	}

	//fmt.Println("parser."+txType+"Init")
	err0 := utils.CallMethod(parser, txType+"Init")
	if i, ok := err0.(error); ok {
		fmt.Println(err0.(error), i)
		return err0.(error)
	}

	if testType == "work_and_rollback" {

		err0 = utils.CallMethod(parser, txType)
		if i, ok := err0.(error); ok {
			fmt.Println(err0.(error), i)
			return err0.(error)
		}

		//fmt.Println("-------------------")
		// узнаем, какие таблицы были затронуты в результате выполнения основного метода
		// get know which tables were affected by the execution of the main method
		hashesMiddle, err := AllHashes(db)
		if err != nil {
			return utils.ErrInfo(err)
		}
		var tables []string
		//fmt.Println("hashesMiddle", hashesMiddle)
		//fmt.Println("hashesStart", hashesStart)
		for table, hash := range hashesMiddle {
			if hash != hashesStart[table] {
				tables = append(tables, table)
			}
		}
		fmt.Println("tables", tables)

		// rollback
		err0 := utils.CallMethod(parser, txType+"Rollback")
		if i, ok := err0.(error); ok {
			fmt.Println(err0.(error), i)
			return err0.(error)
		}

		// сраниим хэши, которые были до начала и те, что получились после роллбэка
		// compare the hashes, which were before the beginning and those which were created after the rollback
		hashesEnd, err := AllHashes(db)
		if err != nil {
			return utils.ErrInfo(err)
		}
		for table, hash := range hashesEnd {
			if hash != hashesStart[table] {
				fmt.Println("ERROR in table ", table)
			}
		}

	} else if (len(os.Args) > 1 && os.Args[1] == "w") || testType == "work" {
		err0 = utils.CallMethod(parser, txType)
		if i, ok := err0.(error); ok {
			fmt.Println(err0.(error), i)
			return err0.(error)
		}
	} else if (len(os.Args) > 1 && os.Args[1] == "r") || testType == "rollback" {
		err0 = utils.CallMethod(parser, txType+"Rollback")
		if i, ok := err0.(error); ok {
			fmt.Println(err0.(error), i)
			return err0.(error)
		}
	}
	return nil
}
