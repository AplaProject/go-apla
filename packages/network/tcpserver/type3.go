package tcpserver

import (
	"errors"
	"net"
	"time"

	"github.com/GenesisKernel/go-genesis/packages/conf"
	"github.com/GenesisKernel/go-genesis/packages/conf/syspar"
	"github.com/GenesisKernel/go-genesis/packages/consts"
	"github.com/GenesisKernel/go-genesis/packages/converter"
	"github.com/GenesisKernel/go-genesis/packages/crypto"
	"github.com/GenesisKernel/go-genesis/packages/model"
	"github.com/GenesisKernel/go-genesis/packages/network"
	"github.com/GenesisKernel/go-genesis/packages/utils"

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
