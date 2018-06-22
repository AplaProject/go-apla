package tcpclient

import (
	"fmt"
	"net"
	"strings"

	"github.com/GenesisKernel/go-genesis/packages/consts"
	"github.com/GenesisKernel/go-genesis/packages/converter"
	log "github.com/sirupsen/logrus"
)

// NormalizeHostAddress get address. if port not defined returns combined string with ip and defaultPort
func NormalizeHostAddress(address string, defaultPort int64) (string, error) {

	_, _, err := net.SplitHostPort(address)
	if err != nil {
		if strings.HasSuffix(err.Error(), "missing port in address") {
			return fmt.Sprintf("%s:%d", address, defaultPort), nil
		}

		return "", err
	}

	return address, nil
}

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
