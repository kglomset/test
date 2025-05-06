package domain

import (
	"time"
)

type Product struct {
	ID              int       `json:"id"`
	Name            string    `json:"name"`
	Brand           string    `json:"brand"`
	EANCode         string    `json:"ean_code"`
	ImageURL        string    `json:"image_url"`
	Comment         string    `json:"comment"`
	IsPublic        bool      `json:"is_public"`
	Type            string    `json:"type"`
	HighTemperature float64   `json:"high_temperature"`
	LowTemperature  float64   `json:"low_temperature"`
	TestingTeam     int       `json:"testing_team"`
	Version         time.Time `json:"version"`
	Status          string    `json:"status"`
}
