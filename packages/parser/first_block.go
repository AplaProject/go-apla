// Copyright 2016 The go-daylight Authors
// This file is part of the go-daylight library.
//
// The go-daylight library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-daylight library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-daylight library. If not, see <http://www.gnu.org/licenses/>.

package parser

import (
	"encoding/hex"

	"github.com/EGaaS/go-egaas-mvp/packages/consts"
	"github.com/EGaaS/go-egaas-mvp/packages/converter"
	"github.com/EGaaS/go-egaas-mvp/packages/crypto"
	"github.com/EGaaS/go-egaas-mvp/packages/model"
	"github.com/EGaaS/go-egaas-mvp/packages/utils/tx"

	"github.com/shopspring/decimal"
)

type FirstBlockParser struct {
	*Parser
}

func (p *FirstBlockParser) Init() error {
	return nil
}

func (p *FirstBlockParser) Validate() error {
	return nil
}

func (p *FirstBlockParser) Action() error {
	data := p.TxPtr.(*consts.FirstBlock)
	//	myAddress := b58.Encode(lib.Address(data.PublicKey)) //utils.HashSha1Hex(p.TxMaps.Bytes["public_key"]);
	myAddress := crypto.Address(data.PublicKey)
	log.Debug("data.PublicKey %s", data.PublicKey)
	log.Debug("data.PublicKey %x", data.PublicKey)
	dltWallet := &model.DltWallet{
		WalletID:      myAddress,
		Host:          data.Host,
		AddressVote:   converter.AddressToString(myAddress),
		PublicKey:     []byte(hex.EncodeToString(data.PublicKey)),
		NodePublicKey: []byte(hex.EncodeToString(data.NodePublicKey)),
		Amount:        decimal.NewFromFloat(consts.FIRST_QDLT).String(),
	}
	err := dltWallet.Create()
	if err != nil {
		return p.ErrInfo(err)
	}
	fullNode := &model.FullNode{WalletID: myAddress, Host: data.Host}
	err = fullNode.Create()
	if err != nil {
		return p.ErrInfo(err)
	}

	return nil
}

func (p *FirstBlockParser) Rollback() error {
	return nil
}

func (p FirstBlockParser) Header() *tx.Header {
	return nil
}
