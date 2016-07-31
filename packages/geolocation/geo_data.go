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