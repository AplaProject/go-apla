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

package transaction

import "sync"

type transactionCache struct {
	mutex sync.RWMutex
	cache map[string]*Transaction
}

func (tc *transactionCache) Get(hash string) (t *Transaction, ok bool) {
	tc.mutex.RLock()
	defer tc.mutex.RUnlock()

	t, ok = tc.cache[hash]
	return
}

func (tc *transactionCache) Set(t *Transaction) {
	tc.mutex.Lock()
	defer tc.mutex.Unlock()

	tc.cache[string(t.TxHash)] = t
}

func (tc *transactionCache) Clean() {
	tc.mutex.Lock()
	defer tc.mutex.Unlock()

	tc.cache = make(map[string]*Transaction)
}
