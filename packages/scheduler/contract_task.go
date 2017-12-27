package scheduler

import (
	"fmt"
	"time"

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

func (ct *ContractTask) Equal(t Task) bool {
	v, ok := t.(*ContractTask)
	return ok && ct.ID == v.ID
}

func (ct *ContractTask) Update(t Task) {
	v, ok := t.(*ContractTask)
	if ok {
		*ct = *v
	}
}

func (ct *ContractTask) Next(tm time.Time) time.Time {
	return ct.schedule.Next(tm)
}

func (ct *ContractTask) Run() {
	log.WithFields(log.Fields{"task": ct.ID, "cron": ct.CronSpec, "contract": ct.Contract}).Info("run contract task")
}
