package tcpclient

import (
	"context"
	"sync"

	"github.com/GenesisKernel/go-genesis/packages/consts"
	"github.com/GenesisKernel/go-genesis/packages/converter"
	"github.com/GenesisKernel/go-genesis/packages/network"
	"github.com/GenesisKernel/go-genesis/packages/utils"
	log "github.com/sirupsen/logrus"
)

func HostWithMaxBlock(hosts []string) (bestHost string, maxBlockID int64, err error) {
	if len(hosts) == 0 {
		return "", -1, nil
	}
	ctx := context.Background()
	return hostWithMaxBlock(ctx, hosts)
}

func GetMaxBlockID(host string) (blockID int64, err error) {
	ctx := context.Background()
	return getMaxBlock(ctx, host)
}

func getMaxBlock(ctx context.Context, host string) (blockID int64, err error) {
	con, err := newConnection(host)
	if err != nil {
		log.WithFields(log.Fields{"error": err, "type": consts.ConnectionError, "host": host}).Debug("error connecting to host")
		return -1, err
	}
	defer con.Close()

	// send max block request
	rt := &network.RequestType{
		Type: network.RequestTypeMaxBlock,
	}

	if err := rt.Write(con); err != nil {
		log.WithFields(log.Fields{"error": err, "type": consts.ConnectionError, "host": host}).Error("on sending Max block request type")
		return -1, err
	}

	// response
	blockIDBin := make([]byte, 4)
	_, err = con.Read(blockIDBin)
	if err != nil {
		log.WithFields(log.Fields{"error": err, "type": consts.ConnectionError, "host": host}).Error("reading max block id from host")
		return -1, err
	}

	return converter.BinToDec(blockIDBin), nil
}

func hostWithMaxBlock(ctx context.Context, hosts []string) (bestHost string, maxBlockID int64, err error) {
	maxBlockID = -1

	type blockAndHost struct {
		host    string
		blockID int64
		err     error
	}

	resultChan := make(chan blockAndHost, len(hosts))

	/* rand.Shuffle(len(hosts), func(i, j int) { hosts[i], hosts[j] = hosts[j], hosts[i] })
	this implementation available only in Golang 1.10
	*/
	utils.ShuffleSlice(hosts)

	var wg sync.WaitGroup
	for _, h := range hosts {
		if ctx.Err() != nil {
			log.WithFields(log.Fields{"error": ctx.Err(), "type": consts.ContextError}).Error("context error")
			return "", maxBlockID, ctx.Err()
		}

		wg.Add(1)

		go func(host string) {
			blockID, err := getMaxBlock(context.TODO(), host)
			defer wg.Done()

			resultChan <- blockAndHost{
				host:    host,
				blockID: blockID,
				err:     err,
			}
		}(h)
	}
	wg.Wait()

	var errCount int
	for i := 0; i < len(hosts); i++ {
		bl := <-resultChan

		if bl.err != nil {
			errCount++
			continue
		}

		// If blockID is maximal then the current host is the best
		if bl.blockID > maxBlockID {
			maxBlockID = bl.blockID
			bestHost = bl.host
		}
	}

	if errCount == len(hosts) {
		return "", 0, ErrNodesUnavailable
	}

	return bestHost, maxBlockID, nil
}
