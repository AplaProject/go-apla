package tcpclient

import (
	"context"

	"github.com/GenesisKernel/go-genesis/packages/consts"
	"github.com/GenesisKernel/go-genesis/packages/network"
	log "github.com/sirupsen/logrus"
)

// GetBlocksBodies send GetBodiesRequest returns channel of binary blocks data
func GetBlocksBodies(ctx context.Context, host string, blockID int64, reverseOrder bool) (chan []byte, error) {
	conn, err := newConnection(host)
	if err != nil {
		return nil, err
	}

	// send the type of data
	rt := &network.RequestType{Type: network.RequestTypeBlockCollection}
	if err = rt.Write(conn); err != nil {
		log.WithFields(log.Fields{"type": consts.IOError, "error": err}).Error("writing data type block body to connection")
		return nil, err
	}

	req := &network.GetBodiesRequest{
		BlockID:      uint32(blockID),
		ReverseOrder: reverseOrder,
	}

	if err = req.Write(conn); err != nil {
		log.WithFields(log.Fields{"type": consts.IOError, "error": err}).Error("on sending blocks bodies request")
		return nil, err
	}

	blocksCount, err := network.ReadInt(conn)
	if err != nil {
		log.WithFields(log.Fields{"type": consts.NetworkError, "error": err}).Error("on getting blocks count")
		return nil, err
	}

	if blocksCount == 0 {
		log.Warnf("host: %s does'nt contains block", host)
		return nil, nil
	}

	rawBlocksCh := make(chan []byte, blocksCount)

	go func() {
		defer func() {
			close(rawBlocksCh)
			conn.Close()
		}()

		for i := 0; i < int(blocksCount); i++ {
			if err := ctx.Err(); err != nil {
				log.Debug(err)
				return
			}

			// receive the data size as a response that server wants to transfer
			resp := &network.GetBodyResponse{}
			if err := resp.Read(conn); err != nil {
				log.WithFields(log.Fields{"type": consts.NetworkError, "error": err, "host": host}).Error("on reading block bodies")
				return
			}

			dataSize := len(resp.Data)
			if dataSize == 0 {
				log.Error("null block")
				return
			}

			rawBlocksCh <- resp.Data
		}
	}()
	return rawBlocksCh, nil

}
