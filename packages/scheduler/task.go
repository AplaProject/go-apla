package scheduler

import (
	"fmt"
	"time"

	"github.com/robfig/cron"
)

var zeroTime time.Time

type Handler interface {
	Run(*Task)
}

type Task struct {
	ID       int64
	CronSpec string

	Handler Handler

	schedule cron.Schedule
}

func (t *Task) String() string {
	return fmt.Sprintf("id %d cron %s", t.ID, t.CronSpec)
}

func (t *Task) ParseCron() error {
	if len(t.CronSpec) == 0 {
		return nil
	}

	var err error
	t.schedule, err = cron.Parse(t.CronSpec)
	return err
}

func (t *Task) Next(tm time.Time) time.Time {
	if len(t.CronSpec) == 0 {
		return zeroTime
	}
	return t.schedule.Next(tm)
}

func (t *Task) Run() {
	t.Handler.Run(t)
}
