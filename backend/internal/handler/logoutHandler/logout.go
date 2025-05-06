package logoutHandler

import (
	"backend/internal/resources"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"strings"
)

// LogoutHandler handles user logout requests.
//
//	@Summary		Logout user
//	@Description	Logs out the user by invalidating the session and CSRF tokens
//	@Tags			Logout
//	@Accept			json
//	@Produce		json
//	@Success		200	{string}	string	"Successfully logged out"
//	@Failure		400	{string}	string	"Request method not allowed"
//	@Failure		401	{string}	string	"Unauthorized"
//	@Failure		500	{string}	string	"Could not log out"
//	@Router			/logout/ [post]
func LogoutHandler(db *sql.DB) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("content-type", "application/json")
		if r.Method != http.MethodPost {
			http.Error(w, resources.MethodNotAllowed, http.StatusMethodNotAllowed)
			return
		}

		// Checking authorization header
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, resources.AuthenticationError, http.StatusUnauthorized)
			log.Println(resources.AuthenticationError + ": Authorization header is missing.")
			return
		}

		// Check if the authorization header is in the correct format
		authToken := strings.TrimPrefix(authHeader, "Bearer ")
		if authToken == authHeader {
			http.Error(w, resources.AuthenticationError, http.StatusUnauthorized)
			log.Println(resources.AuthenticationError + ": Authorization header is not in the correct format.")
			return
		}

		// Mark the session as expired in the database when the user logs out.
		_, err := db.Exec(`UPDATE sessions 
								  SET status = 'expired'
								  WHERE session_token = $1 `, authToken)
		if err != nil {
			http.Error(w, "Could not log out.", http.StatusInternalServerError)
			log.Println("Could not log out: " + err.Error())
			return
		}

		// Write the response to the client.
		fmt.Fprintln(w, "You have successfully logged out!")

		w.WriteHeader(http.StatusOK)
	})
}
