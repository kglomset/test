package usersHandler

import (
	"time"
)

type SessionResponse struct {
	ID         int       `json:"id"`
	IP         string    `json:"ip"`
	LastActive time.Time `json:"last_active"`
}
