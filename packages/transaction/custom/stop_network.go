// Copyright (C) 2017, 2018, 2019 EGAAS S.A.
//
// This program is free software; you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation; either version 2 of the License, or (at
// your option) any later version.
//
// This program is distributed in the hope that it will be useful, but
// WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the GNU
// General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program; if not, write to the Free Software
// Foundation, Inc., 51 Franklin Street, Fifth Floor, Boston, MA 02110-1301, USA.

package custom

import (
	"errors"

	"github.com/AplaProject/go-apla/packages/conf/syspar"
	"github.com/AplaProject/go-apla/packages/consts"
	"github.com/AplaProject/go-apla/packages/service"
	"github.com/AplaProject/go-apla/packages/utils"
	"github.com/AplaProject/go-apla/packages/utils/tx"

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

func (t StopNetworkTransaction) Header() *tx.Header {
	return nil
}
