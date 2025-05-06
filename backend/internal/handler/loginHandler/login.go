package loginHandler

import (
	"backend/internal/domain"
	"backend/internal/resources"
	"backend/internal/utils"
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"time"
)

// generateSecureLoginToken generates a secure login token as a base64 encoded string.
func generateSecureLoginToken(tokenLength int) (string, error) {
	if tokenLength < 16 {
		return "", errors.New("tokenLength must be at least equal to 16")
	}
	bytes := make([]byte, tokenLength)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(bytes), nil
}

// LoginHandler handles user login requests.
//
//	@Summary		Login user
//	@Description	Authenticates the user and creates a session
//	@Tags			Login
//	@Accept			json
//	@Produce		json
//	@Param			credentials	body		LoginRequest	true	"User credentials"
//	@Success		200			{string}	string			"Session created successfully"
//	@Failure		400			{string}	string			"Invalid request body"
//	@Failure		401			{string}	string			"Invalid email or password"
//	@Failure		405			{string}	string			"Method not allowed"
//	@Failure		500			{string}	string			"Could not create session"
//	@Router			/login/ [post]
func LoginHandler(db *sql.DB) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("content-type", "application/json")
		if r.Method != http.MethodPost {
			http.Error(w, resources.MethodNotAllowed, http.StatusMethodNotAllowed)
			return
		}

		// Parse and validate the request body
		credentials, err := utils.ParseAndValidateRequest[LoginRequest](r)
		if err != nil {
			http.Error(w, "Invalid request data", http.StatusBadRequest)
			log.Println("Error parsing login request: ", err)

		}

		// Check if the user exists in the database
		var user domain.User
		user, err = CheckUserExists(db, credentials.Email)
		if err != nil {
			http.Error(w, "Invalid email or password", http.StatusUnauthorized)
			log.Println("Failed login attempt with non-existent or invalid user: ", credentials.Email)
			return
		}

		// Check if the password is correct
		if err = VerifyPassword(credentials.Password, user.Password); err != nil {
			http.Error(w, "Invalid email or password", http.StatusUnauthorized)
			log.Println("Password validation failed for user: ", user.Email)
			return
		}

		// increase randomness and length for better security
		sessionToken, _ := generateSecureLoginToken(32)
		expiresAt := time.Now().Add(24 * time.Hour)

		// Check if the user is an official user
		ip := utils.GetClientIPFromRequest(r)

		// Create a new session in the database
		if err = CreateSession(db, user.ID, sessionToken, expiresAt, ip); err != nil {
			http.Error(w, "Could not create session", http.StatusInternalServerError)
			log.Println("Error creating session: ", err)
			return
		}

		// Set session token in request header for further requests
		r.Header.Set("Authorization", "Bearer "+sessionToken)

		// Create a response struct
		loginResponse := LoginResponse{
			ExpiresAt:    expiresAt,
			SessionToken: sessionToken,
		}

		// Send response
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Authorization", sessionToken)
		err = json.NewEncoder(w).Encode(loginResponse)
		if err != nil {
			http.Error(w, "Could not create session struct.", http.StatusInternalServerError)
			log.Fatal("Could not JSON encode session struct: " + err.Error())
			return
		}
	})
}
