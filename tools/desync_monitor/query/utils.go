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
	"encoding/json"
	"fmt"
	"net/http"
	"sync"

	"io/ioutil"

	log "github.com/sirupsen/logrus"
)

type ConcurrentMap struct {
	m  map[string]interface{}
	mu sync.RWMutex
}

func (c *ConcurrentMap) Set(key string, value interface{}) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.m[key] = value
}

func (c ConcurrentMap) Get(key string) (bool, interface{}) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	res, ok := c.m[key]
	return ok, res
}

func sendGetRequest(url string, v interface{}) error {
	resp, err := http.Get(url)
	if err != nil {
		log.WithFields(log.Fields{"url": url, "error": err}).Error("get requesting url")
		return err
	}
	if resp.StatusCode != http.StatusOK {
		err := fmt.Errorf("status code is not OK %d", resp.StatusCode)
		log.WithFields(log.Fields{"url": url, "error": err}).Error("incorrect status code")
		return err
	}
	defer resp.Body.Close()
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.WithFields(log.Fields{"url": url, "error": err}).Error("reading response body")
		return err
	}
	if err := json.Unmarshal(data, v); err != nil {
		log.WithFields(log.Fields{"data": string(data), "error": err}).Error("unmarshalling json to struct")
		return err
	}
	return nil
}
