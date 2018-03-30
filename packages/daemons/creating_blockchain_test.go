package daemons

import (
	"database/sql"
	"fmt"
	"io/ioutil"
	"os"
	"regexp"
	"testing"

	log "github.com/sirupsen/logrus"
)

func getTmpFile(t *testing.T) string {
	tmpFile, err := ioutil.TempFile("", "chain")
	if err != nil {
		t.Fatalf("can't create test file %s", err)
	}
	fileName := tmpFile.Name()
	tmpFile.Close()
	return fileName
}

func TestEmptyFile(t *testing.T) {
	fileName := getTmpFile(t)
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

	fileName := getTmpFile(t)
	defer os.Remove(fileName)

	db := initGorm(t)
	defer db.Close()

	fileBlockBin := marshallFileBlock(block)
	err := ioutil.WriteFile(fileName, fileBlockBin, os.ModeAppend)
	if err != nil {
		t.Fatalf("can't write to file: %s", err)
	}

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
