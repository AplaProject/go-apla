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

package tcpserver

import (
	"errors"
	"net"
	"time"

	"github.com/AplaProject/go-apla/packages/conf"
	"github.com/AplaProject/go-apla/packages/conf/syspar"
	"github.com/AplaProject/go-apla/packages/consts"
	"github.com/AplaProject/go-apla/packages/converter"
	"github.com/AplaProject/go-apla/packages/crypto"
	"github.com/AplaProject/go-apla/packages/model"
	"github.com/AplaProject/go-apla/packages/network"
	"github.com/AplaProject/go-apla/packages/utils"

	log "github.com/sirupsen/logrus"
)

var errStopCertAlreadyUsed = errors.New("Stop certificate is already used")

// Type3
func Type3(req *network.StopNetworkRequest, w net.Conn) error {
	hash, err := processStopNetwork(req.Data)
	if err != nil {
		return err
	}

	res := &network.StopNetworkResponse{hash}
	if err = res.Write(w); err != nil {
		log.WithFields(log.Fields{"error": err, "type": consts.NetworkError}).Error("sending response")
		return err
	}

	return nil
}

func processStopNetwork(b []byte) ([]byte, error) {
	cert, err := utils.ParseCert(b)
	if err != nil {
		log.WithFields(log.Fields{"error": err, "type": consts.ParseError}).Error("parsing cert")
		return nil, err
	}

	if cert.EqualBytes(consts.UsedStopNetworkCerts...) {
		log.WithFields(log.Fields{"error": errStopCertAlreadyUsed, "type": consts.InvalidObject}).Error("checking cert")
		return nil, errStopCertAlreadyUsed
	}

	fbdata, err := syspar.GetFirstBlockData()
	if err != nil {
		log.WithFields(log.Fields{"error": err, "type": consts.ConfigError}).Error("getting data of first block")
		return nil, err
	}

	if err = cert.Validate(fbdata.StopNetworkCertBundle); err != nil {
		log.WithFields(log.Fields{"error": err, "type": consts.InvalidObject}).Error("validating cert")
		return nil, err
	}

	var data []byte
	_, err = converter.BinMarshal(&data,
		&consts.StopNetwork{
			TxHeader: consts.TxHeader{
				Type:  consts.TxTypeStopNetwork,
				Time:  uint32(time.Now().Unix()),
				KeyID: conf.Config.KeyID,
			},
			StopNetworkCert: b,
		},
	)
	if err != nil {
		log.WithFields(log.Fields{"error": err, "type": consts.MarshallingError}).Error("binary marshaling")
		return nil, err
	}

	hash, err := crypto.Hash(data)
	if err != nil {
		log.WithFields(log.Fields{"error": err, "type": consts.CryptoError}).Error("hashing data")
		return nil, err
	}

	tx := &model.Transaction{
		Hash:     hash,
		Data:     data,
		Type:     consts.TxTypeStopNetwork,
		KeyID:    conf.Config.KeyID,
		HighRate: model.TransactionRateStopNetwork,
	}
	if err = tx.Create(); err != nil {
		log.WithFields(log.Fields{"error": err, "type": consts.DBError}).Error("inserting tx to database")
		return nil, err
	}

	return hash, nil
}
