package productsHandler

type ProductPOSTRequest struct {
	// Name: Only ASCII, non-blank, max 64 characters.
	Name string `json:"name" validate:"required,lte=64,ascii"`

	// Brand: Only ASCII, can be blank, max 64 characters
	Brand string `json:"brand" validate:"omitempty,lte=64,ascii"`

	// ImageURL: Only ASCII, can be blank, and must be a valid URL.
	ImageURL string `json:"image_url" validate:"omitempty,url"`

	// EANCode: Only ASCII, can be blank, max 64 characters,
	// Only capital letters, numbers, and '#' are allowed; non-blank, max 128 characters.
	EANCode string `json:"ean_code" validate:"omitempty,max=128,ascii"`

	// Comment: max 2040 bytes.
	Comment string `json:"comment" validate:"required,max=2040"`

	IsPublic bool `json:"is_public"` // Gets set to false by default and validated in POST handler.

	// Type: Only ASCII, non-blank, max 16 characters.
	Type string `json:"type" validate:"required,oneof=liquid solid spray powder gel bundle"`

	// HighTemperature must be between -100 and 100.
	HighTemperature float64 `json:"high_temperature" validate:"required,lte=100,gte=-100"`

	// LowTemperature must be between -100 and 100.
	LowTemperature float64 `json:"low_temperature" validate:"required,lte=100,gte=-100,ltfield=HighTemperature"`

	TestingTeam int `json:"testing_team"`

	// Status: Only one of the following: active, tested, discontinued, development, retired.
	Status string `json:"status" validate:"required,oneof=active tested discontinued development retired"`
}
