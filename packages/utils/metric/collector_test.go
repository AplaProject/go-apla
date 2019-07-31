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
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func MockValue(v int64) *Value {
	return &Value{Time: 1, Metric: "test_metric", Key: "ecosystem_1", Value: v}
}

func MockCollectorFunc(v int64, err error) CollectorFunc {
	return func() ([]*Value, error) {
		if err != nil {
			return nil, err
		}

		return []*Value{MockValue(v)}, nil
	}
}

func TestValue(t *testing.T) {
	value := MockValue(100)
	result := map[string]interface{}{"time": int64(1), "metric": "test_metric", "key": "ecosystem_1", "value": int64(100)}
	assert.Equal(t, result, value.ToMap())
}

func TestCollector(t *testing.T) {
	c := NewCollector(
		MockCollectorFunc(100, nil),
		MockCollectorFunc(0, errors.New("Test")),
		MockCollectorFunc(200, nil),
	)

	result := []interface{}{
		map[string]interface{}{"time": int64(1), "metric": "test_metric", "key": "ecosystem_1", "value": int64(100)},
		map[string]interface{}{"time": int64(1), "metric": "test_metric", "key": "ecosystem_1", "value": int64(200)},
	}
	assert.Equal(t, result, c.Values())
}
