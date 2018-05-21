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

package api

import (
	"net/http"

	"github.com/GenesisKernel/go-genesis/packages/model"
	//	"github.com/GenesisKernel/go-genesis/packages/converter"

	log "github.com/sirupsen/logrus"
)

type checkResult struct {
	Rollback   int64 `json:"rollback,omitempty"`
	Ecosystems int64 `json:"ecosystems,omitempty"`
	Blockchain int64 `json:"blockchain,omitempty"`
}

func check(w http.ResponseWriter, r *http.Request, data *apiData, logger *log.Entry) (err error) {
	var result checkResult
	var rowsCount int64
	if err := model.DBConn.Table("rollback_tx").Count(&rowsCount).Error; err != nil {
		return err
	}
	result.Rollback = rowsCount
	if err := model.DBConn.Table("1_ecosystems").Count(&rowsCount).Error; err != nil {
		return err
	}
	result.Ecosystems = rowsCount
	if err := model.DBConn.Table("block_chain").Count(&rowsCount).Error; err != nil {
		return err
	}
	result.Blockchain = rowsCount
	data.result = &result

	return
}
