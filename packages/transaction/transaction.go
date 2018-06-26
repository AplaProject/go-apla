package transaction

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	"github.com/GenesisKernel/go-genesis/packages/consts"
	"github.com/GenesisKernel/go-genesis/packages/converter"
	"github.com/GenesisKernel/go-genesis/packages/crypto"
	"github.com/GenesisKernel/go-genesis/packages/model"
	"github.com/GenesisKernel/go-genesis/packages/script"
	"github.com/GenesisKernel/go-genesis/packages/smart"
	"github.com/GenesisKernel/go-genesis/packages/transaction/custom"
	"github.com/GenesisKernel/go-genesis/packages/utils"
	"github.com/GenesisKernel/go-genesis/packages/utils/tx"

	"github.com/shopspring/decimal"
	log "github.com/sirupsen/logrus"
	"gopkg.in/vmihailenco/msgpack.v2"
)

// Transaction is a structure for parsing transactions
type Transaction struct {
	BlockData      *utils.BlockData
	PrevBlock      *utils.BlockData
	dataType       int
	blockData      []byte
	CurrentVersion string
	PublicKeys     [][]byte

	TxBinaryData   []byte // transaction binary data
	TxFullData     []byte // full transaction, with type and data
	TxHash         []byte
	TxKeyID        int64
	TxEcosystemID  int64
	TxNodePosition uint32
	TxTime         int64
	TxType         int64
	TxCost         int64           // Maximum cost of executing contract
	TxFuel         int64           // The fuel cost of executed contract
	TxUsedCost     decimal.Decimal // Used cost of CPU resources
	TxPtr          interface{}     // Pointer to the corresponding struct in consts/struct.go
	TxData         map[string]interface{}
	TxSmart        *tx.SmartContract
	TxContract     *smart.Contract
	TxHeader       *tx.Header
	tx             custom.TransactionInterface
	DbTransaction  *model.DbTransaction
	SysUpdate      bool

	SmartContract smart.SmartContract
}

// GetLogger returns logger
func (t Transaction) GetLogger() *log.Entry {
	if t.BlockData != nil && t.PrevBlock != nil {
		logger := log.WithFields(log.Fields{"block_id": t.BlockData.BlockID, "block_time": t.BlockData.Time, "block_wallet_id": t.BlockData.KeyID, "block_state_id": t.BlockData.EcosystemID, "block_hash": t.BlockData.Hash, "block_version": t.BlockData.Version, "prev_block_id": t.PrevBlock.BlockID, "prev_block_time": t.PrevBlock.Time, "prev_block_wallet_id": t.PrevBlock.KeyID, "prev_block_state_id": t.PrevBlock.EcosystemID, "prev_block_hash": t.PrevBlock.Hash, "prev_block_version": t.PrevBlock.Version, "tx_type": t.TxType, "tx_time": t.TxTime, "tx_state_id": t.TxEcosystemID, "tx_wallet_id": t.TxKeyID})
		return logger
	}
	if t.BlockData != nil {
		logger := log.WithFields(log.Fields{"block_id": t.BlockData.BlockID, "block_time": t.BlockData.Time, "block_wallet_id": t.BlockData.KeyID, "block_state_id": t.BlockData.EcosystemID, "block_hash": t.BlockData.Hash, "block_version": t.BlockData.Version, "tx_type": t.TxType, "tx_time": t.TxTime, "tx_state_id": t.TxEcosystemID, "tx_wallet_id": t.TxKeyID})
		return logger
	}
	if t.PrevBlock != nil {
		logger := log.WithFields(log.Fields{"prev_block_id": t.PrevBlock.BlockID, "prev_block_time": t.PrevBlock.Time, "prev_block_wallet_id": t.PrevBlock.KeyID, "prev_block_state_id": t.PrevBlock.EcosystemID, "prev_block_hash": t.PrevBlock.Hash, "prev_block_version": t.PrevBlock.Version, "tx_type": t.TxType, "tx_time": t.TxTime, "tx_state_id": t.TxEcosystemID, "tx_wallet_id": t.TxKeyID})
		return logger
	}
	logger := log.WithFields(log.Fields{"tx_type": t.TxType, "tx_time": t.TxTime, "tx_state_id": t.TxEcosystemID, "tx_wallet_id": t.TxKeyID})
	return logger
}

var txParserCache = &transactionCache{cache: make(map[string]*Transaction)}

// ParseTransaction is parsing transaction
func ParseTransaction(buffer *bytes.Buffer) (*Transaction, error) {
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

	if t, ok := txParserCache.Get(string(hash)); ok {
		return t, nil
	}

	t := new(Transaction)
	t.TxHash = hash
	t.TxUsedCost = decimal.New(0, 0)
	t.TxFullData = buffer.Bytes()

	txType := int64(buffer.Bytes()[0])
	t.dataType = int(txType)

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
	txParserCache.Set(t)

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

func (t *Transaction) parseFromContract(buf *bytes.Buffer) error {
	smartTx := tx.SmartContract{}
	if err := msgpack.Unmarshal(buf.Bytes(), &smartTx); err != nil {
		log.WithFields(log.Fields{"tx_type": t.dataType, "tx_hash": t.TxHash, "error": err, "type": consts.UnmarshallingError}).Error("unmarshalling smart tx msgpack")
		return err
	}
	t.TxPtr = nil
	t.TxSmart = &smartTx
	t.TxTime = smartTx.Time
	t.TxEcosystemID = (smartTx.EcosystemID)
	t.TxKeyID = smartTx.KeyID

	contract := smart.GetContractByID(int32(smartTx.Type))
	if contract == nil {
		log.WithFields(log.Fields{"contract_type": smartTx.Type, "type": consts.NotFound}).Error("unknown contract")
		return fmt.Errorf(`unknown contract %d`, smartTx.Type)
	}
	forsign := []string{smartTx.ForSign()}

	t.TxContract = contract
	t.TxHeader = &smartTx.Header

	input := smartTx.Data
	t.TxData = make(map[string]interface{})

	if contract.Block.Info.(*script.ContractInfo).Tx != nil {
		for _, fitem := range *contract.Block.Info.(*script.ContractInfo).Tx {
			var err error
			var v interface{}
			var forv string
			var isforv bool

			if fitem.ContainsTag(script.TagFile) {
				var (
					data []byte
					file *tx.File
				)
				if err := converter.BinUnmarshal(&input, &data); err != nil {
					log.WithFields(log.Fields{"error": err, "type": consts.UnmarshallingError}).Error("bin unmarshalling file")
					return err
				}
				if err := msgpack.Unmarshal(data, &file); err != nil {
					log.WithFields(log.Fields{"error": err, "type": consts.UnmarshallingError}).Error("unmarshalling file msgpack")
					return err
				}

				t.TxData[fitem.Name] = file.Data
				t.TxData[fitem.Name+"MimeType"] = file.MimeType

				forsign = append(forsign, file.MimeType, file.Hash)
				continue
			}

			switch fitem.Type.String() {
			case `uint64`:
				var val uint64
				converter.BinUnmarshal(&input, &val)
				v = val
			case `float64`:
				var val float64
				converter.BinUnmarshal(&input, &val)
				v = val
			case `int64`:
				v, err = converter.DecodeLenInt64(&input)
			case script.Decimal:
				var s string
				if err := converter.BinUnmarshal(&input, &s); err != nil {
					log.WithFields(log.Fields{"error": err, "type": consts.UnmarshallingError}).Error("bin unmarshalling script.Decimal")
					return err
				}
				v, err = decimal.NewFromString(s)
			case `string`:
				var s string
				if err := converter.BinUnmarshal(&input, &s); err != nil {
					log.WithFields(log.Fields{"error": err, "type": consts.UnmarshallingError}).Error("bin unmarshalling string")
					return err
				}
				v = s
			case `[]uint8`:
				var b []byte
				if err := converter.BinUnmarshal(&input, &b); err != nil {
					log.WithFields(log.Fields{"error": err, "type": consts.UnmarshallingError}).Error("bin unmarshalling string")
					return err
				}
				v = hex.EncodeToString(b)
			case `[]interface {}`:
				count, err := converter.DecodeLength(&input)
				if err != nil {
					log.WithFields(log.Fields{"error": err, "type": consts.UnmarshallingError}).Error("bin unmarshalling []interface{}")
					return err
				}
				isforv = true
				list := make([]interface{}, 0)
				for count > 0 {
					length, err := converter.DecodeLength(&input)
					if err != nil {
						log.WithFields(log.Fields{"error": err, "type": consts.UnmarshallingError}).Error("bin unmarshalling tx length")
						return err
					}
					if len(input) < int(length) {
						log.WithFields(log.Fields{"error": err, "type": consts.UnmarshallingError, "length": int(length), "slice length": len(input)}).Error("incorrect tx size")
						return fmt.Errorf(`input slice is short`)
					}
					list = append(list, string(input[:length]))
					input = input[length:]
					count--
				}
				if len(list) > 0 {
					slist := make([]string, len(list))
					for j, lval := range list {
						slist[j] = lval.(string)
					}
					forv = strings.Join(slist, `,`)
				}
				v = list
			}
			if t.TxData[fitem.Name] == nil {
				t.TxData[fitem.Name] = v
			}
			if err != nil {
				return err
			}
			if strings.Index(fitem.Tags, `image`) >= 0 {
				continue
			}
			if isforv {
				v = forv
			}
			forsign = append(forsign, fmt.Sprintf("%v", v))
		}
	}
	t.TxData[`forsign`] = strings.Join(forsign, ",")

	return nil
}

// CheckTransaction is checking transaction
func CheckTransaction(data []byte) (*tx.Header, error) {
	trBuff := bytes.NewBuffer(data)
	t, err := ParseTransaction(trBuff)
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
		return utils.ErrInfo(err)
	}
	logger := log.WithFields(log.Fields{"tx_type": t.dataType, "tx_time": t.TxTime, "tx_state_id": t.TxEcosystemID})
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

	err := t.tx.Action()
	if err != nil {
		return "", err
	}

	return "", nil
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
		VM:            smart.GetVM(false, 0),
		TxSmart:       *t.TxSmart,
		TxData:        t.TxData,
		TxContract:    t.TxContract,
		TxCost:        t.TxCost,
		TxUsedCost:    t.TxUsedCost,
		BlockData:     t.BlockData,
		TxHash:        t.TxHash,
		PublicKeys:    t.PublicKeys,
		DbTransaction: t.DbTransaction,
	}
	resultContract, err = sc.CallContract(flags)
	t.SysUpdate = sc.SysUpdate
	return
}

// CleanCache cleans cache of transaction parsers
func CleanCache() {
	txParserCache.Clean()
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
