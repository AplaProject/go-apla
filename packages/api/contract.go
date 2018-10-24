package api

import "github.com/GenesisKernel/go-genesis/packages/blockchain"

type ContractHandler func(rtx *blockchain.Transaction) error

func BlockchainContractHandler(rtx *blockchain.Transaction) error {
	return nil
}

func VDEContractHandler(rtx *blockchain.Transaction) error {
	return nil
}
