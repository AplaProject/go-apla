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
	"net/http"
	"errors"
	"encoding/json"
	"fmt"
)

type Location struct {
	Coordinates *coordinates `json:"location"`
	Accuracy    float32 `json:"accuracy"`
}

type coordinates struct {
	Latitude float64 `json:"lat"`
	Longitude float64 `json:"lng"`
}

func GetLocation() (*coordinates, error) {
	coord, err := getLocation()
	return coord, err
}

func GetInfo(lat, lng float64) (*APIResponse, error) {
	slat := fmt.Sprintf("%.5f", lat)
	slng := fmt.Sprintf("%.5f", lng)
	url := "http://maps.googleapis.com/maps/api/geocode/json?latlng=" + slat + "," + slng + "&components=country"

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	// Check the response
	if res.StatusCode != http.StatusOK {
		err = errors.New(res.Status)
		return nil, err
	}

	var response *APIResponse
	decoder := json.NewDecoder(res.Body)
	err = decoder.Decode(&response)
	if err != nil {
		return nil, err
	}
	return response, nil
}

