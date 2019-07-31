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

package block

import (
	"sync"
	"time"

	"github.com/AplaProject/go-apla/packages/conf"
)

type banKey struct {
	Time time.Time   // banned till
	Bad  []time.Time // time of bad tx
}

var (
	banList = make(map[int64]banKey)
	mutex   = &sync.RWMutex{}
)

// IsBanned returns true if the key has been banned
func IsKeyBanned(keyID int64) bool {
	mutex.RLock()
	if ban, ok := banList[keyID]; ok {
		mutex.RUnlock()
		now := time.Now()
		if now.Before(ban.Time) {
			return true
		}
		for i := 0; i < conf.Config.BanKey.BadTx; i++ {
			if ban.Bad[i].Add(time.Duration(conf.Config.BanKey.BadTime) * time.Minute).After(now) {
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
		for i := 0; i < conf.Config.BanKey.BadTx; i++ {
			if ban.Bad[i].Add(time.Duration(conf.Config.BanKey.BadTime) * time.Minute).After(now) {
				count++
			}
			if i > bMin && ban.Bad[i].Before(ban.Bad[bMin]) {
				bMin = i
			}
		}
		ban.Bad[bMin] = now
		if count >= conf.Config.BanKey.BadTx-1 {
			ban.Time = now.Add(time.Duration(conf.Config.BanKey.BanTime) * time.Minute)
		}
	} else {
		ban = banKey{Bad: make([]time.Time, conf.Config.BanKey.BadTx)}
		ban.Bad[0] = time.Now()
	}
	banList[keyID] = ban
}
