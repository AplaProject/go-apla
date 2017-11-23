package daemons

import (
	"context"
	"testing"
	"time"

	"github.com/AplaProject/go-apla/packages/model"

	"github.com/AplaProject/go-apla/packages/crypto"
	"github.com/AplaProject/go-apla/packages/utils"
)

func TestBlockMarshall(t *testing.T) {
	prevBlock := &model.InfoBlock{BlockID: 1}

	_, priv, err := crypto.GenHexKeys()
	if err != nil {
		t.Fatalf("can't gen keys: %s", err)
	}

	blockTime := time.Now().Unix() - 100
	conf := &model.Config{
		StateID:     1,
		DltWalletID: 100,
	}

	blockBin, err := generateNextBlock(prevBlock, nil, priv, conf, blockTime)
	if err != nil {
		t.Fatalf("generateNextBlock error: %s", err)
	}

	block := blockBin[1:] // skip type
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
		StateID:     1,
		CitizenID:   100,
	}
	if err := config.Save(); err != nil {
		t.Fatalf("can't save config: %s", err)
	}

	nodes := &model.FullNode{
		ID:       1,
		WalletID: 1000,
		StateID:  1,
	}
	if err := nodes.Create(nil); err != nil {
		t.Fatalf("can't create full_nodes config: %s", err)
	}

	prevBlock := &model.InfoBlock{
		StateID:  1,
		WalletID: 1000,
		BlockID:  2,
		Time:     time.Now().Unix() - 100,
		Hash:     []byte("ttt"),
	}
	if err := prevBlock.Create(nil); err != nil {
		t.Fatalf("can't create prevBlock value: %s", err)
	}

	_, public, err := crypto.GenBytesKeys()
	if err != nil {
		t.Fatalf("can't gen keys: %s", err)
	}

	wallet := &model.DltWallet{
		WalletID:      1000,
		PublicKey:     public,
		NodePublicKey: public,
	}
	if err := wallet.Create(nil); err != nil {
		t.Fatalf("can't create wallet: %s", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
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

	if bl.ID != prevBlock.BlockID+1 {
		t.Errorf("bad block_id: wanted %d, got %d", prevBlock.BlockID+1, bl.ID)
	}
}
