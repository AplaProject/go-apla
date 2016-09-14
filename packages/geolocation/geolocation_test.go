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
	"testing"
	"fmt"
	"errors"
)

func TestGetLocation(t *testing.T) {
	if _, err := GetLocation(); err != nil {
		fmt.Println(err.Error())
		t.Fatal(err)
	}
}


func TestGetInfo(t *testing.T) {
//	US
//	lat: 40.622570
//	lng: -73.96258
	resp, err := GetInfo(40.622570, -73.96258)
	if err != nil {
		t.Fatal(err)
	}
	if resp.Status != "OK" {
		t.Fatal(errors.New("Couldn't get data form API"))
	}
	fmt.Println(resp.GetCountryName())
}


func TestGetCountryName(t *testing.T) {
	resp, err := GetInfo(40.622570, -73.96258)
	if err != nil {
		t.Fatal(err)
	}
	if resp.Status != "OK" {
		t.Fatal(errors.New("Couldn't get data form API"))
	}
	if resp.GetCountryName() != "United States" {
		t.Fatal(errors.New("Wrong country name: " + resp.GetCountryName()))
	}
}
