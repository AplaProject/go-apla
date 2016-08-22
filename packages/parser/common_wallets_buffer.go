package parser

import (
	"github.com/DayLightProject/go-daylight/packages/utils"
)

func (p *Parser) getWalletsBufferAmount(currencyId int64) (float64, error) {
	return p.Single("SELECT sum(amount) FROM dlt_wallets_buffer WHERE user_id = ? AND currency_id = ? AND del_block_id = 0", p.TxUserID, currencyId).Float64()
}

func (p *Parser) updateWalletsBuffer(amount float64, currencyId int64) error {
	// добавим нашу сумму в буфер кошельков, чтобы юзер не смог послать запрос на вывод всех DC с кошелька.
	hash, err := p.Single("SELECT hash FROM dlt_wallets_buffer WHERE hex(hash) = ?", p.TxHash).String()
	if len(hash) > 0 {
		err = p.ExecSql("UPDATE wallets_buffer SET user_id = ?, currency_id = ?, amount = ? WHERE hex(hash) = ?", p.TxUserID, currencyId, utils.Round(amount, 2), p.TxHash)
	} else {
		err = p.ExecSql("INSERT INTO wallets_buffer ( hash, user_id, currency_id, amount ) VALUES ( [hex], ?, ?, ? )", p.TxHash, p.TxUserID, currencyId, utils.Round(amount, 2))
	}
	if err != nil {
		return p.ErrInfo(err)
	}
	return nil
}
