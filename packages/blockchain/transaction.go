package blockchain

import (
	"bytes"
	"encoding/hex"
	"fmt"

	"github.com/AplaProject/go-apla/packages/consts"
	"github.com/AplaProject/go-apla/packages/converter"
	"github.com/AplaProject/go-apla/packages/crypto"
	"github.com/AplaProject/go-apla/packages/utils/tx"

	log "github.com/sirupsen/logrus"
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/iterator"
	"github.com/syndtr/goleveldb/leveldb/opt"
	"github.com/syndtr/goleveldb/leveldb/util"
	"gopkg.in/vmihailenco/msgpack.v2"
)

var DB *leveldb.DB

type LevelDBGetterPutterDeleter interface {
	Get([]byte, *opt.ReadOptions) ([]byte, error)
	Put([]byte, []byte, *opt.WriteOptions) error
	Delete([]byte, *opt.WriteOptions) error
	NewIterator(slice *util.Range, ro *opt.ReadOptions) iterator.Iterator
}

func GetDB(tx *leveldb.Transaction) LevelDBGetterPutterDeleter {
	if tx != nil {
		return tx
	}
	return DB
}

const txProcessPrefix = "process-"

var txPrefix func([]byte) []byte = prefixFunc("tx-")
var txStatusPrefix func([]byte) []byte = prefixFunc("txstatus-")
var processTxPrefix func([]byte) []byte = prefixFunc(txProcessPrefix)

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
	Name          string
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
	return fmt.Sprintf("%s,%s,%d,%d,%d,%d,%s,%s,%d", t.RequestID, t.Header.Name, t.Header.Time, t.Header.KeyID, t.Header.EcosystemID,
		t.TokenEcosystem, t.MaxSum, t.PayOver, t.SignedBy)
}

func (t Transaction) Marshal() ([]byte, error) {
	var buf bytes.Buffer
	enc := msgpack.NewEncoder(&buf).SortMapKeys(true)
	if err := enc.Encode(t); err != nil {
		log.WithFields(log.Fields{"type": consts.MarshallingError, "error": err}).Error("marshalling tx")
		return nil, err
	}
	return buf.Bytes(), nil
}

func (t *Transaction) Unmarshal(b []byte) error {
	if err := msgpack.Unmarshal(b, t); err != nil {
		log.WithFields(log.Fields{"type": consts.UnmarshallingError, "error": err}).Error("unmarshalling tx")
		return err
	}
	return nil
}

func (t Transaction) Hash() ([]byte, error) {
	sign := t.Header.BinSignatures
	tokenEcosystem := t.TokenEcosystem
	t.Header.BinSignatures = nil
	t.TokenEcosystem = 0
	b, err := t.Marshal()
	if err != nil {
		return nil, err
	}
	t.Header.BinSignatures = sign
	t.TokenEcosystem = tokenEcosystem
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

func insertTransactionWithPrefix(tx *leveldb.Transaction, t *Transaction, prefix func([]byte) []byte) error {
	hash, err := t.Hash()
	if err != nil {
		return err
	}
	val, err := t.Marshal()
	if err != nil {
		log.WithFields(log.Fields{"type": consts.MarshallingError, "error": err}).Error("marshalling transaction")
		return err
	}
	err = GetDB(tx).Put(prefix(hash), val, nil)
	if err != nil {
		log.WithFields(log.Fields{"type": consts.LevelDBError, "error": err}).Error("inserting transaction")
		return err
	}
	return nil
}

func (t *Transaction) Insert(tx *leveldb.Transaction) error {
	hash, err := t.Hash()
	if err != nil {
		return err
	}
	if err := DeleteProcessedTx(tx, hash); err != nil {
		return err
	}

	return insertTransactionWithPrefix(tx, t, txPrefix)
}

func DeleteProcessedTx(tx *leveldb.Transaction, hash []byte) error {
	err := GetDB(tx).Delete(processTxPrefix(hash), nil)
	if err != nil {
		log.WithFields(log.Fields{"type": consts.LevelDBError, "error": err}).Error("inserting transaction")
		return err
	}
	return nil
}

func GetTxsToProcess(tx *leveldb.Transaction) ([]*Transaction, error) {
	txs := []*Transaction{}
	iter := GetDB(tx).NewIterator(util.BytesPrefix([]byte(txProcessPrefix)), nil)
	for iter.Next() {
		value := iter.Value()
		t := &Transaction{}
		if err := t.Unmarshal(value); err != nil {
			return nil, err
		}
		txs = append(txs, t)
	}
	iter.Release()
	if err := iter.Error(); err != nil {
		return nil, err
	}
	return txs, nil
}

func InsertTxToProcess(tx *leveldb.Transaction, t *Transaction) error {
	return insertTransactionWithPrefix(tx, t, processTxPrefix)
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
	if err := DeleteProcessedTx(tx, hash); err != nil {
		return err
	}
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

func SetTransactionComplete(tx *leveldb.Transaction, hash, blockHash []byte, blockID int64) error {
	ts := &TxStatus{}
	found, err := ts.Get(tx, hash)
	if err != nil {
		return err
	}
	if !found {
		return nil
	}
	ts.BlockHash = blockHash
	ts.BlockID = blockID
	return ts.Insert(tx, hash)
}

// BuildTransaction creates transaction
func BuildTransaction(smartTx Transaction, privKey, pubKey string, params ...string) (*Transaction, error) {
	bytePrivKey, err := hex.DecodeString(privKey)
	if err != nil {
		log.WithFields(log.Fields{"type": consts.ConversionError, "error": err}).Error("decoding private key from hex")
		return nil, err
	}

	if smartTx.Header.PublicKey, err = crypto.HexToPub(pubKey); err != nil {
		log.WithFields(log.Fields{"type": consts.ConversionError, "error": err}).Error("decoding public key from hex")
		return nil, err
	}
	txHash, err := smartTx.Hash()
	if err != nil {
		return nil, err
	}
	signature, err := crypto.Sign(
		[]byte(bytePrivKey),
		txHash,
	)
	if err != nil {
		log.WithFields(log.Fields{"type": consts.CryptoError, "error": err}).Error("signing by node private key")
		return nil, err
	}
	smartTx.Header.BinSignatures = converter.EncodeLengthPlusData(signature)

	return &smartTx, err
}
