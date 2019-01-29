package cmd

import (
	"io/ioutil"
	"time"

	"github.com/spf13/cobra"

	"path/filepath"

	"github.com/AplaProject/go-apla/packages/block"
	"github.com/AplaProject/go-apla/packages/blockchain"
	"github.com/AplaProject/go-apla/packages/conf"
	"github.com/AplaProject/go-apla/packages/conf/syspar"
	"github.com/AplaProject/go-apla/packages/consts"
	"github.com/AplaProject/go-apla/packages/crypto"
	"github.com/AplaProject/go-apla/packages/model"
	"github.com/AplaProject/go-apla/packages/queue"
	"github.com/AplaProject/go-apla/packages/smart"
	"github.com/AplaProject/go-apla/packages/storage/metadb"
	"github.com/GenesisKernel/memdb"

	log "github.com/sirupsen/logrus"
	msgpack "gopkg.in/vmihailenco/msgpack.v2"
)

var stopNetworkBundleFilepath string
var testBlockchain bool
var privateBlockchain bool

// generateFirstBlockCmd represents the generateFirstBlock command
var generateFirstBlockCmd = &cobra.Command{
	Use:    "generateFirstBlock",
	Short:  "First generation",
	PreRun: loadConfigWKey,
	Run: func(cmd *cobra.Command, args []string) {
		now := time.Now().Unix()

		header := &blockchain.BlockHeader{
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

			decodedKey, err := crypto.HexToPub(string(data))
			if err != nil {
				log.WithError(err).Fatalf("converting %s from hex", kName)
			}

			return decodedKey
		}

		var stopNetworkCert []byte
		if len(stopNetworkBundleFilepath) > 0 {
			var err error
			fp := filepath.Join(conf.Config.KeysDir, stopNetworkBundleFilepath)
			if stopNetworkCert, err = ioutil.ReadFile(fp); err != nil {
				log.WithError(err).WithFields(log.Fields{"filepath": fp}).Fatal("Reading cert data")
			}
		}

		if len(stopNetworkCert) == 0 {
			log.Warn("the fullchain of certificates for a network stopping is not specified")
		}
		var tx []byte
		var err error
		var test int64
		var pb uint64
		if testBlockchain == true {
			test = 1
		}
		if privateBlockchain == true {
			pb = 1
		}

		tx, err = msgpack.Marshal(
			&consts.FirstBlock{
				TxHeader: consts.TxHeader{
					Type:  consts.TxTypeFirstBlock,
					Time:  uint32(now),
					KeyID: conf.Config.KeyID,
				},
				PublicKey:             decodeKeyFile(consts.PublicKeyFilename),
				NodePublicKey:         decodeKeyFile(consts.NodePublicKeyFilename),
				StopNetworkCertBundle: stopNetworkCert,
				Test:                  test,
				PrivateBlockchain:     pb,
			},
		)

		if err != nil {
			log.WithFields(log.Fields{"type": consts.MarshallingError, "error": err}).Fatal("first block body bin marshalling")
			return
		}
		dbCfg := conf.Config.DB
		err = model.GormInit(dbCfg.Host, dbCfg.Port, dbCfg.User, dbCfg.Password, dbCfg.Name)
		if err != nil {
			log.WithFields(log.Fields{
				"db_user": dbCfg.User, "db_password": dbCfg.Password, "db_name": dbCfg.Name, "type": consts.DBError,
			}).Error("can't init gorm")
			return
		}

		memdb, err := memdb.OpenDB(filepath.Join(conf.Config.DataDir, "meta.db"), true)
		if err != nil {
			return
		}
		model.MetaStorage = metadb.NewStorage(memdb)

		err = queue.Init()
		if err != nil {
			return
		}
		if err := smart.LoadSysContract(nil); err != nil {
			return
		}
		if err := blockchain.Init(conf.Config.BlockchainDBDir); err != nil {
			log.WithFields(log.Fields{"error": err, "type": consts.LevelDBError}).Error("can't create blockchain db")
			return
		}
		if err := syspar.SysUpdate(nil); err != nil {
			log.WithError(err).Error("updating sys parameters")
			return
		}

		smartTx, err := smart.CallContract("FirstBlock", 1, map[string]string{"Data": string(tx)}, []string{string(tx)})
		if err != nil {
			log.WithFields(log.Fields{"type": consts.ContractError, "error": err}).Fatal("first block contract execution")
			return
		}
		_, found, err := blockchain.GetFirstBlock(nil)
		if err != nil {
			return
		}

		if found {
			log.WithFields(log.Fields{"type": consts.DuplicateObject}).Info("first block already exists")
			return
		}
		txs := []*blockchain.Transaction{smartTx}
		hsh, err := smartTx.Hash()
		if err != nil {
			return
		}

		b := &blockchain.Block{
			Header:   header,
			TxHashes: [][]byte{hsh},
		}
		err = block.InsertBlockWOForks(b, txs, true, false)
		if err != nil {
			log.WithFields(log.Fields{"type": consts.BlockError, "error": err}).Fatal("inserting first block")
			return
		}

		log.Info("first block generated")
	},
}

func init() {
	generateFirstBlockCmd.Flags().StringVar(&stopNetworkBundleFilepath, "stopNetworkCert", "", "Filepath to the fullchain of certificates for network stopping")
	generateFirstBlockCmd.Flags().BoolVar(&testBlockchain, "test", false, "if true - test blockchain")
	generateFirstBlockCmd.Flags().BoolVar(&privateBlockchain, "private", false, "if true - all transactions will be free")
}
