package remote

import (
	"context"

	"github.com/AplaProject/go-apla/packages/network/tcpclient"
	"github.com/AplaProject/go-apla/packages/types"
)

type Node struct {
	*types.FullNode
}

func (node Node) MaxBlock(context.Context) (int64, error) {
	return tcpclient.GetMaxBlockID(node.TCPAddress)
}

func (node Node) GetBlocks(ctx context.Context, lastBlockID int64, reverseOrder bool) (<-chan []byte, error) {
	return tcpclient.GetBlocksBodies(ctx, node.TCPAddress, lastBlockID, reverseOrder)
}

func (node Node) TCPHost() string {
	return node.TCPAddress
}
