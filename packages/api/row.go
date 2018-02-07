// MIT License
//
// Copyright (c) 2016-2018 GenesisKernel
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

package api

import (
	"net/http"

	"github.com/GenesisKernel/go-genesis/packages/consts"
	"github.com/GenesisKernel/go-genesis/packages/converter"
	"github.com/GenesisKernel/go-genesis/packages/model"

	log "github.com/sirupsen/logrus"
)

type rowResult struct {
	Value map[string]string `json:"value"`
}

func row(w http.ResponseWriter, r *http.Request, data *apiData, logger *log.Entry) (err error) {
	cols := `*`
	if len(data.params[`columns`].(string)) > 0 {
		cols = converter.EscapeName(data.params[`columns`].(string))
	}
	table := converter.EscapeName(getPrefix(data) + `_` + data.params[`name`].(string))
	row, err := model.GetOneRow(`SELECT `+cols+` FROM `+table+` WHERE id = ?`, data.params[`id`].(string)).String()
	if err != nil {
		logger.WithFields(log.Fields{"type": consts.DBError, "error": err, "table": data.params["name"].(string), "id": data.params["id"].(string)}).Error("getting one row")
		return errorAPI(w, `E_QUERY`, http.StatusInternalServerError)
	}

	data.result = &rowResult{Value: row}
	return
}
