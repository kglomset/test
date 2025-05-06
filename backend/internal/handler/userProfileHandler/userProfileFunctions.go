package userProfileHandler

import (
	"backend/internal/domain"
	"database/sql"
	"log"
	"net/http"
)

// GetUserAttributes retrieves the email, user role, and team ID for a user from the database.
func GetUserAttributes(db *sql.DB, userID int, user *domain.User) error {
	err := db.QueryRow("SELECT email, user_role, team_id FROM users WHERE id = $1", userID).
		Scan(&user.Email, &user.UserRole, &user.Team)
	if err != nil {
		return err
	}
	return err
}

// getTeamAttributes retrieves the team name and role for a user from the database.
func getTeamAttributes(w http.ResponseWriter, db *sql.DB, user *domain.User, team *domain.Team, err error) {
	err = db.QueryRow("SELECT name, team_role FROM team WHERE id = $1", user.Team).
		Scan(&team.Name, &team.TeamRole)
	if err != nil {
		http.Error(w, "Could not find the team name and role", http.StatusNotFound)
		log.Printf("Could not find team name and role for user with id: %d Error: %v", user.ID, err.Error())
		return
	}
	return
}
