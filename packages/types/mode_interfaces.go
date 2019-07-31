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

package types

import (
	"context"

	log "github.com/sirupsen/logrus"
)

// ClientTxPreprocessor procees tx from client
type ClientTxPreprocessor interface {
	ProcessClientTranstaction([]byte, int64, *log.Entry) (string, error)
}

// SmartContractRunner run serialized contract
type SmartContractRunner interface {
	RunContract(data, hash []byte, keyID int64, le *log.Entry) error
}

type DaemonListFactory interface {
	GetDaemonsList() []string
}

type EcosystemLookupGetter interface {
	GetEcosystemLookup() ([]int64, []string, error)
}

type EcosystemIDValidator interface {
	Validate(id, clientID int64, le *log.Entry) (int64, error)
}

// DaemonLoader allow implement different ways for loading daemons
type DaemonLoader interface {
	Load(context.Context) error
}

type EcosystemNameGetter interface {
	GetEcosystemName(id int64) (string, error)
}
