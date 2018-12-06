package tcpclient

import (
	"github.com/AplaProject/go-apla/packages/consts"
	"github.com/AplaProject/go-apla/packages/converter"
	"github.com/AplaProject/go-apla/packages/network"
	log "github.com/sirupsen/logrus"
)

func CheckConfirmation(host string, blockHash []byte, logger *log.Entry) (hash string) {
	conn, err := newConnection(host)
	if err != nil {
		logger.WithFields(log.Fields{"type": consts.ConnectionError, "error": err, "host": host, "block_hash": blockHash}).Debug("dialing to host")
		return "0"
	}
	defer conn.Close()

	rt := &network.RequestType{Type: network.RequestTypeConfirmation}
	if err = rt.Write(conn); err != nil {
		logger.WithFields(log.Fields{"type": consts.IOError, "error": err, "host": host, "block_hash": blockHash}).Error("sending request type")
		return "0"
	}

	req := &network.ConfirmRequest{
		BlockHash: blockHash,
	}
	if err = req.Write(conn); err != nil {
		logger.WithFields(log.Fields{"type": consts.IOError, "error": err, "host": host, "block_hash": blockHash}).Error("sending confirmation request")
		return "0"
	}

	resp := &network.ConfirmResponse{}

	if err := resp.Read(conn); err != nil {
		logger.WithFields(log.Fields{"type": consts.IOError, "error": err, "host": host, "block_hash": blockHash}).Error("receiving confirmation response")
		return "0"
	}
	return string(converter.BinToHex(resp.Hash))
}
