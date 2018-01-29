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

package parser

// Limits is used for saving current limit information
type Limits struct {
	GenMode  bool                    // equals true when the node is generating a block
	Start    int64                   // the time of the start of generating block
	TxUsers  map[int64]int           // the counter of tx from one user
	Count    int64                   // the count of proceed transactions
	TxEcosys map[int64]map[int64]int // the counter of tx from one user for the each ecosystem
}

type 

// newLimits initializes Limits structure.
func (b *Block) newLimits() (limits *Limits) {
	limits = &Limits{GenMode: b.GenBlock}
	return
}

func (limits *Limits) preProcess(p *Parser) error {
	return nil
}

func (limits *Limits) postProcess(p *Parser) error {
	return nil
}
