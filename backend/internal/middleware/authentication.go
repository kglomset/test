package middleware

import (
	"backend/internal/domain"
	"backend/internal/resources"
	"context"
	"database/sql"
	"errors"
	"log"
	"net/http"
)

const CtxSessionKey string = "session"

type AuthHandler struct {
	db *sql.DB
}

func NewAuthHandler(db *sql.DB) *AuthHandler {
	return &AuthHandler{db: db}
}

func (a *AuthHandler) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		authToken := GetAuthorizationToken(r)
		if authToken == "" {
			http.Error(w, resources.AuthenticationError, http.StatusUnauthorized)
			log.Println(resources.AuthenticationError + ", authorization token is empty")
			return
		}

		var session domain.Session
		err := a.db.QueryRow(`
			SELECT s.user_id, s.session_token, s.expires_at
			FROM sessions s
			LEFT JOIN public.users u ON s.user_id = u.id
			WHERE s.session_token = $1 AND s.expires_at > NOW()
		`, authToken).Scan(&session.UserID, &session.SessionToken, &session.ExpiresAt)

		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				log.Printf("Invalid or expired session token")
				http.Error(w, resources.AuthenticationError, http.StatusUnauthorized)
				return
			}
			log.Printf("Database error during authentication: %v", err)
			http.Error(w, resources.InternalServerError, http.StatusInternalServerError)
			return
		}

		log.Printf("Authenticated user ID: %d", session.UserID)
		ctx := context.WithValue(r.Context(), CtxSessionKey, session)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
