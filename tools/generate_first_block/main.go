package main

import (
	"flag"
	"io/ioutil"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/GenesisKernel/go-genesis/packages/consts"
	"github.com/GenesisKernel/go-genesis/packages/converter"
	"github.com/GenesisKernel/go-genesis/packages/parser"
	"github.com/GenesisKernel/go-genesis/packages/utils"
)

var keyID *int64 = flag.Int64("KeyID", 0, "keyID for the first block")
var publicKey *string = flag.String("publicKey", "", "user public key")
var nodePublicKey *string = flag.String("nodePublicKey", "", "node public key")
var host *string = flag.String("host", "127.0.0.1", "first block host")
var path *string = flag.String("path", "1block", "first block file path")

func main() {
	flag.Parse()
	now := time.Now().Unix()

	header := &utils.BlockData{
		BlockID:      1,
		Time:         now,
		EcosystemID:  0,
		KeyID:        *keyID,
		NodePosition: 0,
		Version:      consts.BLOCK_VERSION,
	}

	var tx []byte
	_, err := converter.BinMarshal(&tx,
		&consts.FirstBlock{
			TxHeader: consts.TxHeader{
				// TODO: move types to enum
				Type: 1, // FirstBlock

				Time:  uint32(now),
				KeyID: *keyID,
			},
			PublicKey:     []byte(*publicKey),
			NodePublicKey: []byte(*nodePublicKey),
			Host:          *host,
		},
	)

	if err != nil {
		log.WithFields(log.Fields{"type": consts.MarshallingError, "error": err}).Fatal("first block body bin marshalling")
		return
	}

	block, err := parser.MarshallBlock(header, [][]byte{tx}, []byte("0"), "")
	if err != nil {
		log.WithFields(log.Fields{"type": consts.MarshallingError, "error": err}).Fatal("first block marshalling")
		return
	}

	ioutil.WriteFile(*path, block, 0644)
	log.Info("first block generated")
}
