//MIT License
//
//Copyright (c) 2016 GenesisKernel
//
//Permission is hereby granted, free of charge, to any person obtaining a copy
//of this software and associated documentation files (the "Software"), to deal
//in the Software without restriction, including without limitation the rights
//to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
//copies of the Software, and to permit persons to whom the Software is
//furnished to do so, subject to the following conditions:
//
//The above copyright notice and this permission notice shall be included in all
//copies or substantial portions of the Software.
//
//THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
//IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
//FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
//AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
//LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
//OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
//SOFTWARE.
package api

import (
	"fmt"
	"net/http"

	"github.com/GenesisCommunity/go-genesis/packages/consts"
	"github.com/GenesisCommunity/go-genesis/packages/converter"
	"github.com/GenesisCommunity/go-genesis/packages/model"

	log "github.com/sirupsen/logrus"
)

type tableInfo struct {
	Name  string `json:"name"`
	Count string `json:"count"`
}

type tablesResult struct {
	Count int64       `json:"count"`
	List  []tableInfo `json:"list"`
}

func tables(w http.ResponseWriter, r *http.Request, data *apiData, logger *log.Entry) (err error) {
	var (
		result tablesResult
		limit  int
	)

	table := getPrefix(data) + `_tables`

	count, err := model.GetRecordsCountTx(nil, table)
	if err != nil {
		logger.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("selecting records count from tables")
		return errorAPI(w, err.Error(), http.StatusInternalServerError)
	}
	if data.params[`limit`].(int64) > 0 {
		limit = int(data.params[`limit`].(int64))
	} else {
		limit = 25
	}
	list, err := model.GetAll(`select name from "`+table+`" order by name`+
		fmt.Sprintf(` offset %d `, data.params[`offset`].(int64)), limit)
	if err != nil {
		logger.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("selecting names from tables")
		return errorAPI(w, err.Error(), http.StatusInternalServerError)
	}

	result = tablesResult{
		Count: count, List: make([]tableInfo, len(list)),
	}
	for i, item := range list {
		var maxid int64
		result.List[i].Name = item[`name`]
		fullname := getPrefix(data) + `_` + item[`name`]
		if item[`name`] == `keys` || item[`name`] == `member` {
			err = model.DBConn.Table(fullname).Count(&maxid).Error
			if err != nil {
				logger.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("selecting count from table")
			}
		} else {
			maxid, err = model.GetNextID(nil, fullname)
			if err != nil {
				logger.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("getting next id from table")
			}
			maxid--
		}
		if err != nil {
			return errorAPI(w, err.Error(), http.StatusInternalServerError)
		}
		result.List[i].Count = converter.Int64ToStr(maxid)
	}
	data.result = &result
	return
}
