package api

import (
	"net/http"

	"github.com/AplaProject/go-apla/packages/script"
	"github.com/AplaProject/go-apla/packages/smart"
)

func getContract(r *http.Request, name string) *smart.Contract {
	vm := smart.GetVM()
	if vm == nil {
		return nil
	}
	client := getClient(r)
	contract := smart.VMGetContract(vm, name, uint32(client.EcosystemID))
	if contract == nil {
		return nil
	}
	return contract
}

func getContractInfo(contract *smart.Contract) *script.ContractInfo {
	return contract.Block.Info.(*script.ContractInfo)
}
