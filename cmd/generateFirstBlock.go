package cmd

import (
	"encoding/hex"
	"io/ioutil"
	"time"

	"github.com/spf13/cobra"

	"github.com/GenesisKernel/go-genesis/packages/consts"
	"github.com/GenesisKernel/go-genesis/packages/converter"
	"github.com/GenesisKernel/go-genesis/packages/parser"
	"github.com/GenesisKernel/go-genesis/packages/utils"
	log "github.com/sirupsen/logrus"
)

var keyID int64
var publicKey string
var nodePublicKey string
var host string
var path string

// generateFirstBlockCmd represents the generateFirstBlock command
var generateFirstBlockCmd = &cobra.Command{
	Use:   "generateFirstBlock",
	Short: "Generate first block",
	Run: func(cmd *cobra.Command, args []string) {
		now := time.Now().Unix()

		header := &utils.BlockData{
			BlockID:      1,
			Time:         now,
			EcosystemID:  0,
			KeyID:        keyID,
			NodePosition: 0,
			Version:      consts.BLOCK_VERSION,
		}

		var tx []byte
		bPublicKey, err := hex.DecodeString(publicKey)
		if err != nil {
			log.WithFields(log.Fields{"value": publicKey, "error": err}).Fatal("converting public key from hex")
		}
		bNodePublicKey, err := hex.DecodeString(nodePublicKey)
		if err != nil {
			log.WithFields(log.Fields{"value": nodePublicKey, "error": err}).Fatal("converting node public key from hex")
		}
		_, err = converter.BinMarshal(&tx,
			&consts.FirstBlock{
				TxHeader: consts.TxHeader{
					// TODO: move types to enum
					Type: 1, // FirstBlock

					Time:  uint32(now),
					KeyID: keyID,
				},
				PublicKey:     []byte(bPublicKey),
				NodePublicKey: []byte(bNodePublicKey),
				Host:          host,
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

		ioutil.WriteFile(path, block, 0644)
		log.Info("first block generated")
	},
}

func init() {
	rootCmd.AddCommand(generateFirstBlockCmd)

	generateFirstBlockCmd.Flags().Int64Var(&keyID, "keyID", 0, "keyID for the first block")
	generateFirstBlockCmd.Flags().StringVar(&publicKey, "publicKey", "", "public key")
	generateFirstBlockCmd.Flags().StringVar(&nodePublicKey, "nodePublicKey", "", "node public key")
	generateFirstBlockCmd.Flags().StringVar(&host, "host", "127.0.0.1", "first block host")
	generateFirstBlockCmd.Flags().StringVar(&path, "path", "1block", "first block file path")

	generateFirstBlockCmd.MarkFlagRequired("KeyID")
	generateFirstBlockCmd.MarkFlagRequired("nodePublicKey")
	generateFirstBlockCmd.MarkFlagRequired("publicKey")

}
