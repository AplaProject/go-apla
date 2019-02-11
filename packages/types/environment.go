package types

import (
	"context"

	log "github.com/sirupsen/logrus"
)

// RemoteNode allow work with remote nodes
type RemoteNode interface {
	MaxBlock(ctx context.Context) (int64, error)
	GetBlocks(ctx context.Context, lastBlockID int64, reverseOrder bool) (<-chan []byte, error)
	TCPHost() string
}

// RemoteEnvironment interface allows to interact with a remote environment
type RemoteEnvironment interface {
	RemoteNodes() []RemoteNode
	NodeWithMaxBlock(context.Context, *log.Entry) (host RemoteNode, maxBlockID int64, err error)
}

type NodesBanService interface {
	RegisterBadBlock(node FullNode, badBlockId, blockTime int64, reason string) error
	IsBanned(FullNode) bool
}
