package cmd

import (
	"errors"
	"io/ioutil"
	"path/filepath"

	"github.com/GenesisKernel/go-genesis/packages/conf"
	"github.com/GenesisKernel/go-genesis/packages/consts"
	"github.com/GenesisKernel/go-genesis/packages/tcpserver"
	"github.com/GenesisKernel/go-genesis/packages/utils"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	addrsForStopping        []string
	stopNetworkCertFilepath string

	errNotAccepted = errors.New("Not accepted")
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

		req := &tcpserver.StopNetworkRequest{
			Data: stopNetworkCert,
		}

		errCount := 0
		for _, addr := range addrsForStopping {
			if err := sendStopNetworkCert(addr, req); err != nil {
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

func sendStopNetworkCert(addr string, req *tcpserver.StopNetworkRequest) error {
	conn, err := utils.TCPConn(addr)
	if err != nil {
		return err
	}
	defer conn.Close()

	if err = tcpserver.SendRequestType(tcpserver.RequestTypeStopNetwork, conn); err != nil {
		return err
	}

	if err = tcpserver.SendRequest(req, conn); err != nil {
		return err
	}

	res := &tcpserver.StopNetworkResponse{}
	if err = tcpserver.ReadRequest(res, conn); err != nil {
		return err
	}

	if len(res.Hash) != consts.HashSize {
		return errNotAccepted
	}

	return nil
}

func init() {
	stopNetworkCmd.Flags().StringVar(&stopNetworkCertFilepath, "stopNetworkCert", "", "Filepath to certificate for network stopping")
	stopNetworkCmd.Flags().StringArrayVar(&addrsForStopping, "addr", []string{}, "Node address")
	stopNetworkCmd.MarkFlagRequired("stopNetworkCert")
	stopNetworkCmd.MarkFlagRequired("addr")
}
