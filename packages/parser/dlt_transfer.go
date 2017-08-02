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

	"github.com/EGaaS/go-egaas-mvp/packages/converter"
	"github.com/EGaaS/go-egaas-mvp/packages/crypto"
	"github.com/EGaaS/go-egaas-mvp/packages/model"
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
	if err := msgpack.Unmarshal(p.TxBinaryData, dltTransfer); err != nil {
		return p.ErrInfo(err)
	}
	p.DLTTransfer = dltTransfer
	p.DLTTransfer.PublicKey = converter.BinToHex(p.DLTTransfer.Header.PublicKey)
	return nil
}

func (p *DLTTransferParser) Validate() error {
	err := p.generalCheck(`dlt_transfer`, &p.DLTTransfer.Header, map[string]string{})
	if err != nil {
		return p.ErrInfo(err)
	}

	verifyData := map[string][]interface{}{"walletAddress": []interface{}{p.DLTTransfer.WalletAddress}, "decimal": []interface{}{p.DLTTransfer.Amount, p.DLTTransfer.Commission}, "comment": []interface{}{p.DLTTransfer.Comment}}
	err = p.CheckInputData(verifyData)
	if err != nil {
		return p.ErrInfo(err)
	}

	// public key need only when we don't have public_key in the dlt_wallets table
	dltWallet := &model.DltWallet{}
	exists, err := dltWallet.IsExists()
	if err != nil {
		return p.ErrInfo(err)
	}
	if !exists {
		bkey, err := hex.DecodeString(string(p.DLTTransfer.Header.PublicKey))
		if err != nil {
			return p.ErrInfo(err)
		}
		if crypto.KeyToAddress(bkey) != converter.AddressToString(p.TxWalletID) {
			return p.ErrInfo("incorrect public_key")
		}
	}

	zero, _ := decimal.NewFromString("0")

	ourAmount, err := decimal.NewFromString(p.DLTTransfer.Amount)
	if err != nil {
		return p.ErrInfo(err)
	}
	if ourAmount.Cmp(zero) <= 0 {
		return p.ErrInfo("amount<=0")
	}

	sp := &model.SystemParameter{}
	fPrice, err := sp.GetValueParameterByName("op_price", "dlt_transfer")
	if err != nil {
		return p.ErrInfo(err)
	}

	fuelRate := model.GetFuel()
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
	ourCommission, err := decimal.NewFromString(p.DLTTransfer.Commission) //fPrice)
	if err != nil {
		return p.ErrInfo(err)
	}

	// проверим, удовлетворяет ли нас комиссия, которую предлагает юзер
	if ourCommission.Cmp(commission) < 0 {
		return p.ErrInfo(fmt.Sprintf("commission %v < dltPrice %v", ourCommission, commission))
	}

	if p.DLTTransfer.Comment == "null" {
		p.DLTTransfer.Comment = ""
	}

	CheckSignResult, err := utils.CheckSign(p.PublicKeys, p.DLTTransfer.ForSign(), p.DLTTransfer.BinSignatures, false)
	if err != nil {
		return p.ErrInfo(err)
	}
	if !CheckSignResult {
		return p.ErrInfo("incorrect sign OOPS")
	}

	wallet := &model.DltWallet{}
	err = wallet.GetWallet(p.TxWalletID)
	if err != nil {
		return p.ErrInfo(err)
	}
	if wallet.Amount.Cmp(ourAmount.Add(ourCommission)) < 0 {
		return p.ErrInfo(fmt.Sprintf("%s + %s < %s)", ourAmount, ourCommission, wallet.Amount))
	}
	return nil
}

func (p *DLTTransferParser) Action() error {
	log.Debug("wallet address %s", p.DLTTransfer.WalletAddress)
	dltWallet := &model.DltWallet{}
	err := dltWallet.GetWallet(p.TxWalletID)
	if err != nil {
		return p.ErrInfo(err)
	}
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
	if len(p.DLTTransfer.Header.PublicKey) > 30 && len(dltWallet.PublicKey) == 0 {
		_, _, err = p.selectiveLoggingAndUpd([]string{"-amount", "public_key_0"}, []interface{}{amountAndCommission.String(), converter.HexToBin(p.DLTTransfer.PublicKey)}, "dlt_wallets", []string{"wallet_id"}, []string{converter.Int64ToStr(p.TxWalletID)}, true)
	} else {
		_, _, err = p.selectiveLoggingAndUpd([]string{"-amount"}, []interface{}{amountAndCommission.String()}, "dlt_wallets", []string{"wallet_id"}, []string{converter.Int64ToStr(p.TxWalletID)}, true)
	}
	if err != nil {
		return p.ErrInfo(err)
	}

	walletID := converter.StringToAddress(p.DLTTransfer.WalletAddress)
	if dltWallet.WalletID == 0 {
		log.Debug("walletId == 0")
		log.Debug("%s", string(p.DLTTransfer.WalletAddress))
		_, _, err = p.selectiveLoggingAndUpd([]string{"+amount"}, []interface{}{amount}, "dlt_wallets",
			[]string{"wallet_id"}, []string{converter.Int64ToStr(walletID)}, true)
	} else {
		_, _, err = p.selectiveLoggingAndUpd([]string{"+amount"}, []interface{}{amount}, "dlt_wallets",
			[]string{"wallet_id"}, []string{converter.Int64ToStr(walletID)}, true)
	}
	if err != nil {
		return p.ErrInfo(err)
	}

	// node commission
	_, _, err = p.selectiveLoggingAndUpd([]string{"+amount"}, []interface{}{commission}, "dlt_wallets", []string{"wallet_id"}, []string{converter.Int64ToStr(p.BlockData.WalletID)}, true)
	if err != nil {
		return p.ErrInfo(err)
	}
	dltTransaction := &model.DltTransaction{
		SenderWalletID:         p.TxWalletID,
		RecepientWalletID:      dltWallet.WalletID,
		RecepientWalletAddress: converter.AddressToString(int64(converter.StrToUint64(p.DLTTransfer.WalletAddress))),
		Amount:                 amount,
		Comission:              commission,
		Comment:                p.DLTTransfer.Comment,
		Time:                   p.BlockData.Time,
		BlockID:                p.BlockData.BlockID,
	}
	err = dltTransaction.Create()
	if err != nil {
		return p.ErrInfo(err)
	}
	rollbackTx := &model.RollbackTx{
		BlockID:   p.BlockData.BlockID,
		TxHash:    []byte(p.TxHash),
		TableName: "dlt_transactions",
		TableID:   converter.Int64ToStr(dltTransaction.ID),
	}
	err = rollbackTx.Create()
	if err != nil {
		return p.ErrInfo(err)
	}
	return nil
}

func (p *DLTTransferParser) Rollback() error {
	return p.autoRollback()
}

func (p DLTTransferParser) Header() *tx.Header {
	return &p.DLTTransfer.Header
}
