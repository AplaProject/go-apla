// Copyright (C) 2017, 2018, 2019 EGAAS S.A.
//
// This program is free software; you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation; either version 2 of the License, or (at
// your option) any later version.
//
// This program is distributed in the hope that it will be useful, but
// WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the GNU
// General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program; if not, write to the Free Software
// Foundation, Inc., 51 Franklin Street, Fifth Floor, Boston, MA 02110-1301, USA.

package metric

import (
	"github.com/AplaProject/go-apla/packages/types"
)

// CollectorFunc represents function for collects values of metrics
type CollectorFunc func(int64) ([]*Value, error)

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
func (c *Collector) Values(timeBlock int64) []interface{} {
	values := make([]interface{}, 0)
	for _, fn := range c.funcs {
		result, err := fn(timeBlock)
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
