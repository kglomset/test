package userProfileHandler

import (
	"backend/internal/domain"
	"backend/internal/middleware"
	"backend/internal/resources"
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"strings"
)

// UserProfileHandler routes HTTP requests for user profiles to the appropriate handler function.
//
// It supports the following methods:
// - GET: Retrieves the user profile for the authenticated user.
func UserProfileHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("content-type", "application/json")
		switch r.Method {
		case http.MethodGet:
			UserProfileRequestGET(w, r, db)
		default:
			http.Error(w, resources.MethodNotAllowed, http.StatusNotImplemented)
			return
		}
	}
}

// UserProfileRequestGET handles GET requests for user profiles.
//
//	@Summary		Get user profile
//	@Description	Retrieves the user profile for the authenticated user.
//	@Tags			UserProfile
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Success		200	{object}	UserProfileResponse	"Successful response with user profile details"
//	@Failure		400	{string}	string				"Invalid request URL"
//	@Failure		404	{string}	string				"User not found"
//	@Failure		401	{string}	string				"Unauthorized"
//	@Failure		500	{string}	string				"Could not create user profile struct."
//	@Router			/user/profile [get]
func UserProfileRequestGET(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	// Check if the request URL is valid
	validPath := "/user/profile"
	if !strings.HasPrefix(r.URL.Path, validPath) {
		http.Error(w, "Invalid request URL", http.StatusBadRequest)
		log.Println("Invalid request URL: " + r.URL.Path)
		return
	}

	// Get the user ID from the URL path
	userID := middleware.GetUserID(w, r, db)

	// Get the email, user role, and team ID for the user.
	var user domain.User
	err := GetUserAttributes(db, userID, &user)
	if err != nil {
		http.Error(w, resources.UserNotFound, http.StatusNotFound)
		log.Println(resources.UserNotFound + ": " + err.Error())
		return
	}

	// Get the team name and role for the user.
	var team domain.Team
	user.ID = userID
	getTeamAttributes(w, db, &user, &team, err)

	roleInt := team.TeamRole
	var teamRole string
	switch domain.TeamRole(roleInt) {
	case domain.Official:
		teamRole = "Official"
	case domain.Researcher:
		teamRole = "Researcher"
	default:
		teamRole = ""
	}

	if teamRole != "" {
		// Create the response struct
		userProfileResponse := UserProfileResponse{
			Email:    user.Email,
			UserRole: user.UserRole,
			Team: TeamResponse{
				Name:     team.Name,
				TeamRole: teamRole,
			},
		}

		// Send the response
		w.WriteHeader(http.StatusOK)
		err = json.NewEncoder(w).Encode(userProfileResponse)
		if err != nil {
			http.Error(w, "Could not create users struct.", http.StatusInternalServerError)
			log.Fatal("Could not JSON encode users struct: " + err.Error())
			return
		}
	} else {
		http.Error(w, "Invalid team role", http.StatusNotFound)
		log.Println("Invalid team role: ", roleInt)
		return
	}
}
