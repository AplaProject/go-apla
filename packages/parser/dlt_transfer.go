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
	"fmt"

	"github.com/EGaaS/go-egaas-mvp/packages/lib"
	"github.com/EGaaS/go-egaas-mvp/packages/utils"
	"github.com/EGaaS/go-egaas-mvp/packages/utils/tx"

	"github.com/shopspring/decimal"
	"gopkg.in/vmihailenco/msgpack.v2"
)

type DLTTransferParser struct {
	*Parser
	DLTTransfer *tx.DLTTransfer
}

func (p *DLTTransferParser) Init() error {
	dltTransfer := &tx.DLTTransfer{}
	if err := msgpack.Unmarshal(p.BinaryData, dltTransfer); err != nil {
		return p.ErrInfo(err)
	}
	p.DLTTransfer = dltTransfer
	p.DLTTransfer.PublicKey = utils.BinToHex(p.DLTTransfer.Header.PublicKey)
	return nil
}

func (p *DLTTransferParser) Validate() error {
	err := p.generalCheck(`dlt_transfer`, &p.DLTTransfer.Header)
	if err != nil {
		return p.ErrInfo(err)
	}

	verifyData := map[string]string{"walletAddress": "walletAddress", "amount": "decimal", "commission": "decimal", "comment": "comment"}
	err = p.CheckInputData(verifyData)
	if err != nil {
		return p.ErrInfo(err)
	}

	// public key need only when we don't have public_key in the dlt_wallets table
	public_key_0, err := p.Single(`SELECT public_key_0 FROM dlt_wallets WHERE wallet_id = ?`, p.TxWalletID).String()
	if err != nil {
		return p.ErrInfo(err)
	}
	if len(public_key_0) == 0 {
		bkey, err := hex.DecodeString(string(p.DLTTransfer.Header.PublicKey))
		if err != nil {
			return p.ErrInfo(err)
		}
		if lib.KeyToAddress(bkey) != lib.AddressToString(p.TxWalletID) {
			return p.ErrInfo("incorrect public_key")
		}
	}

	zero, _ := decimal.NewFromString("0")
	if p.TxMaps.Decimal["amount"].Cmp(zero) <= 0 {
		return p.ErrInfo("amount<=0")
	}

	fPrice, err := p.Single(`SELECT value->'dlt_transfer' FROM system_parameters WHERE name = ?`, "op_price").String()
	if err != nil {
		return p.ErrInfo(err)
	}

	fuelRate := p.GetFuel()
	if fuelRate.Cmp(decimal.New(0, 0)) <= 0 {
		return fmt.Errorf(`fuel rate must be greater than 0`)
	}
	// 1 000 000 000 000 000 000 qDLT = 1 DLT * 100 000 000
	// fuelRate = 1 000 000 000 000 000
	//
	fPriceDecemal, err := decimal.NewFromString(fPrice)
	if err != nil {
		return p.ErrInfo(err)
	}
	commission := fPriceDecemal.Mul(fuelRate)
	ourCommission, err := decimal.NewFromString(fPrice)
	if err != nil {
		return p.ErrInfo(err)
	}

	// проверим, удовлетворяет ли нас комиссия, которую предлагает юзер
	if ourCommission.Cmp(commission) < 0 {
		return p.ErrInfo(fmt.Sprintf("commission %s < dltPrice %d", ourCommission.String(), commission))
	}

	if p.DLTTransfer.Comment == "null" {
		p.DLTTransfer.Comment = ""
	}

	CheckSignResult, err := utils.CheckSign(p.PublicKeys, p.DLTTransfer.ForSign(), p.TxMap["sign"], false)
	if err != nil {
		return p.ErrInfo(err)
	}
	if !CheckSignResult {
		return p.ErrInfo("incorrect sign OOPS")
	}

	totalAmount, err := p.Single(`SELECT amount FROM dlt_wallets WHERE wallet_id = ?`, p.TxWalletID).String()
	if err != nil {
		return p.ErrInfo(err)
	}
	totalAmountDecimal, err := decimal.NewFromString(totalAmount)
	if err != nil {
		return p.ErrInfo(err)
	}
	ourAmount, err := decimal.NewFromString(p.DLTTransfer.Amount)
	if err != nil {
		return p.ErrInfo(err)
	}
	if totalAmountDecimal.Cmp(ourAmount.Add(ourCommission)) < 0 {
		return p.ErrInfo(fmt.Sprintf("%s + %s < %s)", ourAmount, ourCommission, totalAmount))
	}
	return nil
}

func (p *DLTTransferParser) Action() error {
	log.Debug("wallet address %s", p.DLTTransfer.WalletAddress)
	address := lib.StringToAddress(p.DLTTransfer.WalletAddress)
	walletId, err := p.Single(`SELECT wallet_id FROM dlt_wallets WHERE wallet_id = ?`, address).Int64()
	if err != nil {
		return p.ErrInfo(err)
	}
	log.Debug("walletId %d", walletId)
	//if walletId > 0 {
	pkey, err := p.Single(`SELECT public_key_0 FROM dlt_wallets WHERE wallet_id = ?`, p.TxWalletID).String()
	if err != nil {
		return p.ErrInfo(err)
	}
	log.Debug("pkey %x", pkey)

	log.Debug("amount %s", p.DLTTransfer.Amount)
	log.Debug("commission %s", p.DLTTransfer.Commission)
	amount, err := decimal.NewFromString(p.DLTTransfer.Amount)
	if err != nil {
		return p.ErrInfo(err)
	}
	commission, err := decimal.NewFromString(p.DLTTransfer.Commission)
	if err != nil {
		return p.ErrInfo(err)
	}
	amountAndCommission := amount.Add(commission)
	if err != nil {
		return p.ErrInfo(err)
	}
	log.Debug("amountAndCommission %s", amountAndCommission)
	log.Debug("amountAndCommission %s", amountAndCommission.String())
	if len(p.DLTTransfer.Header.PublicKey) > 30 && len(pkey) == 0 {
		_, err = p.selectiveLoggingAndUpd([]string{"-amount", "public_key_0"}, []interface{}{amountAndCommission.String(), utils.HexToBin(p.DLTTransfer.PublicKey)}, "dlt_wallets", []string{"wallet_id"}, []string{utils.Int64ToStr(p.TxWalletID)}, true)
	} else {
		_, err = p.selectiveLoggingAndUpd([]string{"-amount"}, []interface{}{amountAndCommission.String()}, "dlt_wallets", []string{"wallet_id"}, []string{utils.Int64ToStr(p.TxWalletID)}, true)
	}
	if err != nil {
		return p.ErrInfo(err)
	}

	if walletId == 0 {
		log.Debug("walletId == 0")
		log.Debug("%s", string(p.DLTTransfer.WalletAddress))
		walletId = lib.StringToAddress(p.DLTTransfer.WalletAddress)
		_, err = p.selectiveLoggingAndUpd([]string{"+amount"}, []interface{}{amount}, "dlt_wallets",
			[]string{"wallet_id"}, []string{utils.Int64ToStr(walletId)}, true)
	} else {
		_, err = p.selectiveLoggingAndUpd([]string{"+amount"}, []interface{}{amount}, "dlt_wallets",
			[]string{"wallet_id"}, []string{utils.Int64ToStr(walletId)}, true)
	}
	if err != nil {
		return p.ErrInfo(err)
	}

	// node commission
	_, err = p.selectiveLoggingAndUpd([]string{"+amount"}, []interface{}{commission}, "dlt_wallets", []string{"wallet_id"}, []string{utils.Int64ToStr(p.BlockData.WalletId)}, true)
	if err != nil {
		return p.ErrInfo(err)
	}

	// пишем в общую историю тр-ий
	dlt_transactions_id, err := p.ExecSqlGetLastInsertId(`INSERT INTO dlt_transactions ( sender_wallet_id, recipient_wallet_id, recipient_wallet_address, amount, commission, comment, time, block_id ) VALUES ( ?, ?, ?, ?, ?, ?, ?, ? )`, "dlt_transactions",
		p.TxWalletID, walletId, lib.AddressToString(int64(utils.StrToUint64(p.DLTTransfer.WalletAddress))), amount.String(), commission.String(), p.DLTTransfer.Comment, p.BlockData.Time, p.BlockData.BlockId)
	if err != nil {
		return p.ErrInfo(err)
	}
	err = p.ExecSql("INSERT INTO rollback_tx ( block_id, tx_hash, table_name, table_id ) VALUES (?, [hex], ?, ?)", p.BlockData.BlockId, p.TxHash, "dlt_transactions", dlt_transactions_id)
	if err != nil {
		return err
	}

	dlt_transactions_id, err = p.ExecSqlGetLastInsertId(`INSERT INTO dlt_transactions ( sender_wallet_id, recipient_wallet_id, recipient_wallet_address, amount, commission, comment, time, block_id ) VALUES ( ?, ?, ?, ?, ?, ?, ?, ? )`, "dlt_transactions",
		p.TxWalletID, p.BlockData.WalletId, lib.AddressToString(p.BlockData.WalletId), commission.String(), 0, "Commission", p.BlockData.Time, p.BlockData.BlockId)
	if err != nil {
		return p.ErrInfo(err)
	}
	err = p.ExecSql("INSERT INTO rollback_tx ( block_id, tx_hash, table_name, table_id ) VALUES (?, [hex], ?, ?)", p.BlockData.BlockId, p.TxHash, "dlt_transactions", dlt_transactions_id)
	if err != nil {
		return err
	}
	return nil
}

func (p *DLTTransferParser) Rollback() error {
	return p.autoRollback()
}
