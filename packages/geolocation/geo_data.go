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
	"strings"
)

type APIResponse struct {
	Results []Results `json:"results"`
	Status string `json:"status"`
}

type Results struct {
	Address []AddressComponents `json:"address_components"`
	FormattedAddress string `json:"formatted_address"`
	Geometry Geometry `json:"geometry"`
	PlaceID string   `json:"place_id"`
	Types   []string `json:"types"`
}

type AddressComponents struct {
	LongName  string   `json:"long_name"`
	ShortName string   `json:"short_name"`
	Types     []string `json:"types"`
}

type Coordinates struct  {
	Lat float64 `json:"lat"`
	Lng float64 `json:"lng"`
}

type Geometry struct {
	Bounds struct {
		Northeast Coordinates `json:"northeast"`
		Southwest Coordinates `json:"southwest"`
	} `json:"bounds"`
	Location Coordinates `json:"location"`
	LocationType string `json:"location_type"`
	Viewport     struct {
		Northeast Coordinates `json:"northeast"`
		Southwest Coordinates `json:"southwest"`
	} `json:"viewport"`
}

func (resp *APIResponse) GetCountryName() string {
	for _, v := range resp.Results[0].Address {
		if strings.Contains(v.Types[0], "country") {
			return v.LongName
		}
	}

	return ""
}