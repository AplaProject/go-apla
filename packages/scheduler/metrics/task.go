package metrics

import (
	"github.com/GenesisKernel/go-genesis/packages/consts"
	"github.com/GenesisKernel/go-genesis/packages/model"
	"github.com/GenesisKernel/go-genesis/packages/scheduler"

	log "github.com/sirupsen/logrus"
)

const (
	taskCronSpec = "*/1 * * * *" // every 30 minutes
)

var tasks = []*scheduler.Task{
	{ID: metricEcosystemTables, CronSpec: taskCronSpec, Handler: handlerFunc(HandlerEcosystemTables)},
	{ID: metricEcosystemTx, CronSpec: taskCronSpec, Handler: handlerFunc(HandlerEcosystemTx)},
}

type CollectorFunc func() ([]*model.Metric, error)

type TaskHandler struct {
	collector CollectorFunc
}

func (h *TaskHandler) Run(t *scheduler.Task) {
	metricValues, err := h.collector()
	if err != nil {
		return
	}

	err = model.PutMetrics(metricValues)
	if err != nil {
		log.WithFields(log.Fields{"error": err, "type": consts.DBError}).Error("Save metric values")
	}

	log.WithFields(log.Fields{"task": t.String(), "values": len(metricValues)}).Debug("Add metric values")
}

func handlerFunc(fn CollectorFunc) scheduler.Handler {
	return &TaskHandler{
		collector: fn,
	}
}

func Run() {
	for _, task := range tasks {
		scheduler.AddTask(task)
	}
}
