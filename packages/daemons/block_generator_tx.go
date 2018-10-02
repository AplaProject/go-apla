package daemons

import (
	"encoding/hex"
	"time"

	"github.com/GenesisKernel/go-genesis/packages/consts"
	"github.com/GenesisKernel/go-genesis/packages/model"
	"github.com/GenesisKernel/go-genesis/packages/smart"
	"github.com/GenesisKernel/go-genesis/packages/utils/tx"

	log "github.com/sirupsen/logrus"
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
	vm := smart.GetVM()
	contract := smart.VMGetContract(vm, callDelayedContract, uint32(firstEcosystemID))
	info := contract.Info()

	smartTx := tx.SmartContract{
		Header: tx.Header{
			ID:          int(info.ID),
			Time:        time.Now().Unix(),
			EcosystemID: firstEcosystemID,
			KeyID:       keyID,
			NetworkID:   consts.NETWORK_ID,
		},
		SignedBy: smart.PubToID(dtx.publicKey),
		Params: map[string]interface{}{
			"Id": delayedContactID,
		},
	}

	privateKey, err := hex.DecodeString(dtx.privateKey)
	if err != nil {
		return err
	}

	txData, txHash, err := tx.NewInternalTransaction(smartTx, privateKey)
	if err != nil {
		return err
	}

	return tx.CreateTransaction(txData, txHash, keyID)
}
