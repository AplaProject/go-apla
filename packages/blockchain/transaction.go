package blockchain

import (
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/GenesisKernel/go-genesis/packages/consts"
	"github.com/GenesisKernel/go-genesis/packages/converter"
	"github.com/GenesisKernel/go-genesis/packages/crypto"
	"github.com/GenesisKernel/go-genesis/packages/utils/tx"

	log "github.com/sirupsen/logrus"
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/opt"
	"gopkg.in/vmihailenco/msgpack.v2"
)

var DB *leveldb.DB

type levelDBGetterPutterDeleter interface {
	Get([]byte, *opt.ReadOptions) ([]byte, error)
	Put([]byte, []byte, *opt.WriteOptions) error
	Delete([]byte, *opt.WriteOptions) error
}

func GetDB(tx *leveldb.Transaction) levelDBGetterPutterDeleter {
	if tx != nil {
		return tx
	}
	return DB
}

var txPrefix func([]byte) []byte = prefixFunc("tx-")
var txStatusPrefix func([]byte) []byte = prefixFunc("txstatus-")

func Init(filename string) error {
	var err error
	DB, err = leveldb.OpenFile(filename, nil)
	return err
}

type TxStatus struct {
	BlockID   int64
	BlockHash []byte
	Error     string
	Attempts  int64
}

func (ts TxStatus) Marshal() ([]byte, error) {
	b, err := msgpack.Marshal(ts)
	if err != nil {
		log.WithFields(log.Fields{"type": consts.MarshallingError, "error": err}).Error("marshalling tx status")
		return nil, err
	}
	return b, err
}

func (ts *TxStatus) Unmarshal(b []byte) error {
	err := msgpack.Unmarshal(b, ts)
	if err != nil {
		log.WithFields(log.Fields{"type": consts.MarshallingError, "error": err}).Error("unmarshalling tx status")
		return err
	}
	return err
}

func (ts *TxStatus) Get(tx *leveldb.Transaction, hash []byte) (bool, error) {
	val, err := GetDB(tx).Get(txStatusPrefix(hash), nil)
	if err == leveldb.ErrNotFound {
		return false, nil
	}
	if err != nil {
		log.WithFields(log.Fields{"type": consts.LevelDBError, "error": err}).Error("level db error")
		return false, err
	}
	err = ts.Unmarshal(val)
	return true, err
}

func (ts TxStatus) Insert(tx *leveldb.Transaction, hash []byte) error {
	val, err := ts.Marshal()
	if err != nil {
		return err
	}
	err = GetDB(tx).Put(txStatusPrefix(hash), val, nil)
	if err != nil {
		log.WithFields(log.Fields{"type": consts.LevelDBError, "error": err}).Error("level db error")
		return err
	}
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
	Params         map[string]string
	Files          map[string]*tx.File
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

func (t *Transaction) Get(tx *leveldb.Transaction, hash []byte) (bool, error) {
	val, err := GetDB(tx).Get(txPrefix(hash), nil)
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

func (t *Transaction) Insert(tx *leveldb.Transaction) error {
	hash, err := t.Hash()
	if err != nil {
		return err
	}
	val, err := t.Marshal()
	if err != nil {
		log.WithFields(log.Fields{"type": consts.MarshallingError, "error": err}).Error("marshalling transaction")
		return err
	}
	err = GetDB(tx).Put(txPrefix(hash), val, nil)
	if err != nil {
		log.WithFields(log.Fields{"type": consts.LevelDBError, "error": err}).Error("inserting transaction")
		return err
	}
	return nil
}

func SetTransactionError(tx *leveldb.Transaction, hash []byte, errString string) error {
	txWithError := &TxStatus{
		Error: errString,
	}
	txStatus := &TxStatus{}
	found, err := txStatus.Get(tx, hash)
	if err != nil {
		return err
	}
	if !found {
		return txWithError.Insert(tx, hash)
	}
	txStatus.Error = errString
	return txStatus.Insert(tx, hash)
}

func IncrementTxAttemptCount(tx *leveldb.Transaction, hash []byte) error {
	ts := &TxStatus{}
	found, err := ts.Get(tx, hash)
	if err != nil {
		return err
	}
	if !found {
		return nil
	}
	ts.Attempts += 1
	return ts.Insert(tx, hash)
}

func DecrementTxAttemptCount(tx *leveldb.Transaction, hash []byte) error {
	ts := &TxStatus{}
	found, err := ts.Get(tx, hash)
	if err != nil {
		return err
	}
	if !found {
		return nil
	}
	ts.Attempts -= 1
	return ts.Insert(tx, hash)
}

// BuildTransaction creates transaction
func BuildTransaction(smartTx Transaction, privKey, pubKey string, params ...string) (*Transaction, error) {
	signPrms := []string{smartTx.ForSign()}
	signPrms = append(signPrms, params...)
	signature, err := crypto.Sign(
		[]byte(privKey),
		[]byte(strings.Join(signPrms, ",")),
	)
	if err != nil {
		log.WithFields(log.Fields{"type": consts.CryptoError, "error": err}).Error("signing by node private key")
		return nil, err
	}
	smartTx.Header.BinSignatures = converter.EncodeLengthPlusData(signature)

	if smartTx.Header.PublicKey, err = hex.DecodeString(pubKey); err != nil {
		log.WithFields(log.Fields{"type": consts.ConversionError, "error": err}).Error("decoding public key from hex")
		return nil, err
	}

	return &smartTx, err
}
