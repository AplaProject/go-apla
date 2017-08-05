package daemons

import (
	"testing"
	"context"
	"time"
	"github.com/EGaaS/go-egaas-mvp/packages/model"

	"github.com/EGaaS/go-egaas-mvp/packages/crypto"
	"github.com/EGaaS/go-egaas-mvp/packages/utils"
	"github.com/EGaaS/go-egaas-mvp/packages/converter"
)

func TestBlockMarshall(t *testing.T) {
	prevBlock := &model.InfoBlock{BlockID: 1}

	_, priv, err := crypto.GenHexKeys()
	if err != nil {
		t.Fatalf("can't gen keys: %s", err)
	}

	blockTime := time.Now().Unix() - 100
	conf := &model.Config {
		StateID: 1,
		DltWalletID: 100,
	}

	blockBin, err := generateNextBlock(prevBlock, nil, priv, conf, blockTime)
	if err != nil {
		t.Fatalf("generateNextBlock error: %s", err)
	}

	block := blockBin[1:]  // skip type
	data := utils.ParseBlockHeader(&block)
	if data.BlockID != 2 {
		t.Errorf("bad block_id: want 2, got %d", data.BlockID)
	}

	if data.WalletID != conf.DltWalletID {
		t.Errorf("bad wallet value: want %d, got %d", conf.DltWalletID, data.WalletID)
	}

	if data.StateID != conf.StateID {
		t.Errorf("bad state id: want %d, got %d", conf.StateID, data.StateID)
	}

	if data.Time != blockTime {
		t.Errorf("bad time value: want %d, got %d", blockTime, data.Time)
	}
}

func TestBlockGenerator(t *testing.T) {

	db := initGorm(t)

	config := &model.Config{
		DltWalletID: 1000,
		StateID: 1,
		CitizenID: 100,
	}
	if err := config.Save(); err != nil {
		t.Fatalf("can't save config: %s", err)
	}

	nodes := &model.FullNodes {
		ID: 1,
		WalletID: 1000,
		StateID: 1,
	}
	if err := nodes.Create(); err != nil {
		t.Fatalf("can't create full_nodes config: %s", err)
	}

	prevBlock := &model.InfoBlock{
		StateID: 1,
		WalletID: 1000,
		BlockID: 2,
		Time: int32(time.Now().Unix() - 100),
		Hash: []byte("ttt"),
	}
	if err := prevBlock.Create(); err != nil {
		t.Fatalf("can't create prevBlock value: %s", err)
	}

	priv, public, err := crypto.GenHexKeys()
	if err != nil {
		t.Fatalf("can't gen keys: %s", err)
	}

	keys := &model.MyNodeKeys{
		ID: 1,
		BlockID: 1,
		PublicKey: []byte(public),
		PrivateKey: []byte(priv),
	}
	if err := keys.Create(); err != nil {
		t.Fatalf("can't create my_node_keys table: %s", err)
	}

	wallet := &model.Wallet{
		WalletID: 1000,
		PublicKey: []byte(public),
		NodePublicKey: []byte(converter.HexToBin(public)),  // TODO: ????????
	}
	if err := wallet.Create(); err != nil {
		t.Fatalf("can't create wallet: %s", err)
	}

	ctx, cancel:= context.WithTimeout(context.Background(), 1 * time.Second)
	defer  cancel()
	d := createDaemon(db.DB())

	err = BlockGenerator(d, ctx)
	if err != nil {
		t.Fatalf("block generator return: %s", err)
	}

	bl := &model.Block{}
	err = bl.GetMaxBlock()
	if err != nil {
		t.Fatalf("can't get block: %s", err)
	}

	if bl.ID != prevBlock.BlockID + 1 {
		t.Errorf("bad block_id: wanted %d, got %d", prevBlock.BlockID + 1, bl.ID)
	}
}


