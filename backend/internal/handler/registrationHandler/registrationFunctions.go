package registrationHandler

import (
	"backend/internal/services/pwd"
	"database/sql"
	"fmt"
	"log"
)

// RegistrationPOSTRequest represents the request body for the registration endpoint.
func createTeam(tx *sql.Tx, credentials RegistrationPOSTRequest) (int, error) {
	var teamID int
	err := tx.QueryRow(`INSERT INTO team (
							   name, team_role) 
							    VALUES ($1, $2) RETURNING id`,
		credentials.TeamName, credentials.TeamRole).Scan(&teamID)
	return teamID, err
}

func checkExistingUser(tx *sql.Tx, credentials RegistrationPOSTRequest) (bool, error) {
	var email string
	err := tx.QueryRow("SELECT email FROM users WHERE email = $1", credentials.Email).Scan(&email)

	if err == nil {
		// User was found
		return true, nil
	}

	if err == sql.ErrNoRows {
		// User doesn't exist
		return false, nil
	}

	// Some other database error
	return false, err
}

func checkExistingTeam(tx *sql.Tx, credentials RegistrationPOSTRequest) (bool, int, error) {
	var teamID int
	err := tx.QueryRow("SELECT id FROM team WHERE name = $1", credentials.TeamName).Scan(&teamID)
	if err == nil {
		// No error means we found the team
		log.Println("Team already exists.")
		return true, teamID, nil
	}

	if err == sql.ErrNoRows {
		// This specific error means no matching team was found
		return false, 0, nil
	}

	// Any other error is a database/query error
	return false, 0, err
}

func insertNewUser(tx *sql.Tx, credentials RegistrationPOSTRequest, teamID int) error {
	hash, err := pwd.HashAndSalt(credentials.Password)
	if err != nil {
		return fmt.Errorf("error hashing password: %w", err)
	}
	_, err = tx.Exec(`INSERT INTO users (email, password, team_id) 
                     VALUES ($1, $2, $3)`,
		credentials.Email, hash, teamID)
	return err
}

func registerUserAndTeam(tx *sql.Tx, credentials RegistrationPOSTRequest) (int, error) {
	// Check if the user already exists
	userExists, err := checkExistingUser(tx, credentials)
	if err != nil {
		return 0, fmt.Errorf("database error checking user: %w", err)
	}
	if userExists {
		return 0, fmt.Errorf("user already exists")
	}

	// Check if team exists or create new team
	var teamID int
	teamExists, existingTeamID, err := checkExistingTeam(tx, credentials)
	if err != nil {
		return 0, fmt.Errorf("database error checking team: %w", err)
	}

	if teamExists {
		teamID = existingTeamID
	} else {
		teamID, err = createTeam(tx, credentials)
		if err != nil {
			return 0, fmt.Errorf("failed to create team: %w", err)
		}
	}

	// Insert the new user
	err = insertNewUser(tx, credentials, teamID)
	if err != nil {
		return 0, fmt.Errorf("failed to create user: %w", err)
	}

	return teamID, nil
}
