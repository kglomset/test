package domain

import "time"

type TestRank struct {
	TestID         int       `json:"test_id"`
	ProductID      int       `json:"product_id"`
	Rank           int       `json:"rank"`
	DistanceBehind int       `json:"distance_behind"`
	IsPublic       bool      `json:"is_public"`
	Version        time.Time `json:"version"`
}
