// Copyright 2016 The go-daylight Authors
// This file is part of the go-daylight library.
//
// The go-daylight library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-daylight library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-daylight library. If not, see <http://www.gnu.org/licenses/>.

package api

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAppContent(t *testing.T) {
	assert.NoError(t, keyLogin(1))

	var ret appContentResult
	err := sendGet(`appcontent/1`, nil, &ret)
	if err != nil {
		t.Error(err)
		return
	}

	if len(ret.Blocks) == 0 {
		t.Error("incorrect blocks count")
	}

	if len(ret.Contracts) == 0 {
		t.Error("incorrect contracts count")
	}

	if len(ret.Pages) == 0 {
		t.Error("incorrent pages count")
	}
}
