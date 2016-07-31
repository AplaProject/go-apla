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
