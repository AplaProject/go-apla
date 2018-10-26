package metric

import (
	"github.com/GenesisKernel/go-genesis/packages/types"
)

// CollectorFunc represents function for collects values of metrics
type CollectorFunc func() ([]*Value, error)

// Value represents value of metrics
type Value struct {
	Time   int64
	Metric string
	Key    string
	Value  int64
}

// ToMap returns values as map
func (v *Value) ToMap() *types.Map {
	return types.LoadMap(map[string]interface{}{
		"time":   v.Time,
		"metric": v.Metric,
		"key":    v.Key,
		"value":  v.Value,
	})
}

// Collector represents struct that works with the collection of metrics
type Collector struct {
	funcs []CollectorFunc
}

// Values returns values of all metrics
func (c *Collector) Values() []interface{} {
	values := make([]interface{}, 0)
	for _, fn := range c.funcs {
		result, err := fn()
		if err != nil {
			continue
		}

		for _, v := range result {
			values = append(values, v.ToMap())
		}
	}
	return values
}

// NewCollector creates new collector
func NewCollector(funcs ...CollectorFunc) *Collector {
	c := &Collector{}
	c.funcs = make([]CollectorFunc, 0, len(funcs))
	c.funcs = append(c.funcs, funcs...)
	return c
}
