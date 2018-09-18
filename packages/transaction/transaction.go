package transaction

import (
	"bytes"
	"fmt"
	"math/rand"
	"strings"
	"time"

	"github.com/GenesisKernel/go-genesis/packages/consts"
	"github.com/GenesisKernel/go-genesis/packages/converter"
	"github.com/GenesisKernel/go-genesis/packages/crypto"
	"github.com/GenesisKernel/go-genesis/packages/model"
	"github.com/GenesisKernel/go-genesis/packages/script"
	"github.com/GenesisKernel/go-genesis/packages/smart"
	"github.com/GenesisKernel/go-genesis/packages/transaction/custom"
	"github.com/GenesisKernel/go-genesis/packages/types"
	"github.com/GenesisKernel/go-genesis/packages/utils"
	"github.com/GenesisKernel/go-genesis/packages/utils/tx"

	"github.com/shopspring/decimal"
	log "github.com/sirupsen/logrus"
	"gopkg.in/vmihailenco/msgpack.v2"
)

// Transaction is a structure for parsing transactions
type Transaction struct {
	BlockData  *utils.BlockData
	PrevBlock  *utils.BlockData
	PublicKeys [][]byte

	TxBinaryData  []byte // transaction binary data
	TxFullData    []byte // full transaction, with type and data
	TxHash        []byte
	TxKeyID       int64
	TxTime        int64
	TxType        int64
	TxCost        int64 // Maximum cost of executing contract
	TxFuel        int64
	TxUsedCost    decimal.Decimal // Used cost of CPU resources
	TxPtr         interface{}     // Pointer to the corresponding struct in consts/struct.go
	TxData        map[string]interface{}
	TxSmart       *tx.SmartContract
	TxContract    *smart.Contract
	TxHeader      *tx.Header
	tx            custom.TransactionInterface
	DbTransaction *model.DbTransaction
	SysUpdate     bool
	Rand          *rand.Rand

	SmartContract smart.SmartContract
}

// GetLogger returns logger
func (t Transaction) GetLogger() *log.Entry {
	logger := log.WithFields(log.Fields{"tx_type": t.TxType, "tx_time": t.TxTime, "tx_wallet_id": t.TxKeyID})
	if t.BlockData != nil {
		logger = logger.WithFields(log.Fields{"block_id": t.BlockData.BlockID, "block_time": t.BlockData.Time, "block_wallet_id": t.BlockData.KeyID, "block_state_id": t.BlockData.EcosystemID, "block_hash": t.BlockData.Hash, "block_version": t.BlockData.Version})
	}
	if t.PrevBlock != nil {
		logger = logger.WithFields(log.Fields{"block_id": t.BlockData.BlockID, "block_time": t.BlockData.Time, "block_wallet_id": t.BlockData.KeyID, "block_state_id": t.BlockData.EcosystemID, "block_hash": t.BlockData.Hash, "block_version": t.BlockData.Version})
	}
	return logger
}

var txCache = &transactionCache{cache: make(map[string]*Transaction)}

// UnmarshallTransaction is unmarshalling transaction
func UnmarshallTransaction(buffer *bytes.Buffer) (*Transaction, error) {
	if buffer.Len() == 0 {
		log.WithFields(log.Fields{"type": consts.EmptyObject}).Error("empty transaction buffer")
		return nil, fmt.Errorf("empty transaction buffer")
	}

	hash, err := crypto.Hash(buffer.Bytes())
	// or DoubleHash ?
	if err != nil {
		log.WithFields(log.Fields{"type": consts.CryptoError, "error": err}).Error("hashing transaction")
		return nil, err
	}

	if t, ok := txCache.Get(string(hash)); ok {
		return t, nil
	}

	t := new(Transaction)
	t.TxHash = hash
	t.TxUsedCost = decimal.New(0, 0)
	t.TxFullData = buffer.Bytes()

	txType := int64(buffer.Bytes()[0])

	// smart contract transaction
	if IsContractTransaction(int(txType)) {
		// skip byte with transaction type
		buffer.Next(1)
		t.TxBinaryData = buffer.Bytes()
		if err := t.parseFromContract(buffer); err != nil {
			return nil, err
		}

		// struct transaction (only first block transaction for now)
	} else if consts.IsStruct(int(txType)) {
		t.TxBinaryData = buffer.Bytes()
		if err := t.parseFromStruct(buffer, txType); err != nil {
			return t, err
		}

		// all other transactions
	}
	txCache.Set(t)

	return t, nil
}

// IsContractTransaction checks txType
func IsContractTransaction(txType int) bool {
	return txType > 127
}

func (t *Transaction) parseFromStruct(buf *bytes.Buffer, txType int64) error {
	t.TxPtr = consts.MakeStruct(consts.TxTypes[int(txType)])
	input := buf.Bytes()
	if err := converter.BinUnmarshal(&input, t.TxPtr); err != nil {
		log.WithFields(log.Fields{"error": err, "type": consts.UnmarshallingError, "tx_type": int(txType)}).Error("getting parser for tx type")
		return err
	}
	head := consts.Header(t.TxPtr)
	t.TxKeyID = head.KeyID
	t.TxTime = int64(head.Time)
	t.TxType = txType

	trParser, err := GetTransaction(t, consts.TxTypes[int(txType)])
	if err != nil {
		return err
	}
	t.tx = trParser

	err = trParser.Validate()
	if err != nil {
		return utils.ErrInfo(err)
	}

	return nil
}

func (t *Transaction) fillTxData(fieldInfos []*script.FieldInfo, params []interface{}, forsign []string) error {
	if len(params) != len(fieldInfos) {
		return fmt.Errorf("Invalid number of parameters")
	}

	for index, fitem := range fieldInfos {
		var err error
		var v interface{}
		var ok bool
		var forv string
		var isforv bool

		switch fitem.Type.String() {
		case `bool`:
			if v, ok = params[index].(bool); !ok {
				return fmt.Errorf("Incorrect type bool")
			}
		case `uint64`:
			if v, ok = params[index].(uint64); !ok {
				return fmt.Errorf("Incorrect type uint64")
			}
		case `float64`:
			if v, ok = params[index].(float64); !ok {
				return fmt.Errorf("Incorrect type float64")
			}
		case `int64`:
			if v, ok = params[index].(int64); !ok {
				return fmt.Errorf("Incorrect type int64")
			}
		case script.Decimal:
			var s string
			if s, ok = params[index].(string); !ok {
				return fmt.Errorf("Incorrect type money")
			}
			v, err = decimal.NewFromString(s)
			if err != nil {
				return err
			}
		case `string`:
			if v, ok = params[index].(string); !ok {
				return fmt.Errorf("Incorrect type string")
			}
		case `[]uint8`:
			var val []byte
			if val, ok = params[index].([]byte); !ok {
				return fmt.Errorf("Incorrect type []uint8")
			}

			if forv, err = crypto.HashHex(val); err != nil {
				return err
			}

			isforv = true
			v = val
		case `[]interface {}`:
			var val []interface{}
			if val, ok = params[index].([]interface{}); !ok {
				return fmt.Errorf("Incorrect type []interface {}")
			}

			list := make([]string, len(val)+1)
			list[0] = converter.IntToStr(len(val))
			for i, _ := range val {
				list[i+1] = fmt.Sprintf("%v", val[i])
			}

			v = val
			isforv = true
			forv = strings.Join(list, ",")
		case script.File:
			var val map[interface{}]interface{}
			if val, ok = params[index].(map[interface{}]interface{}); !ok {
				return fmt.Errorf("Incorrect type file")
			}

			file := types.File{
				"Name":     val["Name"].(string),
				"MimeType": val["MimeType"].(string),
				"Body":     val["Body"].([]byte),
			}

			v = file
			isforv = true
			forv = "file"
		}

		if _, ok = t.TxData[fitem.Name]; !ok {
			t.TxData[fitem.Name] = v
		}
		if err != nil {
			return err
		}
		if isforv {
			v = forv
		}
		forsign = append(forsign, fmt.Sprintf("%v", v))
	}
	t.TxData[`forsign`] = strings.Join(forsign, ",")
	return nil
}

func (t *Transaction) parseFromContract(buf *bytes.Buffer) error {
	smartTx := tx.SmartContract{}
	if err := msgpack.Unmarshal(buf.Bytes(), &smartTx); err != nil {
		log.WithFields(log.Fields{"tx_hash": t.TxHash, "error": err, "type": consts.UnmarshallingError}).Error("unmarshalling smart tx msgpack")
		return err
	}
	t.TxPtr = nil
	t.TxSmart = &smartTx
	t.TxTime = smartTx.Time
	t.TxKeyID = smartTx.KeyID

	contract := smart.GetContractByID(int32(smartTx.Type))
	if contract == nil {
		log.WithFields(log.Fields{"contract_type": smartTx.Type, "type": consts.NotFound}).Error("unknown contract")
		return fmt.Errorf(`unknown contract %d`, smartTx.Type)
	}
	forsign := []string{smartTx.ForSign()}

	t.TxContract = contract
	t.TxHeader = &smartTx.Header

	t.TxData = make(map[string]interface{})
	txInfo := contract.Block.Info.(*script.ContractInfo).Tx

	if txInfo != nil {
		if err := t.fillTxData(*txInfo, smartTx.Params, forsign); err != nil {
			return err
		}
	} else {
		t.TxData[`forsign`] = strings.Join(forsign, ",")
	}

	return nil
}

// CheckTransaction is checking transaction
func CheckTransaction(data []byte) (*tx.Header, error) {
	trBuff := bytes.NewBuffer(data)
	t, err := UnmarshallTransaction(trBuff)
	if err != nil {
		return nil, err
	}

	err = t.Check(time.Now().Unix(), true)
	if err != nil {
		return nil, err
	}

	return t.TxHeader, nil
}

func (t *Transaction) Check(checkTime int64, checkForDupTr bool) error {
	err := CheckLogTx(t.TxFullData, checkForDupTr, false)
	if err != nil {
		return err
	}
	logger := log.WithFields(log.Fields{"tx_time": t.TxTime})
	// time in the transaction cannot be more than MAX_TX_FORW seconds of block time
	if t.TxTime-consts.MAX_TX_FORW > checkTime {
		logger.WithFields(log.Fields{"tx_max_forw": consts.MAX_TX_FORW, "type": consts.ParameterExceeded}).Error("time in the tx cannot be more than MAX_TX_FORW seconds of block time ")
		return utils.ErrInfo(fmt.Errorf("transaction time is too big"))
	}

	// time in transaction cannot be less than -24 of block time
	if t.TxTime < checkTime-consts.MAX_TX_BACK {
		logger.WithFields(log.Fields{"tx_max_back": consts.MAX_TX_BACK, "type": consts.ParameterExceeded}).Error("time in the tx cannot be less then -24 of block time")
		return utils.ErrInfo(fmt.Errorf("incorrect transaction time"))
	}

	if t.TxContract == nil {
		if t.BlockData != nil && t.BlockData.BlockID != 1 {
			if t.TxKeyID == 0 {
				logger.WithFields(log.Fields{"type": consts.EmptyObject}).Error("Empty user id")
				return utils.ErrInfo(fmt.Errorf("empty user id"))
			}
		}
	}

	return nil
}

func (t *Transaction) Play() (string, error) {
	// smart-contract
	if t.TxContract != nil {
		// check that there are enough money in CallContract
		return t.CallContract(smart.CallInit | smart.CallCondition | smart.CallAction)
	}

	if t.tx == nil {
		return "", utils.ErrInfo(fmt.Errorf("can't find parser for %d", t.TxType))
	}

	return "", t.tx.Action()
}

// AccessRights checks the access right by executing the condition value
func (t *Transaction) AccessRights(condition string, iscondition bool) error {
	logger := t.GetLogger()
	sp := &model.StateParameter{}
	sp.SetTablePrefix(converter.Int64ToStr(t.TxSmart.EcosystemID))
	_, err := sp.Get(t.DbTransaction, condition)
	if err != nil {
		logger.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("getting state parameter by name transaction")
		return err
	}
	conditions := sp.Value
	if iscondition {
		conditions = sp.Conditions
	}
	if len(conditions) > 0 {
		ret, err := t.SmartContract.EvalIf(conditions)
		if err != nil {
			logger.WithFields(log.Fields{"type": consts.EvalError, "error": err, "conditions": conditions}).Error("evaluating conditions")
			return err
		}
		if !ret {
			logger.WithFields(log.Fields{"type": consts.AccessDenied}).Error("Access denied")
			return fmt.Errorf(`Access denied`)
		}
	} else {
		logger.WithFields(log.Fields{"type": consts.EmptyObject, "conditions": condition}).Error("No condition in state_parameters")
		return fmt.Errorf(`There is not %s in state_parameters`, condition)
	}
	return nil
}

// CallContract calls the contract functions according to the specified flags
func (t *Transaction) CallContract(flags int) (resultContract string, err error) {
	sc := smart.SmartContract{
		VDE:           false,
		Rollback:      true,
		SysUpdate:     false,
		VM:            smart.GetVM(),
		TxSmart:       *t.TxSmart,
		TxData:        t.TxData,
		TxContract:    t.TxContract,
		TxCost:        t.TxCost,
		TxUsedCost:    t.TxUsedCost,
		BlockData:     t.BlockData,
		TxHash:        t.TxHash,
		PublicKeys:    t.PublicKeys,
		DbTransaction: t.DbTransaction,
		Rand:          t.Rand,
	}
	resultContract, err = sc.CallContract(flags)
	t.SysUpdate = sc.SysUpdate
	return
}

// CleanCache cleans cache of transaction parsers
func CleanCache() {
	txCache.Clean()
}

// GetTxTypeAndUserID returns tx type, wallet and citizen id from the block data
func GetTxTypeAndUserID(binaryBlock []byte) (txType int64, keyID int64) {
	tmp := binaryBlock[:]
	txType = converter.BinToDecBytesShift(&binaryBlock, 1)
	if consts.IsStruct(int(txType)) {
		var txHead consts.TxHeader
		converter.BinUnmarshal(&tmp, &txHead)
		keyID = txHead.KeyID
	}
	return
}

func GetTransaction(t *Transaction, txType string) (custom.TransactionInterface, error) {
	switch txType {
	case consts.TxTypeParserFirstBlock:
		return &custom.FirstBlockTransaction{t.GetLogger(), t.DbTransaction, t.TxPtr}, nil
	case consts.TxTypeParserStopNetwork:
		return &custom.StopNetworkTransaction{t.GetLogger(), t.TxPtr, nil}, nil
	}
	log.WithFields(log.Fields{"tx_type": txType, "type": consts.UnknownObject}).Error("unknown txType")
	return nil, fmt.Errorf("Unknown txType: %s", txType)
}
