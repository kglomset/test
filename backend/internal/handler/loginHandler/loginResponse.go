package loginHandler

import "time"

type LoginResponse struct {
	ExpiresAt    time.Time `json:"expires_at"`
	SessionToken string    `json:"session_token"`
}
