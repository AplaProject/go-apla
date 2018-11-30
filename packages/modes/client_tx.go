package modes

import (
	"bytes"
	"errors"
	"time"

	"github.com/GenesisKernel/go-genesis/packages/conf"
	"github.com/GenesisKernel/go-genesis/packages/consts"
	"github.com/GenesisKernel/go-genesis/packages/converter"
	"github.com/GenesisKernel/go-genesis/packages/model"
	"github.com/GenesisKernel/go-genesis/packages/transaction"
	"github.com/GenesisKernel/go-genesis/packages/utils/tx"
	log "github.com/sirupsen/logrus"
	msgpack "gopkg.in/vmihailenco/msgpack.v2"
)

var ErrDiffKey = errors.New("Different keys")

type ClientTxPreprocessor interface {
	ProcessClientTranstaction([]byte) (string, error)
}

type blockchainTxPreprocessor struct {
	logger *log.Entry
	keyID  int64
}

func (p blockchainTxPreprocessor) ProcessClientTranstaction(txData []byte) (string, error) {
	rtx := &transaction.RawTransaction{}
	if err := rtx.Unmarshall(bytes.NewBuffer(txData)); err != nil {
		return "", err
	}

	smartTx := tx.SmartContract{}
	if err := msgpack.Unmarshal(rtx.Payload(), &smartTx); err != nil {
		return "", err
	}

	if smartTx.Header.KeyID != p.keyID {
		return "", ErrDiffKey
	}

	if err := model.SendTx(rtx, p.keyID); err != nil {
		p.logger.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("sending tx")
		return "", err
	}

	return string(converter.BinToHex(rtx.Hash())), nil
}

type ObsTxPreprocessor struct {
	Logger *log.Entry
	KeyID  int64
}

func (p ObsTxPreprocessor) ProcessClientTranstaction(txData []byte) (string, error) {

	tx, err := transaction.UnmarshallTransaction(bytes.NewBuffer(txData))
	if err != nil {
		log.WithFields(log.Fields{"type": consts.ParseError, "error": err}).Error("on unmarshaling user tx")
		return "", err
	}

	ts := &model.TransactionStatus{
		BlockID:  1,
		Hash:     tx.TxHash,
		Time:     time.Now().Unix(),
		WalletID: p.KeyID,
		Type:     tx.TxType,
	}

	if err := ts.Create(); err != nil {
		log.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("on creating tx status")
		return "", err
	}

	res, _, err := tx.CallOBSContract()
	if err != nil {
		log.WithFields(log.Fields{"type": consts.ParseError, "error": err}).Error("on execution contract")
		return "", err
	}

	if err := ts.UpdateBlockMsg(nil, 1, res, tx.TxHash); err != nil {
		p.Logger.WithFields(log.Fields{"type": consts.DBError, "error": err, "tx_hash": tx.TxHash}).Error("updating transaction status block id")
		return "", err
	}

	return string(converter.BinToHex(tx.TxHash)), nil
}

func GetClientTxPreprocessor(logger *log.Entry, keyID int64) ClientTxPreprocessor {
	if conf.Config.IsSupportingOBS() {
		return ObsTxPreprocessor{
			Logger: logger,
			KeyID:  keyID,
		}
	}

	return blockchainTxPreprocessor{
		logger: logger,
		keyID:  keyID,
	}
}
