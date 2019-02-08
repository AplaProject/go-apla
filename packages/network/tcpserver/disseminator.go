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

package tcpserver

import (
	"bytes"
	"errors"
	"io"

	"github.com/AplaProject/go-apla/packages/consts"
	"github.com/AplaProject/go-apla/packages/converter"
	"github.com/AplaProject/go-apla/packages/model"
	"github.com/AplaProject/go-apla/packages/network"
	"github.com/AplaProject/go-apla/packages/transaction"
	"github.com/AplaProject/go-apla/packages/utils"

	"github.com/AplaProject/go-apla/packages/conf/syspar"
	"github.com/AplaProject/go-apla/packages/service"
	log "github.com/sirupsen/logrus"
)

// Type1 get the list of transactions which belong to the sender from 'disseminator' daemon
// do not load the blocks here because here could be the chain of blocks that are loaded for a long time
// download the transactions here, because they are small and definitely will be downloaded in 60 sec
func Type1(rw io.ReadWriter) error {
	r := &network.DisRequest{}
	if err := r.Read(rw); err != nil {
		return err
	}

	buf := bytes.NewBuffer(r.Data)

	/*
	 *  data structure
	 *  type - 1 byte. 0 - block, 1 - list of transactions
	 *  {if type==1}:
	 *  <any number of the next sets>
	 *   tx_hash - 32 bytes
	 * </>
	 * {if type==0}:
	 *  block_id - 3 bytes
	 *  hash - 32 bytes
	 * <any number of the next sets>
	 *   tx_hash - 32 bytes
	 * </>
	 * */

	// full_node_id of the sender to know where to take a data when it will be downloaded by another daemon
	fullNodeID := converter.BinToDec(buf.Next(8))
	log.Debug("fullNodeID", fullNodeID)

	n := syspar.GetNode(fullNodeID)
	if n != nil && service.GetNodesBanService().IsBanned(*n) {
		return nil
	}

	// get data type (0 - block and transactions, 1 - only transactions)
	newDataType := converter.BinToDec(buf.Next(1))

	log.Debug("newDataType", newDataType)
	if newDataType == 0 {
		err := processBlock(buf, fullNodeID)
		if err != nil {
			log.WithFields(log.Fields{"error": err}).Error("on process block")
			return err
		}
	}

	// get unknown transactions from received packet
	needTx, err := getUnknownTransactions(buf)
	if err != nil {
		log.WithFields(log.Fields{"error": err}).Error("on getting unknown txes")
		return err
	}

	// send the list of transactions which we want to get
	err = (&network.DisHashResponse{Data: needTx}).Write(rw)
	if err != nil {
		log.WithFields(log.Fields{"error": err}).Error("on sending neeeded tx list")
		return err
	}

	if len(needTx) == 0 {
		return nil
	}

	// get this new transactions
	txBodies, err := resieveTxBodies(rw)
	if err != nil {
		log.WithFields(log.Fields{"error": err}).Error("on reading needed txes from disseminator")
		return err
	}

	// and save them
	return saveNewTransactions(txBodies)
}

func resieveTxBodies(con io.Reader) ([]byte, error) {
	sizeBuf := make([]byte, 4)
	if _, err := io.ReadFull(con, sizeBuf); err != nil {
		log.WithFields(log.Fields{"type": consts.IOError, "error": err}).Error("on getting size of tx bodies")
		return nil, err
	}

	size := converter.BinToDec(sizeBuf)
	txBodies := make([]byte, size)
	if _, err := io.ReadFull(con, txBodies); err != nil {
		log.WithFields(log.Fields{"type": consts.IOError, "error": err}).Error("on getting tx bodies")
		return nil, err
	}

	return txBodies, nil
}

func processBlock(buf *bytes.Buffer, fullNodeID int64) error {
	infoBlock := &model.InfoBlock{}
	found, err := infoBlock.Get()
	if err != nil {
		log.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("Getting cur block ID")
		return utils.ErrInfo(err)
	}
	if !found {
		log.WithFields(log.Fields{"type": consts.NotFound}).Error("cant find info block")
		return errors.New("can't find info block")
	}

	// get block ID
	newBlockID := converter.BinToDec(buf.Next(3))
	log.WithFields(log.Fields{"new_block_id": newBlockID}).Debug("Generated new block id")

	// get block hash
	blockHash := buf.Next(consts.HashSize)
	log.Debug("blockHash %x", blockHash)

	qb := &model.QueueBlock{}
	found, err = qb.GetQueueBlockByHash(blockHash)
	if err != nil {
		log.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("Getting QueueBlock")
		return utils.ErrInfo(err)
	}
	// we accept only new blocks
	if !found && newBlockID >= infoBlock.BlockID {
		queueBlock := &model.QueueBlock{Hash: blockHash, FullNodeID: fullNodeID, BlockID: newBlockID}
		err = queueBlock.Create()
		if err != nil {
			log.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("Creating QueueBlock")
			return nil
		}
	}

	return nil
}

func getUnknownTransactions(buf *bytes.Buffer) ([]byte, error) {
	hashes, err := readHashes(buf)
	if err != nil {
		log.WithFields(log.Fields{"type": consts.ProtocolError, "error": err}).Error("on reading hashes")
		return nil, err
	}

	var needTx []byte
	// TODO: remove cycle, select miltiple txes throw in(?)
	for _, hash := range hashes {
		// check if we have such a transaction
		// check log_transaction
		exists, err := model.GetLogTransactionsCount(hash)
		if err != nil {
			log.WithFields(log.Fields{"type": consts.DBError, "error": err, "txHash": hash}).Error("Getting log tx count")
			return nil, utils.ErrInfo(err)
		}
		if exists > 0 {
			log.WithFields(log.Fields{"txHash": hash, "type": consts.DuplicateObject}).Warning("tx with this hash already exists in log_tx")
			continue
		}

		exists, err = model.GetTransactionsCount(hash)
		if err != nil {
			log.WithFields(log.Fields{"type": consts.DBError, "error": err, "txHash": hash}).Error("Getting tx count")
			return nil, utils.ErrInfo(err)
		}
		if exists > 0 {
			log.WithFields(log.Fields{"txHash": hash, "type": consts.DuplicateObject}).Warning("tx with this hash already exists in tx")
			continue
		}

		// check transaction queue
		exists, err = model.GetQueuedTransactionsCount(hash)
		if err != nil {
			log.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("Getting queue_tx count")
			return nil, utils.ErrInfo(err)
		}
		if exists > 0 {
			log.WithFields(log.Fields{"txHash": hash, "type": consts.DuplicateObject}).Warning("tx with this hash already exists in queue_tx")
			continue
		}
		needTx = append(needTx, hash...)
	}

	return needTx, nil
}

func readHashes(buf *bytes.Buffer) ([][]byte, error) {
	if buf.Len()%consts.HashSize != 0 {
		log.WithFields(log.Fields{"hashes_slice_size": buf.Len(), "tx_size": consts.HashSize, "type": consts.ProtocolError}).Error("incorrect hashes length")
		return nil, errors.New("wrong transactions hashes size")
	}

	hashes := make([][]byte, 0, buf.Len()/consts.HashSize)

	for buf.Len() > 0 {
		hashes = append(hashes, buf.Next(consts.HashSize))
	}

	return hashes, nil
}

func saveNewTransactions(binaryTxs []byte) error {
	queue := []model.BatchModel{}
	log.WithFields(log.Fields{"binaryTxs": binaryTxs}).Debug("trying to save binary txs")

	for len(binaryTxs) > 0 {
		txSize, err := converter.DecodeLength(&binaryTxs)
		if err != nil {
			log.WithFields(log.Fields{"type": consts.ProtocolError, "err": err}).Error("decoding binary txs length")
			return err
		}
		if int64(len(binaryTxs)) < txSize {
			log.WithFields(log.Fields{"type": consts.ProtocolError, "size": txSize, "len": len(binaryTxs)}).Error("incorrect binary txs len")
			return utils.ErrInfo(errors.New("bad transactions packet"))
		}

		txBinData := converter.BytesShift(&binaryTxs, txSize)
		if len(txBinData) == 0 {
			log.WithFields(log.Fields{"type": consts.EmptyObject}).Error("binaryTxs is empty")
			return utils.ErrInfo(errors.New("len(txBinData) == 0"))
		}

		if int64(len(txBinData)) > syspar.GetMaxTxSize() {
			log.WithFields(log.Fields{"type": consts.ParameterExceeded, "len": len(txBinData), "size": syspar.GetMaxTxSize()}).Error("len of tx data exceeds max size")
			return utils.ErrInfo("len(txBinData) > max_tx_size")
		}

		tx := transaction.RawTransaction{}
		if err = tx.Unmarshall(bytes.NewBuffer(txBinData)); err != nil {
			log.WithFields(log.Fields{"type": consts.UnmarshallingError, "error": err}).Error("unmarshalling transaction")
			return err
		}

		queue = append(queue, &model.QueueTx{Hash: tx.Hash(), Data: txBinData, FromGate: 1})
	}

	if err := model.BatchInsert(queue, []string{"hash", "data", "from_gate"}); err != nil {
		log.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("error creating QueueTx")
		return err
	}

	return nil
}
