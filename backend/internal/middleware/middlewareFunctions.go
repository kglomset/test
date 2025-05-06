package middleware

import (
	"backend/internal/resources"
	"database/sql"
	"net/http"
	"strings"
)

// GetAuthorizationToken retrieves the authorization token from the request header.
func GetAuthorizationToken(r *http.Request) string {
	// Checking if the Authorization header is present.
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer") {
		return ""
	}

	authToken := strings.TrimSpace(strings.TrimPrefix(authHeader, "Bearer"))

	return authToken
}

// GetUserID retrieves the user ID from the session token stored in the database.
func GetUserID(w http.ResponseWriter, r *http.Request, db *sql.DB) int {
	// Get the user ID from the session token.
	authorizationToken := GetAuthorizationToken(r)

	// Add an early return if the authorization token is empty.
	if authorizationToken == "" {
		http.Error(w, resources.AuthenticationError, http.StatusUnauthorized)
		return 0
	}

	var userID int
	err := db.QueryRow("SELECT user_id FROM sessions WHERE (session_token = $1 AND expires_at > NOW())",
		authorizationToken).Scan(&userID)
	if err != nil {
		http.Error(w, resources.AuthenticationError, http.StatusUnauthorized)
		return 0
	}

	return userID
}

// GetUserRole retrieves the user role of the authenticated user from the database.
func GetUserRole(w http.ResponseWriter, r *http.Request, db *sql.DB) string {
	// Get the user ID from the session token.
	var userID int
	userID = GetUserID(w, r, db)

	// Add an early return if the userID is 0.
	if userID == 0 {
		return ""
	}

	// Fetch the authenticated user's user role attribute.
	var userRole string
	err := db.QueryRow("SELECT user_role FROM users WHERE id = $1;", userID).Scan(&userRole)
	if err != nil {
		http.Error(w, resources.AuthenticationError, http.StatusUnauthorized)
		return ""
	}

	// Return the user role.
	return userRole
}

// GetUserTeamID retrieves the team ID of the authenticated user from the database.
func GetUserTeamID(w http.ResponseWriter, r *http.Request, db *sql.DB) int {
	// Get the user ID from the session token.
	var userID int
	userID = GetUserID(w, r, db)

	// Add an early return if the userID is 0.
	if userID == 0 {
		return 0
	}

	// Fetch the authenticated user's teamID ID attribute.
	var teamID int
	err := db.QueryRow("SELECT team_id FROM users WHERE id = $1;", userID).Scan(&teamID)
	if err != nil {
		http.Error(w, resources.AuthenticationError, http.StatusUnauthorized)
		return 0
	}

	// Return the teamID ID.
	return teamID
}

// GetUserTeamRole retrieves the team role of the authenticated user from the database.
func GetUserTeamRole(w http.ResponseWriter, r *http.Request, db *sql.DB) int {
	// Fetch the authenticated user's team ID attribute.
	teamID := GetUserTeamID(w, r, db)

	// Add an early return if the teamID is 0.
	if teamID == 0 {
		return 0
	}

	// Fetch the user´s team role attribute.
	var teamRole int
	err := db.QueryRow("SELECT team_role FROM team WHERE id = $1;", teamID).Scan(&teamRole)
	if err != nil {
		http.Error(w, resources.AuthenticationError, http.StatusUnauthorized)
		return 0
	}

	// Return the team role.
	return teamRole
}
