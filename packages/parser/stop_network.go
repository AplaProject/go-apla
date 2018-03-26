package parser

import (
	"errors"

	"github.com/GenesisKernel/go-genesis/packages/conf/syspar"
	"github.com/GenesisKernel/go-genesis/packages/consts"
	"github.com/GenesisKernel/go-genesis/packages/notificator"
	"github.com/GenesisKernel/go-genesis/packages/service"
	"github.com/GenesisKernel/go-genesis/packages/utils"
	"github.com/GenesisKernel/go-genesis/packages/utils/tx"
)

var (
	messageNetworkStopping = "Attention! The network is stopped!"

	errNetworkStopping = errors.New("Network is stopping")
)

type StopNetworkParser struct {
	*Parser

	cert *utils.Cert
}

func (p *StopNetworkParser) Init() error {
	return nil
}

func (p *StopNetworkParser) Validate() error {
	if err := p.validate(); err != nil {
		p.GetLogger().WithError(err).Error("validating tx")
		return err
	}

	return nil
}

func (p *StopNetworkParser) validate() error {
	data := p.TxPtr.(*consts.StopNetwork)

	cert, err := utils.ParseCert(data.StopNetworkCert)
	if err != nil {
		return err
	}

	fbdata, err := syspar.GetFirstBlockData()
	if err != nil {
		return err
	}

	if err = cert.Validate(fbdata.StopNetworkCertBundle); err != nil {
		return err
	}

	p.cert = cert
	return nil
}

func (p *StopNetworkParser) Action() error {
	// Allow execute transaction, If the certificate was used
	if isUsedCert(p.cert) {
		return nil
	}

	// Set the node in a pause state
	service.PauseNodeActivity(service.PauseTypeStopingNetwork)

	notificator.BroadcastMessage(map[string]string{
		"notification_type": "system",
		"body_text":         messageNetworkStopping,
	})

	p.GetLogger().Warn(messageNetworkStopping)
	return errNetworkStopping
}

func (p *StopNetworkParser) Rollback() error {
	return nil
}

func (p StopNetworkParser) Header() *tx.Header {
	return nil
}

func isUsedCert(cert *utils.Cert) bool {
	for _, v := range consts.UsedStopNetworkCerts {
		if cert.EqualBytes(v) {
			return true
		}
	}

	return false
}
