package tcpclient

import (
	"net"

	"github.com/GenesisKernel/go-genesis/packages/consts"
	"github.com/GenesisKernel/go-genesis/packages/converter"
	log "github.com/sirupsen/logrus"
)

func (c *client) sendDisseminatorRequest(con net.Conn, requestType int, packet []byte) (err error) {
	/*
		Packet format:
		type  2 bytes
		len   4 bytes
		data  len bytes
	*/
	// type
	_, err = con.Write(converter.DecToBin(requestType, 2))
	if err != nil {
		c.WithFields(log.Fields{"type": consts.IOError, "error": err}).Error("writing request type to host")
		return err
	}

	// data size
	size := converter.DecToBin(len(packet), 4)
	_, err = con.Write(size)
	if err != nil {
		c.WithFields(log.Fields{"type": consts.IOError, "error": err}).Error("writing data size to host")
		return err
	}

	// data
	_, err = con.Write(packet)
	if err != nil {
		c.WithFields(log.Fields{"type": consts.IOError, "error": err}).Error("writing data to host")
		return err
	}

	return nil
}
