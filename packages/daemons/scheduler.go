package daemons

import (
	"context"
	"fmt"
	"time"

	"github.com/AplaProject/go-apla/packages/consts"
	"github.com/AplaProject/go-apla/packages/model"
	"github.com/AplaProject/go-apla/packages/scheduler"
	"github.com/AplaProject/go-apla/packages/scheduler/contract"

	log "github.com/sirupsen/logrus"
)

func loadContractTasks() error {
	stateIDs, err := model.GetAllSystemStatesIDs()
	if err != nil {
		log.WithFields(log.Fields{"error": err, "type": consts.DBError}).Error("get all system states ids")
		return err
	}

	for _, stateID := range stateIDs {
		if !model.IsTable(fmt.Sprintf("%d_vde_cron", stateID)) {
			return nil
		}

		c := model.Cron{}
		c.SetTablePrefix(fmt.Sprintf("%d_vde", stateID))
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
	}

	return nil
}

// Scheduler starts contracts on schedule
func Scheduler(ctx context.Context, d *daemon) error {
	d.sleepTime = time.Hour
	return loadContractTasks()
}
