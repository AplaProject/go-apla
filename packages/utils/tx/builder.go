package tx

import (
	"encoding/hex"
	"strings"

	"github.com/GenesisKernel/go-genesis/packages/consts"
	"github.com/GenesisKernel/go-genesis/packages/converter"
	"github.com/GenesisKernel/go-genesis/packages/crypto"
	"github.com/GenesisKernel/go-genesis/packages/model"
	log "github.com/sirupsen/logrus"
	"gopkg.in/vmihailenco/msgpack.v2"
)

func BuildTransaction(smartTx SmartContract, privKey, pubKey string, params ...string) error {
	signPrms := []string{smartTx.ForSign()}
	signPrms = append(signPrms, params...)
	signature, err := crypto.Sign(
		privKey,
		strings.Join(signPrms, ","),
	)
	if err != nil {
		log.WithFields(log.Fields{"type": consts.CryptoError, "error": err}).Error("signing by node private key")
		return err
	}
	smartTx.BinSignatures = converter.EncodeLengthPlusData(signature)

	if smartTx.PublicKey, err = hex.DecodeString(pubKey); err != nil {
		log.WithFields(log.Fields{"type": consts.ConversionError, "error": err}).Error("decoding public key from hex")
		return err
	}

	data, err := msgpack.Marshal(smartTx)
	if err != nil {
		log.WithFields(log.Fields{"type": consts.MarshallingError, "error": err}).Error("marshalling smart contract to msgpack")
		return err
	}
	data = append([]byte{128}, data...)

	hash, err := crypto.Hash(data)
	if err != nil {
		log.WithFields(log.Fields{"type": consts.CryptoError, "error": err}).Error("calculating hash of smart contract")
		return err
	}

	tx := &model.Transaction{
		Hash:     hash,
		Data:     data[:],
		Type:     int8(converter.BinToDecBytesShift(&data, 1)),
		KeyID:    smartTx.KeyID,
		HighRate: model.TransactionRateOnBlock,
	}
	if err = tx.Create(); err != nil {
		log.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("creating new transaction")
		return err
	}

	return nil
}
