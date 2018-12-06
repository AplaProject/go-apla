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

	"io/ioutil"

	"github.com/AplaProject/go-apla/packages/consts"
	"github.com/AplaProject/go-apla/packages/converter"
	"github.com/AplaProject/go-apla/packages/model"
)

func encode(x, y []byte) string {
	return string(converter.BinToHex(x))
}
func decode(x, y string) []byte {
	res := converter.HexToBin(x)
	return res
}

func initGorm(t *testing.T) *gorm.DB {
	return nil
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

	host, maxBlockID, err := ChooseBestHost(context.Background(), []string{l.Addr().String()})
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
