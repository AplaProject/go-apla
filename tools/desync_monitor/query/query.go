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

package query

import (
	"fmt"
	"sync"
)

const maxBlockIDEndpoint = "/api/v2/maxblockid"
const blockInfoEndpoint = "/api/v2/block/%d"

type MaxBlockID struct {
	MaxBlockID int64 `json:"max_block_id"`
}

type blockInfoResult struct {
	Hash          []byte `json:"hash"`
	EcosystemID   int64  `json:"ecosystem_id"`
	KeyID         int64  `json:"key_id"`
	Time          int64  `json:"time"`
	Tx            int32  `json:"tx_count"`
	RollbacksHash []byte `json:"rollbacks_hash"`
}

func MaxBlockIDs(nodesList []string) ([]int64, error) {
	wg := sync.WaitGroup{}
	workResults := ConcurrentMap{m: map[string]interface{}{}}
	for _, nodeUrl := range nodesList {
		wg.Add(1)
		go func(url string) {
			defer wg.Done()
			maxBlockID := &MaxBlockID{}
			if err := sendGetRequest(url+maxBlockIDEndpoint, maxBlockID); err != nil {
				workResults.Set(url, err)
				return
			}
			workResults.Set(url, maxBlockID.MaxBlockID)
		}(nodeUrl)
	}
	wg.Wait()
	maxBlockIds := []int64{}
	for _, result := range workResults.m {
		switch res := result.(type) {
		case int64:
			maxBlockIds = append(maxBlockIds, res)
		case error:
			return nil, res
		}
	}
	return maxBlockIds, nil
}

func BlockInfo(nodesList []string, blockID int64) (map[string]*blockInfoResult, error) {
	wg := sync.WaitGroup{}
	workResults := ConcurrentMap{m: map[string]interface{}{}}
	for _, nodeUrl := range nodesList {
		wg.Add(1)
		go func(url string) {
			defer wg.Done()
			blockInfo := &blockInfoResult{}
			if err := sendGetRequest(url+fmt.Sprintf(blockInfoEndpoint, blockID), blockInfo); err != nil {
				workResults.Set(url, err)
				return
			}
			workResults.Set(url, blockInfo)
		}(nodeUrl)
	}
	wg.Wait()
	result := map[string]*blockInfoResult{}
	for nodeUrl, blockInfoOrError := range workResults.m {
		switch res := blockInfoOrError.(type) {
		case error:
			return nil, res
		case *blockInfoResult:
			result[nodeUrl] = res
		}
	}
	return result, nil
}
