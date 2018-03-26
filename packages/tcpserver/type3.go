package tcpserver

import (
	"net"
	"time"

	"github.com/GenesisKernel/go-genesis/packages/conf"
	"github.com/GenesisKernel/go-genesis/packages/consts"
	"github.com/GenesisKernel/go-genesis/packages/converter"
	"github.com/GenesisKernel/go-genesis/packages/crypto"
	"github.com/GenesisKernel/go-genesis/packages/model"

	log "github.com/sirupsen/logrus"
)

// Type3
func Type3(req *StopNetworkRequest, w net.Conn) error {
	// TODO: validate cert

	var data []byte
	_, err := converter.BinMarshal(&data,
		&consts.StopNetwork{
			TxHeader: consts.TxHeader{
				Type:  consts.TxTypeStopNetwork,
				Time:  uint32(time.Now().Unix()),
				KeyID: conf.Config.KeyID,
			},
			StopNetworkCert: req.Data,
		},
	)
	if err != nil {
		log.WithError(err).Error("binary marshaling")
		return err
	}

	hash, err := crypto.Hash(data)
	if err != nil {
		log.WithError(err).Error("hashing data")
		return err
	}

	tx := &model.Transaction{
		Hash:     hash,
		Data:     data,
		Type:     consts.TxTypeStopNetwork,
		KeyID:    conf.Config.KeyID,
		HighRate: model.TransactionRateStopNetwork,
	}
	if err = tx.Create(); err != nil {
		log.WithError(err).Error("inserting tx to database")
		return err
	}

	res := &StopNetworkResponse{hash}
	if err = SendRequest(res, w); err != nil {
		log.WithError(err).Error("sending response")
		return err
	}

	w.Read(make([]byte, 1))

	return nil
}
