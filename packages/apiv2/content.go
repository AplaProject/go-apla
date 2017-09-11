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

package apiv2

import (
	"net/http"
)

type contentResult struct {
	Tree string `json:"tree"`
}

func getPage(w http.ResponseWriter, r *http.Request, data *apiData) error {
	var result contentResult

	result = contentResult{
		Tree: `[{"type":"fn","name":"Title","data":["State info"]},{"type":"fn","name":"Navigation","data":[[{"type":"fn","name":"LiTemplate","data":["government","Government"]}],"State info"]},{"type":"block","name":"Divs","data":["md-4","panel panel-default elastic center"],"children":[{"type":"block","name":"Divs","data":["panel-body"],"children":[{"type":"fn","name":"IfParams","data":["#flag#==\"\"",[{"type":"fn","name":"Image","data":["static/img/noflag.svg","No flag","img-responsive"]}],[{"type":"fn","name":"Image","data":["#flag#","Flag","img-responsive"]}]]},{"type":"fn","name":"DivsEnd"}]},{"type":"fn","name":"DivsEnd"}]}]`,
	}

	data.result = result
	return nil
}

func getMenu(w http.ResponseWriter, r *http.Request, data *apiData) error {
	var result contentResult

	result = contentResult{
		Tree: `[{"type":"fn","name":"Title","data":["State info"]},{"type":"fn","name":"Navigation","data":[[{"type":"fn","name":"LiTemplate","data":["government","Government"]}],"State info"]},{"type":"block","name":"Divs","data":["md-4","panel panel-default elastic center"],"children":[{"type":"block","name":"Divs","data":["panel-body"],"children":[{"type":"fn","name":"IfParams","data":["#flag#==\"\"",[{"type":"fn","name":"Image","data":["static/img/noflag.svg","No flag","img-responsive"]}],[{"type":"fn","name":"Image","data":["#flag#","Flag","img-responsive"]}]]},{"type":"fn","name":"DivsEnd"}]},{"type":"fn","name":"DivsEnd"}]}]`,
	}

	data.result = result
	return nil
}
