package contract

import (
	"github.com/GenesisCommunity/go-genesis/packages/consts"
	"github.com/GenesisCommunity/go-genesis/packages/scheduler"

	log "github.com/sirupsen/logrus"
)

type ContractHandler struct {
	Contract string
}

func (ch *ContractHandler) Run(t *scheduler.Task) {
	_, err := NodeContract(ch.Contract)
	if err != nil {
		log.WithFields(log.Fields{"type": consts.ContractError, "error": err, "task": t.String(), "contract": ch.Contract}).Error("run contract task")
		return
	}

	log.WithFields(log.Fields{"task": t.String(), "contract": ch.Contract}).Info("run contract task")
}
