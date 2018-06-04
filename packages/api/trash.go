package api

import (
	"net/http"
	"time"

	"github.com/GenesisKernel/go-genesis/packages/conf"
	"github.com/GenesisKernel/go-genesis/packages/consts"
	"github.com/GenesisKernel/go-genesis/packages/converter"
	"github.com/GenesisKernel/go-genesis/packages/script"
	"github.com/GenesisKernel/go-genesis/packages/smart"
	"github.com/GenesisKernel/go-genesis/packages/utils"
	"github.com/GenesisKernel/go-genesis/packages/utils/tx"
	log "github.com/sirupsen/logrus"
)

func getContract(r *http.Request, name string) *smart.Contract {
	client := getClient(r)
	vm := smart.GetVM(client.IsVDE, client.EcosystemID)
	if vm == nil {
		return nil
	}
	return smart.VMGetContract(vm, name, uint32(client.EcosystemID))
}

func createTx(contract *smart.Contract, values ...string) error {
	NodePrivateKey, NodePublicKey, err := utils.GetNodeKeys()
	if err != nil || len(NodePrivateKey) < 1 {
		if err == nil {
			log.WithFields(log.Fields{"type": consts.EmptyObject}).Error("node private key is empty")
		}
		return err
	}

	params := make([]byte, 0)
	for _, v := range values {
		params = append(append(params, converter.EncodeLength(int64(len(v)))...), v...)
	}

	info := contract.Block.Info.(*script.ContractInfo)
	err = tx.BuildTransaction(tx.SmartContract{
		Header: tx.Header{
			Type:        int(info.ID),
			Time:        time.Now().Unix(),
			EcosystemID: 1,
			KeyID:       conf.Config.KeyID,
			NetworkID:   consts.NETWORK_ID,
		},
		SignedBy: smart.PubToID(NodePublicKey),
		Data:     params,
	}, NodePrivateKey, NodePublicKey, values...)
	if err != nil {
		log.WithFields(log.Fields{"type": consts.ContractError}).Error("Executing contract")
		return err
	}

	return nil
}
