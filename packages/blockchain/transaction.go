package blockchain

import (
	"encoding/hex"
	"strings"

	"github.com/GenesisKernel/go-genesis/packages/consts"
	"github.com/GenesisKernel/go-genesis/packages/converter"
	"github.com/GenesisKernel/go-genesis/packages/crypto"
	"github.com/GenesisKernel/go-genesis/packages/utils/tx"

	log "github.com/sirupsen/logrus"
	"github.com/syndtr/goleveldb/leveldb"
	"gopkg.in/vmihailenco/msgpack.v2"
)

var db *leveldb.DB

const txPrefix = "tx-"

func Init(filename string) error {
	var err error
	db, err = leveldb.OpenFile(filename, nil)
	return err
}

func GetTransaction(hash []byte) (*tx.SmartContract, bool, error) {
	val, err := db.Get([]byte(txPrefix+string(hash)), nil)
	if err == leveldb.ErrNotFound {
		return nil, false, nil
	}
	if err != nil {
		log.WithFields(log.Fields{"type": consts.LevelDBError, "error": err}).Error("getting transaction")
		return nil, false, err
	}
	var tx tx.SmartContract
	if err := msgpack.Unmarshal(val, &tx); err != nil {
		log.WithFields(log.Fields{"type": consts.UnmarshallingError, "error": err}).Error("unmarshalling transaction")
		return nil, true, err
	}
	return &tx, true, nil
}

func SetTransaction(hash []byte, tx *tx.SmartContract) error {
	val, err := msgpack.Marshal(tx)
	if err != nil {
		log.WithFields(log.Fields{"type": consts.MarshallingError, "error": err}).Error("marshalling transaction")
		return err
	}
	err = db.Put([]byte(txPrefix+string(hash)), val, nil)
	if err != nil {
		log.WithFields(log.Fields{"type": consts.LevelDBError, "error": err}).Error("inserting transaction")
		return err
	}
	return nil
}

func GetTransactionBinary(hash []byte) ([]byte, bool, error) {
	val, err := db.Get([]byte(txPrefix+string(hash)), nil)
	if err == leveldb.ErrNotFound {
		return nil, false, nil
	}
	if err != nil {
		log.WithFields(log.Fields{"type": consts.LevelDBError, "error": err}).Error("getting transaction")
		return nil, false, err
	}
	return val, true, nil
}

func SetTransactionBinary(hash, tx []byte) error {
	err := db.Put([]byte(txPrefix+string(hash)), tx, nil)
	if err != nil {
		log.WithFields(log.Fields{"type": consts.LevelDBError, "error": err}).Error("inserting transaction")
		return err
	}
	return nil
}

func SetTransactionError(hash []byte, errString string) error {
	txWithError := &tx.SmartContract{
		Header: tx.Header{
			Error: errString,
		},
	}
	tx, found, err := GetTransaction(hash)
	if err != nil {
		return err
	}
	if !found {
		return SetTransaction(hash, txWithError)
	}
	tx.Header.Error = errString
	return SetTransaction(hash, tx)
}

func IncrementTxAttemptCount(hash []byte) error {
	tx, found, err := GetTransaction(hash)
	if err != nil {
		return err
	}
	if !found {
		return nil
	}
	tx.Attempts += 1
	return SetTransaction(hash, tx)
}

func DecrementTxAttemptCount(hash []byte) error {
	tx, found, err := GetTransaction(hash)
	if err != nil {
		return err
	}
	if !found {
		return nil
	}
	tx.Attempts -= 1
	return SetTransaction(hash, tx)
}

// BuildTransaction creates transaction
func BuildTransaction(smartTx tx.SmartContract, privKey, pubKey string, params ...string) error {
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

	if err = SetTransactionBinary(hash, data); err != nil {
		return err
	}

	return nil
}
