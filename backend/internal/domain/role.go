package domain

type UserRole string
type TeamRole int

const (
	// User roles
	Admin  UserRole = "admin"
	Member UserRole = "member"

	// Team roles
	Official   TeamRole = 1
	Researcher TeamRole = 2
)
