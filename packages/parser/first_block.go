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
	"github.com/EGaaS/go-egaas-mvp/packages/lib"
	//	b58 "github.com/jbenet/go-base58"
)

func (p *Parser) FirstBlockInit() error {
	/*	err := p.GetTxMaps([]map[string]string{{"public_key": "bytes"}, {"node_public_key": "bytes"}, {"host": "string"}})
		if err != nil {
			return p.ErrInfo(err)
		}
		p.TxMaps.Bytes["public_key"] = utils.BinToHex(p.TxMaps.Bytes["public_key"])
		p.TxMaps.Bytes["node_public_key"] = utils.BinToHex(p.TxMaps.Bytes["node_public_key"])*/
	return nil
}

func (p *Parser) FirstBlockFront() error {
	return nil
}

func (p *Parser) FirstBlock() error {

	data := p.TxPtr.(*consts.FirstBlock)
	//	myAddress := b58.Encode(lib.Address(data.PublicKey)) //utils.HashSha1Hex(p.TxMaps.Bytes["public_key"]);
	myAddress := int64(lib.Address(data.PublicKey))
	log.Debug("data.PublicKey %s", data.PublicKey)
	log.Debug("data.PublicKey %x", data.PublicKey)
	err := p.ExecSql(`INSERT INTO dlt_wallets (wallet_id, host, address_vote, public_key_0, node_public_key, amount) VALUES (?, ?, ?, [hex], [hex], ?)`,
		myAddress, data.Host, lib.AddressToString(uint64(myAddress)), hex.EncodeToString(data.PublicKey), hex.EncodeToString(data.NodePublicKey), consts.FIRST_QDLT)
	//p.TxMaps.String["host"], myAddress, p.TxMaps.Bytes["public_key"], p.TxMaps.Bytes["node_public_key"], consts.FIRST_DLT)
	if err != nil {
		return p.ErrInfo(err)
	}
	err = p.ExecSql(`INSERT INTO full_nodes (wallet_id, host) VALUES (?,?)`, myAddress, data.Host) //p.TxMaps.String["host"])
	if err != nil {
		return p.ErrInfo(err)
	}

	return nil
}

func (p *Parser) FirstBlockRollback() error {
	return nil
}
