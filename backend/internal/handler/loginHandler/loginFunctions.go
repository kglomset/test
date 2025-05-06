package loginHandler

import (
	"backend/internal/domain"
	"backend/internal/services/pwd"
	"database/sql"
	"errors"
	"fmt"
	"time"
)

func CheckUserExists(db *sql.DB, email string) (domain.User, error) {
	var user domain.User
	err := db.QueryRow("SELECT id, email, password FROM users WHERE email = $1", email).
		Scan(&user.ID, &user.Email, &user.Password)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return user, fmt.Errorf("user not found")
		}
		return user, fmt.Errorf("database error: %v", err)
	}
	return user, nil
}

func VerifyPassword(providedPassword, storedHash string) error {
	match, err := pwd.CheckPasswordHash(providedPassword, storedHash)
	if err != nil {
		return fmt.Errorf("error checking password hash: %v", err)
	}
	if !match {
		return fmt.Errorf("password mismatch")
	}
	return nil
}

func CreateSession(db *sql.DB, userID int, sessionToken string, expiresAt time.Time, ip string) error {
	_, err := db.Exec("INSERT INTO sessions (user_id, session_token, expires_at, ip_address) VALUES ($1, $2, $3, $4)",
		userID, sessionToken, expiresAt, ip)
	if err != nil {
		return fmt.Errorf("could not create session: %v", err)
	}
	return nil
}
