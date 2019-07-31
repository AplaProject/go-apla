// Copyright (C) 2017, 2018, 2019 EGAAS S.A.
//
// This program is free software; you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation; either version 2 of the License, or (at
// your option) any later version.
//
// This program is distributed in the hope that it will be useful, but
// WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the GNU
// General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program; if not, write to the Free Software
// Foundation, Inc., 51 Franklin Street, Fifth Floor, Boston, MA 02110-1301, USA.

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
