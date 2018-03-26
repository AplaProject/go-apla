package parser

import (
	"crypto/x509"
	"encoding/pem"
	"errors"

	"github.com/GenesisKernel/go-genesis/packages/consts"
	"github.com/GenesisKernel/go-genesis/packages/service"
	"github.com/GenesisKernel/go-genesis/packages/utils/tx"
)

var (
	errParseNetworkStopCert     = errors.New("Failed to parse certificate of network stop")
	errParseNetworkStopRootCert = errors.New("Failed to parse root certificate of network stop")
	errFirstBlockData           = errors.New("Failed to get data of the first block")
	errNetworkStopping          = errors.New("Network stopping")
)

type StopNetworkParser struct {
	*Parser

	cert *x509.Certificate
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

	cert, err := parseCert(data.StopNetworkCert)
	if err != nil {
		return err
	}
	p.cert = cert

	firstBlockData, ok := GetDataFromFirstBlock()
	if !ok {
		return errFirstBlockData
	}

	roots := x509.NewCertPool()
	if ok := roots.AppendCertsFromPEM(firstBlockData.StopNetworkCertBundle); !ok {
		return errParseNetworkStopRootCert
	}

	if _, err := cert.Verify(x509.VerifyOptions{Roots: roots}); err != nil {
		return err
	}

	return nil
}

func (p *StopNetworkParser) Action() error {
	// Allow execute transaction, If the certificate was used
	if isUsedCert(p.cert) {
		return nil
	}

	// Set the node in a pause state
	p.GetLogger().Warn("Attention! The network is stopped!")
	service.PauseNodeActivity(service.PauseTypeStopingNetwork)
	return errNetworkStopping
}

func (p *StopNetworkParser) Rollback() error {
	return nil
}

func (p StopNetworkParser) Header() *tx.Header {
	return nil
}

func isUsedCert(cert *x509.Certificate) bool {
	for _, v := range consts.UsedStopNetworkCerts {
		usedCert, _ := parseCert(v)
		if cert.Equal(usedCert) {
			return true
		}
	}

	return false
}

func parseCert(b []byte) (*x509.Certificate, error) {
	block, _ := pem.Decode(b)
	if block == nil {
		return nil, errParseNetworkStopCert
	}

	return x509.ParseCertificate(block.Bytes)
}
