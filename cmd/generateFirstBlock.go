package cmd

import (
	"encoding/hex"
	"io/ioutil"
	"time"

	"github.com/spf13/cobra"

	"path/filepath"

	"github.com/GenesisKernel/go-genesis/packages/conf"
	"github.com/GenesisKernel/go-genesis/packages/consts"
	"github.com/GenesisKernel/go-genesis/packages/converter"
	"github.com/GenesisKernel/go-genesis/packages/parser"
	"github.com/GenesisKernel/go-genesis/packages/utils"
	log "github.com/sirupsen/logrus"
)

// generateFirstBlockCmd represents the generateFirstBlock command
var generateFirstBlockCmd = &cobra.Command{
	Use:    "generateFirstBlock",
	Short:  "First generation",
	PreRun: loadConfigWKey,
	Run: func(cmd *cobra.Command, args []string) {
		now := time.Now().Unix()

		header := &utils.BlockData{
			BlockID:      1,
			Time:         now,
			EcosystemID:  0,
			KeyID:        conf.Config.KeyID,
			NodePosition: 0,
			Version:      consts.BLOCK_VERSION,
		}

		decodeKeyFile := func(kName string) []byte {
			filepath := filepath.Join(conf.Config.KeysDir, kName)
			data, err := ioutil.ReadFile(filepath)
			if err != nil {
				log.WithError(err).WithFields(log.Fields{"key": kName, "filepath": filepath}).Fatal("Reading key data")
			}

			decodedKey, err := hex.DecodeString(string(data))
			if err != nil {
				log.WithError(err).Fatalf("converting %s from hex", kName)
			}

			return decodedKey
		}

		var tx []byte
		_, err := converter.BinMarshal(&tx,
			&consts.FirstBlock{
				TxHeader: consts.TxHeader{
					Type:  1, // FirstBlock
					Time:  uint32(now),
					KeyID: conf.Config.KeyID,
				},
				PublicKey:     decodeKeyFile(consts.PublicKeyFilename),
				NodePublicKey: decodeKeyFile(consts.NodePublicKeyFilename),
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

		ioutil.WriteFile(conf.Config.FirstBlockPath, block, 0644)
		log.Info("first block generated")
	},
}
