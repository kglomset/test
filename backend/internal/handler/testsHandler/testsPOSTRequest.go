package testsHandler

import (
	"time"
)

type TestPOSTRequest struct {
	SnowConditions  SnowConditionsPOST  `json:"sc" validate:"required"`
	AirConditions   AirConditionsPOST   `json:"ac" validate:"required"`
	TrackConditions TrackConditionsPOST `json:"tc" validate:"required"`
	Location        string              `json:"location" validate:"required,lte=256,ascii"`
	Date            time.Time           `json:"test_date" validate:"omitempty"` //TODO: Need validation for date format
	Comment         string              `json:"comment" validate:"required,max=2040"`
	IsPublic        bool                `json:"is_public" validate:"omitempty,oneof=true false"`
	TestingTeam     int                 `json:"testing_team"`
	TestRanks       []TestRanksPOST     `json:"test_ranks" validate:"required,dive"`
}

type TestRanksPOST struct {
	ProductID      int  `json:"product_id"  validate:"omitempty,gte=0"`
	Rank           int  `json:"rank" validate:"omitempty,gt=0"`
	DistanceBehind int  `json:"distance_behind" validate:"omitempty,gte=0"`
	IsRankPublic   bool `json:"is_rank_public" validate:"omitempty,oneof=true false"`
}

type SnowConditionsPOST struct {
	Temperature  float32 `json:"temperature" validate:"omitempty,lte=100,gte=-100"`
	SnowType     string  `json:"snow_type" validate:"omitempty,oneof=A1 A2 A3 A4 A5 FS NS IN IT TR"`
	SnowHumidity string  `json:"snow_humidity" validate:"omitempty,oneof=DS W1 W2 W3 W4"`
}

type AirConditionsPOST struct {
	Temperature float32 `json:"temperature" validate:"omitempty,lte=100,gte=-100"`
	Humidity    int     `json:"humidity" validate:"omitempty,lte=100,gte=0"`
	Wind        string  `json:"wind" validate:"omitempty,oneof=S L M ST"`
	Cloud       string  `json:"cloud" validate:"omitempty,oneof=1 2 3 4"`
}

type TrackConditionsPOST struct {
	TrackHardness string `json:"track_hardness" validate:"omitempty,oneof=H1 H2 H3 H4 H5 H6"`
	TrackType     string `json:"track_type" validate:"omitempty,oneof=T1 T2 D1 D2"`
}
