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

// type Task interface {
// 	String() string

// 	Equal(Task) bool
// 	ParseCron() error
// 	Update(Task)

// 	Next(time.Time) time.Time
// 	Run()
// }

// type BaseTask struct {
// 	ID       int64
// 	CronSpec string

// 	schedule cron.Schedule
// }

// func (bt *BaseTask) ParseCron() error {
// 	if len(bt.CronSpec) == 0 {
// 		return nil
// 	}

// 	var err error
// 	bt.schedule, err = cron.ParseStandard(bt.CronSpec)
// 	return err
// }

// func (bt *BaseTask) Equal(t Task) bool {
// 	v, ok := t.(*BaseTask)
// 	return ok && bt.ID == v.ID
// }

// func (bt *BaseTask) Update(t Task) {
// 	v, ok := t.(*BaseTask)
// 	if ok {
// 		*bt = *v
// 	}
// }

// func (bt *BaseTask) Next(tm time.Time) time.Time {
// 	if len(bt.CronSpec) == 0 {
// 		var zeroTime time.Time
// 		return zeroTime
// 	}
// 	return bt.schedule.Next(tm)
// }

// func (bt *BaseTask) Run() {

// }

// func (bt *BaseTask) String() string {
// 	return fmt.Sprintf("id %d cron %s", bt.ID, bt.CronSpec)
// }
