package daemons

import (
	"context"
	"time"

	"github.com/GenesisKernel/go-genesis/packages/consts"
	"github.com/GenesisKernel/go-genesis/packages/converter"
	"github.com/GenesisKernel/go-genesis/packages/model"
	"github.com/GenesisKernel/go-genesis/packages/scheduler"
	"github.com/GenesisKernel/go-genesis/packages/scheduler/contract"

	log "github.com/sirupsen/logrus"
)

func loadContractTasks() error {
	c := model.Cron{}
	c.SetTablePrefix(converter.IntToStr(consts.DefaultVDE))

	if !model.IsTable(c.TableName()) {
		return nil
	}

	tasks, err := c.GetAllCronTasks()
	if err != nil {
		log.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("get all cron tasks")
		return err
	}

	for _, cronTask := range tasks {
		err = scheduler.UpdateTask(&scheduler.Task{
			ID:       cronTask.UID(),
			CronSpec: cronTask.Cron,
			Handler: &contract.ContractHandler{
				Contract: cronTask.Contract,
			},
		})
		if err != nil {
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
