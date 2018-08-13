package queue

import (
	"github.com/beeker1121/goque"

	"github.com/GenesisKernel/go-genesis/packages/consts"

	log "github.com/sirupsen/logrus"
)

var SendTxQueue *goque.Queue
var SendBlockQueue *goque.Queue
var ValidateTxQueue *goque.Queue
var ValidateBlockQueue *goque.Queue

func Init() error {
	var err error
	SendTxQueue, err = goque.OpenQueue("queues/sendTxQueue")
	if err != nil {
		log.WithFields(log.Fields{"error": err, "type": consts.QueueError}).Error("opening sendTxQueue")
		return err
	}
	ValidateTxQueue, err = goque.OpenQueue("queues/validateTxQueue")
	if err != nil {
		log.WithFields(log.Fields{"error": err, "type": consts.QueueError}).Error("opening validateTxQueue")
		return err
	}
	SendBlockQueue, err = goque.OpenQueue("queues/sendBlockQueue")
	if err != nil {
		log.WithFields(log.Fields{"error": err, "type": consts.QueueError}).Error("opening sendBlockQueue")
		return err
	}
	ValidateBlockQueue, err = goque.OpenQueue("queues/validateBlockQueue")
	if err != nil {
		log.WithFields(log.Fields{"error": err, "type": consts.QueueError}).Error("opening validateBlockQueue")
		return err
	}
	return nil
}
