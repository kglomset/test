package registrationHandler

type RegistrationPOSTRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=14"`
	TeamName string `json:"team_name" validate:"required"`
	TeamRole int    `json:"team_role" validate:"required,oneof=1 2"`
}
