package usersHandler

import (
	"backend/internal/domain"
	"backend/internal/middleware"
	"backend/internal/resources"
	"backend/internal/services/pwd"
	"backend/internal/utils"
	"database/sql"
	"github.com/go-playground/validator/v10"
	"log"
	"net/http"
	"regexp"
)

// UsersHandler routes HTTP requests for users to the appropriate handler function.
//
// It supports the following methods:
// - GET: Retrieves the user for a specific user ID.
// - PATCH: Updates the password of a user.
// - DELETE: Deletes an active session of a user.
func UsersHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("content-type", "application/json")
		switch r.Method {
		case http.MethodGet:
			UsersRequestGET(w, r, db)
		case http.MethodPatch:
			UsersRequestPATCH(w, r, db)
		case http.MethodDelete:
			UsersRequestDELETE(w, r, db)
		default:
			http.Error(w, resources.MethodNotAllowed, http.StatusNotImplemented)
			return
		}
	}
}

// UsersRequestGET handles GET requests for users.
//
//	@Summary		Get a user by ID
//	@Description	Retrieves the user for a specific user ID.
//	@Tags			Users
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			user_id	path		int					true	"User ID"
//	@Param			status	query		string				false	"Session status"
//	@Success		200		{object}	UserInfoResponse	"Successful response with user details"
//	@Success		200		{array}		domain.Session		"Successful response with a list of active sessions"
//	@Failure		400		{string}	string				"Invalid request URL, no user_id found."
//	@Failure		404		{string}	string				"User not found"
//	@Failure		401		{string}	string				"Unauthorized"
//	@Failure		500		{string}	string				"Could not create users struct."
//	@Router			/users/{user_id} [get]
//	@Router			/users/{user_id}/sessions [get]
func UsersRequestGET(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	// Check if the authenticated user has the required team_role.
	userRole := middleware.GetUserRole(w, r, db)
	userInfoPath := regexp.MustCompile(`^/users/\d+$`)
	//sessionPath := regexp.MustCompile(`^/users/\d+/sessions\?status=active$`)
	if domain.UserRole(userRole) == domain.Admin {
		switch {
		case userInfoPath.MatchString(r.URL.Path):
			getUserInformation(w, r, db)
			return
		//case sessionPath.MatchString(r.URL.Path):
		default:
			getListOfAllActiveSessions(w, r, db)
			return
		}
	} else {
		http.Error(w, resources.AuthenticationError, http.StatusUnauthorized)
		log.Println(resources.AuthenticationError + ": " + "user is not an admin user.")
		return
	}
}

// UsersRequestPATCH handles PATCH requests for users.
//
//	@Summary		Update a user's password
//	@Description	Updates the password of a user.
//	@Tags			Users
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			user_id	path		int						true	"User ID"
//	@Param			body	body		ChangePasswordRequest	true	"New password"
//	@Success		200		{string}	string					"Password updated successfully"
//	@Failure		400		{string}	string					"Invalid request URL"
//	@Failure		400		{string}	string					"Invalid request body"
//	@Failure		400		{string}	string					"New password cannot be the same as the current password."
//	@Failure		401		{string}	string					"Unauthorized"
//	@Failure		500		{string}	string					"Could not change the current password."
//	@Failure		500		{string}	string					"Could not update the password"
//	@Failure		500		{string}	string					"Failed to commit the transaction."
//	@Router			/users/password [patch]
func UsersRequestPATCH(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	// Check if the URL path is valid
	validPath := regexp.MustCompile("/users/password").FindStringSubmatch(r.URL.Path)
	if len(validPath) == 0 {
		http.Error(w, "Invalid request URL, use '/users/password' to change password.", http.StatusBadRequest)
		log.Println("Invalid request URL: " + r.URL.Path)
		return
	}
	// Get the user ID from the session token.
	userID := middleware.GetUserID(w, r, db)

	// Start a transaction.
	tx, err := db.Begin()
	if err != nil {
		http.Error(w, resources.TransactionStartFailed, http.StatusInternalServerError)
		log.Println(resources.TransactionStartFailed + ": " + err.Error())
		return
	}

	// Decode the request body into the ChangePasswordRequest struct.
	var changePasswordRequest ChangePasswordRequest
	err = utils.DecodeRequestBody(w, r, &changePasswordRequest)
	if err != nil {
		http.Error(w, resources.InvalidPATCHRequest, http.StatusBadRequest)
		log.Println(resources.InvalidPATCHRequest + ": " + err.Error())
		return
	}

	// Validate the request body.
	validate := validator.New()
	err = validate.Struct(changePasswordRequest)
	if err != nil {
		http.Error(w, "Invalid request body.", http.StatusBadRequest)
		log.Println("Invalid request body: " + err.Error())
		return
	}

	var currentPasswordHash string
	err = tx.QueryRow("SELECT password FROM users WHERE id = $1", userID).
		Scan(&currentPasswordHash)
	if err != nil {
		http.Error(w, "Could not change the current password.", http.StatusInternalServerError)
		log.Println("Could not retrieve the current password: " + err.Error())
		return
	}

	// Check if the current password matches the hash of the provided password.
	validCurrentPassword, _ := pwd.CheckPasswordHash(changePasswordRequest.CurrentPassword, currentPasswordHash)
	if !validCurrentPassword {
		http.Error(w, "Current password is incorrect.", http.StatusBadRequest)
		log.Println("Current password is incorrect.")
		return
	}

	// Check if the new password is the same as the current password.
	validNewPassword, _ := pwd.CheckPasswordHash(changePasswordRequest.NewPassword, currentPasswordHash)
	if validNewPassword {
		http.Error(w, "New password cannot be the same as the current password.", http.StatusConflict)
		log.Println("New password cannot be the same as the current password.")
		return
	}

	// Hash and salt the new password.
	var newPasswordHash string
	newPasswordHash, err = pwd.HashAndSalt(changePasswordRequest.NewPassword)
	if err != nil {
		http.Error(w, "Could not hash the new password.", http.StatusInternalServerError)
		log.Println("Could not hash the new password: " + err.Error())
		return
	}

	// Update the password in the database.
	_, err = tx.Exec(`UPDATE users 
								  SET password = $1
								  WHERE id = $2`, newPasswordHash, userID)
	if err != nil {
		http.Error(w, "Could not update the password", http.StatusInternalServerError)
		log.Println("Could not update the password: " + err.Error())
		return
	}

	// Commit the transaction.
	if err = tx.Commit(); err != nil {
		http.Error(w, resources.TransactionCommitFailed, http.StatusInternalServerError)
		log.Println(resources.TransactionCommitFailed + ": " + err.Error())
		return
	}

	// Write the response
	w.WriteHeader(http.StatusOK)
}

// UsersRequestDELETE handles DELETE requests for users.
//
//	@Summary		Delete a user's active session
//	@Description	Deletes an active session of a user.
//	@Tags			Users
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			user_id		path		int		true	"User ID"
//	@Param			team_id		path		int		true	"Team ID"
//	@Param			sessionId	path		int		true	"Session ID"
//	@Success		204			{string}	string	"Session deleted successfully"
//	@Failure		400			{string}	string	"Invalid request URL, no user_id or session_id found."
//	@Failure		404			{string}	string	"Session not found"
//	@Failure		401			{string}	string	"Unauthorized"
//	@Failure		500			{string}	string	"Failed to delete session"
//	@Failure		500			{string}	string	"Failed to start the transaction."
//	@Failure		500			{string}	string	"Failed to commit the transaction."
//	@Router			/users/{user_id}/sessions/{sessionId} [delete]
//	@Router			/users/{user_id} [delete]
func UsersRequestDELETE(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	// Check if the authenticated user has the required team_role.
	userRole := middleware.GetUserRole(w, r, db)
	teamID := middleware.GetUserTeamID(w, r, db)
	removeSessionPath := regexp.MustCompile("/users/\\d+/sessions/\\d+")
	removeUserFromTeamPath := regexp.MustCompile("/users/\\d+")
	if domain.UserRole(userRole) == domain.Admin {
		switch {
		case removeSessionPath.MatchString(r.URL.Path):
			removeActiveSession(w, r, db, teamID, removeSessionPath.FindStringSubmatch(r.URL.Path))
			return
		case removeUserFromTeamPath.MatchString(r.URL.Path):
			removeUserFromTeam(w, r, db, teamID, removeUserFromTeamPath.FindStringSubmatch(r.URL.Path))
			return
		default:
			http.Error(w, "Invalid request URL, no user_id, session_id found.", http.StatusBadRequest)
			log.Println("Invalid request URL, no user_id, session_id found: " + r.URL.Path)
			return
		}
	} else {
		http.Error(w, resources.AuthenticationError, http.StatusUnauthorized)
		log.Println(resources.AuthenticationError + ": " + "user is not an admin.")
		return
	}
}
