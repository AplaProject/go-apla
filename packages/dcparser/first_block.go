package dcparser

import (
	"encoding/json"
)

func (p *Parser) FirstBlockInit() error {
	fields := []string{"data"}
	TxMap := make(map[string][]byte)
	TxMap, err := p.GetTxMap(fields)
	p.TxMap = TxMap
	if err != nil {
		return err
	}
	return nil
}

func (p *Parser) FirstBlockFront() error {
	return nil
}

type firstBlock struct {
	PublicKey          string                 `json:"public_key"`
	NodePublicKey      string                 `json:"node_public_key"`
	Host               string                 `json:"host"`
}

func (p *Parser) FirstBlock() error {
	var firstBlock firstBlock
	err := json.Unmarshal(p.TxMap["data"], &firstBlock)
	if err != nil {
		return err
	}

	err = p.ExecSql(`INSERT INTO full_nodes (wallet_id, host) VALUES (1,?)`, firstBlock.Host)
	if err != nil {
		return p.ErrInfo(err)
	}

	err = p.ExecSql(`INSERT INTO wallets (wallet_id, host, vote, public_key_0, node_public_key) VALUES (?, ?, ?, [hex], [hex])`, 1, firstBlock.Host, 1, firstBlock.PublicKey, firstBlock.NodePublicKey)
	if err != nil {
		return p.ErrInfo(err)
	}

	return nil
}

func (p *Parser) FirstBlockRollback() error {
	return nil
}
