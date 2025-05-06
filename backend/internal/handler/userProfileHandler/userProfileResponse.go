package userProfileHandler

type UserProfileResponse struct {
	Email    string       `json:"email"`
	UserRole string       `json:"user_role"`
	Team     TeamResponse `json:"team"`
}

type TeamResponse struct {
	Name     string `json:"name"`
	TeamRole string `json:"team_role"`
}
