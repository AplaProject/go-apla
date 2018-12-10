package cmd

import (
	"io/ioutil"
	"path/filepath"

	"github.com/AplaProject/go-apla/packages/conf"
	"github.com/AplaProject/go-apla/packages/consts"
	"github.com/AplaProject/go-apla/packages/network"
	"github.com/AplaProject/go-apla/packages/network/tcpclient"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	addrsForStopping        []string
	stopNetworkCertFilepath string
)

// stopNetworkCmd represents the stopNetworkCmd command
var stopNetworkCmd = &cobra.Command{
	Use:    "stopNetwork",
	Short:  "Sending a special transaction to stop the network",
	PreRun: loadConfigWKey,
	Run: func(cmd *cobra.Command, args []string) {
		fp := filepath.Join(conf.Config.KeysDir, stopNetworkCertFilepath)
		stopNetworkCert, err := ioutil.ReadFile(fp)
		if err != nil {
			log.WithFields(log.Fields{"error": err, "type": consts.IOError, "filepath": fp}).Fatal("Reading cert data")
		}

		req := &network.StopNetworkRequest{
			Data: stopNetworkCert,
		}

		errCount := 0
		for _, addr := range addrsForStopping {
			if err := tcpclient.SendStopNetwork(addr, req); err != nil {
				log.WithFields(log.Fields{"error": err, "type": consts.NetworkError, "addr": addr}).Errorf("Sending request")
				errCount++
				continue
			}

			log.WithFields(log.Fields{"addr": addr}).Info("Sending request")
		}

		log.WithFields(log.Fields{
			"successful": len(addrsForStopping) - errCount,
			"failed":     errCount,
		}).Info("Complete")
	},
}

func init() {
	stopNetworkCmd.Flags().StringVar(&stopNetworkCertFilepath, "stopNetworkCert", "", "Filepath to certificate for network stopping")
	stopNetworkCmd.Flags().StringArrayVar(&addrsForStopping, "addr", []string{}, "Node address")
	stopNetworkCmd.MarkFlagRequired("stopNetworkCert")
	stopNetworkCmd.MarkFlagRequired("addr")
}
