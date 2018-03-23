package cmd

import (
	"time"

	"github.com/GenesisKernel/go-genesis/packages/conf"
	"github.com/GenesisKernel/go-genesis/packages/consts"
	"github.com/GenesisKernel/go-genesis/packages/converter"
	"github.com/GenesisKernel/go-genesis/packages/crypto"
	"github.com/GenesisKernel/go-genesis/packages/model"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

// stopNetworkCmd represents the stopNetworkCmd command
var stopNetworkCmd = &cobra.Command{
	Use:    "stopNetwork",
	Short:  "Sending a special transaction to stop the network",
	PreRun: loadConfigWKey,
	Run: func(cmd *cobra.Command, args []string) {
		err := model.GormInit(
			conf.Config.DB.Host,
			conf.Config.DB.Port,
			conf.Config.DB.User,
			conf.Config.DB.Password,
			conf.Config.DB.Name,
		)
		if err != nil {
			log.WithError(err).Fatal("init db")
		}

		var data []byte
		_, err = converter.BinMarshal(&data,
			&consts.StopNetwork{
				TxHeader: consts.TxHeader{
					Type:  consts.TxTypeStopNetwork,
					Time:  uint32(time.Now().Unix()),
					KeyID: conf.Config.KeyID,
				},
				StopNetworkCert: []byte(stopNetworkCert),
			},
		)
		if err != nil {
			log.WithError(err).Fatal("marshal")
		}

		// daemons.SendDRequest(conf.Config.TCPServer.Str(), 1, data, nil, log.WithFields(nil))

		hash, err := crypto.Hash(data)
		if err != nil {
			log.WithError(err).Fatal("hash")
		}

		tx := &model.Transaction{
			Hash:     hash,
			Data:     data[:],
			Type:     int8(converter.BinToDecBytesShift(&data, 1)),
			KeyID:    conf.Config.KeyID,
			HighRate: model.TransactionRateStopNetwork,
		}
		if err = tx.Create(); err != nil {
			log.WithError(err).Fatal("insert tx")
		}

		log.WithFields(log.Fields{"hash": hash}).Info("Sent transaction of network stop")
	},
}
