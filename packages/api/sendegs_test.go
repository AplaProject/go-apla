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
	"fmt"
	"net/url"
	"testing"
)

func TestSendEGS(t *testing.T) {
	if err := keyLogin(0); err != nil {
		t.Error(err)
		return
	}
	inBefore, err := getBalance(`0080-2194-0302-1823-2313`)
	if err != nil {
		t.Error(err)
		return
	}
	outBefore, err := getBalance(gAddress)
	if err != nil {
		t.Error(err)
		return
	}
	fmt.Println("Before", inBefore, outBefore)
	form := url.Values{"recipient": {`0080-2194-0302-1823-2313`}, "amount": {`1234567890000000000`},
		"commission": {`2000000000000000`}, `comment`: {"Test"}, `pubkey`: {gPublic}}
	if err := postTx(`sendegs`, &form); err != nil {
		t.Error(err)
		return
	}
	inAfter, err := getBalance(`0080-2194-0302-1823-2313`)
	if err != nil {
		t.Error(err)
		return
	}
	outAfter, err := getBalance(gAddress)
	if err != nil {
		t.Error(err)
		return
	}
	if inAfter.Sub(inBefore).String() != `1234567890000000000` {
		t.Error(fmt.Errorf(`IN %v != %s`, inAfter.Sub(inBefore), `1234567890000000000`))
		return
	}
	if outBefore.Sub(outAfter).String() != `1234597890000000000` {
		t.Error(fmt.Errorf(`OUT %v != %s`, outBefore.Sub(outAfter), `1234597890000000000`))
		return
	}
}
