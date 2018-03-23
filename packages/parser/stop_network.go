package parser

import (
	"crypto/x509"
	"encoding/pem"
	"errors"
	"os"

	"github.com/GenesisKernel/go-genesis/packages/consts"
	"github.com/GenesisKernel/go-genesis/packages/utils/tx"
)

var (
	errParseNetworkStopCert     = errors.New("Failed to parse certificate of network stop")
	errParseNetworkStopRootCert = errors.New("Failed to parse root certificate of network stop")
)

type StopNetworkParser struct {
	*Parser
}

func (p *StopNetworkParser) Init() error {
	return nil
}

func (p *StopNetworkParser) Validate() error {
	if err := p.validate(); err != nil {
		p.GetLogger().Error(err)
		return err
	}

	return nil
}

func (p *StopNetworkParser) validate() error {
	data := p.TxPtr.(*consts.StopNetwork)

	block, _ := pem.Decode(data.StopNetworkCert)
	if block == nil {
		return errParseNetworkStopCert
	}

	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return err
	}

	roots := x509.NewCertPool()
	if ok := roots.AppendCertsFromPEM([]byte("")); !ok {
		return errParseNetworkStopRootCert
	}

	if _, err := cert.Verify(x509.VerifyOptions{Roots: roots}); err != nil {
		return err
	}

	return nil
}

func (p *StopNetworkParser) Action() error {
	p.GetLogger().Info("Attention! The network is stopped!")
	os.Exit(0)
	return nil
}

func (p *StopNetworkParser) Rollback() error {
	return nil
}

func (p StopNetworkParser) Header() *tx.Header {
	return nil
}
