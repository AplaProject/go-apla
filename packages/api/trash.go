package api

import (
	"fmt"
	"net/http"
	"regexp"
	"strings"
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

func getTestValue(v string) string {
	return smart.GetTestValue(v)
}

type getter interface {
	Get(string) string
}

func validateParamsContract(contract *smart.Contract, params getter) (err error) {
	if contract.Block.Info.(*script.ContractInfo).Tx == nil {
		return
	}

	for _, fitem := range *(*contract).Block.Info.(*script.ContractInfo).Tx {
		if fitem.ContainsTag(script.TagFile) || fitem.ContainsTag(script.TagCrypt) || fitem.ContainsTag(script.TagSignature) {
			continue
		}

		val := strings.TrimSpace(params.Get(fitem.Name))
		if fitem.Type.String() == "[]interface {}" {
			count := params.Get(fitem.Name + "[]")
			if converter.StrToInt(count) > 0 || len(val) > 0 {
				continue
			}
			val = ""
		}
		if len(val) == 0 && !fitem.ContainsTag(script.TagOptional) {
			log.WithFields(log.Fields{"type": consts.EmptyObject, "item_name": fitem.Name}).Error("route item is empty")
			err = fmt.Errorf("%s is empty", fitem.Name)
			break
		}
		if fitem.ContainsTag(script.TagAddress) {
			addr := converter.StringToAddress(val)
			if addr == 0 {
				log.WithFields(log.Fields{"type": consts.ConversionError, "value": val}).Error("converting string to address")
				err = fmt.Errorf("Address %s is not valid", val)
				break
			}
		}
		if fitem.Type.String() == script.Decimal {
			re := regexp.MustCompile(`^\d+$`)
			if !re.Match([]byte(val)) {
				log.WithFields(log.Fields{"type": consts.InvalidObject, "value": val}).Error("The value of money is not valid")
				err = fmt.Errorf("The value of money %s is not valid", val)
				break
			}
		}
	}

	return
}

func packParamsContract(contract *smart.Contract, params getter) {

}
