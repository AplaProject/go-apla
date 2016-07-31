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

