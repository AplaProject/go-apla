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

package utils

import (
	"testing"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
)

func TestBanError(t *testing.T) {
	cases := map[error]bool{
		errors.New("case 1"):                                  false,
		WithBan(errors.New("case 2")):                         true,
		errors.Wrap(errors.New("case 3"), "message"):          false,
		errors.Wrap(WithBan(errors.New("case 4")), "message"): true,
	}

	for err, ok := range cases {
		assert.Equal(t, ok, IsBanError(err), err.Error())
	}
}
