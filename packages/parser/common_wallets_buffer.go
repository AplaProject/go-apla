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