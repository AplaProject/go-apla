package queue

import (
	"path/filepath"

	"github.com/GenesisKernel/go-genesis/packages/blockchain"
	"github.com/GenesisKernel/go-genesis/packages/conf"
	"github.com/GenesisKernel/go-genesis/packages/consts"

	"github.com/beeker1121/goque"
	log "github.com/sirupsen/logrus"
	msgpack "gopkg.in/vmihailenco/msgpack.v2"
)

type TransactionQueue struct {
	queue *goque.Queue
	name  string
}

func (tq *TransactionQueue) Init(name string) error {
	var err error
	tq.name = name
	tq.queue, err = goque.OpenQueue(filepath.Join(conf.Config.QueuesDir, name))
	if err != nil {
		log.WithFields(log.Fields{"error": err, "type": consts.QueueError, "name": name}).Error("opening queue")
		return err
	}
	return nil
}

func (tq *TransactionQueue) Dequeue() (*blockchain.Transaction, bool, error) {
	item, err := tq.queue.Dequeue()
	if err == goque.ErrEmpty {
		return nil, true, nil
	}
	if err != nil {
		log.WithFields(log.Fields{"error": err, "type": consts.QueueError}).Error("dequeuing item from queue")
		return nil, false, err
	}
	tx := &blockchain.Transaction{}
	if err := tx.Unmarshal(item.Value); err != nil {
		return nil, false, err
	}
	return tx, false, nil
}

func (tq *TransactionQueue) Enqueue(tx *blockchain.Transaction) error {
	val, err := tx.Marshal()
	if err != nil {
		return err
	}
	_, err = tq.queue.Enqueue(val)
	if err != nil {
		log.WithFields(log.Fields{"type": consts.QueueError, "error": err}).Error("enqueueing transaction")
		return err
	}
	return nil
}

func (tq *TransactionQueue) ProcessItems(processF func(tx *blockchain.Transaction) error) error {
	for tq.queue.Length() > 0 {
		item, err := tq.queue.Peek()
		if err == goque.ErrEmpty {
			return nil
		}
		if err != nil {
			log.WithFields(log.Fields{"error": err, "type": consts.QueueError}).Error("peeking tx from queue")
			return err
		}
		tx := &blockchain.Transaction{}
		if err := tx.Unmarshal(item.Value); err != nil {
			return err
		}
		err = processF(tx)
		if err != nil {
			return err
		}
		_, err = tq.queue.Dequeue()
		if err == goque.ErrEmpty {
			return nil
		}
		if err != nil {
			log.WithFields(log.Fields{"error": err, "type": consts.QueueError}).Error("dequeueing tx from queue")
			return err
		}
	}
	return nil
}

func (tq *TransactionQueue) peekAll() ([]*blockchain.Transaction, error) {
	txs := []*blockchain.Transaction{}
	var i uint64
	for i = 0; i < tq.queue.Length(); i++ {
		item, err := tq.queue.PeekByOffset(i)
		if err == goque.ErrEmpty {
			break
		}
		if err != nil {
			log.WithFields(log.Fields{"error": err, "type": consts.QueueError}).Error("peeking tx from queue")
			return nil, err
		}
		tx := &blockchain.Transaction{}
		if err := tx.Unmarshal(item.Value); err != nil {
			return nil, err
		}
		txs = append(txs, tx)
	}
	return txs, nil
}

func (tq *TransactionQueue) dequeueAll() error {
	for tq.queue.Length() > 0 {
		_, err := tq.queue.Dequeue()
		if err == goque.ErrEmpty {
			break
		}
		if err != nil {
			log.WithFields(log.Fields{"error": err, "type": consts.QueueError}).Error("dequeueing tx from queue")
			return err
		}
	}
	return nil
}

func (tq *TransactionQueue) ProcessAllItems(processF func(tx []*blockchain.Transaction) error) error {
	txs, err := tq.peekAll()
	if err != nil {
		return err
	}
	if err := processF(txs); err != nil {
		return err
	}
	return tq.dequeueAll()
}

type BlockQueue struct {
	queue *goque.Queue
}

func (bq *BlockQueue) Init(name string) error {
	var err error
	bq.queue, err = goque.OpenQueue(filepath.Join(conf.Config.QueuesDir, name))
	if err != nil {
		log.WithFields(log.Fields{"error": err, "type": consts.QueueError, "name": name}).Error("opening queue")
		return err
	}
	return nil
}

func (bq *BlockQueue) Dequeue() (*blockchain.Block, bool, error) {
	item, err := bq.queue.Dequeue()
	if err == goque.ErrEmpty {
		return nil, true, nil
	}
	if err != nil {
		log.WithFields(log.Fields{"error": err, "type": consts.QueueError}).Error("dequeuing item from queue")
		return nil, false, err
	}
	b := &blockchain.Block{}
	if err := b.Unmarshal(item.Value); err != nil {
		return nil, false, err
	}
	return b, false, nil
}

func (bq *BlockQueue) Enqueue(b *blockchain.Block) error {
	val, err := b.Marshal()
	if err != nil {
		return err
	}
	_, err = bq.queue.Enqueue(val)
	if err != nil {
		log.WithFields(log.Fields{"type": consts.QueueError, "error": err}).Error("enqueueing transaction")
		return err
	}
	return nil
}

func (bq *BlockQueue) ProcessItems(processF func(tx *blockchain.Block) error) error {
	for bq.queue.Length() > 0 {
		item, err := bq.queue.Peek()
		if err == goque.ErrEmpty {
			return nil
		}
		if err != nil {
			log.WithFields(log.Fields{"error": err, "type": consts.QueueError}).Error("peeking tx from queue")
			return err
		}
		b := &blockchain.Block{}
		if err := b.Unmarshal(item.Value); err != nil {
			return err
		}
		err = processF(b)
		if err != nil {
			return err
		}
		_, err = bq.queue.Dequeue()
		if err == goque.ErrEmpty {
			return nil
		}
		if err != nil {
			log.WithFields(log.Fields{"error": err, "type": consts.QueueError}).Error("dequeueing tx from queue")
			return err
		}
	}
	return nil
}

type QueueBlock struct {
	BlockHash  []byte `gorm:"primary_key;not null"`
	BlockID    int64  `gorm:"not null"`
	FullNodeID int64  `gorm:"not null"`
}

func (qb QueueBlock) Marshal() ([]byte, error) {
	return msgpack.Marshal(qb)
}

func (qb *QueueBlock) Unmarshal(b []byte) error {
	return msgpack.Unmarshal(b, qb)
}

type QueueBlockQueue struct {
	queue *goque.Queue
}

func (bq *QueueBlockQueue) Init(name string) error {
	var err error
	bq.queue, err = goque.OpenQueue(filepath.Join(conf.Config.QueuesDir, name))
	if err != nil {
		log.WithFields(log.Fields{"error": err, "type": consts.QueueError, "name": name}).Error("opening queue")
		return err
	}
	return nil
}

func (bq *QueueBlockQueue) Dequeue() (*QueueBlock, bool, error) {
	item, err := bq.queue.Dequeue()
	if err == goque.ErrEmpty {
		return nil, true, nil
	}
	if err != nil {
		log.WithFields(log.Fields{"error": err, "type": consts.QueueError}).Error("dequeuing item from queue")
		return nil, false, err
	}
	b := &QueueBlock{}
	if err := b.Unmarshal(item.Value); err != nil {
		return nil, false, err
	}
	return b, false, nil
}

func (bq *QueueBlockQueue) Enqueue(b *QueueBlock) error {
	val, err := b.Marshal()
	if err != nil {
		return err
	}
	_, err = bq.queue.Enqueue(val)
	if err != nil {
		log.WithFields(log.Fields{"type": consts.QueueError, "error": err}).Error("enqueueing transaction")
		return err
	}
	return nil
}

func (bq *QueueBlockQueue) ProcessItems(processF func(tx *QueueBlock) error) error {
	for bq.queue.Length() > 0 {
		item, err := bq.queue.Peek()
		if err == goque.ErrEmpty {
			return nil
		}
		if err != nil {
			log.WithFields(log.Fields{"error": err, "type": consts.QueueError}).Error("peeking tx from queue")
			return err
		}
		b := &QueueBlock{}
		if err := b.Unmarshal(item.Value); err != nil {
			return err
		}
		err = processF(b)
		if err != nil {
			return err
		}
		_, err = bq.queue.Dequeue()
		if err == goque.ErrEmpty {
			return nil
		}
		if err != nil {
			log.WithFields(log.Fields{"error": err, "type": consts.QueueError}).Error("dequeueing tx from queue")
			return err
		}
	}
	return nil
}

var SendTxQueue *TransactionQueue = &TransactionQueue{}
var ValidateTxQueue *TransactionQueue = &TransactionQueue{}
var SendBlockQueue *BlockQueue = &BlockQueue{}
var ProcessBlockQueue *BlockQueue = &BlockQueue{}
var ValidateBlockQueue *QueueBlockQueue = &QueueBlockQueue{}

func Init() error {
	if err := SendTxQueue.Init("send_tx"); err != nil {
		return err
	}
	if err := ValidateTxQueue.Init("validate_tx"); err != nil {
		return err
	}
	if err := SendBlockQueue.Init("send_block"); err != nil {
		return err
	}
	if err := ValidateBlockQueue.Init("validate_block"); err != nil {
		return err
	}
	if err := ProcessBlockQueue.Init("process_block"); err != nil {
		return err
	}
	return nil
}
