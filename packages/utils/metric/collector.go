package metric

type CollectorFunc func() ([]*Value, error)

type Value struct {
	Time   int64
	Metric string
	Key    string
	Value  int64
}

func (v *Value) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"time":   v.Time,
		"metric": v.Metric,
		"key":    v.Key,
		"value":  v.Value,
	}
}

type Collector struct {
	funcs []CollectorFunc
}

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

func NewCollector(funcs ...CollectorFunc) *Collector {
	c := &Collector{}
	c.funcs = make([]CollectorFunc, 0, len(funcs))
	c.funcs = append(c.funcs, funcs...)
	return c
}
