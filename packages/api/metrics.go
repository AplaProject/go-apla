package api

import (
	"net/http"
	"runtime"

	"github.com/AplaProject/go-apla/packages/conf/syspar"
	"github.com/AplaProject/go-apla/packages/consts"
	"github.com/AplaProject/go-apla/packages/model"
	"github.com/AplaProject/go-apla/packages/service"

	log "github.com/sirupsen/logrus"
)

type blockMetric struct {
	Count int64 `json:"count"`
}

type txMetric struct {
	Count int64 `json:"count"`
}

type ecosysMetric struct {
	Count int64 `json:"count"`
}

type keyMetric struct {
	Count int64 `json:"count"`
}

type fullNodeMetric struct {
	Count int64 `json:"count"`
}

type memMetric struct {
	Alloc uint64 `json:"alloc"`
	Sys   uint64 `json:"sys"`
}

type banMetric struct {
	NodePosition int  `json:"node_position"`
	Status       bool `json:"status"`
}

func blocksCountHandler(w http.ResponseWriter, r *http.Request) {
	b := &model.Block{}
	logger := getLogger(r)

	found, err := b.GetMaxBlock()
	if err != nil {
		logger.WithFields(log.Fields{"error": err, "type": consts.DBError}).Error("on getting max block")
		errorResponse(w, err, http.StatusInternalServerError)
		return
	}

	if !found {
		errorResponse(w, errNotFound)
		return
	}

	bm := blockMetric{Count: b.ID}
	jsonResponse(w, bm)
}

func txCountHandler(w http.ResponseWriter, r *http.Request) {
	c, err := model.GetTxCount()
	if err != nil {
		logger := getLogger(r)
		logger.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("on getting tx count")
		errorResponse(w, err, http.StatusInternalServerError)
		return
	}

	jsonResponse(w, txMetric{Count: c})
}

func (m Mode) ecosysCountHandler(w http.ResponseWriter, r *http.Request) {
	ids, _, err := m.EcosysLookupGetter.GetEcosystemLookup()
	if err != nil {
		logger := getLogger(r)
		logger.WithFields(log.Fields{"error": err}).Error("on getting ecosystem count")
		errorResponse(w, err, http.StatusInternalServerError)
		return
	}

	jsonResponse(w, ecosysMetric{Count: int64(len(ids))})
}

func keysCountHandler(w http.ResponseWriter, r *http.Request) {
	cnt, err := model.GetKeysCount()
	if err != nil {
		logger := getLogger(r)
		logger.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("on getting keys count")
		errorResponse(w, err, http.StatusInternalServerError)
		return
	}

	jsonResponse(w, keyMetric{Count: cnt})
}

func fullNodesCountHandler(w http.ResponseWriter, _ *http.Request) {
	fnMetric := fullNodeMetric{
		Count: syspar.GetNumberOfNodesFromDB(nil),
	}

	jsonResponse(w, fnMetric)
}

func memStatHandler(w http.ResponseWriter, _ *http.Request) {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	jsonResponse(w, memMetric{Alloc: m.Alloc, Sys: m.Sys})
}

func banStatHandler(w http.ResponseWriter, _ *http.Request) {
	nodes := syspar.GetNodes()
	list := make([]banMetric, 0, len(nodes))

	b := service.GetNodesBanService()
	for i, n := range nodes {
		list = append(list, banMetric{
			NodePosition: i,
			Status:       b.IsBanned(n),
		})
	}

	jsonResponse(w, list)
}
