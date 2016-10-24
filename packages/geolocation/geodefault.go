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

package geolocation

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"fmt"
	"encoding/json"
	"errors"
	"github.com/EGaaS/go-mvp/packages/consts"
)

func getLocation() (*coordinates, error) {
	var buf bytes.Buffer
	resp, err := http.Post("https://www.googleapis.com/geolocation/v1/geolocate?key=" + consts.GOOGLE_API_KEY, "json", &buf)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Cannot read body:", err.Error())
		return nil, err
	}

	loc, err := parseResponse(body)
	if err != nil {
		fmt.Println("Cannot parse:", err.Error())
		return nil, err
	}
	
	if loc.Coordinates == nil {
		return nil, errors.New("Couldn't get user's location")
	}

	return loc.Coordinates, nil
}

func parseResponse(b []byte) (*Location, error) {
	var pos *Location
	if err := json.Unmarshal(b, &pos); err != nil {
		return nil, err
	}

	return pos, nil
}
