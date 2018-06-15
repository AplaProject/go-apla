package tcpclient

import (
	"fmt"
	"net"
	"strings"
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
