package tcpclient

import (
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/GenesisKernel/go-genesis/packages/consts"
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

func newConnection(addr string) (net.Conn, error) {
	host, err := NormalizeHostAddress(addr, consts.DEFAULT_TCP_PORT)
	if err != nil {
		log.WithFields(log.Fields{"type": consts.NetworkError, "host": addr, "error": err}).Error("on normalize host address")
		return nil, err
	}

	conn, err := net.DialTimeout("tcp", host, consts.TCPConnTimeout)
	if err != nil {
		log.WithFields(log.Fields{"type": consts.ConnectionError, "error": err, "address": host}).Debug("dialing tcp")
		return nil, err
	}

	conn.SetReadDeadline(time.Now().Add(consts.READ_TIMEOUT))
	conn.SetWriteDeadline(time.Now().Add(consts.WRITE_TIMEOUT))
	return conn, nil
}
