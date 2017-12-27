package daemons

import (
	"context"
	"time"

	"github.com/AplaProject/go-apla/packages/consts"
	"github.com/AplaProject/go-apla/packages/model"
	"github.com/AplaProject/go-apla/packages/scheduler"

	log "github.com/sirupsen/logrus"
)

func loadContractTasks() error {
	if !model.IsTable("1_vde_tables") {
		return nil
	}

	c := model.Cron{}
	c.SetTablePrefix("1_vde")
	tasks, err := c.GetAllCronTasks()
	if err != nil {
		log.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("get all cron tasks")
		return err
	}

	for _, task := range tasks {
		err = scheduler.UpdateTask(&scheduler.ContractTask{
			ID:       task.ID,
			CronSpec: task.Cron,
			Contract: task.Contract,
		})
		if err != nil {
			log.WithFields(log.Fields{"type": consts.SchedulerError, "error": err}).Error("update cron task")
			return err
		}
	}

	return nil
}

// Scheduler starts contracts on schedule
func Scheduler(ctx context.Context, d *daemon) error {
	d.sleepTime = time.Hour
	return loadContractTasks()
}
