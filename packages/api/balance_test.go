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

	"github.com/stretchr/testify/require"
)

func TestBalance(t *testing.T) {
	err := keyLogin(1)
	require.NoError(t, err)

	var ret balanceResult
	err = sendGet(`balance/`+gAddress, nil, &ret)
	require.NoError(t, err)

	expAmountLen := len(ret.Amount)
	require.Truef(t, expAmountLen >= 10, "length of returning amount %d, should be greater or equal 10", expAmountLen)
	err = sendGet(`balance/3434341`, nil, &ret)
	require.NoError(t, err)

	require.Truef(t, len(ret.Amount) > 0, `wrong balance %s`, ret.Amount)
}
