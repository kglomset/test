package productsHandler

import (
	"time"
)

type ProductPATCHRequest struct {
	Updates map[string]interface{} `json:"updates" validate:"required,dive,keys,oneof=name brand ean_code image_url type high_temperature low_temperature comment is_public status,endkeys"`
	Version time.Time              `json:"version"`
}

type ProductUpdateFields struct {
	Name            *string  `json:"name" validate:"omitempty,lte=64,ascii"`
	Brand           *string  `json:"brand" validate:"omitempty,lte=64,ascii"`
	ImageURL        *string  `json:"image_url" validate:"omitempty,url"`
	EANCode         *string  `json:"ean_code" validate:"omitempty,max=128,ascii"`
	Comment         *string  `json:"comment" validate:"omitempty,max=2040"`
	IsPublic        *bool    `json:"is_public" validate:"omitempty,oneof=true false"`
	Type            *string  `json:"type" validate:"omitempty,oneof=liquid solid spray powder gel bundle"`
	HighTemperature *float64 `json:"high_temperature" validate:"omitempty,lte=100,gte=-100"`
	LowTemperature  *float64 `json:"low_temperature" validate:"omitempty,lte=100,gte=-100,ltfield=HighTemperature"`
	Status          *string  `json:"status" validate:"omitempty,oneof=active tested discontinued development retired"`
}
