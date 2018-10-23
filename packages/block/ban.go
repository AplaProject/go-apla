// Copyright 2018 The go-daylight Authors
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

package block

import (
	"sync"
	"time"
)

const (
	badTime  = 5  // time period
	maxBadTx = 3  // maximum bad tx during badTime minutes
	banTime  = 15 // ban time in minutes
)

type banKey struct {
	Time time.Time           // banned till
	Bad  [maxBadTx]time.Time // time of bad tx
}

var (
	banList = make(map[int64]banKey)
	mutex   = &sync.RWMutex{}
)

// IsBanned returns true if the key has been banned
func IsBanned(keyID int64) bool {
	mutex.RLock()
	if ban, ok := banList[keyID]; ok {
		mutex.RUnlock()
		now := time.Now()
		if now.Before(ban.Time) {
			return true
		}
		for i := 0; i < maxBadTx; i++ {
			if ban.Bad[i].Add(badTime * time.Minute).After(now) {
				return false
			}
		}
		// Delete if time of all bad tx is old
		mutex.Lock()
		delete(banList, keyID)
		mutex.Unlock()
	} else {
		mutex.RUnlock()
	}
	return false
}

// BannedTill returns the time that the user has been banned till
func BannedTill(keyID int64) string {
	mutex.RLock()
	defer mutex.RUnlock()
	if ban, ok := banList[keyID]; ok {
		return ban.Time.Format(`2006-01-02 15:04:05`)
	}
	return ``
}

// BadTxForBan adds info about bad tx of the key
func BadTxForBan(keyID int64) {
	var (
		ban banKey
		ok  bool
	)
	mutex.Lock()
	defer mutex.Unlock()
	now := time.Now()
	if ban, ok = banList[keyID]; ok {
		var bMin, count int
		for i := 0; i < maxBadTx; i++ {
			if ban.Bad[i].Add(badTime * time.Minute).After(now) {
				count++
			}
			if i > bMin && ban.Bad[i].Before(ban.Bad[bMin]) {
				bMin = i
			}
		}
		ban.Bad[bMin] = now
		if count >= maxBadTx-1 {
			ban.Time = now.Add(banTime * time.Minute)
		}
	} else {
		ban = banKey{}
		ban.Bad[0] = time.Now()
	}
	banList[keyID] = ban
}
