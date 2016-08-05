package dcparser

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

	err = p.ExecSql(`INSERT INTO wallets (wallet_id, host, vote, public_key_0, node_public_key) VALUES (?, ?, ?, [hex], [hex])`, 1, p.TxMaps.String["host"], 1, p.TxMaps.Bytes["public_key"], p.TxMaps.Bytes["node_public_key"])
	if err != nil {
		return p.ErrInfo(err)
	}

	return nil
}

func (p *Parser) FirstBlockRollback() error {
	return nil
}
