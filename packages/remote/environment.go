package remote

import (
	"context"
	"math/rand"
	"sync"

	"github.com/AplaProject/go-apla/packages/conf/syspar"
	"github.com/AplaProject/go-apla/packages/consts"
	"github.com/AplaProject/go-apla/packages/network/tcpclient"
	"github.com/AplaProject/go-apla/packages/types"
	log "github.com/sirupsen/logrus"
)

type Environtment struct {
	banService types.NodesBanService
}

func (e Environtment) RemoteNodes() []types.RemoteNode {
	nodes := syspar.GetRemoteNodes()
	remoteNodes := make([]types.RemoteNode, len(nodes))

	for i := 0; i < len(nodes); i++ {
		remoteNodes = append(remoteNodes, Node{nodes[i]})
	}

	return remoteNodes
}

func (e Environtment) NodeWithMaxBlock(ctx context.Context, logger *log.Entry) (bestNode types.RemoteNode, maxBlockID int64, err error) {
	nodes := e.RemoteNodes()
	maxBlockID = -1

	type blockAndHost struct {
		node    types.RemoteNode
		blockID int64
		err     error
	}

	resultChan := make(chan blockAndHost, len(nodes))

	rand.Shuffle(len(nodes), func(i, j int) { nodes[i], nodes[j] = nodes[j], nodes[i] })

	var wg sync.WaitGroup
	for _, node := range nodes {
		if ctx.Err() != nil {
			log.WithFields(log.Fields{"error": ctx.Err(), "type": consts.ContextError}).Error("context error")
			return nil, maxBlockID, ctx.Err()
		}

		remoteNode := node.(Node)
		if e.banService.IsBanned(*remoteNode.FullNode) {
			continue
		}

		wg.Add(1)

		go func(rn Node) {
			blockID, err := rn.MaxBlock(ctx)
			defer wg.Done()

			resultChan <- blockAndHost{
				node:    rn,
				blockID: blockID,
				err:     err,
			}
		}(remoteNode)
	}

	wg.Wait()
	close(resultChan)

	var errCount int
	for bl := range resultChan {

		if bl.err != nil {
			errCount++
			continue
		}

		// If blockID is maximal then the current host is the best
		if bl.blockID > maxBlockID {
			maxBlockID = bl.blockID
			bestNode = bl.node
		}
	}

	if errCount == len(nodes) {
		return nil, 0, tcpclient.ErrNodesUnavailable
	}

	return bestNode, maxBlockID, nil
}
