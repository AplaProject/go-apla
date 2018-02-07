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
	"net/http"

	"github.com/GenesisCommunity/go-genesis/packages/consts"
	"github.com/GenesisCommunity/go-genesis/packages/model"

	log "github.com/sirupsen/logrus"
)

type ecosystemsResult struct {
	Number uint32 `json:"number"`
}

func ecosystems(w http.ResponseWriter, r *http.Request, data *apiData, logger *log.Entry) (err error) {

	number, err := model.GetNextID(nil, `system_states`)
	if err != nil {
		logger.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("Error getting next system_states id")
		return err
	}
	data.result = &ecosystemsResult{Number: uint32(number - 1)}
	return
}
