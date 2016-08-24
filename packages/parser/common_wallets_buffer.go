package parser

func (p *Parser) getWalletsBufferAmount() (int64, error) {
	return p.Single("SELECT amount FROM dlt_wallets_buffer WHERE wallet_id = ? AND del_block_id = 0", p.TxWalletID).Int64()
}

func (p *Parser) updateWalletsBuffer(amount int64) error {
	// добавим нашу сумму в буфер кошельков, чтобы юзер не смог послать запрос на вывод всех DC с кошелька.
	hash, err := p.Single("SELECT hash FROM dlt_wallets_buffer WHERE hex(hash) = ?", p.TxHash).String()
	if len(hash) > 0 {
		err = p.ExecSql("UPDATE dlt_wallets_buffer SET wallet_id = ?, amount = ? WHERE hex(hash) = ?", p.TxWalletID, amount, p.TxHash)
	} else {
		err = p.ExecSql("INSERT INTO dlt_wallets_buffer ( hash, wallet_id, amount ) VALUES ( [hex], ?, ? )", p.TxHash, p.TxWalletID, amount)
	}
	if err != nil {
		return p.ErrInfo(err)
	}
	return nil
}