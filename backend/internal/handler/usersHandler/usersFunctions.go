package usersHandler

import (
	"backend/internal/domain"
	"backend/internal/handler/userProfileHandler"
	"backend/internal/resources"
	"backend/internal/utils"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// getTeamName retrieves the team name for a user from the database.
func getTeamName(w http.ResponseWriter, tx *sql.Tx, userIdStr string, teamID int, err error) string {
	var teamName string
	err = tx.QueryRow("SELECT name FROM team WHERE id = $1", teamID).
		Scan(&teamName)
	if err != nil {
		http.Error(w, "Could not find team name for user with id: "+userIdStr, http.StatusNotFound)
		log.Println("Could not find team name for user with id: "+userIdStr, err.Error())
		return ""
	}
	return teamName
}

// getUserInformation retrieves the user information from the database and sends it as a JSON response.
func getUserInformation(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	// Check if the request URL is valid
	validPath := "/users/"
	if !strings.HasPrefix(r.URL.Path, validPath) || len(strings.TrimPrefix(r.URL.Path, validPath)) == 0 {
		http.Error(w, "Invalid request URL, no user_id found.", http.StatusBadRequest)
		log.Println("Invalid request URL, no user_id found: " + r.URL.Path)
		return
	}

	// Get the user ID from the URL path
	userID := strings.TrimPrefix(r.URL.Path, "/users/")
	idStr := regexp.MustCompile(`\d+`).FindString(userID)

	var id int
	if idStr != "" {
		id, _ = utils.GetIDFromURLQuery(w, idStr)
	}

	// Get the email, user role, and team ID for the user.
	var user domain.User
	err := userProfileHandler.GetUserAttributes(db, id, &user)
	if err != nil {
		http.Error(w, resources.UserNotFound, http.StatusInternalServerError)
		log.Println(resources.UserNotFound + ": " + err.Error())
		return
	}

	// Start a transaction.
	var tx *sql.Tx
	tx, err = db.Begin()
	if err != nil {
		http.Error(w, resources.TransactionStartFailed, http.StatusInternalServerError)
		log.Println(resources.TransactionStartFailed + " " + err.Error())
		return
	}

	// Get the team name for the user.
	teamName := getTeamName(w, tx, idStr, user.Team, err)

	// Create the response struct
	userResponse := UserInfoResponse{
		Email:    user.Email,
		TeamName: teamName,
		UserRole: user.UserRole,
	}

	// Send the response
	writeUsersResponse(w, userResponse)
}

// getListOfAllActiveSessions retrieves a list of all active sessions for a user from the database.
func getListOfAllActiveSessions(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	// Check if the request URL is valid
	parsedURL, err := url.Parse(r.URL.Path)
	if err != nil {
		http.Error(w, "Invalid request URL", http.StatusBadRequest)
		log.Println("Invalid request URL: " + err.Error())
		return
	}

	// Split the URL path into segments.
	pathSegments := strings.Split(strings.Trim(parsedURL.Path, "/"), "/")

	var idStr string
	if len(pathSegments) >= 2 {
		idStr = pathSegments[1]
	}
	var userID int
	if idStr != "" {
		userID, _ = utils.GetIDFromURLQuery(w, idStr)
	}

	// Get session status from query parameters
	query := r.URL.Query()
	status := query.Get("status")

	// Query the database for the last active timestamps and IP addresses of the sessions for the user ID.
	var rows *sql.Rows
	rows, err = db.Query("SELECT created_at, ip_address FROM sessions WHERE user_id = $1 AND status = $2", userID, status)
	if err != nil {
		http.Error(w, "Could not find any active sessions for user with id: "+idStr, http.StatusNotFound)
		log.Println("Could not find any active sessions for user with id: "+idStr, err.Error())
		return
	}

	// Create a slice to hold the last active timestamps and one to hold the IP addresses.
	var lastActiveList []time.Time
	var ipList []string
	for rows.Next() {
		var lastActiveAt time.Time
		var ip string
		err = rows.Scan(&lastActiveAt, &ip)
		if err != nil {
			http.Error(w, "Could not scan session row.", http.StatusInternalServerError)
			log.Println("Could not scan session row: " + err.Error())
			return
		}
		lastActiveList = append(lastActiveList, lastActiveAt)
		ipList = append(ipList, ip)
	}

	// Create the response struct list
	var i = 1
	var activeSessions []SessionResponse
	for _, lastActiveAt := range lastActiveList {
		for _, ip := range ipList {
			activeSessions = append(activeSessions, SessionResponse{
				ID:         i,
				IP:         ip,
				LastActive: lastActiveAt,
			})
			i++
		}
	}

	// Send the response
	writeUsersResponse(w, activeSessions)
}

// writeUsersResponse writes the response to the HTTP response writer.
func writeUsersResponse(w http.ResponseWriter, response any) {
	w.WriteHeader(http.StatusOK)
	err := json.NewEncoder(w).Encode(response)
	if err != nil {
		http.Error(w, "Could not create users struct.", http.StatusInternalServerError)
		log.Fatal("Could not JSON encode users struct: " + err.Error())
		return
	}
}

// removeActiveSession removes an active session for a user from the database.
func removeActiveSession(w http.ResponseWriter, r *http.Request, db *sql.DB, teamID int, validPath []string) {
	// Check if the request URL is valid
	if len(validPath) == 0 {
		http.Error(w, "Invalid request URL, no user_id or session_id found.", http.StatusBadRequest)
		log.Println("Invalid request URL, no user_id or session_id found: " + r.URL.Path)
		return
	}

	// Split the URL path into segments.
	pathSegments := strings.Split(r.URL.Path, "/")

	// Extract the string value of the user ID and session ID from the URL.
	userID, _ := strconv.Atoi(pathSegments[2])
	sessionID, _ := strconv.Atoi(pathSegments[4])

	// Start a transaction.
	tx, err := db.Begin()
	if err != nil {
		http.Error(w, resources.TransactionStartFailed, http.StatusInternalServerError)
		log.Println(resources.TransactionStartFailed + " " + err.Error())
		return
	}

	// Delete the session for the user.
	err = deleteUserSession(tx, sessionID, userID, teamID)
	if err != nil {
		http.Error(w, "Failed to delete session: "+err.Error(), http.StatusInternalServerError)
		log.Println("Failed to delete session: " + err.Error())
		return
	}

	// Commit the transaction.
	if err = tx.Commit(); err != nil {
		http.Error(w, resources.TransactionCommitFailed, http.StatusInternalServerError)
		log.Println(resources.TransactionCommitFailed + ": " + err.Error())
		return
	}

	// Write the response
	w.WriteHeader(http.StatusNoContent)
}

// removeUserFromTeam removes a user from a team in the database.
func removeUserFromTeam(w http.ResponseWriter, r *http.Request, db *sql.DB, adminTeamID int, validPath []string) {
	if len(validPath) == 0 {
		http.Error(w, "Invalid request URL, no user_id found.", http.StatusBadRequest)
		log.Println("Invalid request URL, no user_id found: " + r.URL.Path)
		return
	}

	// Split the URL path into segments.
	pathSegments := strings.Split(r.URL.Path, "/")

	// Extract the string value of the user ID from the URL.
	userID, _ := strconv.Atoi(pathSegments[2])

	// Start a transaction.
	tx, err := db.Begin()
	if err != nil {
		http.Error(w, resources.TransactionStartFailed, http.StatusInternalServerError)
		log.Println(resources.TransactionStartFailed + ": " + err.Error())
		return
	}

	var teamID int
	teamID, _ = getAndCheckTeamID(tx, userID)

	if teamID == adminTeamID {
		// Create a new temporary team
		var newTempTeamID int
		err = tx.QueryRow(`INSERT INTO team (name, team_role) 
							VALUES ($1, $2) RETURNING id`,
			"Temporary team", domain.Researcher).Scan(&newTempTeamID)
		if err != nil {
			http.Error(w, "Failed to create temporary team: "+err.Error(), http.StatusInternalServerError)
			log.Println("Failed to create temporary team: " + err.Error())
			return
		}

		// Remove the user from the team.
		_, err = tx.Exec("UPDATE users SET team_id = $1 WHERE id = $2 AND team_id = $3", newTempTeamID, userID, adminTeamID)
		if err != nil {
			http.Error(w, "Failed to remove user from team.", http.StatusBadRequest)
			log.Println("Failed to remove user from team: " + err.Error())
			return
		}

		// Commit the transaction.
		if err = tx.Commit(); err != nil {
			http.Error(w, resources.TransactionCommitFailed, http.StatusInternalServerError)
			log.Println(resources.TransactionCommitFailed + ": " + err.Error())
			return
		}

		// Write the response
		writeUsersResponse(w, map[string]interface{}{
			"message":      "User removed from team successfully.",
			"temp_team_id": newTempTeamID,
		})
	} else {
		http.Error(w, "The user is not part of your team.", http.StatusBadRequest)
		log.Println("The user is not part of your team")
		return
	}
}

// deleteUserSession deletes a user session from the database.
func deleteUserSession(tx *sql.Tx, sessionID int, userID int,
	adminTeamID int) error {

	// Get the team ID of the user.
	teamID, err := getAndCheckTeamID(tx, userID)
	if err != nil {
		return err
	}

	/// Check if the user has permission to delete the session of the user
	// (Only if user is in the same team as the admin).
	if teamID == adminTeamID {
		_, err = tx.Exec("DELETE FROM sessions WHERE id = $1 AND user_id = $2", sessionID, userID)
		if err != nil {
			return fmt.Errorf("failed to delete session: %w", err)
		}
	} else {
		return fmt.Errorf("the user is not part of the same team as the admin")
	}
	return nil
}

// getAndCheckTeamID retrieves the team ID of a user from the database.
func getAndCheckTeamID(tx *sql.Tx, adminID int) (int, error) {
	var teamID int
	err := tx.QueryRow("SELECT team_id FROM users WHERE id = $1", adminID).Scan(&teamID)
	if err != nil {
		return 0, err
	}

	// Return the team ID.
	return teamID, nil
}
