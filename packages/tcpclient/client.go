package tcpclient

import (
	"bytes"
	"context"
	"errors"
	"io"
	"math/rand"
	"net"
	"sync"
	"sync/atomic"
	"time"

	"github.com/GenesisKernel/go-genesis/packages/conf/syspar"
	"github.com/GenesisKernel/go-genesis/packages/model"

	"github.com/GenesisKernel/go-genesis/packages/consts"
	"github.com/GenesisKernel/go-genesis/packages/converter"
	log "github.com/sirupsen/logrus"
)

const (
	// I_AM_FULL_NODE is full node flag
	I_AM_FULL_NODE = 1
	// I_AM_NOT_FULL_NODE is not full node flag
	I_AM_NOT_FULL_NODE = 2
)

var (
	ErrNodesUnavailable = errors.New("All nodes unvailabale")
)

type Config struct {
	DefaultPort  int64
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
}

type client struct {
	*log.Entry
	config Config
}

func NewClient(config Config, logger *log.Entry) *client {
	return &client{
		config: config,
		Entry:  logger,
	}
}

func (c *client) HostWithMaxBlock(hosts []string) (bestHost string, maxBlockID int64, err error) {
	ctx := context.Background()
	return c.hostWithMaxBlock(ctx, hosts)
}

func (c *client) GetMaxBlockID(host string) (blockID int64, err error) {
	ctx := context.Background()
	return c.getMaxBlock(ctx, host)
}

func (c *client) getMaxBlock(ctx context.Context, host string) (blockID int64, err error) {
	con, err := c.newConnection(host)

	if err != nil {
		c.WithFields(log.Fields{"error": err, "type": consts.ConnectionError, "host": host}).Debug("error connecting to host")
		return -1, err
	}
	defer con.Close()

	// send max block request
	_, err = con.Write(converter.DecToBin(consts.DATA_TYPE_MAX_BLOCK_ID, 2))
	if err != nil {
		c.WithFields(log.Fields{"error": err, "type": consts.ConnectionError, "host": host}).Error("writing max block id to host")
		return -1, err
	}

	// response
	blockIDBin := make([]byte, 4)
	_, err = con.Read(blockIDBin)
	if err != nil {
		c.WithFields(log.Fields{"error": err, "type": consts.ConnectionError, "host": host}).Error("reading max block id from host")
		return -1, err
	}

	return converter.BinToDec(blockIDBin), nil
}

func (c *client) hostWithMaxBlock(ctx context.Context, hosts []string) (bestHost string, maxBlockID int64, err error) {
	maxBlockID = -1

	if len(hosts) == 0 {
		return bestHost, maxBlockID, nil
	}

	type blockAndHost struct {
		host    string
		blockID int64
		err     error
	}

	resultChan := make(chan blockAndHost, len(hosts))

	rand.Shuffle(len(hosts), func(i, j int) { hosts[i], hosts[j] = hosts[j], hosts[i] })

	var wg sync.WaitGroup
	for _, h := range hosts {
		if ctx.Err() != nil {
			c.WithFields(log.Fields{"error": ctx.Err(), "type": consts.ContextError}).Error("context error")
			return "", maxBlockID, ctx.Err()
		}

		wg.Add(1)

		go func(host string) {
			blockID, err := c.getMaxBlock(context.TODO(), host)
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

// GetBlocksBody is retrieving `blocksCount` blocks bodies starting with blockID and puts them in the channel
func (c *client) GetBlocksBodies(host string, blockID int64, blocksCount int32, reverseOrder bool) (chan []byte, error) {
	conn, err := c.newConnection(host)
	if err != nil {
		return nil, err
	}

	// send the type of data
	_, err = conn.Write(converter.DecToBin(consts.DATA_TYPE_BLOCK_BODY, 2))
	if err != nil {
		log.WithFields(log.Fields{"type": consts.IOError, "error": err}).Error("writing data type block body to connection")
		return nil, err
	}

	// send the number of a block
	_, err = conn.Write(converter.DecToBin(blockID, 4))
	if err != nil {
		log.WithFields(log.Fields{"type": consts.IOError, "error": err}).Error("writing data type block body to connection")
		return nil, err
	}

	rvBytes := make([]byte, 1)
	if reverseOrder {
		rvBytes[0] = 1
	} else {
		rvBytes[0] = 0
	}

	// send reverse flag
	_, err = conn.Write(rvBytes)
	if err != nil {
		log.WithFields(log.Fields{"type": consts.IOError, "error": err}).Error("writing reverse flag to connection")
		return nil, err
	}

	rawBlocksCh := make(chan []byte, blocksCount)
	go func() {
		defer func() {
			close(rawBlocksCh)
			conn.Close()
		}()

		for {
			// receive the data size as a response that server wants to transfer
			buf := make([]byte, 4)
			_, err = conn.Read(buf)
			if err != nil {
				if err != io.EOF {
					log.WithFields(log.Fields{"type": consts.IOError, "error": err}).Error("reading block data size from connection")
				}
				return
			}
			dataSize := converter.BinToDec(buf)
			var binaryBlock []byte

			// data size must be less than 10mb
			if dataSize >= 10485760 || dataSize == 0 {
				log.Error("null block")
				return
			}

			binaryBlock = make([]byte, dataSize)

			_, err = io.ReadFull(conn, binaryBlock)
			if err != nil {
				log.WithFields(log.Fields{"type": consts.IOError, "error": err}).Error("reading block data from connection")
				return
			}

			rawBlocksCh <- binaryBlock
		}
	}()
	return rawBlocksCh, nil

}

func (c *client) SendTransactionsToHost(host string, txes []model.Transaction) error {
	packet := prepareTxPacket(txes)
	return c.sendRawTransacitionsToHost(host, packet)
}

func (c *client) sendRawTransacitionsToHost(host string, packet []byte) error {
	con, err := c.newConnection(host)
	if err != nil {
		log.WithFields(log.Fields{"type": consts.NetworkError, "error": err, "host": host}).Error("on creating ctp connection")
		return err
	}

	defer con.Close()

	/*
		Packet format:
		type  2 bytes
		len   4 bytes
		data  len bytes
	*/
	// type
	_, err = con.Write(converter.DecToBin(I_AM_NOT_FULL_NODE, 2))
	if err != nil {
		c.WithFields(log.Fields{"type": consts.IOError, "error": err, "host": host}).Error("writing request type to host")
		return err
	}

	// data size
	size := converter.DecToBin(len(packet), 4)
	_, err = con.Write(size)
	if err != nil {
		c.WithFields(log.Fields{"type": consts.IOError, "error": err, "host": host}).Error("writing data size to host")
		return err
	}

	// data
	_, err = con.Write(packet)
	if err != nil {
		c.WithFields(log.Fields{"type": consts.IOError, "error": err, "host": host}).Error("writing data to host")
		return err
	}

	return nil
}

func (c *client) SendTransacitionsToAll(hosts []string, txes []model.Transaction) error {
	packet := prepareTxPacket(txes)

	var wg sync.WaitGroup
	var errCount int32
	for _, h := range hosts {
		wg.Add(1)
		go func(host string, pak []byte) {
			defer wg.Done()

			if err := c.sendRawTransacitionsToHost(host, pak); err != nil {
				atomic.AddInt32(&errCount, 1)
			}
		}(h, packet)
	}

	wg.Wait()

	if int(errCount) == len(hosts) {
		return ErrNodesUnavailable
	}

	return nil
}

func (c *client) SendFullBlockToAll(hosts []string, block *model.InfoBlock, txes []model.Transaction, nodeID int64) error {
	req := prepareFullBlockRequest(block, txes, nodeID)
	txDataMap := make(map[string][]byte, len(txes))
	for _, tx := range txes {
		txDataMap[string(tx.Hash)] = tx.Data
	}

	for _, host := range hosts {
		go func(h string) {
			con, err := c.newConnection(h)
			if err != nil {
				log.WithFields(log.Fields{"type": consts.NetworkError, "error": err, "host": h}).Error("on creating ctp connection")
				return
			}

			response, err := c.sendFullBlockRequest(con, req)
			if err != nil {
				c.WithFields(log.Fields{"type": consts.NetworkError, "error": err, "host": h}).Error("on sending full block request")
				return
			}

			if len(response) == 0 || len(response) < consts.HashSize {
				return
			}

			var buf bytes.Buffer
			requestedHashes := parseTxHashesFromResponse(response)
			for _, rh := range requestedHashes {
				if data, ok := txDataMap[string(rh)]; ok && len(data) > 0 {
					buf.Write(converter.EncodeLengthPlusData(data))
				}
			}

			if _, err := io.Copy(con, bytes.NewReader(buf.Bytes())); err != nil {
				log.WithFields(log.Fields{"type": consts.IOError, "error": err, "host": h}).Error("on writing requested transactions")
			}
		}(host)
	}

	return nil
}

func (c *client) sendFullBlockRequest(con net.Conn, data []byte) (response []byte, err error) {
	// type
	_, err = con.Write(converter.DecToBin(I_AM_FULL_NODE, 2))
	if err != nil {
		c.WithFields(log.Fields{"type": consts.IOError, "error": err}).Error("writing request type to host")
		return nil, err
	}

	// data size
	size := converter.DecToBin(len(data), 4)
	_, err = con.Write(size)
	if err != nil {
		c.WithFields(log.Fields{"type": consts.IOError, "error": err}).Error("writing data size to host")
		return nil, err
	}

	_, err = con.Write(data)
	if err != nil {
		c.WithFields(log.Fields{"type": consts.IOError, "error": err}).Error("writing data to host")
		return nil, err
	}

	//response
	return c.sendRequiredTransactions(con)
}

func (c client) newConnection(addr string) (net.Conn, error) {
	host, err := NormalizeHostAddress(addr, c.config.DefaultPort)
	if err != nil {
		log.WithFields(log.Fields{"type": consts.NetworkError, "host": addr, "error": err}).Error("on normalize host address")
		return nil, err
	}

	conn, err := net.DialTimeout("tcp", host, consts.TCPConnTimeout)
	if err != nil {
		c.WithFields(log.Fields{"type": consts.ConnectionError, "error": err, "address": host}).Debug("dialing tcp")
		return nil, err
	}

	conn.SetReadDeadline(time.Now().Add(c.config.ReadTimeout))
	conn.SetWriteDeadline(time.Now().Add(c.config.WriteTimeout))
	return conn, nil
}

func prepareTxPacket(txes []model.Transaction) []byte {
	// form packet to send
	var buf bytes.Buffer
	for _, tr := range txes {
		buf.Write(tr.Data)
	}

	return buf.Bytes()
}

func prepareFullBlockRequest(block *model.InfoBlock, trs []model.Transaction, nodeID int64) []byte {
	var noBlockFlag byte
	if block == nil {
		noBlockFlag = 1
	}

	var buf bytes.Buffer
	buf.Write(converter.DecToBin(nodeID, 8))
	buf.WriteByte(noBlockFlag)
	if noBlockFlag == 0 {
		buf.Write(block.Marshall())
	}
	if trs != nil {
		for _, tr := range trs {
			buf.Write(tr.Hash)
		}
	}

	return buf.Bytes()
}

func (c client) sendRequiredTransactions(con net.Conn) (response []byte, err error) {
	buf := make([]byte, 4)

	// read data size
	_, err = io.ReadFull(con, buf)
	if err != nil {
		if err == io.EOF {
			c.WithFields(log.Fields{"type": consts.IOError, "error": err}).Warn("connection closed unexpectedly")
		} else {
			c.WithFields(log.Fields{"type": consts.IOError, "error": err}).Error("reading data size")
		}

		return nil, err
	}

	respSize := converter.BinToDec(buf)
	if respSize > syspar.GetMaxTxSize() {
		c.WithFields(log.Fields{"size": respSize, "max_size": syspar.GetMaxTxSize(), "type": consts.ParameterExceeded}).Warning("response size is larger than max tx size")
		return nil, nil
	}
	// read the data
	response = make([]byte, respSize)
	_, err = io.ReadFull(con, response)
	if err != nil {
		c.WithFields(log.Fields{"type": consts.IOError, "error": err}).Error("reading data")
		return nil, err
	}

	return response, err
}

func parseTxHashesFromResponse(resp []byte) (hashes [][]byte) {
	hashes = make([][]byte, 0, len(resp)/consts.HashSize)
	for len(resp) >= consts.HashSize {
		hashes = append(hashes, converter.BytesShift(&resp, consts.HashSize))
	}

	return
}
