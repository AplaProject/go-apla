package parser

import (
	"errors"

	"github.com/GenesisKernel/go-genesis/packages/conf/syspar"
	"github.com/GenesisKernel/go-genesis/packages/consts"
	"github.com/GenesisKernel/go-genesis/packages/service"
	"github.com/GenesisKernel/go-genesis/packages/utils"
	"github.com/GenesisKernel/go-genesis/packages/utils/tx"
)

var (
	messageNetworkStopping = "Attention! The network is stopped!"

	errNetworkStopping = errors.New("Network is stopping")
)

type StopNetworkTransaction struct {
	*Transaction

	cert *utils.Cert
}

func (t *StopNetworkTransaction) Init() error {
	return nil
}

func (t *StopNetworkTransaction) Validate() error {
	if err := t.validate(); err != nil {
		t.GetLogger().WithError(err).Error("validating tx")
		return err
	}

	return nil
}

func (t *StopNetworkTransaction) validate() error {
	data := t.TxPtr.(*consts.StopNetwork)

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

	t.cert = cert
	return nil
}

func (t *StopNetworkTransaction) Action() error {
	// Allow execute transaction, if the certificate was used
	if t.cert.EqualBytes(consts.UsedStopNetworkCerts...) {
		return nil
	}

	// Set the node in a pause state
	service.PauseNodeActivity(service.PauseTypeStopingNetwork)

	t.GetLogger().Warn(messageNetworkStopping)
	return errNetworkStopping
}

func (t *StopNetworkTransaction) Rollback() error {
	return nil
}

func (t StopNetworkTransaction) Header() *tx.Header {
	return nil
}
