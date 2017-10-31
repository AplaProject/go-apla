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
	"bytes"
	"encoding/hex"
	"fmt"
	"strconv"

	"github.com/AplaProject/go-apla/packages/config/syspar"
	"github.com/AplaProject/go-apla/packages/consts"
	"github.com/AplaProject/go-apla/packages/converter"
	"github.com/AplaProject/go-apla/packages/crypto"
	"github.com/AplaProject/go-apla/packages/model"
	"github.com/AplaProject/go-apla/packages/utils"
	"github.com/AplaProject/go-apla/packages/utils/tx"

	"github.com/shopspring/decimal"
	log "github.com/sirupsen/logrus"
	"gopkg.in/vmihailenco/msgpack.v2"
)

type DLTTransferParser struct {
	*Parser
	DLTTransfer *tx.DLTTransfer
}

func (p *DLTTransferParser) Init() error {
	logger := p.GetLogger()
	if p.BlockData.Version == 0 {
		oldSlice, err := ParseOldTransaction(bytes.NewBuffer(p.TxBinaryData))
		if err != nil {
			return fmt.Errorf("old transaction parsing failed")
		}
		if len(oldSlice) < 10 {
			logger.WithFields(log.Fields{"type": consts.InvalidObject}).Error("bad transaction format")
			return fmt.Errorf("bad transaction format")
		}
		p.DLTTransfer = &tx.DLTTransfer{
			Header: tx.Header{
				Type:          int(p.TxType),
				Time:          converter.BytesToInt64(oldSlice[2]),
				UserID:        converter.BytesToInt64(oldSlice[3]),
				StateID:       converter.BytesToInt64(oldSlice[4]),
				PublicKey:     converter.BinToHex(oldSlice[9]),
				BinSignatures: converter.BinToHex(oldSlice[10]),
			},
			WalletAddress: string(oldSlice[5]),
			Amount:        string(oldSlice[6]),
			Commission:    string(oldSlice[7]),
			Comment:       string(oldSlice[8]),
		}
		return nil
	}

	dltTransfer := &tx.DLTTransfer{}
	if err := msgpack.Unmarshal(p.TxBinaryData, dltTransfer); err != nil {
		logger.WithFields(log.Fields{"type": consts.UnmarshallingError, "error": err}).Error("unmarshalling dlt transfer from binary with msgpack")
		return p.ErrInfo(err)
	}
	p.DLTTransfer = dltTransfer
	p.DLTTransfer.PublicKey = converter.BinToHex(p.DLTTransfer.Header.PublicKey)
	return nil
}

func (p *DLTTransferParser) Validate() error {
	logger := p.GetLogger()
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
	found, err := dltWallet.Get(nil, converter.StringToAddress(p.DLTTransfer.WalletAddress))
	if err != nil {
		logger.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("checking that dlt wallet is exists")
		return p.ErrInfo(err)
	}
	if !found {
		bkey, err := hex.DecodeString(string(p.DLTTransfer.Header.PublicKey))
		if err != nil {
			logger.WithFields(log.Fields{"type": consts.ConvertionError, "error": err}).Error("decoding transaction public key from hex")
			return p.ErrInfo(err)
		}
		if crypto.KeyToAddress(bkey) != converter.AddressToString(p.TxWalletID) {
			logger.WithFields(log.Fields{"key_address": crypto.KeyToAddress(bkey), "wallet_address": converter.AddressToString(p.TxWalletID)}).Error("wallet addresses does not match")
			return p.ErrInfo("incorrect public_key")
		}
	}

	zero, _ := decimal.NewFromString("0")

	ourAmount, err := decimal.NewFromString(p.DLTTransfer.Amount)
	if err != nil {
		logger.WithFields(log.Fields{"type": consts.ConvertionError, "error": err, "value": p.DLTTransfer.Amount}).Error("coverting dlt transfer amount from string to decimal")
		return p.ErrInfo(err)
	}
	if ourAmount.Cmp(zero) <= 0 {
		logger.WithFields(log.Fields{"type": consts.ParameterExceeded}).Error("dlt transfer amount is less then zero")
		return p.ErrInfo("amount<=0")
	}

	systemParam := &model.SystemParameter{}
	found, err = systemParam.Get("fuel_rate")
	if err != nil {
		logger.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("getting fuel rate system param")
	}
	if !found {
		return p.ErrInfo("can't find fuel rate")
	}

	fuelRate, err := decimal.NewFromString(systemParam.Value)
	if err != nil {
		logger.WithFields(log.Fields{"type": consts.ParameterExceeded, "error": err, "value": systemParam.Value}).Error("coverting fuel rate system parameter from string to decimal")
		return err
	}
	if fuelRate.Cmp(decimal.New(0, 0)) <= 0 {
		logger.WithFields(log.Fields{"type": consts.ParameterExceeded}).Error("fuel rate param is less than zero")
		return fmt.Errorf(`fuel rate must be greater than 0`)
	}
	// 1 000 000 000 000 000 000 qDLT = 1 DLT * 100 000 000
	// fuelRate = 1 000 000 000 000 000
	//
	fPriceDecimal := decimal.New(syspar.SysCost(`dlt_transfer`), 0)
	if err != nil {
		logger.WithFields(log.Fields{"type": consts.ConvertionError, "error": err}).Error("converting sys cost dlt_transfer to decimal")
		return p.ErrInfo(err)
	}
	commission := fPriceDecimal.Mul(fuelRate)
	ourCommission, err := decimal.NewFromString(p.DLTTransfer.Commission)
	if err != nil {
		logger.WithFields(log.Fields{"type": consts.ConvertionError, "error": err, "value": p.DLTTransfer.Commission}).Error("coverting dlt transfer commission from string to decimal")
		return p.ErrInfo(err)
	}

	// check commission
	if ourCommission.Cmp(commission) < 0 {
		logger.WithFields(log.Fields{"commission": commission, "our_commission": ourCommission, "type": consts.ParameterExceeded}).Error("our commission is less than commission")
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
		logger.WithFields(log.Fields{"type": consts.InvalidObject}).Error("incorrect sign")
		return p.ErrInfo("incorrect sign OOPS")
	}

	wallet := &model.DltWallet{}
	found, err = wallet.Get(nil, p.TxWalletID)
	if err != nil {
		logger.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("getting wallet")
		return p.ErrInfo(err)
	}
	if !found {
		return p.ErrInfo("can't find wallet: ID" + strconv.FormatInt(p.TxWalletID, 10))
	}
	wltAmount, err := decimal.NewFromString(wallet.Amount)
	if err != nil {
		logger.WithFields(log.Fields{"type": consts.ConvertionError, "error": err}).Error("converting wallet amount from string to decimal")
		return p.ErrInfo(err)
	}

	if wltAmount.Cmp(ourAmount.Add(ourCommission)) < 0 {
		logger.Error("wallet amount is less than our amount + our commission")
		return p.ErrInfo(fmt.Sprintf("%s + %s < %s)", ourAmount, ourCommission, wallet.Amount))
	}
	if converter.StringToAddress(p.DLTTransfer.WalletAddress) == 0 {
		logger.WithFields(log.Fields{"type": consts.ConvertionError, "error": err, "value": p.DLTTransfer.WalletAddress}).Error("converting wallet address string to address")
		return p.ErrInfo(fmt.Sprintf(`Wallet %v is invalid`, p.DLTTransfer.WalletAddress))
	}

	return nil
}

func (p *DLTTransferParser) Action() error {
	logger := p.GetLogger()
	dltWallet := &model.DltWallet{}
	found, err := dltWallet.Get(nil, p.TxWalletID)
	if err != nil {
		logger.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("getting wallet")
		return p.ErrInfo(err)
	}
	if !found {
		logger.WithFields(log.Fields{"type": consts.NotFound, "error": err}).Error("wallet not found")
		return p.ErrInfo("can't find wallet. ID: " + strconv.FormatInt(p.TxWalletID, 10))
	}

	amount, err := decimal.NewFromString(p.DLTTransfer.Amount)
	if err != nil {
		logger.WithFields(log.Fields{"type": consts.ConvertionError, "error": err, "value": p.DLTTransfer.Amount}).Error("coverting dlt transfer amount from string to decimal")
		return p.ErrInfo(err)
	}
	commission, err := decimal.NewFromString(p.DLTTransfer.Commission)
	if err != nil {
		logger.WithFields(log.Fields{"type": consts.ConvertionError, "error": err, "value": p.DLTTransfer.Amount}).Error("coverting dlt transfer commission from string to decimal")
		return p.ErrInfo(err)
	}
	amountAndCommission := amount.Add(commission)
	if err != nil {
		logger.WithFields(log.Fields{"type": consts.ConvertionError, "error": err, "value": p.DLTTransfer.Amount}).Error("coverting dlt transfer commission from string to decimal")
		return p.ErrInfo(err)
	}
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
	walletAddress, err := strconv.ParseInt(p.DLTTransfer.WalletAddress, 10, 64)
	if err != nil {
		logger.WithFields(log.Fields{"type": consts.ConvertionError, "error": err, "value": p.DLTTransfer.WalletAddress}).Error("convertion wallet address to int")
	}
	dltTransaction := &model.DltTransaction{
		SenderWalletID:         p.TxWalletID,
		RecipientWalletID:      dltWallet.WalletID,
		RecipientWalletAddress: converter.AddressToString(walletAddress),
		Amount:                 &amount,
		Commission:             &commission,
		Comment:                p.DLTTransfer.Comment,
		Time:                   p.BlockData.Time,
		BlockID:                p.BlockData.BlockID,
	}
	err = dltTransaction.Create(p.DbTransaction)
	if err != nil {
		logger.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("creating dlt transaction")
		return p.ErrInfo(err)
	}
	rollbackTx := &model.RollbackTx{
		BlockID:   p.BlockData.BlockID,
		TxHash:    p.TxHash,
		NameTable: "dlt_transactions",
		TableID:   converter.Int64ToStr(dltTransaction.ID),
	}
	err = rollbackTx.Create(p.DbTransaction)
	if err != nil {
		logger.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("creating rollback transaction")
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
