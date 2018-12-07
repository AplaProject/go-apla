package modes

import (
	"github.com/GenesisKernel/go-genesis/packages/conf"
	"github.com/GenesisKernel/go-genesis/packages/types"
)

func GetDaemonListFactory() types.DaemonListFactory {
	if !conf.Config.IsSupportingOBS() {
		return BlockchainDaemonsListsFactory{}
	}

	return OBSDaemonsListFactory{}
}

type BlockchainDaemonsListsFactory struct{}

func (f BlockchainDaemonsListsFactory) GetDaemonsList() []string {
	return []string{
		"Scheduler",
	}
}

type OBSDaemonsListFactory struct{}

func (f OBSDaemonsListFactory) GetDaemonsList() []string {
	return []string{
		"BlocksCollection",
		"BlockGenerator",
		"QueueParserTx",
		"QueueParserBlocks",
		"Disseminator",
		"Confirmations",
		"Scheduler",
	}
}
