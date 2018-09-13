package transaction

import (
	"encoding/hex"
	"fmt"
	"math/rand"
	"strings"
	"time"

	"github.com/GenesisKernel/go-genesis/packages/blockchain"
	"github.com/GenesisKernel/go-genesis/packages/consts"
	"github.com/GenesisKernel/go-genesis/packages/converter"
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
	BlockData  *blockchain.BlockHeader
	PrevBlock  *blockchain.BlockHeader
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
	TxSmart       *blockchain.Transaction
	TxContract    *smart.Contract
	TxHeader      *blockchain.TxHeader
	tx            custom.TransactionInterface
	DbTransaction *model.DbTransaction
	Rand          *rand.Rand
	SysUpdate     bool

	SmartContract smart.SmartContract
}

// GetLogger returns logger
func (t Transaction) GetLogger() *log.Entry {
	logger := log.WithFields(log.Fields{"tx_type": t.TxType, "tx_time": t.TxTime, "tx_wallet_id": t.TxKeyID})
	if t.BlockData != nil {
		logger = logger.WithFields(log.Fields{"block_id": t.BlockData.BlockID, "block_time": t.BlockData.Time, "block_wallet_id": t.BlockData.KeyID, "block_state_id": t.BlockData.EcosystemID, "block_version": t.BlockData.Version})
	}
	if t.PrevBlock != nil {
		logger = logger.WithFields(log.Fields{"block_id": t.BlockData.BlockID, "block_time": t.BlockData.Time, "block_wallet_id": t.BlockData.KeyID, "block_state_id": t.BlockData.EcosystemID, "block_version": t.BlockData.Version})
	}
	return logger
}

var txCache = &transactionCache{cache: make(map[string]*Transaction)}

// UnmarshallTransaction is unmarshalling transaction
func FromBlockchainTransaction(tx *blockchain.Transaction) (*Transaction, error) {
	hash, err := tx.Hash()
	if err != nil {
		log.WithFields(log.Fields{"type": consts.CryptoError, "error": err}).Error("hashing transaction")
		return nil, err
	}

	if t, ok := txCache.Get(string(hash)); ok {
		return t, nil
	}
	bytes, err := tx.Marshal()
	if err != nil {
		return nil, err
	}

	t := new(Transaction)
	t.TxHash = hash
	t.TxUsedCost = decimal.New(0, 0)
	t.TxFullData = bytes

	// skip byte with transaction type
	t.TxBinaryData = bytes
	if err := t.parseFromContract(tx); err != nil {
		return nil, err
	}

	txCache.Set(t)

	return t, nil
}

func (t *Transaction) ToBlockchainTransaction() (*blockchain.Transaction, error) {
	tx := &blockchain.Transaction{}
	if err := tx.Unmarshal(t.TxFullData); err != nil {
		return nil, err
	}
	return tx, nil
}

func (t *Transaction) fillTxData(fieldInfos []*script.FieldInfo, params map[string]string, forsign []string) error {
	for _, fitem := range fieldInfos {
		var err error
		var v interface{}
		var forv string
		var isforv bool

		if fitem.ContainsTag(script.TagFile) {
			var (
				data []byte
				file *tx.File
			)
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
			val = converter.StrToUint64(params[fitem.Name])
			v = val
		case `float64`:
			var val float64
			val = converter.StrToFloat64(params[fitem.Name])
			v = val
		case `int64`:
			v = converter.StrToInt64(params[fitem.Name])
		case script.Decimal:
			v, err = decimal.NewFromString(params[fitem.Name])
		case `string`:
			v = params[fitem.Name]
		case `[]uint8`:
			v, err = hex.DecodeString(params[fitem.Name])
		case `[]interface {}`:
			var list []string
			for key, value := range params {
				if key == fitem.Name+`[]` && len(value) > 0 {
					count := converter.StrToInt(value)
					for i := 0; i < count; i++ {
						list = append(list, params[fmt.Sprintf(`%s[%d]`, fitem.Name, i)])
					}
				}
			}
			if len(list) > 0 {
				forv = strings.Join(list, `,`)
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
	t.TxData[`forsign`] = strings.Join(forsign, ",")
	return nil
}

func (t *Transaction) parseFromContract(smartTx *blockchain.Transaction) error {
	t.TxPtr = nil
	t.TxSmart = smartTx
	t.TxTime = smartTx.Header.Time
	t.TxKeyID = smartTx.Header.KeyID

	contract := smart.GetContractByID(int32(smartTx.Header.Type))
	if contract == nil {
		log.WithFields(log.Fields{"contract_type": smartTx.Header.Type, "type": consts.NotFound}).Error("unknown contract")
		return fmt.Errorf(`unknown contract %d`, smartTx.Header.Type)
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
	}

	return nil
}

// CheckTransaction is checking transaction
func CheckTransaction(bTx *blockchain.Transaction) error {
	t, err := FromBlockchainTransaction(bTx)
	if err != nil {
		return err
	}

	err = t.Check(time.Now().Unix())
	if err != nil {
		return err
	}

	return nil
}

func (t *Transaction) Check(checkTime int64) error {
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
	// check that there are enough money in CallContract
	return t.CallContract(smart.CallInit | smart.CallCondition | smart.CallAction)
}

// AccessRights checks the access right by executing the condition value
func (t *Transaction) AccessRights(condition string, iscondition bool) error {
	logger := t.GetLogger()
	sp := &model.StateParameter{}
	sp.SetTablePrefix(converter.Int64ToStr(t.TxSmart.Header.EcosystemID))
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
	}
	resultContract, err = sc.CallContract(flags)
	t.SysUpdate = sc.SysUpdate
	return
}

// CleanCache cleans cache of transaction parsers
func CleanCache() {
	txCache.Clean()
}
