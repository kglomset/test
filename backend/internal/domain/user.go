package domain

type User struct {
	ID       int    `json:"id"`
	Email    string `json:"email"`
	Password string `json:"password"`
	Team     int    `json:"team"`
	UserRole string `json:"user_role"`
}
