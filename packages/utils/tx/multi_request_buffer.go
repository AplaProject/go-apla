package tx

import (
	"sync"
	"time"

	"github.com/GenesisKernel/go-genesis/packages/utils"
)

type MultiRequest struct {
	ID        string
	Time      time.Time
	Contracts []MultiRequestContract
}

func (mr *MultiRequest) AddContract(contract string, params map[string]string) {
	mr.Contracts = append(mr.Contracts, MultiRequestContract{
		Contract: contract,
		Params:   params,
	})
}

type MultiRequestContract struct {
	Contract string
	Params   map[string]string
}

type MultiRequestBuffer struct {
	mutex sync.Mutex

	requestExpire time.Duration
	requests      map[string]*MultiRequest
}

func (mrb *MultiRequestBuffer) NewMultiRequest() *MultiRequest {
	r := &MultiRequest{
		ID:        utils.UUID(),
		Time:      time.Now(),
		Contracts: make([]MultiRequestContract, 0),
	}

	return r
}

func (mrb *MultiRequestBuffer) AddRequest(mr *MultiRequest) {
	mrb.mutex.Lock()
	defer mrb.mutex.Unlock()

	mrb.requests[mr.ID] = mr
}

func (mrb *MultiRequestBuffer) GetRequest(id string) (*MultiRequest, bool) {
	mrb.mutex.Lock()
	defer mrb.mutex.Unlock()

	r, ok := mrb.requests[id]
	if !ok {
		return nil, false
	}

	return r, true
}

func (mrb *MultiRequestBuffer) waitForCleaning() {
	ticker := time.NewTicker(mrb.requestExpire)

	for t := range ticker.C {
		mrb.clean(t)
	}
}

func (mrb *MultiRequestBuffer) clean(t time.Time) {
	mrb.mutex.Lock()
	defer mrb.mutex.Unlock()

	for id, r := range mrb.requests {
		if t.Sub(r.Time) > mrb.requestExpire {
			delete(mrb.requests, id)
		}
	}
}

func NewMultiRequestBuffer(requestExpire time.Duration) *MultiRequestBuffer {
	mrb := &MultiRequestBuffer{
		requests:      make(map[string]*MultiRequest),
		requestExpire: requestExpire,
	}

	go mrb.waitForCleaning()

	return mrb
}
