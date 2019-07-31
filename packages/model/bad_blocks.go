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

package model

import (
	"time"
)

type BadBlocks struct {
	ID             int64
	ProducerNodeId int64
	BlockId        int64
	ConsumerNodeId int64
	BlockTime      time.Time
	Deleted        bool
}

// TableName returns name of table
func (r BadBlocks) TableName() string {
	return "1_bad_blocks"
}

// BanRequests represents count of unique ban requests for node
type BanRequests struct {
	ProducerNodeId int64
	Count          int64
}

// GetNeedToBanNodes is returns list of ban requests for each node
func (r *BadBlocks) GetNeedToBanNodes(now time.Time, blocksPerNode int) ([]BanRequests, error) {
	var res []BanRequests

	err := DBConn.
		Raw(
			`SELECT
				producer_node_id,
				COUNT(consumer_node_id) as count
			FROM (
				SELECT
					producer_node_id,
					consumer_node_id,
					count(DISTINCT block_id)
				FROM
				"1_bad_blocks"
				WHERE
					block_time > ?::date - interval '24 hours'
					AND deleted = 0
				GROUP BY
					producer_node_id,
					consumer_node_id
				HAVING
					count(DISTINCT block_id) >= ?) AS tbl
			GROUP BY
			producer_node_id`,
			now,
			blocksPerNode,
		).
		Scan(&res).
		Error

	return res, err
}

func (r *BadBlocks) GetNodeBlocks(nodeId int64, now time.Time) ([]BadBlocks, error) {
	var res []BadBlocks
	err := DBConn.
		Table(r.TableName()).
		Model(&BadBlocks{}).
		Where(
			"producer_node_id = ? AND block_time > ?::date - interval '24 hours' AND deleted = ?",
			nodeId,
			now,
			false,
		).
		Scan(&res).
		Error

	return res, err
}
