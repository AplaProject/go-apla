package packages

import (
	"github.com/GenesisKernel/go-genesis/packages/model"
)

// RemoteBlockService allow returns max block from host and find host with maxBlockID
type RemoteBlockService interface {
	GetMaxBlockID(host string) (blockID int64, err error)
	HostWithMaxBlock(hosts []string) (host string, maxBlockID int64, err error)
	GetBlocksBodies(host string, startBlock int64, blocksCount int, reverseOrder bool) (chan []byte, error)
	SendTransactions(host string, txes []model.Transaction) error
}
