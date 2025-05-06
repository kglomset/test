package usersHandler

type UserInfoResponse struct {
	Email    string `json:"email"`
	TeamName string `json:"team_name"`
	UserRole string `json:"user_role"`
}
