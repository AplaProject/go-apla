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
	"encoding/json"
	"net/http"

	"github.com/GenesisCommunity/go-genesis/packages/consts"
	"github.com/GenesisCommunity/go-genesis/packages/model"

	log "github.com/sirupsen/logrus"
)

type columnInfo struct {
	Name string `json:"name"`
	Type string `json:"type"`
	Perm string `json:"perm"`
}

type tableResult struct {
	Name       string       `json:"name"`
	Insert     string       `json:"insert"`
	NewColumn  string       `json:"new_column"`
	Update     string       `json:"update"`
	Read       string       `json:"read,omitempty"`
	Filter     string       `json:"filter,omitempty"`
	Conditions string       `json:"conditions"`
	Columns    []columnInfo `json:"columns"`
}

func table(w http.ResponseWriter, r *http.Request, data *apiData, logger *log.Entry) (err error) {
	var result tableResult

	prefix := getPrefix(data)
	table := &model.Table{}
	table.SetTablePrefix(prefix)
	_, err = table.Get(nil, data.params[`name`].(string))
	if err != nil {
		logger.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("Getting table")
		return errorAPI(w, err.Error(), http.StatusInternalServerError)
	}

	if len(table.Name) > 0 {
		var perm map[string]string
		err := json.Unmarshal([]byte(table.Permissions), &perm)
		if err != nil {
			logger.WithFields(log.Fields{"type": consts.JSONUnmarshallError, "error": err}).Error("Unmarshalling table permissions to json")
			return errorAPI(w, err.Error(), http.StatusInternalServerError)
		}
		var cols map[string]string
		err = json.Unmarshal([]byte(table.Columns), &cols)
		if err != nil {
			logger.WithFields(log.Fields{"type": consts.JSONUnmarshallError, "error": err}).Error("Unmarshalling table columns to json")
			return errorAPI(w, err.Error(), http.StatusInternalServerError)
		}
		columns := make([]columnInfo, 0)
		for key, value := range cols {
			colType, err := model.GetColumnType(prefix+`_`+data.params[`name`].(string), key)
			if err != nil {
				logger.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("getting column type from db")
				return errorAPI(w, err.Error(), http.StatusInternalServerError)
			}
			columns = append(columns, columnInfo{Name: key, Perm: value,
				Type: colType})
		}
		result = tableResult{
			Name:       table.Name,
			Insert:     perm[`insert`],
			NewColumn:  perm[`new_column`],
			Update:     perm[`update`],
			Read:       perm[`read`],
			Filter:     perm[`filter`],
			Conditions: table.Conditions,
			Columns:    columns,
		}
	} else {
		return errorAPI(w, `E_TABLENOTFOUND`, http.StatusBadRequest, data.params[`name`].(string))
	}
	data.result = &result
	return
}
