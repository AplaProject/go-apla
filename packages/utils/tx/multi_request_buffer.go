package tx

import (
	"sync"
	"time"

	"github.com/GenesisKernel/go-genesis/packages/utils"
)

type MultiRequest struct {
	ID       string
	Time     time.Time
	Contract string
	Values   []map[string]string
}

type MultiRequestBuffer struct {
	mutex sync.Mutex

	timer         *time.Timer
	requestExpire time.Duration

	requests map[string]*MultiRequest
}

func (mrb *MultiRequestBuffer) NewMultiRequest(contract string) *MultiRequest {
	r := &MultiRequest{
		ID:       utils.UUID(),
		Time:     time.Now(),
		Contract: contract,
	}

	return r
}

func (mrb *MultiRequestBuffer) AddRequest(mr *MultiRequest) {
	mrb.mutex.Lock()
	defer mrb.mutex.Unlock()

	mrb.requests[mr.ID] = mr
	mrb.timer.Reset(mrb.requestExpire)
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
	for t := range mrb.timer.C {
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
		timer:         time.NewTimer(-1),
		requestExpire: requestExpire,
	}

	go mrb.waitForCleaning()

	return mrb
}
