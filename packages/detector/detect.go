package detector

import (
	"encoding/json"
	"log"
	"errors"
)


func Race(path string) (int, error) {
	raceInt := 2

	base := GetURL(DETECT)
	req, err := POSTRequest(base.String(), path)
	if err != nil {
		return raceInt, err
	}

	res, err := Send(req)
	if err != nil {
		return raceInt, err
	}

	var data Data
	decoder := json.NewDecoder(res.Body)
	err = decoder.Decode(&data)
	if err != nil {
		return raceInt, err
	}

	if len(data.Face) < 1 {
		return raceInt, errors.New("UNKNOWN race")
	}


	switch data.Face[0].Attr.Race.Value {
	case "Asian":
		raceInt = 1
	case "Black":
		raceInt = 3
	default:
		raceInt = 2
	}

	return raceInt, err
}

func DetectSimilarity(faceId1, faceId2 string) (float64, error) {
	base := GetURL(COMPARE)
	params := base.Query()
	params.Add("face_id1", faceId1)
	params.Add("face_id2", faceId2)
	base.RawQuery = params.Encode()
	log.Println(base.String())
	req, err := GETRequest(base.String())
	if err != nil {
		return 0.0, err
	}

	res, err := Send(req)
	if err != nil {
		return 0.0, err
	}

	var sim Similarity
	decoder := json.NewDecoder(res.Body)
	if err := decoder.Decode(&sim); err != nil {
		return 0.0, err
	}
	return sim.Similarity, err
}