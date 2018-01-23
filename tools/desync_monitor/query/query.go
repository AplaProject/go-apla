package query

import (
	"fmt"
	"sync"

	"github.com/AplaProject/go-apla/packages/api"
)

const maxBlockIDEndpoint = "/api/v2/maxblockid"
const blockInfoEndpoint = "/api/v2/block/%d"

type MaxBlockID struct {
	MaxBlockID int64 `json:"max_block_id"`
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

func BlockInfo(nodesList []string, blockID int64) (map[string]*api.GetBlockInfoResult, error) {
	wg := sync.WaitGroup{}
	workResults := ConcurrentMap{m: map[string]interface{}{}}
	for _, nodeUrl := range nodesList {
		wg.Add(1)
		go func(url string) {
			defer wg.Done()
			blockInfo := &api.GetBlockInfoResult{}
			if err := sendGetRequest(url+fmt.Sprintf(blockInfoEndpoint, blockID), blockInfo); err != nil {
				workResults.Set(url, err)
				return
			}
			workResults.Set(url, blockInfo)
		}(nodeUrl)
	}
	wg.Wait()
	result := map[string]*api.GetBlockInfoResult{}
	for nodeUrl, blockInfoOrError := range workResults.m {
		switch res := blockInfoOrError.(type) {
		case error:
			return nil, res
		case *api.GetBlockInfoResult:
			result[nodeUrl] = res
		}
	}
	return result, nil
}
