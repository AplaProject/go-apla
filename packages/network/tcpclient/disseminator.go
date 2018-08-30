package tcpclient

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net"
	"sync"
	"sync/atomic"

	"github.com/GenesisKernel/go-genesis/packages/network"

	"github.com/GenesisKernel/go-genesis/packages/blockchain"
	"github.com/GenesisKernel/go-genesis/packages/conf/syspar"

	"github.com/GenesisKernel/go-genesis/packages/consts"
	"github.com/GenesisKernel/go-genesis/packages/converter"
	"github.com/GenesisKernel/go-genesis/packages/utils/tx"
	log "github.com/sirupsen/logrus"
)

var (
	ErrNodesUnavailable = errors.New("All nodes unvailabale")
)

func SendTransactionsToHost(host string, txes []*tx.SmartContract) error {
	packet, err := prepareTxPacket(txes)
	if err != nil {
		return err
	}
	return sendRawTransacitionsToHost(host, packet)
}

func sendRawTransacitionsToHost(host string, packet []byte) error {
	con, err := newConnection(host)
	if err != nil {
		log.WithFields(log.Fields{"type": consts.NetworkError, "error": err, "host": host}).Error("on creating tcp connection")
		return err
	}

	defer con.Close()

	if err := sendDisseminatorRequest(con, network.RequestTypeNotFullNode, packet); err != nil {
		log.WithFields(log.Fields{"type": consts.TCPClientError, "error": err, "host": host}).Error("on sending disseminator request")
		return err
	}
	return nil
}

func SendTransacitionsToAll(ctx context.Context, hosts []string, txes []*tx.SmartContract) error {
	if len(hosts) == 0 || len(txes) == 0 {
		return nil
	}

	packet, err := prepareTxPacket(txes)
	if err != nil {
		return err
	}

	var wg sync.WaitGroup
	var errCount int32
	for _, h := range hosts {
		if err := ctx.Err(); err != nil {
			log.Debug("exit by context error")
			return err
		}

		wg.Add(1)
		go func(host string, pak []byte) {
			defer wg.Done()

			if err := sendRawTransacitionsToHost(host, pak); err != nil {
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

func SendFullBlockToAll(ctx context.Context, hosts []string, block *blockchain.Block, txes []*tx.SmartContract, nodeID int64) error {
	if len(hosts) == 0 {
		return nil
	}

	req, err := prepareFullBlockRequest(block, txes, nodeID)
	if err != nil {
		return err
	}
	txDataMap := make(map[string][]byte, len(txes))
	for _, tx := range txes {
		hash, err := tx.Hash()
		data, err := tx.Marshal()
		if err != nil {
			return err
		}
		txDataMap[string(hash)] = data
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

			con, err := newConnection(h)
			if err != nil {
				increaseErrCount()
				log.WithFields(log.Fields{"type": consts.NetworkError, "error": err, "host": h}).Error("on creating tcp connection")
				return
			}

			defer con.Close()

			response, err := sendFullBlockRequest(con, req)
			if err != nil {
				increaseErrCount()
				log.WithFields(log.Fields{"type": consts.NetworkError, "error": err, "host": h}).Error("on sending full block request")
				return
			}

			if len(response) == 0 || len(response) < consts.HashSize {
				return
			}

			var buf bytes.Buffer
			requestedHashes := parseTxHashesFromResponse(response)
			for _, txhash := range requestedHashes {
				if data, ok := txDataMap[string(txhash)]; ok && len(data) > 0 {
					log.WithFields(log.Fields{"len_of_tx": len(data)}).Debug("on prepare full tx package")
					if _, err := buf.Write(converter.EncodeLengthPlusData(data)); err != nil {
						log.WithFields(log.Fields{"type": consts.IOError, "error": err}).Warn("on write tx hash to response buffer")
					}
				}
			}

			if _, err = con.Write(converter.DecToBin(buf.Len(), 4)); err != nil {
				increaseErrCount()
				log.WithFields(log.Fields{"type": consts.IOError, "error": err, "host": h}).Error("on writing requested transactions buf length")
				return
			}

			if _, err = con.Write(buf.Bytes()); err != nil {
				increaseErrCount()
				log.WithFields(log.Fields{"type": consts.IOError, "error": err, "host": h}).Error("on writing requested transactions")
				return
			}
		}(host)
	}

	wg.Wait()

	if int(errCount) == len(hosts) {
		return ErrNodesUnavailable
	}

	return nil
}

func sendFullBlockRequest(con net.Conn, data []byte) (response []byte, err error) {

	if err := sendDisseminatorRequest(con, network.RequestTypeFullNode, data); err != nil {
		log.WithFields(log.Fields{"type": consts.TCPClientError, "error": err}).Error("on sending disseminator request")
		return nil, err
	}

	//response
	return sendRequiredTransactions(con)
}

func prepareTxPacket(txes []*tx.SmartContract) ([]byte, error) {
	// form packet to send
	var buf bytes.Buffer
	for _, tr := range txes {
		data, err := tr.Marshal()
		if err != nil {
			return nil, err
		}
		buf.Write(data)
	}

	return buf.Bytes(), nil
}

func prepareFullBlockRequest(block *blockchain.Block, trs []*tx.SmartContract, nodeID int64) ([]byte, error) {
	var noBlockFlag byte
	if block == nil {
		noBlockFlag = 1
	}

	var buf bytes.Buffer
	buf.Write(converter.DecToBin(nodeID, 8))
	buf.WriteByte(noBlockFlag)
	if noBlockFlag == 0 {
		b, err := block.Marshal()
		if err != nil {
			return nil, err
		}
		buf.Write(b)
	}
	if trs != nil {
		for _, tr := range trs {
			hash, err := tr.Hash()
			if err != nil {
				return nil, err
			}
			buf.Write(hash)
		}
	}

	return buf.Bytes(), nil
}

func sendRequiredTransactions(con net.Conn) (response []byte, err error) {
	buf := make([]byte, 4)

	// read data size
	_, err = io.ReadFull(con, buf)
	if err != nil {
		if err == io.EOF {
			log.WithFields(log.Fields{"type": consts.IOError, "error": err}).Warn("connection closed unexpectedly")
		} else {
			log.WithFields(log.Fields{"type": consts.IOError, "error": err}).Error("reading data size")
		}

		return nil, err
	}

	respSize := converter.BinToDec(buf)
	if respSize > syspar.GetMaxTxSize() {
		log.WithFields(log.Fields{"size": respSize, "max_size": syspar.GetMaxTxSize(), "type": consts.ParameterExceeded}).Warning("response size is larger than max tx size")
		return nil, nil
	}
	// read the data
	response = make([]byte, respSize)
	_, err = io.ReadFull(con, response)
	if err != nil {
		log.WithFields(log.Fields{"type": consts.IOError, "error": err}).Error("reading data")
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

func sendDisseminatorRequest(con net.Conn, requestType int, packet []byte) (err error) {
	/*
		Packet format:
		type  2 bytes
		len   4 bytes
		data  len bytes
	*/
	// type
	_, err = con.Write(converter.DecToBin(requestType, 2))
	if err != nil {
		log.WithFields(log.Fields{"type": consts.IOError, "error": err}).Error("writing request type to host")
		return err
	}

	// data size
	size := converter.DecToBin(len(packet), 4)
	_, err = con.Write(size)
	if err != nil {
		log.WithFields(log.Fields{"type": consts.IOError, "error": err}).Error("writing data size to host")
		return err
	}

	// data
	_, err = con.Write(packet)
	if err != nil {
		log.WithFields(log.Fields{"type": consts.IOError, "error": err}).Error("writing data to host")
		return err
	}

	return nil
}
