package rankingsHandler

import "time"

type RankingsPOSTRequest struct {
	Wins     int       `json:"wins" validate:"required"`
	Rank     int       `json:"rank" validate:"required"`
	TestID   int       `json:"test_id" validate:"required"`
	IsPublic bool      `json:"is_public"`
	Version  time.Time `json:"version"`
}
