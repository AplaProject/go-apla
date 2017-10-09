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

	"github.com/EGaaS/go-egaas-mvp/packages/config/syspar"
	"github.com/EGaaS/go-egaas-mvp/packages/consts"
	"github.com/EGaaS/go-egaas-mvp/packages/converter"
	"github.com/EGaaS/go-egaas-mvp/packages/crypto"
	"github.com/EGaaS/go-egaas-mvp/packages/model"
	"github.com/EGaaS/go-egaas-mvp/packages/template"
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

	err := model.ExecSchemaEcosystem(1, myAddress, ``)
	if err != nil {
		return p.ErrInfo(err)
	}
	key := &model.Key{
		ID:        myAddress,
		PublicKey: data.PublicKey,
		Amount:    decimal.NewFromFloat(consts.FIRST_QDLT).String(),
	}
	if err = key.SetTablePrefix(consts.MainEco).Create(); err != nil {
		return p.ErrInfo(err)
	}
	err = model.DBConn.Exec(`insert into "1_pages" (id,name,menu,value,conditions) values('1', 'default_page',
		  'default_menu', ?, 'ContractAccess("@1EditPage")')`, syspar.SysString(`default_ecosystem_page`)).Error
	if err != nil {
		return p.ErrInfo(err)
	}
	err = model.DBConn.Exec(`insert into "1_menu" (id,name,value,conditions) values('1', 'default_menu', ?, 'ContractAccess("@1EditMenu")')`,
		syspar.SysString(`default_ecosystem_menu`)).Error
	if err != nil {
		return p.ErrInfo(err)
	}
	err = template.LoadContract(`1`)
	if err != nil {
		return p.ErrInfo(err)
	}
	node := &model.SystemParameterV2{Name: `full_nodes`}
	if err = node.SaveArray([][]string{{data.Host, converter.Int64ToStr(myAddress),
		hex.EncodeToString(data.NodePublicKey)}}); err != nil {
		return p.ErrInfo(err)
	}
	syspar.SysUpdate()
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
