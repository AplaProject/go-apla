package custom

import (
	"errors"

	"github.com/GenesisKernel/go-genesis/packages/blockchain"
	"github.com/GenesisKernel/go-genesis/packages/conf/syspar"
	"github.com/GenesisKernel/go-genesis/packages/consts"
	"github.com/GenesisKernel/go-genesis/packages/service"
	"github.com/GenesisKernel/go-genesis/packages/utils"

	log "github.com/sirupsen/logrus"
)

var (
	messageNetworkStopping = "Attention! The network is stopped!"

	ErrNetworkStopping = errors.New("Network is stopping")
)

type StopNetworkTransaction struct {
	Logger *log.Entry
	Data   interface{}

	Cert *utils.Cert
}

func (t *StopNetworkTransaction) Init() error {
	return nil
}

func (t *StopNetworkTransaction) Validate() error {
	if err := t.validate(); err != nil {
		t.Logger.WithError(err).Error("validating tx")
		return err
	}

	return nil
}

func (t *StopNetworkTransaction) validate() error {
	data := t.Data.(*consts.StopNetwork)
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

	t.Cert = cert
	return nil
}

func (t *StopNetworkTransaction) Action() error {
	// Allow execute transaction, if the certificate was used
	if t.Cert.EqualBytes(consts.UsedStopNetworkCerts...) {
		return nil
	}

	// Set the node in a pause state
	service.PauseNodeActivity(service.PauseTypeStopingNetwork)

	t.Logger.Warn(messageNetworkStopping)
	return ErrNetworkStopping
}

func (t *StopNetworkTransaction) Rollback() error {
	return nil
}

func (t StopNetworkTransaction) Header() *blockchain.TxHeader {
	return nil
}
