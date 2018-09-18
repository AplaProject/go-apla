package tcpclient

import (
	"context"
	"errors"
	"io"

	"github.com/GenesisKernel/go-genesis/packages/consts"
	"github.com/GenesisKernel/go-genesis/packages/network"
	log "github.com/sirupsen/logrus"
)

var ErrorEmptyBlockBody = errors.New("block is empty")

var BlockBodyPool BytesPull

type BytesPull struct {
	pull     chan []byte
	sliceCap int
}

const hasVal = "has value"
const hasntVal = "has not value"

func (bbf BytesPull) getBytes() []byte {
	select {
	case slice, ok := <-bbf.pull:
		if ok {
			// fmt.Println("buf exists cap", cap(slice))
			return slice
		}
	default:
	}

	return nil
}

func (bbf BytesPull) putBytes(slice []byte) {
	if cap(slice) > bbf.sliceCap {
		return
	}

	slice = slice[:0]
	bbf.pull <- slice
}

func InitBlockBodyBuffer(cap, maxSliceCap int) {
	BlockBodyPool = BytesPull{
		pull:     make(chan []byte, cap),
		sliceCap: maxSliceCap,
	}
}

// GetBlocksBodies send GetBodiesRequest returns channel of binary blocks data
func GetBlocksBodies(ctx context.Context, host string, blockID int64, reverseOrder bool) (<-chan []byte, error) {
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

	blocksChan, errChan := GetBlockBodiesChan(ctx, conn, blocksCount)
	go func() {
		if err := <-errChan; err != nil {
			log.WithFields(log.Fields{"type": "dbError", consts.NetworkError: err}).Error("on reading block bodies")
		}
	}()

	return blocksChan, nil
}

func GetBlockBodiesChan(ctx context.Context, src io.ReadCloser, blocksCount int64) (<-chan []byte, <-chan error) {
	rawBlocksCh := make(chan []byte, blocksCount)
	errChan := make(chan error, 1)

	go func() {
		defer func() {
			close(rawBlocksCh)
			close(errChan)
			src.Close()
		}()

		for i := 0; i < int(blocksCount); i++ {
			if err := ctx.Err(); err != nil {
				log.Debug(err)
				return
			}

			// receive the data size as a response that server wants to transfer
			resp := &network.GetBodyResponse{}
			if err := resp.Read(src); err != nil {
				log.WithFields(log.Fields{"type": consts.NetworkError, "error": err}).Error("on reading block bodies")
				errChan <- err
				return
			}

			if len(resp.Data) == 0 {
				errChan <- ErrorEmptyBlockBody
				return
			}

			rawBlocksCh <- resp.Data
		}
	}()

	return rawBlocksCh, errChan
}

func GetBlockBodiesChanWithPool(ctx context.Context, src io.ReadCloser, blocksCount int64) (<-chan []byte, <-chan error) {
	rawBlocksCh := make(chan []byte, blocksCount)
	errChan := make(chan error, 1)

	go func() {
		defer func() {
			close(rawBlocksCh)
			close(errChan)
			src.Close()
		}()

		for i := 0; i < int(blocksCount); i++ {
			if err := ctx.Err(); err != nil {
				log.Debug(err)
				return
			}

			// receive the data size as a response that server wants to transfer
			resp := &network.BodyResponse{}
			if err := resp.Read(src, BlockBodyPool.getBytes()); err != nil {
				log.WithFields(log.Fields{"type": consts.NetworkError, "error": err}).Error("on reading block bodies")
				errChan <- err
				return
			}

			if len(resp.Data) == 0 {
				errChan <- ErrorEmptyBlockBody
				return
			}

			rawBlocksCh <- resp.Data
		}
	}()

	return rawBlocksCh, errChan
}
