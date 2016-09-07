package parser

import (
	"github.com/DayLightProject/go-daylight/packages/lib"
	"github.com/DayLightProject/go-daylight/packages/consts"
	"encoding/hex"
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
	err := p.ExecSql(`INSERT INTO full_nodes (wallet_id, host) VALUES (1,?)`, data.Host )//p.TxMaps.String["host"])
	if err != nil {
		return p.ErrInfo(err)
	}
	myAddress := hex.EncodeToString(lib.Address(data.PublicKey))//utils.HashSha1Hex(p.TxMaps.Bytes["public_key"]);
	err = p.ExecSql(`INSERT INTO dlt_wallets (wallet_id, address, host, addressVote, public_key_0, node_public_key, amount) VALUES (?, [hex], ?, [hex], [hex], [hex], ?)`, 
	    1, myAddress, data.Host, myAddress, hex.EncodeToString(data.PublicKey), hex.EncodeToString(data.NodePublicKey), consts.FIRST_DLT )
		//p.TxMaps.String["host"], myAddress, p.TxMaps.Bytes["public_key"], p.TxMaps.Bytes["node_public_key"], consts.FIRST_DLT)
	if err != nil {
		return p.ErrInfo(err)
	}

	return nil
}

func (p *Parser) FirstBlockRollback() error {
	return nil
}
