package blockchain

import (
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/GenesisKernel/go-genesis/packages/consts"
	"github.com/GenesisKernel/go-genesis/packages/converter"
	"github.com/GenesisKernel/go-genesis/packages/crypto"

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

// Header is contain header data
type TxHeader struct {
	Type          int
	Time          int64
	EcosystemID   int64
	KeyID         int64
	RoleID        int64
	NetworkID     int64
	NodePosition  int64
	BlockID       int64
	Attempts      int64
	Error         string
	PublicKey     []byte
	BinSignatures []byte
}

// SmartContract is storing smart contract data
type Transaction struct {
	Header         TxHeader
	RequestID      string
	TokenEcosystem int64
	MaxSum         string
	PayOver        string
	SignedBy       int64
	Attempts       int64
	Error          string
	Params         map[string]string
}

// ForSign is converting SmartContract to string
func (t Transaction) ForSign() string {
	return fmt.Sprintf("%s,%d,%d,%d,%d,%d,%s,%s,%d", t.RequestID, t.Header.Type, t.Header.Time, t.Header.KeyID, t.Header.EcosystemID,
		t.TokenEcosystem, t.MaxSum, t.PayOver, t.SignedBy)
}

func (t Transaction) Marshal() ([]byte, error) {
	var b []byte
	var err error
	if b, err = msgpack.Marshal(t); err != nil {
		log.WithFields(log.Fields{"type": consts.MarshallingError, "error": err}).Error("marshalling tx")
		return nil, err
	}
	return b, err
}

func (t *Transaction) Unmarshal(b []byte) error {
	if err := msgpack.Unmarshal(b, t); err != nil {
		log.WithFields(log.Fields{"type": consts.UnmarshallingError, "error": err}).Error("unmarshalling tx")
		return err
	}
	return nil
}

func (t Transaction) Hash() ([]byte, error) {
	b, err := t.Marshal()
	if err != nil {
		return nil, err
	}
	return crypto.DoubleHash(b)
}

func (t *Transaction) Get(hash []byte) (bool, error) {
	val, err := db.Get([]byte(txPrefix+string(hash)), nil)
	if err == leveldb.ErrNotFound {
		return false, nil
	}
	if err != nil {
		log.WithFields(log.Fields{"type": consts.LevelDBError, "error": err}).Error("getting transaction")
		return false, err
	}
	if err := t.Unmarshal(val); err != nil {
		log.WithFields(log.Fields{"type": consts.UnmarshallingError, "error": err}).Error("unmarshalling transaction")
		return true, err
	}
	return true, nil
}

func (t *Transaction) Insert(hash []byte) error {
	val, err := t.Marshal()
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

func InsertTransactionBinary(hash, tx []byte) error {
	err := db.Put([]byte(txPrefix+string(hash)), tx, nil)
	if err != nil {
		log.WithFields(log.Fields{"type": consts.LevelDBError, "error": err}).Error("inserting transaction")
		return err
	}
	return nil
}

func SetTransactionError(hash []byte, errString string) error {
	txWithError := &Transaction{
		Header: TxHeader{
			Error: errString,
		},
	}
	tx := &Transaction{}
	found, err := tx.Get(hash)
	if err != nil {
		return err
	}
	if !found {
		return txWithError.Insert(hash)
	}
	tx.Header.Error = errString
	return tx.Insert(hash)
}

func IncrementTxAttemptCount(hash []byte) error {
	tx := &Transaction{}
	found, err := tx.Get(hash)
	if err != nil {
		return err
	}
	if !found {
		return nil
	}
	tx.Attempts += 1
	return tx.Insert(hash)
}

func DecrementTxAttemptCount(hash []byte) error {
	tx := &Transaction{}
	found, err := tx.Get(hash)
	if err != nil {
		return err
	}
	if !found {
		return nil
	}
	tx.Attempts -= 1
	return tx.Insert(hash)
}

// BuildTransaction creates transaction
func BuildTransaction(smartTx Transaction, privKey, pubKey string, params ...string) error {
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
	smartTx.Header.BinSignatures = converter.EncodeLengthPlusData(signature)

	if smartTx.Header.PublicKey, err = hex.DecodeString(pubKey); err != nil {
		log.WithFields(log.Fields{"type": consts.ConversionError, "error": err}).Error("decoding public key from hex")
		return err
	}
	hash, err := smartTx.Hash()
	if err != nil {
		return err
	}

	return smartTx.Insert(hash)
}
