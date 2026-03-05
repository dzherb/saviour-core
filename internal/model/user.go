package model

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	UUID         uuid.UUID
	CreatedAt    time.Time
	Username     string
	PasswordHash string
	IsAdmin      bool
}
