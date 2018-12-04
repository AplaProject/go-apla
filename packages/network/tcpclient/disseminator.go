// Apla Software includes an integrated development
// environment with a multi-level system for the management
// of access rights to data, interfaces, and Smart contracts. The
// technical characteristics of the Apla Software are indicated in
// Apla Technical Paper.

// Apla Users are granted a permission to deal in the Apla
// Software without restrictions, including without limitation the
// rights to use, copy, modify, merge, publish, distribute, sublicense,
// and/or sell copies of Apla Software, and to permit persons
// to whom Apla Software is furnished to do so, subject to the
// following conditions:
// * the copyright notice of GenesisKernel and EGAAS S.A.
// and this permission notice shall be included in all copies or
// substantial portions of the software;
// * a result of the dealing in Apla Software cannot be
// implemented outside of the Apla Platform environment.

// THE APLA SOFTWARE IS PROVIDED “AS IS”, WITHOUT WARRANTY
// OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED
// TO THE WARRANTIES OF MERCHANTABILITY, FITNESS FOR A
// PARTICULAR PURPOSE, ERROR FREE AND NONINFRINGEMENT. IN
// NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE
// LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER
// IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING
// FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR
// THE USE OR OTHER DEALINGS IN THE APLA SOFTWARE.

package tcpclient

import (
	"bytes"
	"context"
	"errors"
	"net"
	"sync"
	"sync/atomic"

	"github.com/AplaProject/go-apla/packages/conf/syspar"
	"github.com/AplaProject/go-apla/packages/network"

	"github.com/AplaProject/go-apla/packages/model"

	"github.com/AplaProject/go-apla/packages/consts"
	"github.com/AplaProject/go-apla/packages/converter"
	log "github.com/sirupsen/logrus"
)

var (
	ErrNodesUnavailable = errors.New("All nodes unvailabale")
)

func SendTransactionsToHost(host string, txes []model.Transaction) error {
	packet := prepareTxPacket(txes)
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

func SendTransacitionsToAll(ctx context.Context, hosts []string, txes []model.Transaction) error {
	if len(hosts) == 0 || len(txes) == 0 {
		return nil
	}

	packet := prepareTxPacket(txes)

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

func SendFullBlockToAll(ctx context.Context, hosts []string, block *model.InfoBlock, txes []model.Transaction, nodeID int64) error {
	if len(hosts) == 0 {
		return nil
	}

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
	return resieveRequiredTransactions(con)
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

func resieveRequiredTransactions(con net.Conn) (response []byte, err error) {
	needTxResp := network.DisHashResponse{}
	if err := needTxResp.Read(con); err != nil {
		if err == network.ErrMaxSize {
			log.WithFields(log.Fields{"max_size": syspar.GetMaxTxSize(), "type": consts.ParameterExceeded}).Warning("response size is larger than max tx size")
			return nil, nil
		}

		log.WithFields(log.Fields{"type": consts.IOError, "error": err}).Error("reading data")
		return nil, err
	}

	return needTxResp.Data, err
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
	rt := network.RequestType{
		Type: uint16(requestType),
	}
	err = rt.Write(con)
	if err != nil {
		log.WithFields(log.Fields{"type": consts.IOError, "error": err}).Error("writing request type to host")
		return err
	}

	// data size
	// size := converter.DecToBin(len(packet), 4)
	// _, err = con.Write(size)
	// if err != nil {
	// 	log.WithFields(log.Fields{"type": consts.IOError, "error": err}).Error("writing data size to host")
	// 	return err
	// }

	// // data
	// _, err = con.Write(packet)
	// if err != nil {
	// 	log.WithFields(log.Fields{"type": consts.IOError, "error": err}).Error("writing data to host")
	// 	return err
	// }

	req := network.DisRequest{
		Data: packet,
	}

	return req.Write(con)
}
