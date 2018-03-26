package cmd

import (
	"errors"
	"io/ioutil"
	"path/filepath"

	"github.com/GenesisKernel/go-genesis/packages/conf"
	"github.com/GenesisKernel/go-genesis/packages/consts"

	"github.com/GenesisKernel/go-genesis/packages/converter"
	"github.com/GenesisKernel/go-genesis/packages/tcpserver"
	"github.com/GenesisKernel/go-genesis/packages/utils"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

const reqType = 3

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
			log.WithError(err).WithFields(log.Fields{"filepath": fp}).Fatal("Reading cert data")
		}

		req := &tcpserver.StopNetworkRequest{
			Data: stopNetworkCert,
		}

		errCount := 0
		for _, addr := range addrsForStopping {
			if err := sendStopNetworkCert(addr, req); err != nil {
				log.WithError(err).Errorf("Sending request to %s", addr)
				errCount++
				continue
			}

			log.Infof("Sending request to %s", addr)
		}

		log.Infof("Successful: %d, failed: %d", len(addrsForStopping)-errCount, errCount)
	},
}

func sendStopNetworkCert(addr string, req *tcpserver.StopNetworkRequest) error {
	conn, err := utils.TCPConn(addr)
	if err != nil {
		return err
	}
	defer conn.Close()

	if _, err = conn.Write(converter.DecToBin(reqType, 2)); err != nil {
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
