package queue

import (
	"github.com/beeker1121/goque"

	"github.com/GenesisKernel/go-genesis/packages/blockchain"
	"github.com/GenesisKernel/go-genesis/packages/consts"

	log "github.com/sirupsen/logrus"
	msgpack "gopkg.in/vmihailenco/msgpack.v2"
)

var queue *goque.PrefixQueue

type TransactionQueue struct {
	prefix []byte
}

func (tq *TransactionQueue) Dequeue() (*blockchain.Transaction, bool, error) {
	item, err := queue.Dequeue(tq.prefix)
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
	_, err = queue.Enqueue(tq.prefix, val)
	if err != nil {
		log.WithFields(log.Fields{"type": consts.QueueError, "error": err}).Error("enqueueing transaction")
		return err
	}
	return nil
}

func (tq *TransactionQueue) ProcessItems(processF func(tx *blockchain.Transaction) error) error {
	if queue.Length() > 0 {
		item, err := queue.Peek(tq.prefix)
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
		_, err = queue.Dequeue(tq.prefix)
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

type BlockQueue struct {
	prefix []byte
}

func (bq *BlockQueue) Dequeue() (*blockchain.Block, bool, error) {
	item, err := queue.Dequeue(bq.prefix)
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
	_, err = queue.Enqueue(bq.prefix, val)
	if err != nil {
		log.WithFields(log.Fields{"type": consts.QueueError, "error": err}).Error("enqueueing transaction")
		return err
	}
	return nil
}

func (bq *BlockQueue) ProcessItems(processF func(tx *blockchain.Block) error) error {
	if queue.Length() > 0 {
		item, err := queue.Peek(bq.prefix)
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
		_, err = queue.Dequeue(bq.prefix)
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
	prefix []byte
}

func (bq *QueueBlockQueue) Dequeue() (*QueueBlock, bool, error) {
	item, err := queue.Dequeue(bq.prefix)
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
	_, err = queue.Enqueue(bq.prefix, val)
	if err != nil {
		log.WithFields(log.Fields{"type": consts.QueueError, "error": err}).Error("enqueueing transaction")
		return err
	}
	return nil
}

func (bq *QueueBlockQueue) ProcessItems(processF func(tx *QueueBlock) error) error {
	if queue.Length() > 0 {
		item, err := queue.Peek(bq.prefix)
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
		_, err = queue.Dequeue(bq.prefix)
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

var SendTxQueue *TransactionQueue = &TransactionQueue{[]byte("send_tx-")}
var ValidateTxQueue *TransactionQueue = &TransactionQueue{[]byte("validate_tx-")}
var ProcessTxQueue *TransactionQueue = &TransactionQueue{[]byte("process_tx-")}
var SendBlockQueue *BlockQueue = &BlockQueue{[]byte("send_block-")}
var ProcessBlockQueue *BlockQueue = &BlockQueue{[]byte("process_block-")}
var ValidateBlockQueue *QueueBlockQueue = &QueueBlockQueue{[]byte("validate_block-")}

func Init() error {
	var err error
	queue, err = goque.OpenPrefixQueue("queues/queue")
	if err != nil {
		log.WithFields(log.Fields{"error": err, "type": consts.QueueError}).Error("opening sendTxQueue")
		return err
	}
	return nil
}
