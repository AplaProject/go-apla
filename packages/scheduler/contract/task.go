package contract

import (
	"fmt"
	"time"

	"github.com/AplaProject/go-apla/packages/consts"
	"github.com/AplaProject/go-apla/packages/scheduler"

	"github.com/robfig/cron"
	log "github.com/sirupsen/logrus"
)

type ContractTask struct {
	ID       int64
	CronSpec string
	Contract string

	schedule cron.Schedule
}

func (ct *ContractTask) String() string {
	return fmt.Sprintf("id %d cron %s contract %s", ct.ID, ct.CronSpec, ct.Contract)
}

func (ct *ContractTask) ParseCron() error {
	var err error
	ct.schedule, err = cron.Parse(ct.CronSpec)
	return err
}

func (ct *ContractTask) Equal(t scheduler.Task) bool {
	v, ok := t.(*ContractTask)
	return ok && ct.ID == v.ID
}

func (ct *ContractTask) Update(t scheduler.Task) {
	v, ok := t.(*ContractTask)
	if ok {
		*ct = *v
	}
}

func (ct *ContractTask) Next(tm time.Time) time.Time {
	return ct.schedule.Next(tm)
}

func (ct *ContractTask) Run() {
	_, err := NodeContract(ct.Contract)
	if err != nil {
		log.WithFields(log.Fields{"type": consts.ContractError, "error": err, "task": ct.ID, "cron": ct.CronSpec, "contract": ct.Contract}).Error("run contract task")
		return
	}

	log.WithFields(log.Fields{"task": ct.ID, "cron": ct.CronSpec, "contract": ct.Contract}).Info("run contract task")
}
