package daemons

import (
	"database/sql"
	"io/ioutil"
	"os"
	"testing"

	"github.com/EGaaS/go-egaas-mvp/packages/static"

	"regexp"

	"log"

	"github.com/EGaaS/go-egaas-mvp/packages/model"
	//"github.com/erikstmartin/go-testdb"
	_ "github.com/jinzhu/gorm/dialects/sqlite"

	"fmt"

	"github.com/jinzhu/gorm"
)

func initAndGetFile(t *testing.T) (string, *gorm.DB) {
	os.Remove("./db_test")
	db, err := gorm.Open("sqlite3", "./db_test")
	if err != nil {
		t.Fatalf("gorm init failed %s", err)
	}
	db.SetLogger(log.New(os.Stdout, "\r\n", 0))
	model.GormSet(db)

	tmpFile, err := ioutil.TempFile("", "chain")
	if err != nil {
		t.Fatalf("can't create test file %s", err)
	}
	fileName := tmpFile.Name()
	tmpFile.Close()
	return fileName, db
}

func TestEmptyFile(t *testing.T) {
	fileName, _ := initAndGetFile(t)
	defer os.Remove(fileName)

	err := writeNextBlocks(fileName, 1)
	if err == nil {
		t.Errorf("should be emty_file error")
	}
	matched, regErr := regexp.Match("empty blockchain file", []byte(err.Error()))
	if regErr != nil || !matched {
		t.Errorf("bad error %s", err)
	}
}
func getFirstBlock(t *testing.T) blockData {
	newBlock, err := static.Asset("static/1block")
	if err != nil {
		t.Fatalf("Can't get first block")
	}

	block, err := unmarshalBlockData(newBlock)
	if err != nil {
		t.Fatalf("readBlock error: %s", err)
	}

	return block
}

func TestBlockUnmarshal(t *testing.T) {
	block := getFirstBlock(t)

	if block.ID != 1 {
		t.Errorf("bad blockID, want 1, got %d", block.ID)
	}
}

func TestLastBlock(t *testing.T) {
	block := getFirstBlock(t)

	fileName, _ := initAndGetFile(t)
	defer os.Remove(fileName)

	fileBlockBin := marshallFileBlock(block)
	err := ioutil.WriteFile(fileName, fileBlockBin, os.ModeAppend)
	if err != nil {
		t.Fatalf("can't write to file: %s", err)
	}

	blockID, err := getLastBlockID(fileName)
	if err != nil {
		t.Fatalf("can't get last id: %s", err)
	}

	if blockID != 1 {
		t.Errorf("bad id, want 1, got %d", blockID)
	}
}

func createTables(t *testing.T, db *sql.DB) {
	sql := `
	CREATE TABLE "info_block" (
		"hash" blob  NOT NULL DEFAULT '',
		"block_id" integer NOT NULL DEFAULT '0',
		"state_id" integer  NOT NULL DEFAULT '0',
		"wallet_id" integer  NOT NULL DEFAULT '0',
		"time" integer  NOT NULL DEFAULT '0',
		"level" integer  NOT NULL DEFAULT '0',
		"current_version" text NOT NULL DEFAULT '0.0.1',
	"sent" integer NOT NULL DEFAULT '0'
	);
	CREATE TABLE "block_chain" (
		"id" integer NOT NULL DEFAULT '0',
		"hash" blob  NOT NULL DEFAULT '',
		"data" blob NOT NULL DEFAULT '',
		"state_id" integer  NOT NULL DEFAULT '0',
		"wallet_id" integer NOT NULL DEFAULT '0',
		"time" integer NOT NULL DEFAULT '0',
		"tx" integer NOT NULL DEFAULT '0',
		"cur_0l_miner_id" integer NOT NULL DEFAULT '0',
		"max_miner_id" integer NOT NULL DEFAULT '0'
	);
	`

	var err error
	_, err = db.Exec(sql)
	if err != nil {
		t.Fatalf("%s", err)
	}
}

func addBlockInfo(t *testing.T, blockID int64, db *sql.DB) {
	var err error
	_, err = db.Exec(fmt.Sprintf("insert into info_block(block_id) values(%d)", blockID))
	if err != nil {
		t.Fatal(err)
	}
}

func addBlock(t *testing.T, blockID int64, data []byte, db *sql.DB) {

	stmt, err := db.Prepare("insert into block_chain(id, data) values(?, ?)")
	if err != nil {
		t.Fatal(err)
	}
	defer stmt.Close()

	_, err = stmt.Exec(blockID, data)
	if err != nil {
		t.Fatal(err)
	}

}

func TestWriteNext(t *testing.T) {
	block := getFirstBlock(t)

	fileName, db := initAndGetFile(t)
	defer os.Remove(fileName)

	fileBlockBin := marshallFileBlock(block)
	err := ioutil.WriteFile(fileName, fileBlockBin, os.ModeAppend)
	if err != nil {
		t.Fatalf("can't write to file: %s", err)
	}

	createTables(t, db.DB())
	addBlockInfo(t, 2, db.DB())
	addBlock(t, 2, []byte("test"), db.DB())

	err = writeNextBlocks(fileName, 1)
	if err != nil {
		log.Fatalf("writeNextBlocks error: %s", err)
	}

	file, err := os.Open(fileName)
	if err != nil {
		t.Fatalf("can't open file: %s", err)
	}

	for i := 0; i < 2; i++ {
		blockData, err := readBlock(file)
		if err != nil {
			t.Fatalf("readBlock error: %s", err)
		}
		if blockData.ID != int64(i+1) {
			t.Errorf("bad block id: want %d, got %d", i, blockData.ID)
		}

		if i == 1 {
			if string(blockData.Data) != "test" {
				t.Errorf("bad block data: want test, got %s", string(blockData.Data))
			}
		}
	}

}
