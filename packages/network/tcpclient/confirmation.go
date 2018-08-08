package tcpclient

import (
	"github.com/GenesisKernel/go-genesis/packages/consts"
	"github.com/GenesisKernel/go-genesis/packages/converter"
	"github.com/GenesisKernel/go-genesis/packages/network"
	log "github.com/sirupsen/logrus"
)

func CheckConfirmation(host string, blockID int64, logger *log.Entry) (hash string) {
	conn, err := newConnection(host)
	if err != nil {
		logger.WithFields(log.Fields{"type": consts.ConnectionError, "error": err, "host": host, "block_id": blockID}).Debug("dialing to host")
		return "0"
	}
	defer conn.Close()

	rt := &network.RequestType{Type: network.RequestTypeConfirmation}
	if err = rt.Write(conn); err != nil {
		logger.WithFields(log.Fields{"type": consts.IOError, "error": err, "host": host, "block_id": blockID}).Error("sending request type")
		return "0"
	}

	req := &network.ConfirmRequest{
		BlockID: uint32(blockID),
	}
	if err = req.Write(conn); err != nil {
		logger.WithFields(log.Fields{"type": consts.IOError, "error": err, "host": host, "block_id": blockID}).Error("sending confirmation request")
		return "0"
	}

	resp := &network.ConfirmResponse{}

	if err := resp.Read(conn); err != nil {
		logger.WithFields(log.Fields{"type": consts.IOError, "error": err, "host": host, "block_id": blockID}).Error("receiving confirmation response")
		return "0"
	}
	return string(converter.BinToHex(resp.Hash))
}
