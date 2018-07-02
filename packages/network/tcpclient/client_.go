package tcpclient

import (
	"bytes"
	"errors"
	"io"
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

	if err := c.sendDisseminatorRequest(con, I_AM_NOT_FULL_NODE, packet); err != nil {
		log.WithFields(log.Fields{"type": consts.TCPClientError, "error": err, "host": host}).Error("on sending disseminator request")
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

func (c *client) newConnection(host string) (net.Conn, error) {
	return net.Dial("tcp", host)
}

func (c *client) SendFullBlockToAll(hosts []string, block *model.InfoBlock, txes []model.Transaction, nodeID int64) error {
	req := prepareFullBlockRequest(block, txes, nodeID)
	txDataMap := make(map[string][]byte, len(txes))
	for _, tx := range txes {
		txDataMap[string(tx.Hash)] = tx.Data
	}

	var errCount int32
	increaseErrCount := func() {
		atomic.AddInt32(&errCount, 1)
	}

	var wg sync.WaitGroup
	for _, host := range hosts {
		wg.Add(1)

		go func(h string) {
			defer wg.Done()

			con, err := c.newConnection(h)
			if err != nil {
				increaseErrCount()
				log.WithFields(log.Fields{"type": consts.NetworkError, "error": err, "host": h}).Error("on creating ctp connection")
				return
			}

			defer con.Close()

			response, err := c.sendFullBlockRequest(con, req)
			if err != nil {
				increaseErrCount()
				c.WithFields(log.Fields{"type": consts.NetworkError, "error": err, "host": h}).Error("on sending full block request")
				return
			}

			if len(response) == 0 || len(response) < consts.HashSize {
				return
			}

			var buf bytes.Buffer
			requestedHashes := parseTxHashesFromResponse(response)
			for _, txhash := range requestedHashes {
				if data, ok := txDataMap[string(txhash)]; ok && len(data) > 0 {
					if _, err := buf.Write(converter.EncodeLengthPlusData(data)); err != nil {
						log.WithFields(log.Fields{"type": consts.IOError, "error": err}).Warn("on write tx hash to response buffer")
					}
				}
			}

			if _, err := io.Copy(con, bytes.NewReader(buf.Bytes())); err != nil {
				increaseErrCount()
				log.WithFields(log.Fields{"type": consts.IOError, "error": err, "host": h}).Error("on writing requested transactions")
			}
		}(host)
	}

	wg.Wait()

	if int(errCount) == len(hosts) {
		return ErrNodesUnavailable
	}

	return nil
}

func (c *client) sendFullBlockRequest(con net.Conn, data []byte) (response []byte, err error) {

	if err := c.sendDisseminatorRequest(con, I_AM_FULL_NODE, data); err != nil {
		c.WithFields(log.Fields{"type": consts.TCPClientError, "error": err}).Error("on sending disseminator request")
		return nil, err
	}

	//response
	return c.sendRequiredTransactions(con)
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
