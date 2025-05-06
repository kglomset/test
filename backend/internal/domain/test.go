package domain

import "time"

type Test struct {
	ID              int       `json:"id"`
	Date            time.Time `json:"test_date"`
	Location        string    `json:"location"`
	Comment         string    `json:"comment"`
	SnowConditions  int       `json:"sc_id"`
	TrackConditions int       `json:"tc_id"`
	AirConditions   int       `json:"ac_id"`
	Version         time.Time `json:"version"`
	IsPublic        bool      `json:"is_public"`
	TestingTeam     int       `json:"testing_team"`
}
