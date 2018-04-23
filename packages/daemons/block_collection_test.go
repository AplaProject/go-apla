package daemons

import (
	"context"
	"database/sql"
	"net"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/stretchr/testify/require"

	"github.com/jinzhu/gorm"

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
	require.NoErrorf(t, err, "can't start daemon: %s", err)
	defer l.Close()

	var wg sync.WaitGroup

	go func() {
		wg.Add(1)
		getAndResponse(t, l, converter.DecToBin(consts.DATA_TYPE_MAX_BLOCK_ID, 2), converter.DecToBin(100, 4))
		wg.Done()

	}()

	host, maxBlockID, err := ChooseBestHost(context.Background(), []string{l.Addr().String()})
	require.NoErrorf(t, err, "choose best host return: %s", err)
	require.Equal(t, host, l.Addr().String(), "return bad host, want %s, got %s", l.Addr().String(), host)
	require.Equal(t, 100, maxBlockID, "bad block id: want %d, got %d", 100, maxBlockID)

	wg.Wait()
}

func checkBlock(t *testing.T, id int64) {
	b := &model.Block{}
	require.NoErrorf(t, b.GetBlock(1), "get block failed: %s", err)
	require.Equalf(t, id, b.ID, "bad blockID want %d, got %d", id, b.ID)
}

func checkInfoBlock(t *testing.T, id int64) {
	ib := &model.InfoBlock{}
	require.NoError(t, ib.GetInfoBlock())

	require.Equalf(t, id, ib.BlockID, "bad info block: want %d, got %d", id, ib.BlockID)
}

func TestFirstBlock(t *testing.T) {

	g := initGorm(t)
	defer g.Close()

	entry := logrus.WithFields(logrus.Fields{"type": "test"})
	require.NoError(t, loadFirstBlock(entry), "loadFirstBlock return error:")
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

	require.NoError(t, loadFromFile(context.Background(), fileName), "load from file return error")
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
