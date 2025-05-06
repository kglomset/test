package testsHandler

import "time"

type TestPATCHRequest struct {
	Updates map[string]interface{} `json:"updates" validate:"required,dive,keys,oneof=product_id rank distance_behind is_rank_public track_hardness track_type ac_temperature air_humidity wind cloud sc_temperature snow_type snow_humidity location test_date comment is_public,endkeys"`
	Version time.Time              `json:"version"`
}

type TestUpdateFields struct {
	ProductID      *int       `json:"product_id" validate:"omitempty,gte=0"`
	Rank           *int       `json:"rank" validate:"omitempty,gt=0"`
	DistanceBehind *int       `json:"distance_behind" validate:"omitempty,gte=0"`
	IsRankPublic   *bool      `json:"is_rank_public" validate:"omitempty,oneof=true false"`
	SCTemperature  *float32   `json:"sc_temperature" validate:"omitempty,lte=100,gte=-100"`
	SnowType       *string    `json:"snow_type" validate:"omitempty,oneof=A1 A2 A3 A4 A5 FS NS IN IT TR"`
	SnowHumidity   *string    `json:"snow_humidity" validate:"omitempty,oneof=DS W1 W2 W3 W4"`
	ACTemperature  *float32   `json:"ac_temperature" validate:"omitempty,lte=100,gte=-100"`
	AirHumidity    *int       `json:"air_humidity" validate:"omitempty,gte=0,lte=100"`
	Wind           *string    `json:"wind" validate:"omitempty,oneof=S L M ST"`
	Cloud          *string    `json:"cloud" validate:"omitempty,oneof=1 2 3 4"`
	TrackHardness  *string    `json:"track_hardness" validate:"omitempty,oneof=H1 H2 H3 H4 H5 H6"`
	TrackType      *string    `json:"track_type" validate:"omitempty,oneof=T1 T2 D1 D2"`
	Location       *string    `json:"location" validate:"omitempty,lte=256,ascii"`
	Date           *time.Time `json:"test_date" validate:"omitempty"` //TODO: Need validation for date format
	Comment        *string    `json:"comment" validate:"omitempty,max=2040"`
	IsPublic       *bool      `json:"is_public" validate:"omitempty,oneof=true false"`
}

var validTestFields = map[string]bool{
	"location":  true,
	"test_date": true,
	"comment":   true,
	"is_public": true,
}

var validRankFields = map[string]bool{
	"rank":            true,
	"distance_behind": true,
	"is_rank_public":  true,
}

var validACFields = map[string]bool{
	"ac_temperature": true,
	"air_humidity":   true,
	"wind":           true,
	"cloud":          true,
}

var validTCFields = map[string]bool{
	"track_hardness": true,
	"track_type":     true,
}

var validSCFields = map[string]bool{
	"sc_temperature": true,
	"snow_type":      true,
	"snow_humidity":  true,
}

var actualFieldNamesAC = map[string]string{
	"ac_temperature": "temperature",
	"air_humidity":   "humidity",
	"wind":           "wind",
	"cloud":          "cloud",
}

var actualFieldNamesSC = map[string]string{
	"sc_temperature": "temperature",
	"snow_humidity":  "snow_humidity",
	"snow_type":      "snow_type",
}

var actualFieldNamesTC = map[string]string{
	"track_hardness": "track_hardness",
	"track_type":     "track_type",
}
