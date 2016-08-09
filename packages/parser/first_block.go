package parser

import "github.com/DayLightProject/go-daylight/packages/utils"


func (p *Parser) FirstBlockInit() error {
	err := p.GetTxMaps([]map[string]string{{"public_key": "bytes"}, {"node_public_key": "bytes"}, {"host": "string"}})
	if err != nil {
		return p.ErrInfo(err)
	}
	p.TxMaps.Bytes["public_key"] = utils.BinToHex(p.TxMaps.Bytes["public_key"])
	p.TxMaps.Bytes["node_public_key"] = utils.BinToHex(p.TxMaps.Bytes["node_public_key"])
	return nil
}

func (p *Parser) FirstBlockFront() error {
	return nil
}


func (p *Parser) FirstBlock() error {

	err := p.ExecSql(`INSERT INTO full_nodes (wallet_id, host) VALUES (1,?)`, p.TxMaps.String["host"])
	if err != nil {
		return p.ErrInfo(err)
	}

	myAddress := utils.HashSha1Hex(p.TxMaps.Bytes["public_key"]);
	err = p.ExecSql(`INSERT INTO dlt_wallets (wallet_id, address, host, vote, public_key_0, node_public_key) VALUES (?, [hex], ?, [hex], [hex], [hex])`, 1, myAddress, p.TxMaps.String["host"], myAddress, p.TxMaps.Bytes["public_key"], p.TxMaps.Bytes["node_public_key"])
	if err != nil {
		return p.ErrInfo(err)
	}

	return nil
}

func (p *Parser) FirstBlockRollback() error {
	return nil
}
