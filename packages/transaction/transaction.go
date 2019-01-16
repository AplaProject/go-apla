// Apla Software includes an integrated development
// environment with a multi-level system for the management
// of access rights to data, interfaces, and Smart contracts. The
// technical characteristics of the Apla Software are indicated in
// Apla Technical Paper.

// Apla Users are granted a permission to deal in the Apla
// Software without restrictions, including without limitation the
// rights to use, copy, modify, merge, publish, distribute, sublicense,
// and/or sell copies of Apla Software, and to permit persons
// to whom Apla Software is furnished to do so, subject to the
// following conditions:
// * the copyright notice of GenesisKernel and EGAAS S.A.
// and this permission notice shall be included in all copies or
// substantial portions of the software;
// * a result of the dealing in Apla Software cannot be
// implemented outside of the Apla Platform environment.

// THE APLA SOFTWARE IS PROVIDED “AS IS”, WITHOUT WARRANTY
// OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED
// TO THE WARRANTIES OF MERCHANTABILITY, FITNESS FOR A
// PARTICULAR PURPOSE, ERROR FREE AND NONINFRINGEMENT. IN
// NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE
// LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER
// IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING
// FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR
// THE USE OR OTHER DEALINGS IN THE APLA SOFTWARE.

package transaction

import (
	"bytes"
	"fmt"
	"math/rand"
	"time"

	"github.com/AplaProject/go-apla/packages/consts"
	"github.com/AplaProject/go-apla/packages/converter"
	"github.com/AplaProject/go-apla/packages/crypto"
	"github.com/AplaProject/go-apla/packages/model"
	"github.com/AplaProject/go-apla/packages/script"
	"github.com/AplaProject/go-apla/packages/smart"
	"github.com/AplaProject/go-apla/packages/transaction/custom"
	"github.com/AplaProject/go-apla/packages/utils"
	"github.com/AplaProject/go-apla/packages/utils/tx"

	"github.com/shopspring/decimal"
	log "github.com/sirupsen/logrus"
	"gopkg.in/vmihailenco/msgpack.v2"
)

type RawTransaction struct {
	txType    int64
	hash      []byte
	data      []byte
	payload   []byte
	signature []byte
}

func (rtx *RawTransaction) Unmarshall(buffer *bytes.Buffer) error {
	if buffer.Len() == 0 {
		log.WithFields(log.Fields{"type": consts.EmptyObject}).Error("empty transaction buffer")
		return fmt.Errorf("empty transaction buffer")
	}

	rtx.data = buffer.Bytes()

	b, err := buffer.ReadByte()
	if err != nil {
		return err
	}
	rtx.txType = int64(b)

	if IsContractTransaction(rtx.txType) {
		if err = converter.BinUnmarshalBuff(buffer, &rtx.payload); err != nil {
			return err
		}
		rtx.signature = buffer.Bytes()
	} else {
		buffer.UnreadByte()
		rtx.payload = buffer.Bytes()
	}

	if rtx.hash, err = crypto.DoubleHash(rtx.payload); err != nil {
		log.WithFields(log.Fields{"type": consts.CryptoError, "error": err}).Error("hashing transaction")
		return err
	}

	return nil
}

func (rtx *RawTransaction) Type() int64 {
	return rtx.txType
}

func (rtx *RawTransaction) Hash() []byte {
	return rtx.hash
}

func (rtx *RawTransaction) Bytes() []byte {
	return rtx.data
}

func (rtx *RawTransaction) Payload() []byte {
	return rtx.payload
}

func (rtx *RawTransaction) Signature() []byte {
	return rtx.signature
}

// Transaction is a structure for parsing transactions
type Transaction struct {
	BlockData  *utils.BlockData
	PrevBlock  *utils.BlockData
	PublicKeys [][]byte

	TxBinaryData  []byte // transaction binary data
	TxFullData    []byte // full transaction, with type and data
	TxHash        []byte
	TxSignature   []byte
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
	Notifications []smart.NotifyInfo

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
func UnmarshallTransaction(buffer *bytes.Buffer, fillData bool) (*Transaction, error) {
	rtx := &RawTransaction{}
	if err := rtx.Unmarshall(buffer); err != nil {
		return nil, err
	}

	if t, ok := txCache.Get(string(rtx.Hash())); ok {
		return t, nil
	}

	t := new(Transaction)
	t.TxFullData = rtx.Bytes()
	t.TxType = rtx.Type()
	t.TxHash = rtx.Hash()
	t.TxBinaryData = rtx.Payload()
	t.TxUsedCost = decimal.New(0, 0)

	// smart contract transaction
	if IsContractTransaction(rtx.Type()) {
		t.TxSignature = rtx.Signature()
		// skip byte with transaction type
		if err := t.parseFromContract(fillData); err != nil {
			return nil, err
		}
		// struct transaction (only first block transaction for now)
	} else if consts.IsStruct(rtx.Type()) {
		if err := t.parseFromStruct(); err != nil {
			return t, err
		}

		// all other transactions
	}
	txCache.Set(t)

	return t, nil
}

// IsContractTransaction checks txType
func IsContractTransaction(txType int64) bool {
	return txType > 127
}

func (t *Transaction) parseFromStruct() error {
	t.TxPtr = consts.MakeStruct(consts.TxTypes[t.TxType])
	if err := converter.BinUnmarshal(&t.TxBinaryData, t.TxPtr); err != nil {
		log.WithFields(log.Fields{"error": err, "type": consts.UnmarshallingError, "tx_type": t.TxType}).Error("getting parser for tx type")
		return err
	}
	head := consts.Header(t.TxPtr)
	t.TxKeyID = head.KeyID
	t.TxTime = int64(head.Time)

	trParser, err := GetTransaction(t, consts.TxTypes[t.TxType])
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

func (t *Transaction) fillTxData(fieldInfos []*script.FieldInfo, params map[string]interface{}) error {
	var err error
	t.TxData, err = smart.FillTxData(fieldInfos, params)
	if err != nil {
		return err
	}
	return nil
}

func (t *Transaction) parseFromContract(fillData bool) error {
	smartTx := tx.SmartContract{}
	if err := msgpack.Unmarshal(t.TxBinaryData, &smartTx); err != nil {
		log.WithFields(log.Fields{"tx_hash": t.TxHash, "error": err, "type": consts.UnmarshallingError}).Error("unmarshalling smart tx msgpack")
		return err
	}
	t.TxPtr = nil
	t.TxSmart = &smartTx
	t.TxTime = smartTx.Time
	t.TxKeyID = smartTx.KeyID

	key := &model.Key{}
	key.SetTablePrefix(smartTx.EcosystemID)
	found, err := key.Get(smartTx.KeyID)
	if !found {
		log.WithFields(log.Fields{"type": consts.DBError, "error": err, "tx_key": t.TxKeyID}).Error("key ID not found")
		return err
	}
	if key.ReadOnly == 1 {
		return fmt.Errorf("transaction aborted because of read_only key")
	}

	contract := smart.GetContractByID(int32(smartTx.ID))
	if contract == nil {
		log.WithFields(log.Fields{"contract_id": smartTx.ID, "type": consts.NotFound}).Error("unknown contract")
		return fmt.Errorf(`unknown contract %d`, smartTx.ID)
	}

	t.TxContract = contract
	t.TxHeader = &smartTx.Header

	t.TxData = make(map[string]interface{})
	txInfo := contract.Block.Info.(*script.ContractInfo).Tx

	if txInfo != nil {
		if fillData {
			if err := t.fillTxData(*txInfo, smartTx.Params); err != nil {
				return err
			}
		} else {
			t.TxData = smartTx.Params
			for key, item := range t.TxData {
				if v, ok := item.(map[interface{}]interface{}); ok {
					imap := make(map[string]interface{})
					for ikey, ival := range v {
						imap[fmt.Sprint(ikey)] = ival
					}
					t.TxData[key] = imap
				}
			}
		}
	}

	return nil
}

// CheckTransaction is checking transaction
func CheckTransaction(data []byte) (*tx.Header, error) {
	trBuff := bytes.NewBuffer(data)
	t, err := UnmarshallTransaction(trBuff, true)
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
	err := CheckLogTx(t.TxHash, checkForDupTr, false)
	if err != nil {
		return err
	}
	logger := log.WithFields(log.Fields{"tx_time": t.TxTime})
	// time in the transaction cannot be more than MAX_TX_FORW seconds of block time
	if t.TxTime > checkTime {
		if t.TxTime-consts.MAX_TX_FORW > checkTime {
			logger.WithFields(log.Fields{"tx_max_forw": consts.MAX_TX_FORW, "type": consts.ParameterExceeded}).Error("time in the tx cannot be more than MAX_TX_FORW seconds of block time ")
			return utils.ErrInfo(fmt.Errorf("transaction time is too big"))
		}
		return ErrEarlyTime
	}

	// time in transaction cannot be less than -24 of block time
	if t.TxTime < checkTime-consts.MAX_TX_BACK {
		logger.WithFields(log.Fields{"tx_max_back": consts.MAX_TX_BACK, "type": consts.ParameterExceeded, "tx_time": t.TxTime}).Error("time in the tx cannot be less then -24 of block time")
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

func (t *Transaction) Play() (string, []smart.FlushInfo, error) {
	// smart-contract
	if t.TxContract != nil {
		// check that there are enough money in CallContract
		return t.CallContract()
	}

	if t.tx == nil {
		return "", nil, utils.ErrInfo(fmt.Errorf("can't find parser for %d", t.TxType))
	}

	return "", nil, t.tx.Action()
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
func (t *Transaction) CallContract() (resultContract string, flushRollback []smart.FlushInfo, err error) {
	sc := smart.SmartContract{
		OBS:           false,
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
		TxSignature:   t.TxSignature,
		TxSize:        int64(len(t.TxBinaryData)),
		PublicKeys:    t.PublicKeys,
		DbTransaction: t.DbTransaction,
		Rand:          t.Rand,
	}
	resultContract, err = sc.CallContract()
	t.TxFuel = sc.TxFuel
	t.SysUpdate = sc.SysUpdate
	t.Notifications = sc.Notifications
	if sc.FlushRollback != nil {
		flushRollback = make([]smart.FlushInfo, len(sc.FlushRollback))
		copy(flushRollback, sc.FlushRollback)
	}
	return
}

func (t *Transaction) CallOBSContract() (resultContract string, flushRollback []smart.FlushInfo, err error) {
	sc := smart.SmartContract{
		OBS:           true,
		Rollback:      false,
		SysUpdate:     false,
		VM:            smart.GetVM(),
		TxSmart:       *t.TxSmart,
		TxData:        t.TxData,
		TxContract:    t.TxContract,
		TxCost:        t.TxCost,
		TxUsedCost:    t.TxUsedCost,
		BlockData:     t.BlockData,
		TxHash:        t.TxHash,
		TxSignature:   t.TxSignature,
		TxSize:        int64(len(t.TxBinaryData)),
		PublicKeys:    t.PublicKeys,
		DbTransaction: t.DbTransaction,
		Rand:          t.Rand,
	}
	resultContract, err = sc.CallContract()
	t.SysUpdate = sc.SysUpdate
	t.Notifications = sc.Notifications
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
	if consts.IsStruct(txType) {
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
