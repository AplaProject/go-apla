package daemons

import (
	"encoding/hex"
	"fmt"
	"time"

	"github.com/GenesisKernel/go-genesis/packages/consts"
	"github.com/GenesisKernel/go-genesis/packages/converter"
	"github.com/GenesisKernel/go-genesis/packages/crypto"
	"github.com/GenesisKernel/go-genesis/packages/model"
	"github.com/GenesisKernel/go-genesis/packages/script"
	"github.com/GenesisKernel/go-genesis/packages/smart"
	"github.com/GenesisKernel/go-genesis/packages/utils/tx"

	log "github.com/sirupsen/logrus"
	msgpack "gopkg.in/vmihailenco/msgpack.v2"
)

const (
	callDelayedContract = "CallDelayedContract"
	firstEcosystemID    = 1
)

// DelayedTx represents struct which works with delayed contracts
type DelayedTx struct {
	logger     *log.Entry
	privateKey string
	publicKey  string
}

// RunForBlockID creates the transactions that need to be run for blockID
func (dtx *DelayedTx) RunForBlockID(blockID int64) {
	contracts, err := model.GetAllDelayedContractsForBlockID(blockID)
	if err != nil {
		dtx.logger.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("getting delayed contracts for block")
		return
	}

	for _, c := range contracts {
		if err := dtx.createTx(c.ID, c.KeyID); err != nil {
			dtx.logger.WithFields(log.Fields{"error": err}).Debug("can't create transaction for delayed contract")
		}
	}
}

func (dtx *DelayedTx) createTx(delayedContactID, keyID int64) error {
	vm := smart.GetVM(false, 0)
	contract := smart.VMGetContract(vm, callDelayedContract, uint32(firstEcosystemID))
	info := contract.Block.Info.(*script.ContractInfo)

	params := make([]byte, 0)
	converter.EncodeLenInt64(&params, delayedContactID)

	smartTx := tx.SmartContract{
		Header: tx.Header{
			Type:        int(info.ID),
			Time:        time.Now().Unix(),
			EcosystemID: firstEcosystemID,
			KeyID:       keyID,
		},
		SignedBy: smart.PubToID(dtx.publicKey),
		Data:     params,
	}

	signature, err := crypto.Sign(
		dtx.privateKey,
		fmt.Sprintf("%s,%d", smartTx.ForSign(), delayedContactID),
	)
	if err != nil {
		dtx.logger.WithFields(log.Fields{"type": consts.CryptoError, "error": err}).Error("signing by node private key")
		return err
	}
	smartTx.BinSignatures = converter.EncodeLengthPlusData(signature)

	if smartTx.PublicKey, err = hex.DecodeString(dtx.publicKey); err != nil {
		dtx.logger.WithFields(log.Fields{"type": consts.ConversionError, "error": err}).Error("decoding public key from hex")
		return err
	}

	data, err := msgpack.Marshal(smartTx)
	if err != nil {
		dtx.logger.WithFields(log.Fields{"type": consts.MarshallingError, "error": err}).Error("marshalling smart contract to msgpack")
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
		KeyID:    keyID,
		HighRate: model.TransactionRateOnBlock,
	}
	if err = tx.Create(); err != nil {
		dtx.logger.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("creating new transaction")
		return err
	}

	return nil
}
