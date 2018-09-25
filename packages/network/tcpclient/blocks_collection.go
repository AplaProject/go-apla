package tcpclient

import (
	"bytes"
	"context"
	"encoding/binary"
	"errors"
	"io"
	"sync"

	"github.com/GenesisKernel/go-genesis/packages/consts"
	"github.com/GenesisKernel/go-genesis/packages/network"
	log "github.com/sirupsen/logrus"
)

var ErrorEmptyBlockBody = errors.New("block is empty")
var ErrorWrongSizeBytes = errors.New("wrong size bytes")

const hasVal = "has value"
const hasntVal = "has not value"

const sizeBytesLength = 4

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

func GetBlockBodiesChan(ctx context.Context, src io.Reader, blocksCount int64) (<-chan []byte, <-chan error) {
	rawBlocksCh := make(chan []byte, blocksCount)
	errChan := make(chan error, 1)

	go func() {
		defer func() {
			close(rawBlocksCh)
			close(errChan)
			// src.Close()
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

var poolCounter int
var blockBodyPool = sync.Pool{
	New: func() interface{} {
		poolCounter++
		// fmt.Println("new buffer", poolCounter)
		return bytes.NewBuffer(make([]byte, 0, 12832256))
	},
}

func getBlockBuf() *bytes.Buffer {
	return blockBodyPool.Get().(*bytes.Buffer)
}

var returningBufCounter int

func putBlockBuf(b *bytes.Buffer) {
	b.Reset()
	blockBodyPool.Put(b)
	returningBufCounter++
	//fmt.Println("returned buffer", returningBufCounter, "CAP", b.Cap())
}

// var iterCounter int

func GetBlockBodiesChanReadAll(ctx context.Context, src io.ReadCloser, blocksCount int64) (<-chan []byte, <-chan error) {
	rawBlocksCh := make(chan []byte, blocksCount)
	errChan := make(chan error, 1)
	bodyBuf := getBlockBuf()
	afterBodyProcessed := func(done <-chan struct{}) {
		<-done
		putBlockBuf(bodyBuf)
	}

	go func() {
		defer func() {
			close(rawBlocksCh)
			close(errChan)
			// src.Close()
			go afterBodyProcessed(ctx.Done())
		}()

		_, err := io.Copy(bodyBuf, src)
		if err != nil {
			log.WithFields(log.Fields{"type": consts.NetworkError, "error": err}).Error("on reading block bodies packet")
			errChan <- err
			return
		}

		resp := bodyBuf.Bytes()

		for i := 0; i < int(blocksCount); i++ {

			size, intErr := binary.Uvarint(resp[:4])
			if intErr < 0 {
				log.WithFields(log.Fields{"type": consts.ConversionError, "bufLen": len(resp), "error": ErrorWrongSizeBytes}).Error("on convert size body")
				errChan <- ErrorWrongSizeBytes
				return
			}

			endPos := sizeBytesLength + size
			rawBlocksCh <- resp[sizeBytesLength:endPos]
			resp = resp[endPos:]
		}
	}()

	return rawBlocksCh, errChan
}

//===========================
func GetBlockBodiesChanByBlock(ctx context.Context, src io.ReadCloser, blocksCount int64) (<-chan []byte, <-chan error) {
	rawBlocksCh := make(chan []byte, blocksCount)
	errChan := make(chan error, 1)
	bodyBuf := getBlockBuf()
	sizeBuf := make([]byte, 4)

	afterBodyProcessed := func(done <-chan struct{}) {
		<-done
		putBlockBuf(bodyBuf)
	}

	go func() {
		defer func() {
			close(rawBlocksCh)
			close(errChan)
			// src.Close()
			go afterBodyProcessed(ctx.Done())
		}()

		for i := 0; i < int(blocksCount); i++ {

			_, err := io.ReadFull(src, sizeBuf)
			if err != nil {
				log.WithFields(log.Fields{"type": consts.IOError, "error": err}).Error("on reading size of block data")
			}

			size, intErr := binary.Uvarint(sizeBuf)
			if intErr < 0 {
				log.WithFields(log.Fields{"type": consts.ConversionError, "error": ErrorWrongSizeBytes}).Error("on convert size body")
				errChan <- ErrorWrongSizeBytes
				return
			}

			if readed, err := io.CopyN(bodyBuf, src, int64(size)); err != nil {
				log.WithFields(log.Fields{"type": consts.IOError, "size": size, "readed": readed, "error": err}).Error("on reading block body")
				errChan <- err
				return
			}

			data := bodyBuf.Bytes()
			startPos := len(data) - int(size)
			rawBlocksCh <- data[startPos:]
		}
	}()

	return rawBlocksCh, errChan
}

//============================
func GetBlockBodiesChanByBlockWithBytePool(ctx context.Context, src io.ReadCloser, blocksCount int64) (<-chan []byte, <-chan error) {
	rawBlocksCh := make(chan []byte, blocksCount)
	errChan := make(chan error, 1)

	sizeBuf := make([]byte, 4)
	var bodyBuf []byte
	afterBodyProcessed := func(done <-chan struct{}) {
		<-done
		BytesPool.Put(bodyBuf)
	}

	go func() {
		defer func() {
			close(rawBlocksCh)
			close(errChan)
			// src.Close()
			go afterBodyProcessed(ctx.Done())
		}()

		dataSize, err := network.ReadInt(src)
		if err != nil {
			errChan <- err
			return
		}
		// fmt.Println("Recieved dataSize", dataSize)

		bodyBuf = BytesPool.Get(dataSize)
		var bodyStartIndx int64

		for i := 0; i < int(blocksCount); i++ {

			_, err := io.ReadFull(src, sizeBuf)
			if err != nil {
				log.WithFields(log.Fields{"type": consts.IOError, "error": err}).Error("on reading size of block data")
				errChan <- err
				return
			}

			size, intErr := binary.Uvarint(sizeBuf)
			if intErr < 0 {
				log.WithFields(log.Fields{"type": consts.ConversionError, "error": ErrorWrongSizeBytes}).Error("on convert size body")
				errChan <- ErrorWrongSizeBytes
				return
			}

			bodyEndIndx := bodyStartIndx + int64(size)
			body := bodyBuf[bodyStartIndx:bodyEndIndx]
			if readed, err := io.ReadFull(src, body); err != nil {
				log.WithFields(log.Fields{"type": consts.IOError, "size": size, "readed": readed, "error": err}).Error("on reading block body")
				errChan <- err
				return
			}

			bodyStartIndx = bodyEndIndx
			rawBlocksCh <- body
		}
	}()

	return rawBlocksCh, errChan
}

//===========================
func GetBlockBodiesReadAll(ctx context.Context, src io.ReadCloser, blocksCount int64) ([][]byte, error) {
	rawBlocks := make([][]byte, 0, blocksCount)
	// errChan := make(chan error, 1)
	bodyBuf := getBlockBuf()
	afterBodyProcessed := func(done <-chan struct{}) {
		<-done
		putBlockBuf(bodyBuf)
	}

	defer func() {
		// src.Close()
		go afterBodyProcessed(ctx.Done())
	}()

	_, err := io.Copy(bodyBuf, src)
	if err != nil {
		log.WithFields(log.Fields{"type": consts.NetworkError, "error": err}).Error("on reading block bodies packet")
		return nil, err
	}

	resp := bodyBuf.Bytes()

	for i := 0; i < int(blocksCount); i++ {
		size, intErr := binary.Uvarint(resp[:4])
		if intErr < 0 {
			log.WithFields(log.Fields{"type": consts.ConversionError, "bufLen": len(resp), "error": ErrorWrongSizeBytes}).Error("on convert size body")
			return nil, err
		}

		endPos := sizeBytesLength + size
		rawBlocks = append(rawBlocks, resp[sizeBytesLength:endPos])
		resp = resp[endPos:]
	}

	return rawBlocks, nil
}
