package api

import "github.com/GenesisKernel/go-genesis/packages/transaction"

type ContractHandler func(rtx *transaction.RawTransaction) error

func BlockchainContractHandler(rtx *transaction.RawTransaction) error {
	return nil
}

func VDEContractHandler(rtx *transaction.RawTransaction) error {
	return nil
}
