package detector


type Data struct {
	Face []Face

	ImgHeight int `json:"img_height"`
	ImgID     string `json:"img_id"`
	ImgWidth  int `json:"img_width"`
	SessionID string `json:"session_id"`
	URL       string `json:"url"`
}

type Face struct {
	FaceID    string `json:"face_id"`
	Attr 	Attribute `json:"attribute"`
	Pos 	Position `json:"position"`
	Tag string
}

type Attribute struct {
	Age struct {
		Range int
		Value int
	} `json:"age"`
	Gender struct {
		Confidence float64
		Value      string
	} `json:"gender"`
	Race struct {
		Confidence float64
		Value      string
	} `json:"race"`
	Smiling struct {
		Value float64
	}
}

type Position struct {
	Center Point `json:"center"`
	EyeLeft Point `json:"eye_left"`
	EyeRight Point `json:"eye_right"`
	Height    float64 `json:"height"`
	MouthLeft Point `json:"mouth_left"`
	MouthRight Point `json:"mouth_right"`
	Nose Point `json:"nose"`
	Width float64 `json:"width"`
}

type Point struct  {
	X, Y float64
}

type Similarity struct {
	Attr Components `json:"component_similarity"`
	SessionID  string  `json:"session_id"`
	Similarity float64 `json:"similarity"`
}

type Components struct {
	Eye     float64 `json:"eye"`
	Eyebrow float64 `json:"eyebrow"`
	Mouth   float64 `json:"mouth"`
	Nose    float64 `json:"nose"`
}