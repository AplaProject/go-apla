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

package service

import "sync"

const (
	NoPause PauseType = 0

	PauseTypeUpdatingBlockchain PauseType = 1 + iota
	PauseTypeStopingNetwork
)

// np contains the reason why a node should not generating blocks
var np = &NodePaused{PauseType: NoPause}

type PauseType int

type NodePaused struct {
	mutex sync.RWMutex

	PauseType PauseType
}

func (np *NodePaused) Set(pt PauseType) {
	np.mutex.Lock()
	defer np.mutex.Unlock()

	np.PauseType = pt
}

func (np *NodePaused) Unset() {
	np.mutex.Lock()
	defer np.mutex.Unlock()

	np.PauseType = NoPause
}

func (np *NodePaused) Get() PauseType {
	np.mutex.RLock()
	defer np.mutex.RUnlock()

	return np.PauseType
}

func (np *NodePaused) IsSet() bool {
	np.mutex.RLock()
	defer np.mutex.RUnlock()

	return np.PauseType != NoPause
}

func IsNodePaused() bool {
	return np.IsSet()
}

func PauseNodeActivity(pt PauseType) {
	np.Set(pt)
}

func NodePauseType() PauseType {
	return np.Get()
}
