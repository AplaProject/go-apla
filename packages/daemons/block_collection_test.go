package daemons

import (
	"context"
	"database/sql"
	"net"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/jinzhu/gorm"
	sqlite "github.com/mattn/go-sqlite3"

	"io/ioutil"

	"github.com/GenesisKernel/go-genesis/packages/consts"
	"github.com/GenesisKernel/go-genesis/packages/converter"
	"github.com/GenesisKernel/go-genesis/packages/model"
)

func encode(x, y []byte) string {
	return string(converter.BinToHex(x))
}
func decode(x, y string) []byte {
	res := converter.HexToBin(x)
	return res
}

func initGorm(t *testing.T) *gorm.DB {
	os.Remove("./db_test")
	os.Remove("./schema")

	drivers := sql.Drivers()
	found := false
	for _, d := range drivers {
		if d == "sqlite3_custom" {
			found = true
			break
		}
	}
	if !found {
		sql.Register("sqlite3_custom", &sqlite.SQLiteDriver{
			ConnectHook: func(conn *sqlite.SQLiteConn) error {
				if err := conn.RegisterFunc("encode", encode, true); err != nil {
					return err
				}
				return conn.RegisterFunc("decode", decode, true)
			},
		})
	}

	schema, err := sql.Open("sqlite3", "./schema")
	if err != nil {
		t.Fatalf("can't create schema base %s", err)
	}
	defer schema.Close()

	gormDb, err := sql.Open("sqlite3_custom", "./db_test")
	if err != nil {
		t.Fatalf("sqlite failed %s", err)
	}
	db, err := gorm.Open("sqlite3", gormDb)
	if err != nil {
		t.Fatalf("gorm init failed: %s", err)
	}
	model.DBConn = db

	err = model.FullNodeCreateTable()
	if err != nil {
		t.Fatalf("can't create table: %s", err)
	}

	err = model.MainLockCreateTable()
	if err != nil {
		t.Fatalf("can't create table: %s", err)
	}

	err = model.BlockChainCreateTable()
	if err != nil {
		t.Fatalf("can't create table: %s", err)
	}

	err = model.DBConn.CreateTable(&testDltWallet{}).Error

	//	err = model.WalletCreateTable()
	if err != nil {
		t.Fatalf("can't create table: %s", err)
	}

	err = model.TransactionsCreateTable()
	if err != nil {
		t.Fatalf("can't create table: %s", err)
	}

	err = model.LogTransactionsCreateTable()
	if err != nil {
		t.Fatalf("can't create table: %s", err)
	}

	err = model.ConfigCreateTable()
	if err != nil {
		t.Fatalf("can't create table: %s", err)
	}

	err = model.SystemRecognizedStatesCreateTable()
	if err != nil {
		t.Fatalf("can't create table: %s", err)
	}

	err = model.InfoBlockCreateTable()
	if err != nil {
		t.Fatalf("can't create table: %s", err)
	}

	sql := `
	CREATE TABLE "tables" (
		"table_type" text NOT NULL DEFAULT '',
		"table_name" text NOT NULL DEFAULT '',
		"table_schema" text NOT NULL DEFAULT ''
	);
	`
	_, err = schema.Exec(sql)
	if err != nil {
		t.Fatalf("%s", err)
	}

	sql = `ATTACH DATABASE './schema' AS 'information_schema';`
	_, err = db.DB().Exec(sql)
	if err != nil {
		t.Fatalf("%s", err)
	}

	return db
}

func createDaemon(db *sql.DB) *daemon {

	config := make(map[string]string)
	config["db_type"] = "sqlite"

	return &daemon{
		goRoutineName: "test",
	}
}

func getAndResponse(t *testing.T, l net.Listener, getRequest, sendRequest []byte) {

	conn, err := l.Accept()
	if err != nil {
		t.Errorf("accept error %s", err)
		return
	}
	defer conn.Close()

	conn.SetReadDeadline(time.Now().Add(time.Second))
	conn.SetWriteDeadline(time.Now().Add(time.Second))

	if getRequest != nil {
		toRead := make([]byte, len(getRequest))
		_, err = conn.Read(toRead)
		if err != nil {
			t.Errorf("read error: %s", err)
			return
		}
	}

	_, err = conn.Write(sendRequest)
	if err != nil {
		t.Errorf("write error: %s", err)
	}
}

func TestChooseBlock(t *testing.T) {
	l, err := net.Listen("tcp4", "localhost:0")
	if err != nil {
		t.Fatalf("can't start daemon: %s", err)
	}
	defer l.Close()

	var wg sync.WaitGroup

	go func() {
		wg.Add(1)
		getAndResponse(t, l, converter.DecToBin(consts.DATA_TYPE_MAX_BLOCK_ID, 2), converter.DecToBin(100, 4))
		wg.Done()

	}()

	host, maxBlockID, err := chooseBestHost(context.Background(), []string{l.Addr().String()})
	if err != nil {
		t.Fatalf("choose best host return: %s", err)
	}

	if host != l.Addr().String() {
		t.Errorf("return bad host, want %s, got %s", l.Addr().String(), host)
	}

	if maxBlockID != 100 {
		t.Errorf("bad block id: want %d, got %d", 100, maxBlockID)
	}
	wg.Wait()
}

func checkBlock(t *testing.T, id int64) {
	b := &model.Block{}
	err := b.GetBlock(1)
	if err != nil {
		t.Errorf("get block failed: %s", err)
	} else {
		if b.ID != id {
			t.Errorf("bad blockID want %d, got %d", id, b.ID)
		}
	}
}

func checkInfoBlock(t *testing.T, id int64) {
	ib := &model.InfoBlock{}
	err := ib.GetInfoBlock()
	if err != nil {
		t.Errorf("can't get info block: %s", err)
	}

	if ib.BlockID != id {
		t.Errorf("bad info block: want %d, got %d", id, ib.BlockID)
	}
}

func TestFirstBlock(t *testing.T) {

	g := initGorm(t)
	defer g.Close()

	err := loadFirstBlock()
	if err != nil {
		t.Errorf("loadFirstBlock return error: %s", err)
	}

	checkBlock(t, 1)
	checkInfoBlock(t, 1)

}

func TestLoadFromFile(t *testing.T) {
	g := initGorm(t)
	defer g.Close()

	fileName := getTmpFile(t)
	defer os.Remove(fileName)
	fileBlockBin := marshallFileBlock(getFirstBlock(t))
	err := ioutil.WriteFile(fileName, fileBlockBin, os.ModeAppend)
	if err != nil {
		t.Fatalf("can't write to file: %s", err)
	}

	err = loadFromFile(context.Background(), fileName)
	if err != nil {
		t.Fatalf("load from file return error: %s", err)
	}
}

type testDltWallet struct {
	WalletID           int64  `gorm:"primary_key;not null"`
	Amount             int64  `gorm:"not null"`
	PublicKey          []byte `gorm:"column:public_key_0;not null"`
	NodePublicKey      []byte `gorm:"not null"`
	LastForgingDataUpd int64  `gorm:"not null default 0"`
	Host               string `gorm:"not null default ''"`
	AddressVote        string `gorm:"not null default ''"`
	FuelRate           int64  `gorm:"not null default 0"`
	SpendingContract   string `gorm:"not null default ''"`
	ConditionsChange   string `gorm:"not null default ''"`
	RollbackID         int64  `gorm:"not null default 0"`
}
